package system_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	googleauth_service "github.com/zigzagapp-org/zigzag-api/internal/services/googleauth"
)

type GoogleAuthSettingsResponse struct {
	Body struct {
		Data *models.GoogleAuthSettingsResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetGoogleAuthSettings(ctx context.Context, _ *server.BaseAuthInput) (*GoogleAuthSettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.GoogleAuthService.GetSettings(ctx, user.ID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GoogleAuthSettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type UpdateGoogleAuthSettingsInput struct {
	server.BaseAuthInput
	Body *googleauth_service.UpdateGoogleAuthSettingsInput `required:"true"`
}

func (self *HandlerGroup) UpdateGoogleAuthSettings(ctx context.Context, input *UpdateGoogleAuthSettingsInput) (*GoogleAuthSettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.GoogleAuthService.UpdateSettings(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GoogleAuthSettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

func (self *HandlerGroup) ClearGoogleAuthSettings(ctx context.Context, _ *server.BaseAuthInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.GoogleAuthService.ClearSettings(ctx, user.ID); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}
