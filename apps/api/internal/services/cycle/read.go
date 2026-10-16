package cycle_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
)

type ListCyclesInput struct {
	TeamID          uuid.UUID `query:"team_id" required:"true" format:"uuid" doc:"The team"`
	IncludeArchived bool      `query:"include_archived" required:"false" doc:"Include archived cycles"`
}

func (self *CycleService) List(ctx context.Context, userID uuid.UUID, input *ListCyclesInput) ([]*models.CycleResponse, error) {
	team, err := self.visibleTeam(ctx, userID, input.TeamID)
	if err != nil {
		return nil, err
	}
	cycles, err := self.repo.Cycle().ListByTeam(ctx, team.ID, input.IncludeArchived)
	if err != nil {
		return nil, err
	}
	return self.respond(ctx, team, cycles)
}

type GetCycleInput struct {
	CycleID uuid.UUID `query:"cycle_id" required:"false" format:"uuid" doc:"Cycle id; give this, or team_id and number"`
	TeamID  uuid.UUID `query:"team_id" required:"false" format:"uuid" doc:"Team of the cycle when looking up by number"`
	Number  int       `query:"number" required:"false" minimum:"0" doc:"Cycle number within the team, with team_id"`
}

func (self *CycleService) Get(ctx context.Context, userID uuid.UUID, input *GetCycleInput) (*models.CycleResponse, error) {
	found, team, err := self.lookup(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	response, err := self.respondOne(ctx, team, found)
	if err != nil {
		return nil, err
	}
	history, err := self.repo.Cycle().ScopeHistory(ctx, found.ID)
	if err != nil {
		return nil, err
	}
	response.ScopeHistory = models.TransformCycleScopeHistory(history)
	if response.IsActive {
		response.ScopeHistory = withLivePoint(response.ScopeHistory, response.Progress, localDate(self.now(), teamLocation(team)))
	}
	return response, nil
}

func (self *CycleService) lookup(ctx context.Context, userID uuid.UUID, input *GetCycleInput) (*ent.Cycle, *ent.Team, error) {
	if input.CycleID != uuid.Nil {
		return self.visibleCycle(ctx, userID, input.CycleID)
	}
	if input.TeamID == uuid.Nil || input.Number == 0 {
		return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Give cycle_id, or team_id and number")
	}
	team, err := self.visibleTeam(ctx, userID, input.TeamID)
	if err != nil {
		return nil, nil, err
	}
	found, err := self.repo.Cycle().GetByNumber(ctx, team.ID, input.Number)
	if ent.IsNotFound(err) {
		return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Cycle not found")
	}
	if err != nil {
		return nil, nil, err
	}
	return found, team, nil
}

func withLivePoint(points []models.CycleScopePoint, progress models.CycleProgress, today time.Time) []models.CycleScopePoint {
	live := models.CycleScopePoint{Date: today.Format(time.DateOnly), Scope: progress.Scope, Completed: progress.Completed, Started: progress.Started}
	if len(points) > 0 && points[len(points)-1].Date == live.Date {
		points[len(points)-1] = live
		return points
	}
	return append(points, live)
}
