package cycle_repo

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/cycle"
	"github.com/unbindapp/zigzag-api/ent/team"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type ImportedCycle struct {
	ID          uuid.UUID
	TeamID      uuid.UUID
	Number      int
	Name        string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
	CompletedAt *time.Time
}

func (self *CycleRepository) CyclesOfWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*ent.Cycle, error) {
	return self.base.DB.Cycle.Query().Where(cycle.HasTeamWith(team.WorkspaceID(workspaceID))).All(ctx)
}

func (self *CycleRepository) ImportCycles(ctx context.Context, tx repository.TxInterface, cycles []ImportedCycle) error {
	db := self.base.Client(tx)
	builders := make([]*ent.CycleCreate, len(cycles))
	for i, c := range cycles {
		builders[i] = db.Cycle.Create().
			SetID(c.ID).
			SetTeamID(c.TeamID).
			SetNumber(c.Number).
			SetName(c.Name).
			SetDescription(c.Description).
			SetStartsAt(c.StartsAt).
			SetEndsAt(c.EndsAt).
			SetNillableCompletedAt(c.CompletedAt)
	}
	for batch := range slices.Chunk(builders, 500) {
		if err := db.Cycle.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
