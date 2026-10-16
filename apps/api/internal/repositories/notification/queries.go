package notification_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/notification"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type CreateNotificationInput struct {
	UserID          uuid.UUID
	ActorID         *uuid.UUID
	WorkspaceID     uuid.UUID
	Type            schema.NotificationType
	IssueID         *uuid.UUID
	CommentID       *uuid.UUID
	ProjectID       *uuid.UUID
	ProjectUpdateID *uuid.UUID
	Title           string
	Body            string
}

func (self *NotificationRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateNotificationInput) (*ent.Notification, error) {
	return self.base.Client(tx).Notification.Create().
		SetUserID(input.UserID).
		SetNillableActorID(input.ActorID).
		SetWorkspaceID(input.WorkspaceID).
		SetType(input.Type).
		SetNillableIssueID(input.IssueID).
		SetNillableCommentID(input.CommentID).
		SetNillableProjectID(input.ProjectID).
		SetNillableProjectUpdateID(input.ProjectUpdateID).
		SetTitle(input.Title).
		SetBody(input.Body).
		Save(ctx)
}

func (self *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Notification, error) {
	return self.base.DB.Notification.Get(ctx, id)
}

func (self *NotificationRepository) ResolveWorkspaceOfNotification(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Notification.Query().Where(notification.ID(id)).Select(notification.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
