package notification_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	notification_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/notification"
)

const notFoundMessage = "Notification not found"

type ListNotificationsInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose notifications to list"`
	Filter      string    `query:"filter" required:"false" enum:"inbox,unread,read,archived,snoozed,all" default:"inbox" doc:"inbox (default): read and unread, not archived, not snoozed; unread or read: that part of the inbox; archived; snoozed: hidden until snoozed_until; all: everything"`
	models.PaginationInput
}

type WorkspaceInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace to count notifications in"`
}

type NotificationIDsInput struct {
	NotificationIDs []uuid.UUID `json:"notification_ids" format:"uuid" required:"true" nullable:"false" minItems:"1" maxItems:"250" doc:"Ids of your notifications, from list-notifications"`
}

type SnoozeNotificationsInput struct {
	NotificationIDs []uuid.UUID `json:"notification_ids" format:"uuid" required:"true" nullable:"false" minItems:"1" maxItems:"250" doc:"Ids of your notifications, from list-notifications"`
	Until           time.Time   `json:"until" required:"true" doc:"RFC 3339 time in the future when the notifications return to the inbox as unread"`
}

type MarkAllReadInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose inbox to mark read"`
}

func (self *NotificationService) List(ctx context.Context, userID uuid.UUID, input *ListNotificationsInput) ([]*models.NotificationResponse, *string, error) {
	if err := self.checkWorkspace(ctx, userID, input.WorkspaceID); err != nil {
		return nil, nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	var cursor *repository.TimeCursor
	if input.Cursor != "" {
		decoded, err := repository.DecodeTimeCursor(input.Cursor)
		if err != nil {
			return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "cursor is not one returned by list-notifications")
		}
		cursor = decoded
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	filter := notification_repo.Filter(input.Filter)
	if filter == "" {
		filter = notification_repo.FilterInbox
	}

	rows, err := self.repo.Notification().List(ctx, &notification_repo.ListInput{
		UserID:      userID,
		WorkspaceID: input.WorkspaceID,
		Filter:      filter,
		Limit:       limit,
		Cursor:      cursor,
		Now:         self.now(),
	})
	if err != nil {
		return nil, nil, err
	}
	if len(rows) <= limit {
		return models.TransformNotifications(rows), nil, nil
	}
	page := rows[:limit]
	last := page[len(page)-1]
	next := repository.TimeCursor{CreatedAt: last.CreatedAt, ID: last.ID}.Encode()
	return models.TransformNotifications(page), &next, nil
}

func (self *NotificationService) Counts(ctx context.Context, userID uuid.UUID, input *WorkspaceInput) (*models.NotificationCountsResponse, error) {
	if err := self.checkWorkspace(ctx, userID, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	unread, err := self.repo.Notification().CountUnread(ctx, userID, input.WorkspaceID, self.now())
	if err != nil {
		return nil, err
	}
	return &models.NotificationCountsResponse{Unread: unread}, nil
}

// ensureOwned refuses the whole batch when one id is not reachable.
func (self *NotificationService) ensureOwned(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "notification_ids is empty")
	}
	workspaces, err := self.repo.Notification().OwnedWorkspaces(ctx, userID, ids)
	if err != nil {
		return err
	}
	reached := map[uuid.UUID]bool{}
	for _, id := range ids {
		workspaceID, found := workspaces[id]
		if !found {
			return errdefs.NewCustomError(errdefs.ErrTypeNotFound, notFoundMessage)
		}
		reached[workspaceID] = true
	}
	for workspaceID := range reached {
		if err := self.checkWorkspace(ctx, userID, workspaceID); err != nil {
			return err
		}
	}
	return nil
}

func (self *NotificationService) MarkRead(ctx context.Context, userID uuid.UUID, input *NotificationIDsInput) error {
	ids := input.NotificationIDs
	if err := self.ensureOwned(ctx, userID, ids); err != nil {
		return err
	}
	now := self.now()
	return self.repo.Notification().SetRead(ctx, userID, ids, &now)
}

func (self *NotificationService) MarkUnread(ctx context.Context, userID uuid.UUID, input *NotificationIDsInput) error {
	ids := input.NotificationIDs
	if err := self.ensureOwned(ctx, userID, ids); err != nil {
		return err
	}
	return self.repo.Notification().SetRead(ctx, userID, ids, nil)
}

func (self *NotificationService) Archive(ctx context.Context, userID uuid.UUID, input *NotificationIDsInput) error {
	ids := input.NotificationIDs
	if err := self.ensureOwned(ctx, userID, ids); err != nil {
		return err
	}
	now := self.now()
	return self.repo.Notification().SetArchived(ctx, userID, ids, &now)
}

func (self *NotificationService) Unarchive(ctx context.Context, userID uuid.UUID, input *NotificationIDsInput) error {
	ids := input.NotificationIDs
	if err := self.ensureOwned(ctx, userID, ids); err != nil {
		return err
	}
	return self.repo.Notification().SetArchived(ctx, userID, ids, nil)
}

func (self *NotificationService) Snooze(ctx context.Context, userID uuid.UUID, input *SnoozeNotificationsInput) error {
	if !input.Until.After(self.now()) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "until must be in the future")
	}
	ids := input.NotificationIDs
	if err := self.ensureOwned(ctx, userID, ids); err != nil {
		return err
	}
	return self.repo.Notification().Snooze(ctx, userID, ids, input.Until)
}

func (self *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID, input *MarkAllReadInput) error {
	if err := self.checkWorkspace(ctx, userID, input.WorkspaceID); err != nil {
		return errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	return self.repo.Notification().MarkAllRead(ctx, userID, input.WorkspaceID, self.now())
}
