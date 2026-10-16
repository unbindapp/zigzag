package notification_repo

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/notification"
	"github.com/unbindapp/zigzag-api/ent/predicate"
	"github.com/unbindapp/zigzag-api/ent/schema"
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

var ErrInvalidCursor = errors.New("invalid cursor")

type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

func (self Cursor) Encode() string {
	raw := strconv.FormatInt(self.CreatedAt.UnixNano(), 10) + "|" + self.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(value string) (*Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	nanos, id, found := strings.Cut(string(raw), "|")
	if !found {
		return nil, ErrInvalidCursor
	}
	unixNano, err := strconv.ParseInt(nanos, 10, 64)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	return &Cursor{CreatedAt: time.Unix(0, unixNano).UTC(), ID: parsedID}, nil
}

type ListInput struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Filter      Filter
	Limit       int
	Cursor      *Cursor
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

// OwnedWorkspaces leaves out ids that belong to other users.
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
