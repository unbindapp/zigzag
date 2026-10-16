package cycle_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/cycle"
	"github.com/unbindapp/zigzag-api/ent/team"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateCycleInput struct {
	TeamID      uuid.UUID
	Number      int
	Name        string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
}

func (self *CycleRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateCycleInput) (*ent.Cycle, error) {
	return self.base.Client(tx).Cycle.Create().
		SetTeamID(input.TeamID).
		SetNumber(input.Number).
		SetName(input.Name).
		SetDescription(input.Description).
		SetStartsAt(input.StartsAt).
		SetEndsAt(input.EndsAt).
		Save(ctx)
}

func (self *CycleRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Cycle, error) {
	return self.base.DB.Cycle.Get(ctx, id)
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
