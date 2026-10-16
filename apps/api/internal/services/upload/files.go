package upload_service

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/svg"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	upload_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/upload"
)

const headerBytes = 256 << 10

var (
	extensionPattern = regexp.MustCompile(`^\.[a-z0-9]{1,10}$`)
	inlineTypes      = map[string]bool{
		"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true, "image/avif": true,
		"application/pdf": true, "text/plain": true,
		"video/mp4": true, "video/webm": true, "audio/mpeg": true, "audio/ogg": true, "audio/wav": true,
	}
)

type CreateUploadInput struct {
	WorkspaceID uuid.UUID
	FileName    string
	ContentType string
	Body        io.Reader
}

func (self *UploadService) Create(ctx context.Context, userID uuid.UUID, input *CreateUploadInput) (*models.UploadResponse, error) {
	if err := self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: input.WorkspaceID}}); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	if err := self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleMember, WorkspaceID: input.WorkspaceID}}); err != nil {
		return nil, err
	}

	header := make([]byte, headerBytes)
	n, err := io.ReadFull(input.Body, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Could not read the uploaded file")
	}
	header = header[:n]
	if n == 0 {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The file is empty")
	}

	contentType := detectContentType(header, input.ContentType)
	width, height := imageSize(header, contentType)
	body := io.MultiReader(bytes.NewReader(header), input.Body)
	if contentType == svg.ContentType {
		data, svgWidth, svgHeight, err := readSVG(body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
		width, height = positive(svgWidth), positive(svgHeight)
	}
	data, err := io.ReadAll(io.LimitReader(body, self.cfg.MaxUploadBytes+1))
	if err != nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Could not read the uploaded file")
	}
	if int64(len(data)) > self.cfg.MaxUploadBytes {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The file is larger than the upload limit")
	}

	id := uuid.New()
	created, err := self.repo.Upload().Create(ctx, &upload_repo.CreateUploadInput{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		UploaderID:  userID,
		FileName:    cleanFileName(input.FileName),
		ContentType: contentType,
		StorageKey:  input.WorkspaceID.String() + "/" + id.String() + extensionOf(input.FileName),
		Width:       width,
		Height:      height,
		Data:        data,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformUploadEntity(created), nil
}

type OpenedUpload struct {
	Upload *ent.Upload
	Data   []byte
	Inline bool
}

func (self *UploadService) Open(ctx context.Context, userID, uploadID uuid.UUID) (*OpenedUpload, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Upload not found")
	upload, err := self.repo.Upload().GetByID(ctx, uploadID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if err := self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: upload.WorkspaceID}}); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Upload not found")
	}
	data, err := self.repo.Upload().GetData(ctx, upload.ID)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to read the stored file")
	}
	return &OpenedUpload{Upload: upload, Data: data, Inline: servedInline(upload.ContentType, data)}, nil
}

func servedInline(contentType string, data []byte) bool {
	if contentType != svg.ContentType {
		return inlineTypes[contentType]
	}
	_, _, err := svg.Validate(data)
	return err == nil
}

func readSVG(body io.Reader) ([]byte, int, int, error) {
	data, err := io.ReadAll(io.LimitReader(body, svg.MaxBytes+1))
	if err != nil {
		return nil, 0, 0, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Could not read the uploaded file")
	}
	width, height, err := svg.Validate(data)
	if err != nil {
		return nil, 0, 0, err
	}
	return data, width, height, nil
}

func positive(value int) *int {
	if value <= 0 {
		return nil
	}
	return &value
}

func detectContentType(header []byte, declared string) string {
	sniffed := http.DetectContentType(header)
	if strings.HasPrefix(sniffed, "image/") || sniffed == "application/pdf" {
		return sniffed
	}
	if isWebP(header) {
		return "image/webp"
	}
	mediaType, _, err := mime.ParseMediaType(declared)
	if err != nil || mediaType == "" {
		mediaType, _, _ = mime.ParseMediaType(sniffed)
	}
	mediaType = strings.ToLower(mediaType)
	if mediaType == svg.ContentType || startsWithSVG(header) {
		return svg.ContentType
	}
	return mediaType
}

func startsWithSVG(header []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(header))
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		if element, ok := token.(xml.StartElement); ok {
			return strings.EqualFold(element.Name.Local, "svg")
		}
	}
}

func imageSize(header []byte, contentType string) (*int, *int) {
	if contentType == "image/webp" {
		return webPSize(header)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, nil
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(header))
	if err != nil {
		return nil, nil
	}
	return &config.Width, &config.Height
}

func extensionOf(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if !extensionPattern.MatchString(ext) {
		return ""
	}
	return ext
}

func cleanFileName(fileName string) string {
	name := filepath.Base(strings.ReplaceAll(fileName, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == "/" {
		return "file"
	}
	if len(name) > 255 {
		return name[len(name)-255:]
	}
	return name
}
