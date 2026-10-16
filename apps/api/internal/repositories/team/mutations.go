package team_repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/teammember"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type defaultState struct {
	Name  string
	Type  schema.StateType
	Color string
}

var (
	triageState   = defaultState{Name: "Triage", Type: schema.StateTypeTriage, Color: "#fc7840"}
	todoStateName = "Todo"
	DefaultStates = []defaultState{
		{Name: "Backlog", Type: schema.StateTypeBacklog, Color: "#95a2b3"},
		{Name: todoStateName, Type: schema.StateTypeUnstarted, Color: "#e2e2e2"},
		{Name: "In Progress", Type: schema.StateTypeStarted, Color: "#f2c94c"},
		{Name: "In Review", Type: schema.StateTypeStarted, Color: "#0f783c"},
		{Name: "Done", Type: schema.StateTypeCompleted, Color: "#5e6ad2"},
		{Name: "Canceled", Type: schema.StateTypeCanceled, Color: "#95a2b3"},
		{Name: "Duplicate", Type: schema.StateTypeCanceled, Color: "#95a2b3"},
	}
)

type CreateTeamInput struct {
	WorkspaceID   uuid.UUID
	Name          string
	Key           string
	Description   string
	Icon          string
	Color         string
	Private       bool
	Timezone      string
	TriageEnabled bool
	CreatorID     uuid.UUID
}

func DefaultTeamInput(workspace *ent.Workspace, creatorID uuid.UUID) *CreateTeamInput {
	return &CreateTeamInput{
		WorkspaceID: workspace.ID,
		Name:        workspace.Name,
		Key:         names.TeamKeyFromName(workspace.Name),
		CreatorID:   creatorID,
	}
}

// Writes several tables: call it inside a transaction.
func (self *TeamRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateTeamInput) (*ent.Team, error) {
	db := self.base.Client(tx)
	timezone := input.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	created, err := db.Team.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetName(input.Name).
		SetKey(input.Key).
		SetDescription(input.Description).
		SetIcon(input.Icon).
		SetColor(input.Color).
		SetPrivate(input.Private).
		SetTimezone(timezone).
		SetTriageEnabled(input.TriageEnabled).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	states := DefaultStates
	if input.TriageEnabled {
		states = append([]defaultState{triageState}, DefaultStates...)
	}
	positions := map[schema.StateType]float64{}
	builders := make([]*ent.WorkflowStateCreate, len(states))
	for i, state := range states {
		builders[i] = db.WorkflowState.Create().
			SetTeamID(created.ID).
			SetName(state.Name).
			SetType(state.Type).
			SetColor(state.Color).
			SetPosition(positions[state.Type])
		positions[state.Type]++
	}
	saved, err := db.WorkflowState.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, err
	}

	update := db.Team.UpdateOneID(created.ID)
	for _, state := range saved {
		switch {
		case state.Name == todoStateName:
			update.SetDefaultStateID(state.ID)
		case state.Type == schema.StateTypeTriage:
			update.SetTriageStateID(state.ID)
		}
	}
	if err := update.Exec(ctx); err != nil {
		return nil, err
	}

	if input.CreatorID != uuid.Nil {
		if err := self.AddMember(ctx, tx, created.ID, input.CreatorID, true); err != nil {
			return nil, err
		}
	}
	return db.Team.Query().Where(team.ID(created.ID)).WithStates(withActiveStates).Only(ctx)
}

type UpdateTeamInput struct {
	Name                     *string
	Key                      *string
	Description              *string
	Icon                     *string
	Color                    *string
	Private                  *bool
	Timezone                 *string
	CyclesEnabled            *bool
	CycleDurationWeeks       *int
	CycleCooldownWeeks       *int
	CycleStartDay            *int
	CycleAutoMove            *bool
	CycleAutoAddStarted      *bool
	EstimateType             *schema.EstimateType
	EstimateAllowZero        *bool
	EstimateExtended         *bool
	TriageEnabled            *bool
	DefaultStateID           *uuid.UUID
	TriageStateID            *uuid.UUID
	GithubPrOpenedStateID    *uuid.UUID
	ClearGithubPrOpenedState bool
	GithubPrMergedStateID    *uuid.UUID
	ClearGithubPrMergedState bool
	AutoCloseMonths          *int
	ClearAutoClose           bool
	AutoArchiveMonths        *int
}

