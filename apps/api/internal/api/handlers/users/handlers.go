package user_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	user_service "github.com/unbindapp/zigzag-api/internal/services/user"
)

type WhoamiResponse struct {
	Body struct {
		Data *models.WhoamiResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Whoami(ctx context.Context, _ *server.BaseAuthInput) (*WhoamiResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	whoami, err := self.srv.UserService.Whoami(ctx, user.ID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WhoamiResponse{}
	resp.Body.Data = whoami
	return resp, nil
}

type UpdateMeInput struct {
	server.BaseAuthInput
	Body *user_service.UpdateMeInput `required:"true"`
}

type UserResponse struct {
	Body struct {
		Data *models.UserResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) UpdateMe(ctx context.Context, input *UpdateMeInput) (*UserResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := self.srv.UserService.UpdateMe(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UserResponse{}
	resp.Body.Data = updated
	return resp, nil
}

type ChangePasswordInput struct {
	server.BaseAuthInput
	Body *user_service.ChangePasswordInput `required:"true"`
}

type SuccessResponse struct {
	Body struct {
		Data models.SuccessResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ChangePassword(ctx context.Context, input *ChangePasswordInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.UserService.ChangePassword(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}

type NotificationSettingsResponse struct {
	Body struct {
		Data *models.NotificationSettingsResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetNotificationSettings(ctx context.Context, _ *server.BaseAuthInput) (*NotificationSettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.UserService.GetNotificationSettings(ctx, user.ID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &NotificationSettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type UpdateNotificationSettingsInput struct {
	server.BaseAuthInput
	Body *user_service.UpdateNotificationSettingsInput `required:"true"`
}

func (self *HandlerGroup) UpdateNotificationSettings(ctx context.Context, input *UpdateNotificationSettingsInput) (*NotificationSettingsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := self.srv.UserService.UpdateNotificationSettings(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &NotificationSettingsResponse{}
	resp.Body.Data = settings
	return resp, nil
}

type ListUsersInput struct {
	server.BaseAuthInput
	user_service.ListUsersInput
}

type ListUsersResponse struct {
	Body struct {
		Data []*models.UserResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ListUsers(ctx context.Context, input *ListUsersInput) (*ListUsersResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	users, err := self.srv.UserService.ListUsers(ctx, user.ID, &input.ListUsersInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ListUsersResponse{}
	resp.Body.Data = users
	return resp, nil
}

type CreateUserInput struct {
	server.BaseAuthInput
	Body *user_service.CreateUserInput `required:"true"`
}

func (self *HandlerGroup) CreateUser(ctx context.Context, input *CreateUserInput) (*UserResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	created, err := self.srv.UserService.CreateUser(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UserResponse{}
	resp.Body.Data = created
	return resp, nil
}

type UpdateUserInput struct {
	server.BaseAuthInput
	Body *user_service.UpdateUserInput `required:"true"`
}

func (self *HandlerGroup) UpdateUser(ctx context.Context, input *UpdateUserInput) (*UserResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := self.srv.UserService.UpdateUser(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UserResponse{}
	resp.Body.Data = updated
	return resp, nil
}

type DeleteUserInput struct {
	server.BaseAuthInput
	Body *user_service.DeleteUserInput `required:"true"`
}

type DeletedResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) DeleteUser(ctx context.Context, input *DeleteUserInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.UserService.DeleteUser(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.UserID, Deleted: true}
	return resp, nil
}
