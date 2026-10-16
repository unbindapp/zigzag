package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
)

type DocumentResponse struct {
	ID          uuid.UUID          `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID          `json:"workspace_id" format:"uuid"`
	Title       string             `json:"title"`
	Slug        string             `json:"slug" doc:"URL key, unique in the workspace"`
	Icon        string             `json:"icon"`
	Content     string             `json:"content" doc:"Markdown"`
	Project     *ProjectSummary    `json:"project,omitempty" required:"false"`
	Initiative  *InitiativeSummary `json:"initiative,omitempty" required:"false"`
	IssueID     *uuid.UUID         `json:"issue_id" nullable:"true" format:"uuid"`
	Creator     *UserSummary       `json:"creator,omitempty" required:"false"`
	UpdatedBy   *UserSummary       `json:"updated_by,omitempty" required:"false"`
	ArchivedAt  *time.Time         `json:"archived_at" nullable:"true"`
	TrashedAt   *time.Time         `json:"trashed_at" nullable:"true"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func TransformDocument(entity *ent.Document) *DocumentResponse {
	return &DocumentResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Title:       entity.Title,
		Slug:        entity.Slug,
		Icon:        entity.Icon,
		Content:     entity.Content,
		Project:     TransformProjectSummary(entity.Edges.Project),
		Initiative:  TransformInitiativeSummary(entity.Edges.Initiative),
		IssueID:     entity.IssueID,
		Creator:     TransformUserSummary(entity.Edges.Creator),
		UpdatedBy:   TransformUserSummary(entity.Edges.UpdatedBy),
		ArchivedAt:  entity.ArchivedAt,
		TrashedAt:   entity.TrashedAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

func TransformDocuments(entities []*ent.Document) []*DocumentResponse {
	responses := make([]*DocumentResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformDocument(entity)
	}
	return responses
}
