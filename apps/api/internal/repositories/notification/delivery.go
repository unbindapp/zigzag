package notification_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/notification"
	"github.com/unbindapp/zigzag-api/ent/notificationsetting"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/user"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/ent/workspacemember"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

func (self *NotificationRepository) CreateMany(ctx context.Context, tx repository.TxInterface, inputs []*CreateNotificationInput) ([]*ent.Notification, error) {
	client := self.base.Client(tx)
	builders := make([]*ent.NotificationCreate, len(inputs))
	for i, input := range inputs {
		builders[i] = client.Notification.Create().
			SetUserID(input.UserID).
			SetNillableActorID(input.ActorID).
			SetWorkspaceID(input.WorkspaceID).
			SetType(input.Type).
			SetNillableIssueID(input.IssueID).
			SetNillableCommentID(input.CommentID).
			SetNillableProjectID(input.ProjectID).
			SetNillableProjectUpdateID(input.ProjectUpdateID).
			SetTitle(input.Title).
			SetBody(input.Body)
	}
	return client.Notification.CreateBulk(builders...).Save(ctx)
}

func (self *NotificationRepository) ActiveMembers(ctx context.Context, workspaceID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := self.base.DB.WorkspaceMember.Query().
		Where(
			workspacemember.WorkspaceID(workspaceID),
			workspacemember.UserIDIn(userIDs...),
			workspacemember.HasUserWith(user.Active(true)),
		).
		Select(workspacemember.FieldUserID).
		Scan(ctx, &ids)
	return ids, err
}

// Settings has no entry for users who kept the defaults.
func (self *NotificationRepository) Settings(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*ent.NotificationSetting, error) {
	rows, err := self.base.DB.NotificationSetting.Query().Where(notificationsetting.UserIDIn(userIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	settings := make(map[uuid.UUID]*ent.NotificationSetting, len(rows))
	for _, row := range rows {
		settings[row.UserID] = row
	}
	return settings, nil
}

func (self *NotificationRepository) PendingEmail(ctx context.Context, createdBefore, createdAfter time.Time) ([]*ent.Notification, error) {
	return self.base.DB.Notification.Query().
		Where(
			notification.EmailedAtIsNil(),
			notification.ReadAtIsNil(),
			notification.ArchivedAtIsNil(),
			notification.Or(notification.SnoozedUntilIsNil(), notification.SnoozedUntilLTE(createdBefore)),
			notification.CreatedAtLT(createdBefore),
			notification.CreatedAtGT(createdAfter),
		).
		Order(ent.Asc(notification.FieldCreatedAt)).
		WithUser().
		WithActor().
		WithWorkspace().
		WithIssue(func(q *ent.IssueQuery) { q.WithTeam() }).
		WithProject().
		All(ctx)
}

func (self *NotificationRepository) MarkEmailed(ctx context.Context, ids []uuid.UUID, at time.Time) error {
	return self.base.DB.Notification.Update().Where(notification.IDIn(ids...)).SetEmailedAt(at).Exec(ctx)
}

func (self *NotificationRepository) DueBetween(ctx context.Context, from, to time.Time) ([]*ent.Issue, error) {
	return self.base.DB.Issue.Query().
		Where(
			issue.DueDateGTE(from),
			issue.DueDateLTE(to),
			issue.AssigneeIDNotNil(),
			issue.ArchivedAtIsNil(),
			issue.TrashedAtIsNil(),
			issue.HasStateWith(workflowstate.TypeNotIn(schema.StateTypeCompleted, schema.StateTypeCanceled)),
		).
		WithTeam().
		All(ctx)
}
