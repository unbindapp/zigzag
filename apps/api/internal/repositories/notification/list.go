package notification_repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/notification"
	"github.com/zigzagapp-org/zigzag-api/ent/predicate"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type Filter string

const (
	FilterInbox    Filter = "inbox"
	FilterUnread   Filter = "unread"
	FilterRead     Filter = "read"
	FilterArchived Filter = "archived"
	FilterSnoozed  Filter = "snoozed"
	FilterAll      Filter = "all"
)

type ListInput struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Filter      Filter
	Limit       int
	Cursor      *repository.TimeCursor
	Now         time.Time
}

func filterPredicates(filter Filter, now time.Time) []predicate.Notification {
	awake := notification.Or(notification.SnoozedUntilIsNil(), notification.SnoozedUntilLTE(now))
	switch filter {
	case FilterUnread:
		return []predicate.Notification{notification.ArchivedAtIsNil(), awake, notification.ReadAtIsNil()}
	case FilterRead:
		return []predicate.Notification{notification.ArchivedAtIsNil(), awake, notification.ReadAtNotNil()}
	case FilterArchived:
		return []predicate.Notification{notification.ArchivedAtNotNil()}
	case FilterSnoozed:
		return []predicate.Notification{notification.ArchivedAtIsNil(), notification.SnoozedUntilGT(now)}
	case FilterAll:
		return nil
	default:
		return []predicate.Notification{notification.ArchivedAtIsNil(), awake}
	}
}

// List returns one row more than the limit when another page exists.
func (self *NotificationRepository) List(ctx context.Context, input *ListInput) ([]*ent.Notification, error) {
	query := self.base.DB.Notification.Query().
		Where(notification.UserID(input.UserID), notification.WorkspaceID(input.WorkspaceID)).
		Where(filterPredicates(input.Filter, input.Now)...)
	if input.Cursor != nil {
		query = query.Where(notification.Or(
			notification.CreatedAtLT(input.Cursor.CreatedAt),
			notification.And(notification.CreatedAt(input.Cursor.CreatedAt), notification.IDLT(input.Cursor.ID)),
		))
	}
	return query.
		Order(notification.ByCreatedAt(sql.OrderDesc()), notification.ByID(sql.OrderDesc())).
		Limit(input.Limit + 1).
		WithActor().
		WithIssue(func(q *ent.IssueQuery) { q.WithTeam().WithState() }).
		WithProject().
		All(ctx)
}

func (self *NotificationRepository) CountUnread(ctx context.Context, userID, workspaceID uuid.UUID, now time.Time) (int, error) {
	return self.base.DB.Notification.Query().
		Where(notification.UserID(userID), notification.WorkspaceID(workspaceID)).
		Where(filterPredicates(FilterUnread, now)...).
		Count(ctx)
}

func (self *NotificationRepository) OwnedWorkspaces(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	found, err := self.base.DB.Notification.Query().
		Where(notification.UserID(userID), notification.IDIn(ids...)).
		Select(notification.FieldID, notification.FieldWorkspaceID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	owned := make(map[uuid.UUID]uuid.UUID, len(found))
	for _, row := range found {
		owned[row.ID] = row.WorkspaceID
	}
	return owned, nil
}

func (self *NotificationRepository) SetRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID, readAt *time.Time) error {
	update := self.base.DB.Notification.Update().Where(notification.UserID(userID), notification.IDIn(ids...))
	if readAt == nil {
		return update.ClearReadAt().Exec(ctx)
	}
	return update.SetReadAt(*readAt).Exec(ctx)
}

func (self *NotificationRepository) SetArchived(ctx context.Context, userID uuid.UUID, ids []uuid.UUID, archivedAt *time.Time) error {
	update := self.base.DB.Notification.Update().Where(notification.UserID(userID), notification.IDIn(ids...))
	if archivedAt == nil {
		return update.ClearArchivedAt().Exec(ctx)
	}
	return update.SetArchivedAt(*archivedAt).SetNillableReadAt(archivedAt).Exec(ctx)
}

func (self *NotificationRepository) Snooze(ctx context.Context, userID uuid.UUID, ids []uuid.UUID, until time.Time) error {
	return self.base.DB.Notification.Update().
		Where(notification.UserID(userID), notification.IDIn(ids...)).
		SetSnoozedUntil(until).
		ClearReadAt().
		ClearArchivedAt().
		Exec(ctx)
}

func (self *NotificationRepository) MarkAllRead(ctx context.Context, userID, workspaceID uuid.UUID, now time.Time) error {
	return self.base.DB.Notification.Update().
		Where(notification.UserID(userID), notification.WorkspaceID(workspaceID)).
		Where(filterPredicates(FilterUnread, now)...).
		SetReadAt(now).
		Exec(ctx)
}

func (self *NotificationRepository) HasRecent(ctx context.Context, userID, issueID uuid.UUID, kind schema.NotificationType, since time.Time) (bool, error) {
	return self.base.DB.Notification.Query().
		Where(
			notification.UserID(userID),
			notification.IssueID(issueID),
			notification.TypeEQ(kind),
			notification.CreatedAtGTE(since),
		).
		Exist(ctx)
}
