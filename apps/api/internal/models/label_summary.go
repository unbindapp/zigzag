package models

import (
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
)

type LabelSummary struct {
	ID       uuid.UUID  `json:"id" format:"uuid"`
	Name     string     `json:"name"`
	Color    string     `json:"color"`
	ParentID *uuid.UUID `json:"parent_id" nullable:"true" format:"uuid" doc:"The label group this label belongs to"`
}

func TransformLabelSummary(entity *ent.Label) *LabelSummary {
	if entity == nil {
		return nil
	}
	return &LabelSummary{ID: entity.ID, Name: entity.Name, Color: entity.Color, ParentID: entity.ParentID}
}

func TransformLabelSummaries(entities []*ent.Label) []*LabelSummary {
	summaries := make([]*LabelSummary, len(entities))
	for i, entity := range entities {
		summaries[i] = TransformLabelSummary(entity)
	}
	return summaries
}
