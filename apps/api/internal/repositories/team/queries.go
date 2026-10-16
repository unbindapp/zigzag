package team_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/team"
	"github.com/zigzagapp-org/zigzag-api/ent/teammember"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

var stateTypeOrder = []string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}

func orderStates(s *sql.Selector) {
	expr := "CASE " + s.C(workflowstate.FieldType)
	for i, t := range stateTypeOrder {
		expr += " WHEN '" + t + "' THEN " + string(rune('0'+i))
	}
	expr += " END"
	s.OrderExpr(sql.Expr(expr))
	s.OrderBy(sql.Asc(s.C(workflowstate.FieldPosition)), sql.Asc(s.C(workflowstate.FieldName)))
}

func withActiveStates(q *ent.WorkflowStateQuery) {
	q.Where(workflowstate.ArchivedAtIsNil()).Order(orderStates)
}

func (self *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Team, error) {
	return self.base.DB.Team.Query().Where(team.ID(id)).WithStates(withActiveStates).Only(ctx)
}

func (self *TeamRepository) GetByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*ent.Team, error) {
	return self.base.DB.Team.Query().
		Where(team.WorkspaceID(workspaceID), team.KeyEqualFold(key)).
		WithStates(withActiveStates).
		Only(ctx)
}

func (self *TeamRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, includeArchived bool) ([]*ent.Team, error) {
	q := self.base.DB.Team.Query().Where(team.WorkspaceID(workspaceID))
	if !includeArchived {
		q = q.Where(team.ArchivedAtIsNil())
	}
	return q.WithStates(withActiveStates).Order(ent.Asc(team.FieldName)).All(ctx)
}

func (self *TeamRepository) ResolveWorkspaceOfTeam(ctx context.Context, teamID uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Team.Query().Where(team.ID(teamID)).Select(team.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *TeamRepository) KeyOrNameTaken(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, key, name string, except *uuid.UUID) (keyTaken, nameTaken bool, err error) {
	q := self.base.Client(tx).Team.Query().Where(team.WorkspaceID(workspaceID))
	if except != nil {
		q = q.Where(team.IDNEQ(*except))
	}
	teams, err := q.Select(team.FieldKey, team.FieldName).All(ctx)
	if err != nil {
		return false, false, err
	}
	for _, t := range teams {
		keyTaken = keyTaken || (key != "" && t.Key == key)
		nameTaken = nameTaken || (name != "" && t.Name == name)
	}
	return keyTaken, nameTaken, nil
}

func (self *TeamRepository) ListMembers(ctx context.Context, teamID uuid.UUID) ([]*ent.TeamMember, error) {
	return self.base.DB.TeamMember.Query().
		Where(teammember.TeamID(teamID)).
		WithUser().
		Order(ent.Asc(teammember.FieldCreatedAt)).
		All(ctx)
}

func (self *TeamRepository) GetMember(ctx context.Context, teamID, userID uuid.UUID) (*ent.TeamMember, error) {
	return self.base.DB.TeamMember.Query().
		Where(teammember.TeamID(teamID), teammember.UserID(userID)).
		WithUser().
		Only(ctx)
}

func (self *TeamRepository) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	return self.base.DB.TeamMember.Query().Where(teammember.TeamID(teamID), teammember.UserID(userID)).Exist(ctx)
}

func (self *TeamRepository) TeamIDsOfUser(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	memberships, err := self.base.DB.TeamMember.Query().
		Where(teammember.UserID(userID), teammember.HasTeamWith(team.WorkspaceID(workspaceID))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(memberships))
	for i, membership := range memberships {
		ids[i] = membership.TeamID
	}
	return ids, nil
}

func (self *TeamRepository) MemberCounts(ctx context.Context, teamIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	counts := make(map[uuid.UUID]int, len(teamIDs))
	if len(teamIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		TeamID uuid.UUID `json:"team_id"`
		Count  int       `json:"count"`
	}
	if err := self.base.DB.TeamMember.Query().
		Where(teammember.TeamIDIn(teamIDs...)).
		GroupBy(teammember.FieldTeamID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.TeamID] = row.Count
	}
	return counts, nil
}

func (self *TeamRepository) TeamsInWorkspace(ctx context.Context, workspaceID uuid.UUID, teamIDs []uuid.UUID) ([]*ent.Team, error) {
	return self.base.DB.Team.Query().Where(team.WorkspaceID(workspaceID), team.IDIn(teamIDs...)).All(ctx)
}
