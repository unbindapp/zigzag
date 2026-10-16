package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type ViewResponse struct {
	ID          uuid.UUID          `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID          `json:"workspace_id" format:"uuid"`
	TeamID      *uuid.UUID         `json:"team_id" nullable:"true" format:"uuid" doc:"Team the view belongs to, null for a workspace view"`
	Owner       *UserSummary       `json:"owner,omitempty" required:"false"`
	IsOwner     bool               `json:"is_owner" doc:"Whether the caller owns the view"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Icon        string             `json:"icon"`
	Color       string             `json:"color"`
	Kind        schema.ViewKind    `json:"kind"`
	Shared      bool               `json:"shared" doc:"Shared views are visible to everyone in the workspace"`
	Filters     map[string]any     `json:"filters" doc:"list-issues parameters the view applies, such as state_types or label_ids"`
	Display     schema.ViewDisplay `json:"display"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func TransformViewEntity(entity *ent.View, viewerID uuid.UUID) *ViewResponse {
	filters := entity.Filters
	if filters == nil {
		filters = map[string]any{}
	}
	return &ViewResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		TeamID:      entity.TeamID,
		Owner:       TransformUserSummary(entity.Edges.Owner),
		IsOwner:     entity.OwnerID == viewerID,
		Name:        entity.Name,
		Description: entity.Description,
		Icon:        entity.Icon,
		Color:       entity.Color,
		Kind:        entity.Kind,
		Shared:      entity.Shared,
		Filters:     filters,
		Display:     entity.Display,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}
