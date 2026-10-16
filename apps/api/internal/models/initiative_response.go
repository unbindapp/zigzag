package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type InitiativeSummary struct {
	ID    uuid.UUID `json:"id" format:"uuid"`
	Name  string    `json:"name"`
	Slug  string    `json:"slug"`
	Icon  string    `json:"icon"`
	Color string    `json:"color"`
}

func TransformInitiativeSummary(entity *ent.Initiative) *InitiativeSummary {
	if entity == nil {
		return nil
	}
	return &InitiativeSummary{ID: entity.ID, Name: entity.Name, Slug: entity.Slug, Icon: entity.Icon, Color: entity.Color}
}

type InitiativeResponse struct {
	ID          uuid.UUID               `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID               `json:"workspace_id" format:"uuid"`
	Name        string                  `json:"name"`
	Slug        string                  `json:"slug" doc:"URL key, unique in the workspace"`
	Description string                  `json:"description" doc:"Markdown"`
	Icon        string                  `json:"icon"`
	Color       string                  `json:"color"`
	Owner       *UserSummary            `json:"owner,omitempty" required:"false"`
	Status      schema.InitiativeStatus `json:"status"`
	TargetDate  *string                 `json:"target_date" nullable:"true" format:"date"`
	SortOrder   float64                 `json:"sort_order"`
	Health      *schema.Health          `json:"health" nullable:"true" doc:"Health of the latest initiative update"`
	Projects    []*ProjectSummary       `json:"projects" nullable:"false" doc:"Projects in the initiative, trashed ones left out"`
	Progress    ProgressResponse        `json:"progress" doc:"Progress over the issues of all its projects"`
	ArchivedAt  *time.Time              `json:"archived_at" nullable:"true"`
	TrashedAt   *time.Time              `json:"trashed_at" nullable:"true"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

func TransformInitiative(entity *ent.Initiative, health *schema.Health, totals *IssueTotals) *InitiativeResponse {
	projects := make([]*ProjectSummary, len(entity.Edges.Projects))
	for i, project := range entity.Edges.Projects {
		projects[i] = TransformProjectSummary(project)
	}
	return &InitiativeResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Name:        entity.Name,
		Slug:        entity.Slug,
		Description: entity.Description,
		Icon:        entity.Icon,
		Color:       entity.Color,
		Owner:       TransformUserSummary(entity.Edges.Owner),
		Status:      entity.Status,
		TargetDate:  FormatDate(entity.TargetDate),
		SortOrder:   entity.SortOrder,
		Health:      health,
		Projects:    projects,
		Progress:    TransformProgress(totals),
		ArchivedAt:  entity.ArchivedAt,
		TrashedAt:   entity.TrashedAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

// SumIssueTotals merges the totals of several projects.
func SumIssueTotals(parts ...*IssueTotals) *IssueTotals {
	sum := &IssueTotals{Counts: map[schema.StateType]int{}, Scopes: map[schema.StateType]float64{}}
	for _, part := range parts {
		if part == nil {
			continue
		}
		for stateType, count := range part.Counts {
			sum.Counts[stateType] += count
		}
		for stateType, scope := range part.Scopes {
			sum.Scopes[stateType] += scope
		}
	}
	return sum
}

type InitiativeUpdateResponse struct {
	ID           uuid.UUID     `json:"id" format:"uuid"`
	InitiativeID uuid.UUID     `json:"initiative_id" format:"uuid"`
	Author       *UserSummary  `json:"author,omitempty" required:"false"`
	Body         string        `json:"body" doc:"Markdown"`
	Health       schema.Health `json:"health"`
	EditedAt     *time.Time    `json:"edited_at" nullable:"true"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

func TransformInitiativeUpdate(entity *ent.InitiativePost) *InitiativeUpdateResponse {
	return &InitiativeUpdateResponse{
		ID:           entity.ID,
		InitiativeID: entity.InitiativeID,
		Author:       TransformUserSummary(entity.Edges.Author),
		Body:         entity.Body,
		Health:       entity.Health,
		EditedAt:     entity.EditedAt,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
}

func TransformInitiativeUpdates(entities []*ent.InitiativePost) []*InitiativeUpdateResponse {
	responses := make([]*InitiativeUpdateResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformInitiativeUpdate(entity)
	}
	return responses
}