func (self *TeamRepository) Update(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID, input *UpdateTeamInput) (*ent.Team, error) {
	db := self.base.Client(tx)
	m := db.Team.UpdateOneID(teamID)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.Key != nil {
		m.SetKey(*input.Key)
	}
	if input.Description != nil {
		m.SetDescription(*input.Description)
	}
	if input.Icon != nil {
		m.SetIcon(*input.Icon)
	}
	if input.Color != nil {
		m.SetColor(*input.Color)
	}
	if input.Private != nil {
		m.SetPrivate(*input.Private)
	}
	if input.Timezone != nil {
		m.SetTimezone(*input.Timezone)
	}
	if input.CyclesEnabled != nil {
		m.SetCyclesEnabled(*input.CyclesEnabled)
	}
	if input.CycleDurationWeeks != nil {
		m.SetCycleDurationWeeks(*input.CycleDurationWeeks)
	}
	if input.CycleCooldownWeeks != nil {
		m.SetCycleCooldownWeeks(*input.CycleCooldownWeeks)
	}
	if input.CycleStartDay != nil {
		m.SetCycleStartDay(*input.CycleStartDay)
	}
	if input.CycleAutoMove != nil {
		m.SetCycleAutoMove(*input.CycleAutoMove)
	}
	if input.CycleAutoAddStarted != nil {
		m.SetCycleAutoAddStarted(*input.CycleAutoAddStarted)
	}
	if input.EstimateType != nil {
		m.SetEstimateType(*input.EstimateType)
	}
	if input.EstimateAllowZero != nil {
		m.SetEstimateAllowZero(*input.EstimateAllowZero)
	}
	if input.EstimateExtended != nil {
		m.SetEstimateExtended(*input.EstimateExtended)
	}
	if input.TriageEnabled != nil {
		m.SetTriageEnabled(*input.TriageEnabled)
	}
	if input.DefaultStateID != nil {
		m.SetDefaultStateID(*input.DefaultStateID)
	}
	if input.TriageStateID != nil {
		m.SetTriageStateID(*input.TriageStateID)
	}
	if input.GithubPrOpenedStateID != nil {
		m.SetGithubPrOpenedStateID(*input.GithubPrOpenedStateID)
	}
	if input.ClearGithubPrOpenedState {
		m.ClearGithubPrOpenedStateID()
	}
	if input.GithubPrMergedStateID != nil {
		m.SetGithubPrMergedStateID(*input.GithubPrMergedStateID)
	}
	if input.ClearGithubPrMergedState {
		m.ClearGithubPrMergedStateID()
	}
	if input.AutoCloseMonths != nil {
		m.SetAutoCloseMonths(*input.AutoCloseMonths)
	}
	if input.ClearAutoClose {
		m.ClearAutoCloseMonths()
	}
	if input.AutoArchiveMonths != nil {
		m.SetAutoArchiveMonths(*input.AutoArchiveMonths)
	}
	if err := m.Exec(ctx); err != nil {
		return nil, err
	}
	return db.Team.Query().Where(team.ID(teamID)).WithStates(withActiveStates).Only(ctx)
}

func (self *TeamRepository) Delete(ctx context.Context, teamID uuid.UUID) error {
	return self.base.DB.Team.DeleteOneID(teamID).Exec(ctx)
}

func (self *TeamRepository) AddMember(ctx context.Context, tx repository.TxInterface, teamID, userID uuid.UUID, owner bool) error {
	err := self.base.Client(tx).TeamMember.Create().
		SetTeamID(teamID).
		SetUserID(userID).
		SetOwner(owner).
		OnConflictColumns(teammember.FieldTeamID, teammember.FieldUserID).
		DoNothing().
		Exec(ctx)
	return repository.IgnoreExisting(err)
}

func (self *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	deleted, err := self.base.DB.TeamMember.Delete().
		Where(teammember.TeamID(teamID), teammember.UserID(userID)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// NextIssueNumber must run in the transaction that inserts the issue, so a rollback returns the number.
func (self *TeamRepository) NextIssueNumber(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID) (int, error) {
	rows, err := self.base.Client(tx).QueryContext(ctx,
		"UPDATE teams SET issue_counter = issue_counter + 1 WHERE id = $1 RETURNING issue_counter", teamID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, &ent.NotFoundError{}
	}
	var number int
	if err := rows.Scan(&number); err != nil {
		return 0, fmt.Errorf("scan issue counter: %w", err)
	}
	return number, rows.Err()
}
