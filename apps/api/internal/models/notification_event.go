package models

import (
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type NotificationType = schema.NotificationType

const (
	StateChangedBodyPrefix    = "Moved to "
	PriorityChangedBodyPrefix = "Priority set to "
	BlockedByBodyPrefix       = "Blocked by "
)

type NotificationEvent struct {
	Type            NotificationType
	WorkspaceID     uuid.UUID
	ActorID         *uuid.UUID
	RecipientIDs    []uuid.UUID
	IssueID         *uuid.UUID
	CommentID       *uuid.UUID
	ProjectID       *uuid.UUID
	ProjectUpdateID *uuid.UUID
	Title           string
	Body            string
}
