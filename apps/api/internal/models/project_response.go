package models

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type ProjectSummary struct {
	ID    uuid.UUID `json:"id" format:"uuid"`
	Name  string    `json:"name"`
	Slug  string    `json:"slug"`
	Icon  string    `json:"icon"`
	Color string    `json:"color"`
}

func TransformProjectSummary(entity *ent.Project) *ProjectSummary {
	if entity == nil {
		return nil
	}
	return &ProjectSummary{ID: entity.ID, Name: entity.Name, Slug: entity.Slug, Icon: entity.Icon, Color: entity.Color}
}

type TeamSummary struct {
	ID    uuid.UUID `json:"id" format:"uuid"`
	Key   string    `json:"key"`
	Name  string    `json:"name"`
	Icon  string    `json:"icon"`
	Color string    `json:"color"`
}

func TransformTeamSummaries(entities []*ent.Team) []*TeamSummary {
	summaries := make([]*TeamSummary, len(entities))
	for i, entity := range entities {
		summaries[i] = &TeamSummary{ID: entity.ID, Key: entity.Key, Name: entity.Name, Icon: entity.Icon, Color: entity.Color}
	}
	return summaries
}

type ProjectStatusResponse struct {
	ID          uuid.UUID                `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID                `json:"workspace_id" format:"uuid"`
	Name        string                   `json:"name"`
	Type        schema.ProjectStatusType `json:"type"`
	Color       string                   `json:"color"`
	Description string                   `json:"description"`
	Position    float64                  `json:"position" doc:"Order within its type"`
}

func TransformProjectStatus(entity *ent.ProjectStatus) *ProjectStatusResponse {
	if entity == nil {
		return nil
	}
	return &ProjectStatusResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Name:        entity.Name,
		Type:        entity.Type,
		Color:       entity.Color,
		Description: entity.Description,
		Position:    entity.Position,
	}
}

func TransformProjectStatuses(entities []*ent.ProjectStatus) []*ProjectStatusResponse {
	responses := make([]*ProjectStatusResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformProjectStatus(entity)
	}
	return responses
}

// IssueTotals is the issue count and scope (sum of estimates, 1 for unestimated issues) per state type.
type IssueTotals struct {
	Counts map[schema.StateType]int
	Scopes map[schema.StateType]float64
}

type ProgressResponse struct {
	Scope     float64 `json:"scope" doc:"Sum of estimates of issues that are not canceled; an issue without an estimate counts 1"`
	Completed float64 `json:"completed" doc:"Scope of completed issues"`
	Started   float64 `json:"started" doc:"Scope of started issues"`
	Percent   int     `json:"percent" doc:"Completed scope as a percentage of the scope, 0 to 100"`
}

func TransformProgress(totals *IssueTotals) ProgressResponse {
	if totals == nil {
		return ProgressResponse{}
	}
	progress := ProgressResponse{Completed: totals.Scopes[schema.StateTypeCompleted], Started: totals.Scopes[schema.StateTypeStarted]}
	for stateType, scope := range totals.Scopes {
		if stateType == schema.StateTypeCanceled {
			continue
		}
		progress.Scope += scope
	}
	if progress.Scope > 0 {
		progress.Percent = int(math.Round(progress.Completed / progress.Scope * 100))
	}
	return progress
}

type IssueCountsByStateType struct {
	Triage    int `json:"triage"`
	Backlog   int `json:"backlog"`
	Unstarted int `json:"unstarted"`
	Started   int `json:"started"`
	Completed int `json:"completed"`
	Canceled  int `json:"canceled"`
}

type IssueCountsResponse struct {
	Total       int                    `json:"total"`
	ByStateType IssueCountsByStateType `json:"by_state_type"`
}

func TransformIssueCounts(totals *IssueTotals) IssueCountsResponse {
	if totals == nil {
		return IssueCountsResponse{}
	}
	counts := totals.Counts
	response := IssueCountsResponse{ByStateType: IssueCountsByStateType{
		Triage:    counts[schema.StateTypeTriage],
		Backlog:   counts[schema.StateTypeBacklog],
		Unstarted: counts[schema.StateTypeUnstarted],
		Started:   counts[schema.StateTypeStarted],
		Completed: counts[schema.StateTypeCompleted],
		Canceled:  counts[schema.StateTypeCanceled],
	}}
	for _, count := range counts {
		response.Total += count
	}
	return response
}

type MilestoneResponse struct {
	ID          uuid.UUID        `json:"id" format:"uuid"`
	ProjectID   uuid.UUID        `json:"project_id" format:"uuid"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	TargetDate  *string          `json:"target_date" nullable:"true" format:"date"`
	SortOrder   float64          `json:"sort_order"`
	Progress    ProgressResponse `json:"progress"`
	IssueCount  int              `json:"issue_count"`
	CreatedAt   time.Time        `json:"created_at"`
}

