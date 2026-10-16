package connectedapps_handler

import (
	"context"
	"errors"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	"github.com/zigzagapp-org/zigzag-api/internal/oauthserver"
)

type ApproveInput struct {
	server.BaseAuthInput
	Body *models.ConnectedAppApproveInput `required:"true"`
}

type DenyInput struct {
	server.BaseAuthInput
	Body *models.ConnectedAppDenyInput `required:"true"`
}

type RedirectResponse struct {
	Body struct {
		Data *models.ConnectedAppRedirectResponse `json:"data"`
	}
}

func (self *HandlerGroup) Approve(ctx context.Context, input *ApproveInput) (*RedirectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	redirect, err := self.srv.OAuthServerService.Approve(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}

	resp := &RedirectResponse{}
	resp.Body.Data = redirect
	return resp, nil
}

func (self *HandlerGroup) Deny(ctx context.Context, input *DenyInput) (*RedirectResponse, error) {
	if _, err := self.srv.AuthenticatedUser(ctx); err != nil {
		return nil, err
	}

	redirect, err := self.srv.OAuthServerService.Deny(ctx, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}

	resp := &RedirectResponse{}
	resp.Body.Data = redirect
	return resp, nil
}

func asInputError(err error) error {
	var oauthErr *oauthserver.Error
	if errors.As(err, &oauthErr) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, oauthErr.Error())
	}
	return err
}
