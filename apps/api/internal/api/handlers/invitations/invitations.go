package invitations_handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	workspace_service "github.com/unbindapp/zigzag-api/internal/services/workspace"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-invitation",
		Summary:     "Get Invitation",
		Description: "Show who an invitation is for, the workspace and the role, before it is accepted. The token is the only credential.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.GetInvitation, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "accept-invitation",
		Summary:     "Accept Invitation",
		Description: "Accept an invitation and sign in. Creates the account when the email has none; an existing account must give its password.",
		Path:        "/accept",
		Method:      http.MethodPost,
	}, handlers.AcceptInvitation, oapi.Public, oapi.SessionOnly)
}

type GetInvitationInput struct {
	Token string `query:"token" required:"true" minLength:"1" maxLength:"200" doc:"Token from the invitation link"`
}

type GetInvitationResponse struct {
	Body struct {
		Data *models.InvitationPreviewResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetInvitation(ctx context.Context, input *GetInvitationInput) (*GetInvitationResponse, error) {
	preview, err := self.srv.WorkspaceService.GetInvitation(ctx, input.Token)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GetInvitationResponse{}
	resp.Body.Data = preview
	return resp, nil
}

type AcceptInvitationInput struct {
	Body *workspace_service.AcceptInvitationInput
}

type AcceptInvitationResponse struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Data *models.SessionWorkspaceResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) AcceptInvitation(ctx context.Context, input *AcceptInvitationInput) (*AcceptInvitationResponse, error) {
	user, workspace, err := self.srv.WorkspaceService.AcceptInvitation(ctx, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	cookies, err := self.srv.IssueSession(ctx, user)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &AcceptInvitationResponse{SetCookie: cookies}
	resp.Body.Data = &models.SessionWorkspaceResponse{User: models.TransformUserEntity(user), Workspace: workspace}
	return resp, nil
}