func TransformMilestone(entity *ent.ProjectMilestone, totals *IssueTotals) *MilestoneResponse {
	return &MilestoneResponse{
		ID:          entity.ID,
		ProjectID:   entity.ProjectID,
		Name:        entity.Name,
		Description: entity.Description,
		TargetDate:  FormatDate(entity.TargetDate),
		SortOrder:   entity.SortOrder,
		Progress:    TransformProgress(totals),
		IssueCount:  TransformIssueCounts(totals).Total,
		CreatedAt:   entity.CreatedAt,
	}
}

func TransformMilestones(entities []*ent.ProjectMilestone, totals map[uuid.UUID]*IssueTotals) []*MilestoneResponse {
	responses := make([]*MilestoneResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformMilestone(entity, totals[entity.ID])
	}
	return responses
}

type ProjectLinkResponse struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	ProjectID uuid.UUID `json:"project_id" format:"uuid"`
	URL       string    `json:"url"`
	Label     string    `json:"label"`
	SortOrder float64   `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func TransformProjectLink(entity *ent.ProjectLink) *ProjectLinkResponse {
	return &ProjectLinkResponse{
		ID:        entity.ID,
		ProjectID: entity.ProjectID,
		URL:       entity.URL,
		Label:     entity.Label,
		SortOrder: entity.SortOrder,
		CreatedAt: entity.CreatedAt,
	}
}

func TransformProjectLinks(entities []*ent.ProjectLink) []*ProjectLinkResponse {
	responses := make([]*ProjectLinkResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformProjectLink(entity)
	}
	return responses
}

type ProjectUpdateResponse struct {
	ID        uuid.UUID        `json:"id" format:"uuid"`
	ProjectID uuid.UUID        `json:"project_id" format:"uuid"`
	Author    *UserSummary     `json:"author,omitempty" required:"false"`
	Body      string           `json:"body" doc:"Markdown"`
	Health    schema.Health    `json:"health"`
	EditedAt  *time.Time       `json:"edited_at" nullable:"true"`
	Reactions []*ReactionGroup `json:"reactions" nullable:"false"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func TransformProjectUpdate(entity *ent.ProjectPost, callerID uuid.UUID) *ProjectUpdateResponse {
	if entity == nil {
		return nil
	}
	return &ProjectUpdateResponse{
		ID:        entity.ID,
		ProjectID: entity.ProjectID,
		Author:    TransformUserSummary(entity.Edges.Author),
		Body:      entity.Body,
		Health:    entity.Health,
		EditedAt:  entity.EditedAt,
		Reactions: TransformReactionGroups(entity.Edges.Reactions, callerID),
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func TransformProjectUpdates(entities []*ent.ProjectPost, callerID uuid.UUID) []*ProjectUpdateResponse {
	responses := make([]*ProjectUpdateResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformProjectUpdate(entity, callerID)
	}
	return responses
}

type ProjectResponse struct {
	ID                   uuid.UUID              `json:"id" format:"uuid"`
	WorkspaceID          uuid.UUID              `json:"workspace_id" format:"uuid"`
	Name                 string                 `json:"name"`
	Slug                 string                 `json:"slug" doc:"URL key, unique in the workspace"`
	Summary              string                 `json:"summary" doc:"One line summary"`
	Description          string                 `json:"description" doc:"Markdown"`
	Icon                 string                 `json:"icon"`
	Color                string                 `json:"color"`
	Status               *ProjectStatusResponse `json:"status"`
	Lead                 *UserSummary           `json:"lead,omitempty" required:"false"`
	Priority             int                    `json:"priority" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	StartDate            *string                `json:"start_date" nullable:"true" format:"date"`
	TargetDate           *string                `json:"target_date" nullable:"true" format:"date"`
	StartDateResolution  schema.DateResolution  `json:"start_date_resolution" doc:"How precise start_date is: day, month, quarter, half_year or year"`
	TargetDateResolution schema.DateResolution  `json:"target_date_resolution" doc:"How precise target_date is: day, month, quarter, half_year or year"`
	SortOrder            float64                `json:"sort_order"`
	Health               *schema.Health         `json:"health" nullable:"true" doc:"Health of the latest project update"`
	Teams                []*TeamSummary         `json:"teams" nullable:"false"`
	Members              []*UserSummary         `json:"members" nullable:"false"`
	Labels               []*LabelSummary        `json:"labels" nullable:"false"`
	Milestones           []*MilestoneResponse   `json:"milestones" nullable:"false"`
	Links                []*ProjectLinkResponse `json:"links" nullable:"false"`
	InitiativeIDs        []uuid.UUID            `json:"initiative_ids" nullable:"false"`
	Progress             ProgressResponse       `json:"progress"`
	IssueCounts          IssueCountsResponse    `json:"issue_counts"`
	LastUpdate           *ProjectUpdateResponse `json:"last_update,omitempty" required:"false" doc:"Latest status update, on get-project only"`
	StartedAt            *time.Time             `json:"started_at" nullable:"true"`
	CompletedAt          *time.Time             `json:"completed_at" nullable:"true"`
	CanceledAt           *time.Time             `json:"canceled_at" nullable:"true"`
	ArchivedAt           *time.Time             `json:"archived_at" nullable:"true"`
	TrashedAt            *time.Time             `json:"trashed_at" nullable:"true"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func TransformProject(entity *ent.Project, totals *IssueTotals, milestoneTotals map[uuid.UUID]*IssueTotals) *ProjectResponse {
	members := make([]*UserSummary, len(entity.Edges.Members))
	for i, member := range entity.Edges.Members {
		members[i] = TransformUserSummary(member)
	}
	initiativeIDs := make([]uuid.UUID, len(entity.Edges.Initiatives))
	for i, initiative := range entity.Edges.Initiatives {
		initiativeIDs[i] = initiative.ID
	}
	return &ProjectResponse{
		ID:                   entity.ID,
		WorkspaceID:          entity.WorkspaceID,
		Name:                 entity.Name,
		Slug:                 entity.Slug,
		Summary:              entity.Summary,
		Description:          entity.Description,
		Icon:                 entity.Icon,
		Color:                entity.Color,
		Status:               TransformProjectStatus(entity.Edges.Status),
		Lead:                 TransformUserSummary(entity.Edges.Lead),
		Priority:             entity.Priority,
		StartDate:            FormatDate(entity.StartDate),
		TargetDate:           FormatDate(entity.TargetDate),
		StartDateResolution:  entity.StartDateResolution,
		TargetDateResolution: entity.TargetDateResolution,
		SortOrder:            entity.SortOrder,
		Health:               entity.Health,
		Teams:                TransformTeamSummaries(entity.Edges.Teams),
		Members:              members,
		Labels:               TransformLabelSummaries(entity.Edges.Labels),
		Milestones:           TransformMilestones(entity.Edges.Milestones, milestoneTotals),
		Links:                TransformProjectLinks(entity.Edges.Links),
		InitiativeIDs:        initiativeIDs,
		Progress:             TransformProgress(totals),
		IssueCounts:          TransformIssueCounts(totals),
		StartedAt:            entity.StartedAt,
		CompletedAt:          entity.CompletedAt,
		CanceledAt:           entity.CanceledAt,
		ArchivedAt:           entity.ArchivedAt,
		TrashedAt:            entity.TrashedAt,
		CreatedAt:            entity.CreatedAt,
		UpdatedAt:            entity.UpdatedAt,
	}
}

func TransformProjects(entities []*ent.Project, totals, milestoneTotals map[uuid.UUID]*IssueTotals) []*ProjectResponse {
	responses := make([]*ProjectResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformProject(entity, totals[entity.ID], milestoneTotals)
	}
	return responses
}
