package uploads_handler

import (
	"context"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	upload_service "github.com/zigzagapp-org/zigzag-api/internal/services/upload"
)

const (
	defaultMaxUploadBytes = 26214400
	multipartOverhead     = 1 << 20
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}
	maxBytes := int64(defaultMaxUploadBytes)
	if server.Cfg != nil {
		maxBytes = server.Cfg.MaxUploadBytes
	}

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID:  "create-upload",
		Summary:      "Create Upload",
		Description:  "Upload a file to a workspace as multipart form data with workspace_id and file. Images get their pixel size recorded. Returns the url to use in markdown.",
		Path:         "/create",
		Method:       http.MethodPost,
		MaxBodyBytes: maxBytes + multipartOverhead,
		Errors:       []int{http.StatusRequestEntityTooLarge},
	}, handlers.Create, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-upload",
		Summary:     "Get Upload",
		Description: "Stream an uploaded file. Safe types such as images and pdf are served inline, everything else as a download.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.SessionOnly)
}

type CreateUploadForm struct {
	WorkspaceID string        `form:"workspace_id" required:"true" format:"uuid" doc:"Workspace the file belongs to"`
	File        huma.FormFile `form:"file" required:"true" contentType:"application/octet-stream" doc:"The file"`
}

type CreateUploadInput struct {
	server.BaseAuthInput
	RawBody huma.MultipartFormFiles[CreateUploadForm]
}

type UploadResponse struct {
	Body struct {
		Data *models.UploadResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateUploadInput) (*UploadResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	form := input.RawBody.Data()
	workspaceID, err := uuid.Parse(form.WorkspaceID)
	if err != nil {
		return nil, oapi.MapError(errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "workspace_id must be a uuid"))
	}
	defer form.File.Close()

	upload, err := self.srv.UploadService.Create(ctx, user.ID, &upload_service.CreateUploadInput{
		WorkspaceID: workspaceID,
		FileName:    form.File.Filename,
		ContentType: form.File.ContentType,
		Body:        form.File,
	})
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UploadResponse{}
	resp.Body.Data = upload
	return resp, nil
}

type GetUploadInput struct {
	server.BaseAuthInput
	UploadID uuid.UUID `query:"upload_id" required:"true" format:"uuid" doc:"The upload"`
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetUploadInput) (*huma.StreamResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	opened, err := self.srv.UploadService.Open(ctx, user.ID, input.UploadID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		defer opened.Body.Close()
		disposition := "attachment"
		contentType := "application/octet-stream"
		if opened.Inline {
			disposition = "inline"
			contentType = opened.Upload.ContentType
		}
		hctx.SetHeader("Content-Type", contentType)
		hctx.SetHeader("Content-Length", strconv.FormatInt(opened.Upload.SizeBytes, 10))
		hctx.SetHeader("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": opened.Upload.FileName}))
		hctx.SetHeader("Cache-Control", "private, max-age=31536000, immutable")
		hctx.SetHeader("X-Content-Type-Options", "nosniff")
		hctx.SetHeader("Content-Security-Policy", "default-src 'none'; sandbox")
		hctx.SetStatus(http.StatusOK)
		if _, err := io.Copy(hctx.BodyWriter(), opened.Body); err != nil {
			log.Warn("Failed to stream an upload", "upload_id", opened.Upload.ID, "err", err)
		}
	}}, nil
}
