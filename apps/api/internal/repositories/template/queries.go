package template_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/predicate"
	"github.com/zigzagapp-org/zigzag-api/ent/project"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/template"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
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
	created, err := self.base.Client(tx).Template.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetNillableTeamID(input.TeamID).
		SetKind(input.Kind).
		SetName(input.Name).
		SetDescription(input.Description).
		SetData(input.Data).
		SetSortOrder(input.SortOrder).
		SetNillableCreatorID(input.CreatorID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.base.Client(tx).Template.Query().Where(template.ID(created.ID)).WithCreator().Only(ctx)
}

func (self *TemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Template, error) {
	return self.base.DB.Template.Query().Where(template.ID(id)).WithCreator().Only(ctx)
}

func (self *TemplateRepository) ResolveWorkspaceOfTemplate(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Template.Query().Where(template.ID(id)).Select(template.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *TemplateRepository) List(ctx context.Context, workspaceID uuid.UUID, teamIDs []uuid.UUID, kind *schema.TemplateKind) ([]*ent.Template, error) {
	scope := []predicate.Template{template.TeamIDIsNil()}
	if len(teamIDs) > 0 {
		scope = append(scope, template.TeamIDIn(teamIDs...))
	}
	q := self.base.DB.Template.Query().Where(template.WorkspaceID(workspaceID), template.Or(scope...))
	if kind != nil {
		q = q.Where(template.KindEQ(*kind))
	}
	return q.WithCreator().Order(ent.Asc(template.FieldSortOrder), ent.Asc(template.FieldName)).All(ctx)
}

func (self *TemplateRepository) NextSortOrder(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID) (float64, error) {
	last, err := self.base.Client(tx).Template.Query().Where(template.WorkspaceID(workspaceID)).Order(ent.Desc(template.FieldSortOrder)).First(ctx)
	if ent.IsNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder + 1, nil
}

func (self *TemplateRepository) ProjectInWorkspace(ctx context.Context, workspaceID, projectID uuid.UUID) (bool, error) {
	return self.base.DB.Project.Query().Where(project.ID(projectID), project.WorkspaceID(workspaceID)).Exist(ctx)
}

type UpdateTemplateInput struct {
	Name        *string
	Description *string
	Data        *schema.TemplateData
	SortOrder   *float64
}

func (self *TemplateRepository) Update(ctx context.Context, id uuid.UUID, input *UpdateTemplateInput) (*ent.Template, error) {
	update := self.base.DB.Template.UpdateOneID(id).
		SetNillableName(input.Name).
		SetNillableDescription(input.Description).
		SetNillableSortOrder(input.SortOrder)
	if input.Data != nil {
		update.SetData(*input.Data)
	}
	if err := update.Exec(ctx); err != nil {
		return nil, err
	}
	return self.GetByID(ctx, id)
}

func (self *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Template.DeleteOneID(id).Exec(ctx)
}
