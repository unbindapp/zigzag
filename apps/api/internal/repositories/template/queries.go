package template_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/template"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateTemplateInput struct {
	WorkspaceID uuid.UUID
	TeamID      *uuid.UUID
	Kind        schema.TemplateKind
	Name        string
	Description string
	Data        schema.TemplateData
	SortOrder   float64
	CreatorID   *uuid.UUID
}

func (self *TemplateRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateTemplateInput) (*ent.Template, error) {
	return self.base.Client(tx).Template.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetNillableTeamID(input.TeamID).
		SetKind(input.Kind).
		SetName(input.Name).
		SetDescription(input.Description).
		SetData(input.Data).
		SetSortOrder(input.SortOrder).
		SetNillableCreatorID(input.CreatorID).
		Save(ctx)
}

func (self *TemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Template, error) {
	return self.base.DB.Template.Get(ctx, id)
}

func (self *TemplateRepository) ResolveWorkspaceOfTemplate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Template.Query().Where(template.ID(id)).Select(template.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
