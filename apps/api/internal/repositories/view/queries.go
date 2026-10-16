package view_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
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
	return self.base.Client(tx).View.Create().
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
}

func (self *ViewRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.View, error) {
	return self.base.DB.View.Get(ctx, id)
}

func (self *ViewRepository) ResolveWorkspaceOfView(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.View.Query().Where(view.ID(id)).Select(view.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
