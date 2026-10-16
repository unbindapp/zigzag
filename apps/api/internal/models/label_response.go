package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
)

type LabelResponse struct {
	ID                uuid.UUID  `json:"id" format:"uuid"`
	WorkspaceID       uuid.UUID  `json:"workspace_id" format:"uuid"`
	TeamID            *uuid.UUID `json:"team_id" nullable:"true" format:"uuid" doc:"Null for workspace labels shared by every team"`
	Name              string     `json:"name"`
	Color             string     `json:"color"`
	Description       string     `json:"description"`
	ParentID          *uuid.UUID `json:"parent_id" nullable:"true" format:"uuid" doc:"The label group this label belongs to"`
	IsGroup           bool       `json:"is_group" doc:"Groups hold labels and are never applied to issues themselves"`
	GroupSingleSelect bool       `json:"group_single_select" doc:"Only one label of this group may be on an issue at a time"`
	SortOrder         float64    `json:"sort_order"`
	ArchivedAt        *time.Time `json:"archived_at" nullable:"true"`
}

type LabelTreeResponse struct {
	LabelResponse
	Children []*LabelResponse `json:"children" nullable:"false" doc:"Labels of this group, empty for plain labels"`
}

func TransformLabelEntity(entity *ent.Label) *LabelResponse {
	return &LabelResponse{
		ID:                entity.ID,
		WorkspaceID:       entity.WorkspaceID,
		TeamID:            entity.TeamID,
		Name:              entity.Name,
		Color:             entity.Color,
		Description:       entity.Description,
		ParentID:          entity.ParentID,
		IsGroup:           entity.IsGroup,
		GroupSingleSelect: entity.GroupSingleSelect,
		SortOrder:         entity.SortOrder,
		ArchivedAt:        entity.ArchivedAt,
	}
}

func TransformLabelTree(entities []*ent.Label) []*LabelTreeResponse {
	byID := make(map[uuid.UUID]*LabelTreeResponse, len(entities))
	for _, entity := range entities {
		byID[entity.ID] = &LabelTreeResponse{LabelResponse: *TransformLabelEntity(entity), Children: []*LabelResponse{}}
	}
	roots := make([]*LabelTreeResponse, 0, len(entities))
	for _, entity := range entities {
		response := byID[entity.ID]
		if entity.ParentID == nil || byID[*entity.ParentID] == nil {
			roots = append(roots, response)
			continue
		}
		parent := byID[*entity.ParentID]
		parent.Children = append(parent.Children, &response.LabelResponse)
	}
	return roots
}
