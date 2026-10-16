package team_repo

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/teammember"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

const importBatchSize = 500

type ImportedTeam struct {
	ID             uuid.UUID
	Key            string
	Name           string
	Description    string
	Icon           string
	Color          string
	Timezone       string
	CyclesEnabled  bool
	TriageEnabled  bool
	DefaultStateID *uuid.UUID
	TriageStateID  *uuid.UUID
}

type ImportedState struct {
	ID       uuid.UUID
	TeamID   uuid.UUID
	Name     string
	Type     schema.StateType
	Color    string
	Position float64
}

type ImportedTeamMember struct {
	TeamID uuid.UUID
	UserID uuid.UUID
	Owner  bool
}

func (self *TeamRepository) StatesOfWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*ent.WorkflowState, error) {
	return self.base.DB.WorkflowState.Query().
		Where(workflowstate.HasTeamWith(team.WorkspaceID(workspaceID))).
		Order(ent.Asc(workflowstate.FieldPosition)).
		All(ctx)
}

func (self *TeamRepository) MembersOfWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*ent.TeamMember, error) {
	return self.base.DB.TeamMember.Query().Where(teammember.HasTeamWith(team.WorkspaceID(workspaceID))).All(ctx)
}

func (self *TeamRepository) ImportTeams(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, teams []ImportedTeam, states []ImportedState) error {
	db := self.base.Client(tx)
	teamBuilders := make([]*ent.TeamCreate, len(teams))
	for i, t := range teams {
		teamBuilders[i] = db.Team.Create().
			SetID(t.ID).
			SetWorkspaceID(workspaceID).
			SetKey(t.Key).
			SetName(t.Name).
			SetDescription(t.Description).
			SetIcon(t.Icon).
			SetColor(t.Color).
			SetTimezone(t.Timezone).
			SetCyclesEnabled(t.CyclesEnabled).
			SetTriageEnabled(t.TriageEnabled)
	}
	for batch := range slices.Chunk(teamBuilders, importBatchSize) {
		if err := db.Team.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}

	stateBuilders := make([]*ent.WorkflowStateCreate, len(states))
	for i, s := range states {
		stateBuilders[i] = db.WorkflowState.Create().
			SetID(s.ID).
			SetTeamID(s.TeamID).
			SetName(s.Name).
			SetType(s.Type).
			SetColor(s.Color).
			SetPosition(s.Position)
	}
	for batch := range slices.Chunk(stateBuilders, importBatchSize) {
		if err := db.WorkflowState.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}

	for _, t := range teams {
		if t.DefaultStateID == nil && t.TriageStateID == nil {
			continue
		}
		err := db.Team.UpdateOneID(t.ID).
			SetNillableDefaultStateID(t.DefaultStateID).
			SetNillableTriageStateID(t.TriageStateID).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *TeamRepository) AddMembers(ctx context.Context, tx repository.TxInterface, members []ImportedTeamMember) error {
	db := self.base.Client(tx)
	builders := make([]*ent.TeamMemberCreate, len(members))
	for i, member := range members {
		builders[i] = db.TeamMember.Create().SetTeamID(member.TeamID).SetUserID(member.UserID).SetOwner(member.Owner)
	}
	for batch := range slices.Chunk(builders, importBatchSize) {
		err := db.TeamMember.CreateBulk(batch...).
			OnConflictColumns(teammember.FieldTeamID, teammember.FieldUserID).
			DoNothing().
			Exec(ctx)
		if err := repository.IgnoreExisting(err); err != nil {
			return err
		}
	}
	return nil
}

func (self *TeamRepository) RaiseIssueCounters(ctx context.Context, tx repository.TxInterface, counters map[uuid.UUID]int) error {
	db := self.base.Client(tx)
	for teamID, counter := range counters {
		err := db.Team.Update().
			Where(team.ID(teamID), team.IssueCounterLT(counter)).
			SetIssueCounter(counter).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
