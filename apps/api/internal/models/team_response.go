package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type WorkflowStateResponse struct {
	ID          uuid.UUID        `json:"id" format:"uuid"`
	TeamID      uuid.UUID        `json:"team_id" format:"uuid"`
	Name        string           `json:"name"`
	Type        schema.StateType `json:"type"`
	Color       string           `json:"color"`
	Description string           `json:"description"`
	Position    float64          `json:"position"`
}

func TransformWorkflowState(entity *ent.WorkflowState) *WorkflowStateResponse {
	if entity == nil {
		return nil
	}
	return &WorkflowStateResponse{
		ID:          entity.ID,
		TeamID:      entity.TeamID,
		Name:        entity.Name,
		Type:        entity.Type,
		Color:       entity.Color,
		Description: entity.Description,
		Position:    entity.Position,
	}
}

func TransformWorkflowStates(entities []*ent.WorkflowState) []*WorkflowStateResponse {
	responses := make([]*WorkflowStateResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformWorkflowState(entity)
	}
	return responses
}

type TeamResponse struct {
	ID                    uuid.UUID                `json:"id" format:"uuid"`
	WorkspaceID           uuid.UUID                `json:"workspace_id" format:"uuid"`
	Name                  string                   `json:"name"`
	Key                   string                   `json:"key" doc:"Prefix of issue identifiers, like ENG in ENG-123"`
	Description           string                   `json:"description"`
	Icon                  string                   `json:"icon"`
	Color                 string                   `json:"color"`
	Private               bool                     `json:"private"`
	Timezone              string                   `json:"timezone"`
	IssueCounter          int                      `json:"issue_counter" doc:"Last issue number handed out"`
	MemberCount           int                      `json:"member_count"`
	IsMember              bool                     `json:"is_member" doc:"Whether the caller is on this team"`
	CyclesEnabled         bool                     `json:"cycles_enabled"`
	CycleDurationWeeks    int                      `json:"cycle_duration_weeks"`
	CycleCooldownWeeks    int                      `json:"cycle_cooldown_weeks"`
	CycleStartDay         int                      `json:"cycle_start_day" doc:"1 Monday through 7 Sunday"`
	CycleAutoMove         bool                     `json:"cycle_auto_move"`
	CycleAutoAddStarted   bool                     `json:"cycle_auto_add_started"`
	EstimateType          schema.EstimateType      `json:"estimate_type"`
	EstimateAllowZero     bool                     `json:"estimate_allow_zero"`
	EstimateExtended      bool                     `json:"estimate_extended"`
	TriageEnabled         bool                     `json:"triage_enabled"`
	DefaultStateID        *uuid.UUID               `json:"default_state_id" nullable:"true" format:"uuid" doc:"State new issues get"`
	TriageStateID         *uuid.UUID               `json:"triage_state_id" nullable:"true" format:"uuid"`
	GithubPrOpenedStateID *uuid.UUID               `json:"github_pr_opened_state_id" nullable:"true" format:"uuid" doc:"State a linked pull request opening moves issues to; null means the first started state"`
	GithubPrMergedStateID *uuid.UUID               `json:"github_pr_merged_state_id" nullable:"true" format:"uuid" doc:"State a merged pull request moves the issues it fixes to; null means the first completed state"`
	AutoCloseMonths       *int                     `json:"auto_close_months" nullable:"true"`
	AutoArchiveMonths     int                      `json:"auto_archive_months"`
	ActiveCycleID         *uuid.UUID               `json:"active_cycle_id" nullable:"true" format:"uuid"`
	States                []*WorkflowStateResponse `json:"states" nullable:"false"`
	CreatedAt             time.Time                `json:"created_at"`
}

func TransformTeamEntity(entity *ent.Team, memberCount int, isMember bool, activeCycleID *uuid.UUID) *TeamResponse {
	states := TransformWorkflowStates(entity.Edges.States)
	return &TeamResponse{
		ID:                    entity.ID,
		WorkspaceID:           entity.WorkspaceID,
		Name:                  entity.Name,
		Key:                   entity.Key,
		Description:           entity.Description,
		Icon:                  entity.Icon,
		Color:                 entity.Color,
		Private:               entity.Private,
		Timezone:              entity.Timezone,
		IssueCounter:          entity.IssueCounter,
		MemberCount:           memberCount,
		IsMember:              isMember,
		CyclesEnabled:         entity.CyclesEnabled,
		CycleDurationWeeks:    entity.CycleDurationWeeks,
		CycleCooldownWeeks:    entity.CycleCooldownWeeks,
		CycleStartDay:         entity.CycleStartDay,
		CycleAutoMove:         entity.CycleAutoMove,
		CycleAutoAddStarted:   entity.CycleAutoAddStarted,
		EstimateType:          entity.EstimateType,
		EstimateAllowZero:     entity.EstimateAllowZero,
		EstimateExtended:      entity.EstimateExtended,
		TriageEnabled:         entity.TriageEnabled,
		DefaultStateID:        entity.DefaultStateID,
		TriageStateID:         entity.TriageStateID,
		GithubPrOpenedStateID: entity.GithubPrOpenedStateID,
		GithubPrMergedStateID: entity.GithubPrMergedStateID,
		AutoCloseMonths:       entity.AutoCloseMonths,
		AutoArchiveMonths:     entity.AutoArchiveMonths,
		ActiveCycleID:         activeCycleID,
		States:                states,
		CreatedAt:             entity.CreatedAt,
	}
}

type TeamMemberResponse struct {
	User     *UserSummary `json:"user"`
	Owner    bool         `json:"owner"`
	JoinedAt time.Time    `json:"joined_at"`
}

func TransformTeamMember(entity *ent.TeamMember) *TeamMemberResponse {
	return &TeamMemberResponse{User: TransformUserSummary(entity.Edges.User), Owner: entity.Owner, JoinedAt: entity.CreatedAt}
}

func TransformTeamMembers(entities []*ent.TeamMember) []*TeamMemberResponse {
	responses := make([]*TeamMemberResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformTeamMember(entity)
	}
	return responses
}
