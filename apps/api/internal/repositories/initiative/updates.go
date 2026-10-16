package initiative_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/initiativepost"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

func (self *InitiativeRepository) ListUpdates(ctx context.Context, initiativeID uuid.UUID) ([]*ent.InitiativePost, error) {
	return self.base.DB.InitiativePost.Query().
		Where(initiativepost.InitiativeID(initiativeID)).
		WithAuthor().
		Order(initiativepost.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
}

func (self *InitiativeRepository) GetUpdate(ctx context.Context, updateID uuid.UUID) (*ent.InitiativePost, error) {
	return self.base.DB.InitiativePost.Query().Where(initiativepost.ID(updateID)).WithAuthor().Only(ctx)
}

type CreateUpdateInput struct {
	InitiativeID uuid.UUID
	AuthorID     uuid.UUID
	Body         string
	Health       schema.Health
}

func (self *InitiativeRepository) CreateUpdate(ctx context.Context, input *CreateUpdateInput) (*ent.InitiativePost, error) {
	return self.base.DB.InitiativePost.Create().
		SetInitiativeID(input.InitiativeID).
		SetAuthorID(input.AuthorID).
		SetBody(input.Body).
		SetHealth(input.Health).
		Save(ctx)
}

func (self *InitiativeRepository) DeleteUpdate(ctx context.Context, updateID uuid.UUID) error {
	return self.base.DB.InitiativePost.DeleteOneID(updateID).Exec(ctx)
}
