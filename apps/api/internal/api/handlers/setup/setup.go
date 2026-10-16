package setup_handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	user_service "github.com/unbindapp/zigzag-api/internal/services/user"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-setup-status",
		Summary:     "Get Setup Status",
		Description: "Report whether the first user has been created. Until then the welcome page creates it.",
		Path:        "/status",
		Method:      http.MethodGet,
	}, handlers.GetStatus, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-bootstrap-user",
		Summary:     "Create Bootstrap User",
		Description: "Create the first user as system admin together with the first workspace, and sign them in. Only works while no user exists.",
		Path:        "/create-user",
		Method:      http.MethodPost,
	}, handlers.CreateUser, oapi.Public, oapi.SessionOnly)
}

type SetupStatus struct {
	IsFirstUserCreated bool `json:"is_first_user_created"`
}

type SetupStatusResponse struct {
	Body struct {
		Data *SetupStatus `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetStatus(ctx context.Context, _ *server.EmptyInput) (*SetupStatusResponse, error) {
	created, err := self.srv.UserService.SetupStatus(ctx)
	if err != nil {
		return nil, oapi.MapError(errdefs.NewInternalError(err, "Failed to check whether Zigzag is set up"))
	}
	resp := &SetupStatusResponse{}
	resp.Body.Data = &SetupStatus{IsFirstUserCreated: created}
	return resp, nil
}

type CreateUserInput struct {
	Body *user_service.BootstrapInput `required:"true"`
}

type CreateUserResponse struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Data *models.SessionWorkspaceResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) CreateUser(ctx context.Context, input *CreateUserInput) (*CreateUserResponse, error) {
	user, workspace, err := self.srv.UserService.Bootstrap(ctx, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	cookies, err := self.srv.IssueSession(ctx, user)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &CreateUserResponse{SetCookie: cookies}
	resp.Body.Data = &models.SessionWorkspaceResponse{
		User:      models.TransformUserEntity(user),
		Workspace: models.TransformWorkspaceEntity(workspace, schema.WorkspaceRoleAdmin, 1, 0),
	}
	return resp, nil
}
