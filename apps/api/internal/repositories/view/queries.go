package view_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/favorite"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/view"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateViewInput struct {
	WorkspaceID uuid.UUID
	TeamID      *uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
	Icon        string
	Color       string
	Kind        schema.ViewKind
	Shared      bool
	Filters     map[string]any
	Display     schema.ViewDisplay
}

func (self *ViewRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateViewInput) (*ent.View, error) {
	created, err := self.base.Client(tx).View.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetNillableTeamID(input.TeamID).
		SetOwnerID(input.OwnerID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetIcon(input.Icon).
		SetColor(input.Color).
		SetKind(input.Kind).
		SetShared(input.Shared).
		SetFilters(input.Filters).
		SetDisplay(input.Display).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.base.Client(tx).View.Query().Where(view.ID(created.ID)).WithOwner().Only(ctx)
}

func (self *ViewRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.View, error) {
	return self.base.DB.View.Query().Where(view.ID(id)).WithOwner().Only(ctx)
}

func (self *ViewRepository) ResolveWorkspaceOfView(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.View.Query().Where(view.ID(id)).Select(view.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

type ListViewsFilter struct {
	WorkspaceID uuid.UUID
	ViewerID    uuid.UUID
	TeamID      *uuid.UUID
	Kind        *schema.ViewKind
}

func (self *ViewRepository) ListVisible(ctx context.Context, filter ListViewsFilter) ([]*ent.View, error) {
	q := self.base.DB.View.Query().Where(
		view.WorkspaceID(filter.WorkspaceID),
		view.Or(view.OwnerID(filter.ViewerID), view.Shared(true)),
	)
	if filter.TeamID != nil {
		q = q.Where(view.TeamID(*filter.TeamID))
	}
	if filter.Kind != nil {
		q = q.Where(view.KindEQ(*filter.Kind))
	}
	return q.WithOwner().Order(ent.Asc(view.FieldName)).All(ctx)
}

type UpdateViewInput struct {
	Name        *string
	Description *string
	Icon        *string
	Color       *string
	Shared      *bool
	Filters     map[string]any
	Display     *schema.ViewDisplay
}

func (self *ViewRepository) Update(ctx context.Context, id uuid.UUID, input *UpdateViewInput) (*ent.View, error) {
	update := self.base.DB.View.UpdateOneID(id).
		SetNillableName(input.Name).
		SetNillableDescription(input.Description).
		SetNillableIcon(input.Icon).
		SetNillableColor(input.Color).
		SetNillableShared(input.Shared)
	if input.Filters != nil {
		update.SetFilters(input.Filters)
	}
	if input.Display != nil {
		update.SetDisplay(*input.Display)
	}
	if err := update.Exec(ctx); err != nil {
		return nil, err
	}
	return self.GetByID(ctx, id)
}

func (self *ViewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		client := tx.Client()
		if _, err := client.Favorite.Delete().Where(favorite.KindEQ(schema.FavoriteKindView), favorite.EntityID(id)).Exec(ctx); err != nil {
			return err
		}
		return client.View.DeleteOneID(id).Exec(ctx)
	})
}
