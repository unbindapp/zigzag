package initiative_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/initiative"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateInitiativeInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Slug        string
	Description string
	Icon        string
	Color       string
	OwnerID     *uuid.UUID
	Status      schema.InitiativeStatus
}

func (self *InitiativeRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateInitiativeInput) (*ent.Initiative, error) {
	return self.base.Client(tx).Initiative.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetName(input.Name).
		SetSlug(input.Slug).
		SetDescription(input.Description).
		SetIcon(input.Icon).
		SetColor(input.Color).
		SetNillableOwnerID(input.OwnerID).
		SetStatus(input.Status).
		Save(ctx)
}

func (self *InitiativeRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Initiative, error) {
	return self.base.DB.Initiative.Get(ctx, id)
}

func (self *InitiativeRepository) ResolveWorkspaceOfInitiative(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Initiative.Query().Where(initiative.ID(id)).Select(initiative.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
