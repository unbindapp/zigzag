package label_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/favorite"
	"github.com/unbindapp/zigzag-api/ent/label"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateLabelInput struct {
	WorkspaceID       uuid.UUID
	TeamID            *uuid.UUID
	Name              string
	Color             string
	Description       string
	ParentID          *uuid.UUID
	IsGroup           bool
	GroupSingleSelect bool
	SortOrder         float64
}

func (self *LabelRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateLabelInput) (*ent.Label, error) {
	return self.base.Client(tx).Label.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetNillableTeamID(input.TeamID).
		SetName(input.Name).
		SetColor(input.Color).
		SetDescription(input.Description).
		SetNillableParentID(input.ParentID).
		SetIsGroup(input.IsGroup).
		SetGroupSingleSelect(input.GroupSingleSelect).
		SetSortOrder(input.SortOrder).
		Save(ctx)
}

type UpdateLabelInput struct {
	Name              *string
	Color             *string
	Description       *string
	ParentID          *uuid.UUID
	ClearParent       bool
	GroupSingleSelect *bool
	SortOrder         *float64
}

func (self *LabelRepository) Update(ctx context.Context, tx repository.TxInterface, id uuid.UUID, input *UpdateLabelInput) (*ent.Label, error) {
	update := self.base.Client(tx).Label.UpdateOneID(id).
		SetNillableName(input.Name).
		SetNillableColor(input.Color).
		SetNillableDescription(input.Description).
		SetNillableGroupSingleSelect(input.GroupSingleSelect).
		SetNillableSortOrder(input.SortOrder)
	if input.ClearParent {
		update.ClearParentID()
	}
	if input.ParentID != nil {
		update.SetParentID(*input.ParentID)
	}
	updated, err := update.Save(ctx)
	if err != nil || input.SortOrder == nil {
		return updated, err
	}
	respaced, err := self.base.RespaceIfCrowded(ctx, tx, repository.LabelSortOrder, updated.WorkspaceID, updated.ID, updated.SortOrder)
	if err != nil || !respaced {
		return updated, err
	}
	return self.base.Client(tx).Label.Get(ctx, id)
}

func (self *LabelRepository) SetArchived(ctx context.Context, id uuid.UUID, archivedAt *time.Time) (*ent.Label, error) {
	var updated *ent.Label
	err := self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		client := tx.Client()
		children := client.Label.Update().Where(label.ParentID(id))
		one := client.Label.UpdateOneID(id)
		if archivedAt == nil {
			children.ClearArchivedAt()
			one.ClearArchivedAt()
		} else {
			children.SetArchivedAt(*archivedAt)
			one.SetArchivedAt(*archivedAt)
		}
		if _, err := children.Save(ctx); err != nil {
			return err
		}
		var err error
		updated, err = one.Save(ctx)
		return err
	})
	return updated, err
}

func (self *LabelRepository) Delete(ctx context.Context, tx repository.TxInterface, ids []uuid.UUID) error {
	client := self.base.Client(tx)
	if _, err := client.Favorite.Delete().Where(favorite.KindEQ(schema.FavoriteKindLabel), favorite.EntityIDIn(ids...)).Exec(ctx); err != nil {
		return err
	}
	if _, err := client.Label.Delete().Where(label.ParentIDIn(ids...)).Exec(ctx); err != nil {
		return err
	}
	_, err := client.Label.Delete().Where(label.IDIn(ids...)).Exec(ctx)
	return err
}

type HistoryEntry struct {
	IssueID uuid.UUID
	ActorID *uuid.UUID
	Source  schema.HistorySource
	Changes []schema.FieldChange
}

func (self *LabelRepository) RecordHistory(ctx context.Context, tx repository.TxInterface, entries []HistoryEntry) error {
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
