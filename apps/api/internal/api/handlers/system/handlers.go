package system_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	email_service "github.com/unbindapp/zigzag-api/internal/services/email"
	system_service "github.com/unbindapp/zigzag-api/internal/services/system"
)

type SettingsResponse struct {
	Body struct {
		Data *models.SystemSettingsResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetSettings(ctx context.Context, _ *server.BaseAuthInput) (*SettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.SystemService.GetSettings(ctx, user.ID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type UpdateSettingsInput struct {
	server.BaseAuthInput
	Body *system_service.UpdateSettingsInput
}

func (self *HandlerGroup) UpdateSettings(ctx context.Context, input *UpdateSettingsInput) (*SettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.SystemService.UpdateSettings(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type EmailSettingsResponse struct {
	Body struct {
		Data *models.EmailSettingsResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetEmailSettings(ctx context.Context, _ *server.BaseAuthInput) (*EmailSettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.EmailService.GetSettings(ctx, user.ID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &EmailSettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type UpdateEmailSettingsInput struct {
	server.BaseAuthInput
	Body *email_service.EmailSettingsUpdateInput
}

func (self *HandlerGroup) UpdateEmailSettings(ctx context.Context, input *UpdateEmailSettingsInput) (*EmailSettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.EmailService.UpdateSettings(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &EmailSettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type SuccessResponse struct {
	Body struct {
		Data server.SuccessResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ClearEmailSettings(ctx context.Context, _ *server.BaseAuthInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.EmailService.ClearSettings(ctx, user.ID); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}

func (self *HandlerGroup) SendTestEmail(ctx context.Context, _ *server.BaseAuthInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.EmailService.SendTest(ctx, user.ID); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}
