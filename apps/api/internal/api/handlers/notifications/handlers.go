package notifications_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	notification_service "github.com/unbindapp/zigzag-api/internal/services/notification"
)

type ListInput struct {
	server.BaseAuthInput
	notification_service.ListNotificationsInput
}

type NotificationListResponse struct {
	Body struct {
		Data       []*models.NotificationResponse `json:"data" nullable:"false"`
		NextCursor *string                        `json:"next_cursor" nullable:"true" doc:"Pass as cursor for the next page; null on the last page"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*NotificationListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	notifications, next, err := self.srv.NotificationService.List(ctx, user.ID, &input.ListNotificationsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &NotificationListResponse{}
	resp.Body.Data = notifications
	resp.Body.NextCursor = next
	return resp, nil
}

type CountsInput struct {
	server.BaseAuthInput
	notification_service.WorkspaceInput
}

type NotificationCountsResult struct {
	Body struct {
		Data *models.NotificationCountsResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Counts(ctx context.Context, input *CountsInput) (*NotificationCountsResult, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := self.srv.NotificationService.Counts(ctx, user.ID, &input.WorkspaceInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &NotificationCountsResult{}
	resp.Body.Data = counts
	return resp, nil
}

type IDsInput struct {
	server.BaseAuthInput
	Body *notification_service.NotificationIDsInput `required:"true"`
}

type SuccessResponse struct {
	Body struct {
		Data models.SuccessResponse `json:"data" nullable:"false"`
	}
}

func success(err error) (*SuccessResponse, error) {
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}

func (self *HandlerGroup) MarkRead(ctx context.Context, input *IDsInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return success(self.srv.NotificationService.MarkRead(ctx, user.ID, input.Body))
}

func (self *HandlerGroup) MarkUnread(ctx context.Context, input *IDsInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return success(self.srv.NotificationService.MarkUnread(ctx, user.ID, input.Body))
}

func (self *HandlerGroup) Archive(ctx context.Context, input *IDsInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return success(self.srv.NotificationService.Archive(ctx, user.ID, input.Body))
}

func (self *HandlerGroup) Unarchive(ctx context.Context, input *IDsInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return success(self.srv.NotificationService.Unarchive(ctx, user.ID, input.Body))
}

type SnoozeInput struct {
	server.BaseAuthInput
	Body *notification_service.SnoozeNotificationsInput `required:"true"`
}

func (self *HandlerGroup) Snooze(ctx context.Context, input *SnoozeInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return success(self.srv.NotificationService.Snooze(ctx, user.ID, input.Body))
}

type MarkAllReadInput struct {
	server.BaseAuthInput
	Body *notification_service.MarkAllReadInput `required:"true"`
}

func (self *HandlerGroup) MarkAllRead(ctx context.Context, input *MarkAllReadInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return success(self.srv.NotificationService.MarkAllRead(ctx, user.ID, input.Body))
}
