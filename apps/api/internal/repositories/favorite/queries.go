package favorite_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/favorite"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateFavoriteInput struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Kind        schema.FavoriteKind
	EntityID    uuid.UUID
	SortOrder   float64
}

func (self *FavoriteRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateFavoriteInput) (*ent.Favorite, error) {
	return self.base.Client(tx).Favorite.Create().
		SetUserID(input.UserID).
		SetWorkspaceID(input.WorkspaceID).
		SetKind(input.Kind).
		SetEntityID(input.EntityID).
		SetSortOrder(input.SortOrder).
		Save(ctx)
}

func (self *FavoriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Favorite, error) {
	return self.base.DB.Favorite.Get(ctx, id)
}

func (self *FavoriteRepository) ResolveWorkspaceOfFavorite(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Favorite.Query().Where(favorite.ID(id)).Select(favorite.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
