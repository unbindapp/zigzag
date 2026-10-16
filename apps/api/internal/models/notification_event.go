package models

import (
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type NotificationType = schema.NotificationType

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
