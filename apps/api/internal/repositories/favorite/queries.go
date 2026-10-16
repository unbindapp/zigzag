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

func (self *FavoriteRepository) ListByUser(ctx context.Context, userID, workspaceID uuid.UUID) ([]*ent.Favorite, error) {
	return self.base.DB.Favorite.Query().
		Where(favorite.UserID(userID), favorite.WorkspaceID(workspaceID)).
		Order(ent.Asc(favorite.FieldSortOrder), ent.Asc(favorite.FieldCreatedAt)).
		All(ctx)
}

func (self *FavoriteRepository) Find(ctx context.Context, tx repository.TxInterface, userID uuid.UUID, kind schema.FavoriteKind, entityID uuid.UUID) (*ent.Favorite, error) {
	found, err := self.base.Client(tx).Favorite.Query().
		Where(favorite.UserID(userID), favorite.KindEQ(kind), favorite.EntityID(entityID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return found, err
}

func (self *FavoriteRepository) NextSortOrder(ctx context.Context, tx repository.TxInterface, userID, workspaceID uuid.UUID) (float64, error) {
	last, err := self.base.Client(tx).Favorite.Query().
		Where(favorite.UserID(userID), favorite.WorkspaceID(workspaceID)).
		Order(ent.Desc(favorite.FieldSortOrder)).
		First(ctx)
	if ent.IsNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder + 1, nil
}

func (self *FavoriteRepository) Delete(ctx context.Context, userID uuid.UUID, kind schema.FavoriteKind, entityID uuid.UUID) (int, error) {
	return self.base.DB.Favorite.Delete().
		Where(favorite.UserID(userID), favorite.KindEQ(kind), favorite.EntityID(entityID)).
		Exec(ctx)
}

func (self *FavoriteRepository) Reorder(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	return self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		for i, id := range ids {
			_, err := tx.Client().Favorite.Update().
				Where(favorite.ID(id), favorite.UserID(userID)).
				SetSortOrder(float64(i + 1)).
				Save(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
