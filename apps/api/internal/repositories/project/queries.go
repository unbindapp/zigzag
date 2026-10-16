package project_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/project"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateProjectInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Slug        string
	StatusID    uuid.UUID
	Summary     string
	Description string
	Icon        string
	Color       string
	LeadID      *uuid.UUID
	Priority    int
	CreatorID   *uuid.UUID
}

func (self *ProjectRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateProjectInput) (*ent.Project, error) {
	return self.base.Client(tx).Project.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetName(input.Name).
		SetSlug(input.Slug).
		SetStatusID(input.StatusID).
		SetSummary(input.Summary).
		SetDescription(input.Description).
		SetIcon(input.Icon).
		SetColor(input.Color).
		SetNillableLeadID(input.LeadID).
		SetPriority(input.Priority).
		SetNillableCreatorID(input.CreatorID).
		Save(ctx)
}

func (self *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Project, error) {
	return self.base.DB.Project.Get(ctx, id)
}

func (self *ProjectRepository) ResolveWorkspaceOfProject(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Project.Query().Where(project.ID(id)).Select(project.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
