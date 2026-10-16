package cycle_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/cyclescopehistory"
	"github.com/unbindapp/zigzag-api/ent/favorite"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/schema"
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

type UpdateCycleInput struct {
	Name        *string
	Description *string
	StartsAt    *time.Time
	EndsAt      *time.Time
}

func (self *CycleRepository) Update(ctx context.Context, tx repository.TxInterface, id uuid.UUID, input *UpdateCycleInput) (*ent.Cycle, error) {
	return self.base.Client(tx).Cycle.UpdateOneID(id).
		SetNillableName(input.Name).
		SetNillableDescription(input.Description).
		SetNillableStartsAt(input.StartsAt).
		SetNillableEndsAt(input.EndsAt).
		Save(ctx)
}

func (self *CycleRepository) Archive(ctx context.Context, id uuid.UUID, at time.Time) (*ent.Cycle, error) {
	return self.base.DB.Cycle.UpdateOneID(id).SetArchivedAt(at).Save(ctx)
}

func (self *CycleRepository) Complete(ctx context.Context, tx repository.TxInterface, id uuid.UUID, at time.Time) error {
	return self.base.Client(tx).Cycle.UpdateOneID(id).SetCompletedAt(at).Exec(ctx)
}

func (self *CycleRepository) Delete(ctx context.Context, tx repository.TxInterface, id uuid.UUID) error {
	client := self.base.Client(tx)
	if _, err := client.Favorite.Delete().Where(favorite.KindEQ(schema.FavoriteKindCycle), favorite.EntityID(id)).Exec(ctx); err != nil {
		return err
	}
	if _, err := client.Issue.Update().Where(issue.CycleID(id)).ClearCycleID().Save(ctx); err != nil {
		return err
	}
	return client.Cycle.DeleteOneID(id).Exec(ctx)
}

func (self *CycleRepository) MoveIssues(ctx context.Context, tx repository.TxInterface, issueIDs []uuid.UUID, cycleID uuid.UUID) error {
	if len(issueIDs) == 0 {
		return nil
	}
	_, err := self.base.Client(tx).Issue.Update().Where(issue.IDIn(issueIDs...)).SetCycleID(cycleID).Save(ctx)
	return err
}

type ScopeSnapshot struct {
	CycleID   uuid.UUID
	Date      time.Time
	Scope     float64
	Completed float64
	Started   float64
}

func (self *CycleRepository) UpsertScopeSnapshot(ctx context.Context, tx repository.TxInterface, snapshot ScopeSnapshot) error {
	return self.base.Client(tx).CycleScopeHistory.Create().
		SetCycleID(snapshot.CycleID).
		SetDate(snapshot.Date).
		SetScope(snapshot.Scope).
		SetCompleted(snapshot.Completed).
		SetStarted(snapshot.Started).
		OnConflictColumns(cyclescopehistory.FieldCycleID, cyclescopehistory.FieldDate).
		UpdateScope().
		UpdateCompleted().
		UpdateStarted().
		UpdateUpdatedAt().
		Exec(ctx)
}

type HistoryEntry struct {
	IssueID uuid.UUID
	ActorID *uuid.UUID
	Source  schema.HistorySource
	Changes []schema.FieldChange
}

func (self *CycleRepository) RecordHistory(ctx context.Context, tx repository.TxInterface, entries []HistoryEntry) error {
	if len(entries) == 0 {
		return nil
	}
	client := self.base.Client(tx)
	builders := make([]*ent.IssueHistoryCreate, len(entries))
	for i, entry := range entries {
		builders[i] = client.IssueHistory.Create().
			SetIssueID(entry.IssueID).
			SetNillableActorID(entry.ActorID).
			SetSource(entry.Source).
			SetChanges(entry.Changes)
	}
	return client.IssueHistory.CreateBulk(builders...).Exec(ctx)
}
