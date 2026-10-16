package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type NotificationResponse struct {
	ID              uuid.UUID               `json:"id" format:"uuid"`
	Type            schema.NotificationType `json:"type"`
	WorkspaceID     uuid.UUID               `json:"workspace_id" format:"uuid"`
	Title           string                  `json:"title"`
	Body            string                  `json:"body" doc:"Short plain text snippet, may be empty"`
	Actor           *UserSummary            `json:"actor,omitempty" required:"false" doc:"Who caused the notification; absent for system notifications"`
	Issue           *IssueSummary           `json:"issue,omitempty" required:"false"`
	CommentID       *uuid.UUID              `json:"comment_id" nullable:"true" format:"uuid"`
	Project         *ProjectSummary         `json:"project,omitempty" required:"false"`
	ProjectUpdateID *uuid.UUID              `json:"project_update_id" nullable:"true" format:"uuid"`
	ReadAt          *time.Time              `json:"read_at" nullable:"true"`
	ArchivedAt      *time.Time              `json:"archived_at" nullable:"true"`
	SnoozedUntil    *time.Time              `json:"snoozed_until" nullable:"true"`
	CreatedAt       time.Time               `json:"created_at"`
}

func TransformNotification(entity *ent.Notification) *NotificationResponse {
	if entity == nil {
		return nil
	}
	return &NotificationResponse{
		ID:              entity.ID,
		Type:            entity.Type,
		WorkspaceID:     entity.WorkspaceID,
		Title:           entity.Title,
		Body:            entity.Body,
		Actor:           TransformUserSummary(entity.Edges.Actor),
		Issue:           TransformIssueSummary(entity.Edges.Issue),
		CommentID:       entity.CommentID,
		Project:         TransformProjectSummary(entity.Edges.Project),
		ProjectUpdateID: entity.ProjectUpdateID,
		ReadAt:          entity.ReadAt,
		ArchivedAt:      entity.ArchivedAt,
		SnoozedUntil:    entity.SnoozedUntil,
		CreatedAt:       entity.CreatedAt,
	}
}

func TransformNotifications(entities []*ent.Notification) []*NotificationResponse {
	responses := make([]*NotificationResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformNotification(entity)
	}
	return responses
}

type NotificationCountsResponse struct {
	Unread int `json:"unread" doc:"Unread notifications in the inbox, not counting archived or snoozed ones"`
}
