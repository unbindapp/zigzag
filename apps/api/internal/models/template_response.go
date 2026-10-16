package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type TemplateResponse struct {
	ID          uuid.UUID           `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID           `json:"workspace_id" format:"uuid"`
	TeamID      *uuid.UUID          `json:"team_id" nullable:"true" format:"uuid" doc:"Team the template belongs to, null for a workspace template"`
	Kind        schema.TemplateKind `json:"kind"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Data        schema.TemplateData `json:"data" doc:"Values a new issue starts with"`
	SortOrder   float64             `json:"sort_order"`
	Creator     *UserSummary        `json:"creator,omitempty" required:"false"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func TransformTemplateEntity(entity *ent.Template) *TemplateResponse {
	return &TemplateResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		TeamID:      entity.TeamID,
		Kind:        entity.Kind,
		Name:        entity.Name,
		Description: entity.Description,
		Data:        entity.Data,
		SortOrder:   entity.SortOrder,
		Creator:     TransformUserSummary(entity.Edges.Creator),
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

func TransformTemplateEntities(entities []*ent.Template) []*TemplateResponse {
	responses := make([]*TemplateResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformTemplateEntity(entity)
	}
	return responses
}
