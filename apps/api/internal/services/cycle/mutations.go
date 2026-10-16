package cycle_service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
)

const maxCycleLength = 26 * 7 * 24 * time.Hour

type CreateCycleInput struct {
	TeamID      uuid.UUID `json:"team_id" required:"true" format:"uuid" doc:"The team; cycles must be turned on for it"`
	Name        string    `json:"name,omitempty" required:"false" maxLength:"64" doc:"Optional name; the cycle shows as Cycle <number> without one"`
	Description string    `json:"description,omitempty" required:"false" maxLength:"2000" doc:"What the cycle is about"`
	StartsAt    time.Time `json:"starts_at" required:"true" doc:"Start, RFC 3339 such as 2026-10-19T00:00:00Z"`
	EndsAt      time.Time `json:"ends_at" required:"true" doc:"End, after starts_at; cycles of a team cannot overlap"`
}

func (self *CycleService) Create(ctx context.Context, userID uuid.UUID, input *CreateCycleInput) (*models.CycleResponse, error) {
	team, err := self.adminTeam(ctx, userID, input.TeamID)
	if err != nil {
		return nil, err
	}
	if !team.CyclesEnabled {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Turn on cycles for this team first")
	}
	if err := validateRange(input.StartsAt, input.EndsAt); err != nil {
		return nil, err
	}

	var created *ent.Cycle
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Cycle().LockTeam(ctx, tx, team.ID); err != nil {
			return err
		}
		if err := self.ensureNoOverlap(ctx, tx, team.ID, input.StartsAt, input.EndsAt, nil); err != nil {
			return err
		}
		number, err := self.repo.Cycle().NextNumber(ctx, tx, team.ID)
		if err != nil {
			return err
		}
		created, err = self.repo.Cycle().Create(ctx, tx, &cycle_repo.CreateCycleInput{
			TeamID:      team.ID,
			Number:      number,
			Name:        strings.TrimSpace(input.Name),
			Description: strings.TrimSpace(input.Description),
			StartsAt:    input.StartsAt,
			EndsAt:      input.EndsAt,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return self.respondAndEmit(ctx, team, created, "created")
}

type UpdateCycleInput struct {
	CycleID     uuid.UUID  `json:"cycle_id" required:"true" format:"uuid" doc:"The cycle to change"`
	Name        *string    `json:"name,omitempty" required:"false" maxLength:"64" doc:"Name, empty to clear"`
	Description *string    `json:"description,omitempty" required:"false" maxLength:"2000" doc:"Description, empty to clear"`
	StartsAt    *time.Time `json:"starts_at,omitempty" required:"false" doc:"New start, RFC 3339"`
	EndsAt      *time.Time `json:"ends_at,omitempty" required:"false" doc:"New end, RFC 3339, after the start"`
}

func (self *CycleService) Update(ctx context.Context, userID uuid.UUID, input *UpdateCycleInput) (*models.CycleResponse, error) {
	current, team, err := self.adminCycle(ctx, userID, input.CycleID)
	if err != nil {
		return nil, err
	}
	update := &cycle_repo.UpdateCycleInput{StartsAt: input.StartsAt, EndsAt: input.EndsAt}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		update.Name = &name
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		update.Description = &description
	}
	startsAt, endsAt := current.StartsAt, current.EndsAt
	if input.StartsAt != nil {
		startsAt = *input.StartsAt
	}
	if input.EndsAt != nil {
		endsAt = *input.EndsAt
	}
	if err := validateRange(startsAt, endsAt); err != nil {
		return nil, err
	}

	var updated *ent.Cycle
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Cycle().LockTeam(ctx, tx, team.ID); err != nil {
			return err
		}
		if err := self.ensureNoOverlap(ctx, tx, team.ID, startsAt, endsAt, &current.ID); err != nil {
			return err
		}
		updated, err = self.repo.Cycle().Update(ctx, tx, current.ID, update)
		return err
	})
	if err != nil {
		return nil, err
	}
	return self.respondAndEmit(ctx, team, updated, "updated")
}

func (self *CycleService) respondAndEmit(ctx context.Context, team *ent.Team, found *ent.Cycle, action string) (*models.CycleResponse, error) {
	response, err := self.respondOne(ctx, team, found)
	if err != nil {
		return nil, err
	}
	self.emitResponse(ctx, team, response, action)
	return response, nil
}

type CycleRefInput struct {
	CycleID uuid.UUID `json:"cycle_id" required:"true" format:"uuid" doc:"The cycle"`
}

func (self *CycleService) Archive(ctx context.Context, userID uuid.UUID, input *CycleRefInput) (*models.CycleResponse, error) {
	current, team, err := self.adminCycle(ctx, userID, input.CycleID)
	if err != nil {
		return nil, err
	}
	archived, err := self.repo.Cycle().Archive(ctx, current.ID, self.now())
	if err != nil {
		return nil, err
	}
	return self.respondAndEmit(ctx, team, archived, "archived")
}

func (self *CycleService) Delete(ctx context.Context, userID uuid.UUID, input *CycleRefInput) error {
	current, team, err := self.adminCycle(ctx, userID, input.CycleID)
	if err != nil {
		return err
	}
	response, err := self.respondOne(ctx, team, current)
	if err != nil {
		return err
	}
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		issueIDs, err := self.repo.Cycle().IssueIDsInCycle(ctx, tx, current.ID)
		if err != nil {
			return err
		}
		change := schema.FieldChange{Field: "cycle", From: current.ID.String(), FromLabel: cycleLabel(current)}
		if err := self.repo.Cycle().RecordHistory(ctx, tx, cycleHistory(issueIDs, &userID, apictx.HistorySource(ctx), change)); err != nil {
			return err
		}
		return self.repo.Cycle().Delete(ctx, tx, current.ID)
	})
	if err != nil {
		return err
	}
	self.emitResponse(ctx, team, response, "deleted")
	return nil
}

func (self *CycleService) ensureNoOverlap(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID, startsAt, endsAt time.Time, except *uuid.UUID) error {
	overlaps, err := self.repo.Cycle().Overlaps(ctx, tx, teamID, startsAt, endsAt, except)
	if err != nil {
		return err
	}
	if overlaps {
		return errdefs.NewCustomError(errdefs.ErrTypeConflict, "The dates overlap another cycle of this team")
	}
	return nil
}

func validateRange(startsAt, endsAt time.Time) error {
	if !endsAt.After(startsAt) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "ends_at must be after starts_at")
	}
	if endsAt.Sub(startsAt) > maxCycleLength {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A cycle can last at most 26 weeks")
	}
	return nil
}

func cycleHistory(issueIDs []uuid.UUID, actorID *uuid.UUID, source schema.HistorySource, change schema.FieldChange) []cycle_repo.HistoryEntry {
	entries := make([]cycle_repo.HistoryEntry, len(issueIDs))
	for i, issueID := range issueIDs {
		entries[i] = cycle_repo.HistoryEntry{IssueID: issueID, ActorID: actorID, Source: source, Changes: []schema.FieldChange{change}}
	}
	return entries
}
