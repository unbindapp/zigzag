package cycle_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/cycle"
	"github.com/unbindapp/zigzag-api/ent/cyclescopehistory"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

func (self *CycleRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Cycle, error) {
	return self.base.DB.Cycle.Get(ctx, id)
}

func (self *CycleRepository) GetByNumber(ctx context.Context, teamID uuid.UUID, number int) (*ent.Cycle, error) {
	return self.base.DB.Cycle.Query().Where(cycle.TeamID(teamID), cycle.Number(number)).Only(ctx)
}

func (self *CycleRepository) ResolveWorkspaceOfCycle(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Cycle.Query().Where(cycle.ID(id)).QueryTeam().Select(team.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *CycleRepository) ActiveCycleIDs(ctx context.Context, teamIDs []uuid.UUID, now time.Time) (map[uuid.UUID]uuid.UUID, error) {
	active := make(map[uuid.UUID]uuid.UUID, len(teamIDs))
	if len(teamIDs) == 0 {
		return active, nil
	}
	cycles, err := self.base.DB.Cycle.Query().
		Where(cycle.TeamIDIn(teamIDs...), cycle.StartsAtLTE(now), cycle.EndsAtGT(now), cycle.CompletedAtIsNil(), cycle.ArchivedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cycles {
		active[c.TeamID] = c.ID
	}
	return active, nil
}

func (self *CycleRepository) ListByTeam(ctx context.Context, teamID uuid.UUID, includeArchived bool) ([]*ent.Cycle, error) {
	q := self.base.DB.Cycle.Query().Where(cycle.TeamID(teamID))
	if !includeArchived {
		q = q.Where(cycle.ArchivedAtIsNil())
	}
	return q.Order(ent.Desc(cycle.FieldNumber)).All(ctx)
}

func (self *CycleRepository) LockTeam(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID) error {
	_, err := self.base.Client(tx).Team.Query().Where(team.ID(teamID), repository.LockRowsOnPostgres(team.Table)).Unique(false).Only(ctx)
	return err
}

func (self *CycleRepository) NextNumber(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID) (int, error) {
	last, err := self.base.Client(tx).Cycle.Query().Where(cycle.TeamID(teamID)).Order(ent.Desc(cycle.FieldNumber)).First(ctx)
	if ent.IsNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.Number + 1, nil
}

// Ranges are half open, so a cycle may start exactly when another ends.
func (self *CycleRepository) Overlaps(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID, startsAt, endsAt time.Time, except *uuid.UUID) (bool, error) {
	q := self.base.Client(tx).Cycle.Query().
		Where(cycle.TeamID(teamID), cycle.ArchivedAtIsNil(), cycle.StartsAtLT(endsAt), cycle.EndsAtGT(startsAt))
	if except != nil {
		q = q.Where(cycle.IDNEQ(*except))
	}
	return q.Exist(ctx)
}

type StateTally struct {
	Count  int
	Points float64
}

// An unestimated issue counts as 1 point.
func (self *CycleRepository) Tally(ctx context.Context, tx repository.TxInterface, cycleIDs []uuid.UUID) (map[uuid.UUID]map[schema.StateType]StateTally, error) {
	tallies := make(map[uuid.UUID]map[schema.StateType]StateTally, len(cycleIDs))
	for _, id := range cycleIDs {
		tallies[id] = map[schema.StateType]StateTally{}
	}
	if len(cycleIDs) == 0 {
		return tallies, nil
	}
	issues, err := self.base.Client(tx).Issue.Query().
		Where(issue.CycleIDIn(cycleIDs...), issue.TrashedAtIsNil()).
		WithState(func(q *ent.WorkflowStateQuery) { q.Select(workflowstate.FieldType) }).
		Select(issue.FieldCycleID, issue.FieldStateID, issue.FieldEstimate).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, found := range issues {
		if found.CycleID == nil || found.Edges.State == nil {
			continue
		}
		points := 1.0
		if found.Estimate != nil {
			points = *found.Estimate
		}
		byType := tallies[*found.CycleID]
		tally := byType[found.Edges.State.Type]
		tally.Count++
		tally.Points += points
		byType[found.Edges.State.Type] = tally
	}
	return tallies, nil
}

func (self *CycleRepository) ScopeHistory(ctx context.Context, cycleID uuid.UUID) ([]*ent.CycleScopeHistory, error) {
	return self.base.DB.CycleScopeHistory.Query().
		Where(cyclescopehistory.CycleID(cycleID)).
		Order(ent.Asc(cyclescopehistory.FieldDate)).
		All(ctx)
}

func (self *CycleRepository) IssueIDsInCycle(ctx context.Context, tx repository.TxInterface, cycleID uuid.UUID) ([]uuid.UUID, error) {
	return self.base.Client(tx).Issue.Query().Where(issue.CycleID(cycleID)).IDs(ctx)
}

func (self *CycleRepository) UnfinishedIssueIDs(ctx context.Context, tx repository.TxInterface, cycleID uuid.UUID) ([]uuid.UUID, error) {
	return self.base.Client(tx).Issue.Query().
		Where(
			issue.CycleID(cycleID),
			issue.TrashedAtIsNil(),
			issue.HasStateWith(workflowstate.TypeNotIn(schema.StateTypeCompleted, schema.StateTypeCanceled)),
		).
		IDs(ctx)
}

func (self *CycleRepository) TeamsWithCyclesEnabled(ctx context.Context) ([]*ent.Team, error) {
	return self.base.DB.Team.Query().Where(team.CyclesEnabled(true), team.ArchivedAtIsNil()).All(ctx)
}

func (self *CycleRepository) LatestCycle(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID) (*ent.Cycle, error) {
	found, err := self.base.Client(tx).Cycle.Query().Where(cycle.TeamID(teamID)).Order(ent.Desc(cycle.FieldEndsAt)).First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return found, err
}

func (self *CycleRepository) CountEndingAfter(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID, at time.Time) (int, error) {
	return self.base.Client(tx).Cycle.Query().
		Where(cycle.TeamID(teamID), cycle.ArchivedAtIsNil(), cycle.CompletedAtIsNil(), cycle.EndsAtGT(at)).
		Count(ctx)
}

func (self *CycleRepository) EndedOpenCycles(ctx context.Context, now time.Time) ([]*ent.Cycle, error) {
	return self.base.DB.Cycle.Query().
		Where(cycle.EndsAtLTE(now), cycle.CompletedAtIsNil()).
		WithTeam().
		Order(ent.Asc(cycle.FieldEndsAt)).
		All(ctx)
}

func (self *CycleRepository) ActiveCycles(ctx context.Context, now time.Time) ([]*ent.Cycle, error) {
	return self.base.DB.Cycle.Query().
		Where(cycle.StartsAtLTE(now), cycle.EndsAtGT(now), cycle.CompletedAtIsNil(), cycle.ArchivedAtIsNil()).
		WithTeam().
		All(ctx)
}

func (self *CycleRepository) NextOpenCycle(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID, after time.Time) (*ent.Cycle, error) {
	found, err := self.base.Client(tx).Cycle.Query().
		Where(cycle.TeamID(teamID), cycle.StartsAtGTE(after), cycle.CompletedAtIsNil(), cycle.ArchivedAtIsNil()).
		Order(ent.Asc(cycle.FieldStartsAt)).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return found, err
}
