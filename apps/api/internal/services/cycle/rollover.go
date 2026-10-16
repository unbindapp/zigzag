package cycle_service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
)

const plannedOpenCycles = 2

func (self *CycleService) ProcessCycles(ctx context.Context) error {
	now := self.now()
	var failures []error

	teams, err := self.repo.Cycle().TeamsWithCyclesEnabled(ctx)
	if err != nil {
		return err
	}
	for _, team := range teams {
		if err := self.planCycles(ctx, team, now); err != nil {
			log.Error("Planning cycles failed", "team", team.ID, "err", err)
			failures = append(failures, err)
		}
	}

	ended, err := self.repo.Cycle().EndedOpenCycles(ctx, now)
	if err != nil {
		return errors.Join(append(failures, err)...)
	}
	for _, found := range ended {
		if err := self.closeCycle(ctx, found, now); err != nil {
			log.Error("Closing a cycle failed", "cycle", found.ID, "err", err)
			failures = append(failures, err)
		}
	}

	active, err := self.repo.Cycle().ActiveCycles(ctx, now)
	if err != nil {
		return errors.Join(append(failures, err)...)
	}
	for _, found := range active {
		if err := self.snapshot(ctx, nil, found, found.Edges.Team, now); err != nil {
			log.Error("Recording cycle scope failed", "cycle", found.ID, "err", err)
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (self *CycleService) planCycles(ctx context.Context, team *ent.Team, now time.Time) error {
	var created []*ent.Cycle
	err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Cycle().LockTeam(ctx, tx, team.ID); err != nil {
			return err
		}
		open, err := self.repo.Cycle().CountEndingAfter(ctx, tx, team.ID, now)
		if err != nil {
			return err
		}
		latest, err := self.repo.Cycle().LatestCycle(ctx, tx, team.ID)
		if err != nil {
			return err
		}
		loc := teamLocation(team)
		for ; open < plannedOpenCycles; open++ {
			startsAt := nextStart(team, latest, now, loc)
			endsAt := startsAt.In(loc).AddDate(0, 0, 7*max(team.CycleDurationWeeks, 1))
			number, err := self.repo.Cycle().NextNumber(ctx, tx, team.ID)
			if err != nil {
				return err
			}
			latest, err = self.repo.Cycle().Create(ctx, tx, &cycle_repo.CreateCycleInput{TeamID: team.ID, Number: number, StartsAt: startsAt, EndsAt: endsAt})
			if err != nil {
				return err
			}
			created = append(created, latest)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, found := range created {
		self.emitUpdated(ctx, team, found, "created")
	}
	return nil
}

func (self *CycleService) PlanTeamCycles(ctx context.Context, teamID uuid.UUID) error {
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if err != nil {
		return err
	}
	if !team.CyclesEnabled {
		return nil
	}
	return self.planCycles(ctx, team, self.now())
}

func nextStart(team *ent.Team, latest *ent.Cycle, now time.Time, loc *time.Location) time.Time {
	aligned := weekStart(now, loc, team.CycleStartDay)
	if latest == nil {
		return aligned
	}
	chained := latest.EndsAt.In(loc).AddDate(0, 0, 7*team.CycleCooldownWeeks)
	if chained.In(loc).AddDate(0, 0, 7*max(team.CycleDurationWeeks, 1)).After(now) {
		return chained
	}
	if aligned.Before(latest.EndsAt) {
		return latest.EndsAt
	}
	return aligned
}

// startDay is 1 Monday through 7 Sunday, unlike time.Weekday.
func weekStart(now time.Time, loc *time.Location, startDay int) time.Time {
	local := now.In(loc)
	weekday := int(local.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	back := (weekday - startDay + 7) % 7
	return time.Date(local.Year(), local.Month(), local.Day()-back, 0, 0, 0, 0, loc)
}

func (self *CycleService) closeCycle(ctx context.Context, found *ent.Cycle, now time.Time) error {
	team := found.Edges.Team
	err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.snapshot(ctx, tx, found, team, now); err != nil {
			return err
		}
		if err := self.repo.Cycle().Complete(ctx, tx, found.ID, now); err != nil {
			return err
		}
		if !team.CyclesEnabled || !team.CycleAutoMove {
			return nil
		}
		next, err := self.repo.Cycle().NextOpenCycle(ctx, tx, team.ID, found.EndsAt)
		if err != nil || next == nil {
			return err
		}
		unfinished, err := self.repo.Cycle().UnfinishedIssueIDs(ctx, tx, found.ID)
		if err != nil {
			return err
		}
		if err := self.repo.Cycle().MoveIssues(ctx, tx, unfinished, next.ID); err != nil {
			return err
		}
		change := schema.FieldChange{Field: "cycle", From: found.ID.String(), To: next.ID.String(), FromLabel: cycleLabel(found), ToLabel: cycleLabel(next)}
		return self.repo.Cycle().RecordHistory(ctx, tx, cycleHistory(unfinished, nil, schema.HistorySourceSystem, change))
	})
	if err != nil {
		return err
	}
	completed := *found
	completed.CompletedAt = &now
	self.emitUpdated(ctx, team, &completed, "completed")
	return nil
}

func (self *CycleService) snapshot(ctx context.Context, tx repository.TxInterface, found *ent.Cycle, team *ent.Team, now time.Time) error {
	tallies, err := self.repo.Cycle().Tally(ctx, tx, []uuid.UUID{found.ID})
	if err != nil {
		return err
	}
	progress := progressOf(team, tallies[found.ID])
	at := now
	if !at.Before(found.EndsAt) {
		at = found.EndsAt.Add(-time.Second)
	}
	return self.repo.Cycle().UpsertScopeSnapshot(ctx, tx, cycle_repo.ScopeSnapshot{
		CycleID:   found.ID,
		Date:      localDate(at, teamLocation(team)),
		Scope:     progress.Scope,
		Completed: progress.Completed,
		Started:   progress.Started,
	})
}
