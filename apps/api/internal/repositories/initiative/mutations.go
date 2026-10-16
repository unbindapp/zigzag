package initiative_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
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
	TargetDate  *time.Time
	SortOrder   float64
	ProjectIDs  []uuid.UUID
}

func (self *InitiativeRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateInitiativeInput) (*ent.Initiative, error) {
	m := self.base.Client(tx).Initiative.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetName(input.Name).
		SetSlug(input.Slug).
		SetDescription(input.Description).
		SetIcon(input.Icon).
		SetColor(input.Color).
		SetNillableOwnerID(input.OwnerID).
		SetNillableTargetDate(input.TargetDate).
		SetSortOrder(input.SortOrder).
		AddProjectIDs(input.ProjectIDs...)
	if input.Status != "" {
		m.SetStatus(input.Status)
	}
	return m.Save(ctx)
}

type UpdateInitiativeInput struct {
	Name            *string
	Description     *string
	Icon            *string
	Color           *string
	OwnerID         *uuid.UUID
	ClearOwner      bool
	Status          *schema.InitiativeStatus
	TargetDate      *time.Time
	ClearTargetDate bool
	SortOrder       *float64
	ProjectIDs      []uuid.UUID
	ReplaceProjects bool
}

func (self *InitiativeRepository) Update(ctx context.Context, tx repository.TxInterface, id uuid.UUID, input *UpdateInitiativeInput) error {
	m := self.base.Client(tx).Initiative.UpdateOneID(id)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.Description != nil {
		m.SetDescription(*input.Description)
	}
	if input.Icon != nil {
		m.SetIcon(*input.Icon)
	}
	if input.Color != nil {
		m.SetColor(*input.Color)
	}
	if input.OwnerID != nil {
		m.SetOwnerID(*input.OwnerID)
	}
	if input.ClearOwner {
		m.ClearOwnerID()
	}
	if input.Status != nil {
		m.SetStatus(*input.Status)
	}
	if input.TargetDate != nil {
		m.SetTargetDate(*input.TargetDate)
	}
	if input.ClearTargetDate {
		m.ClearTargetDate()
	}
	if input.SortOrder != nil {
		m.SetSortOrder(*input.SortOrder)
	}
	if input.ReplaceProjects {
		m.ClearProjects().AddProjectIDs(input.ProjectIDs...)
	}
	return m.Exec(ctx)
}

func (self *InitiativeRepository) SetArchivedAt(ctx context.Context, id uuid.UUID, at *time.Time) error {
	m := self.base.DB.Initiative.UpdateOneID(id)
	if at == nil {
		return m.ClearArchivedAt().Exec(ctx)
	}
	return m.SetArchivedAt(*at).Exec(ctx)
}

func (self *InitiativeRepository) AddProject(ctx context.Context, id, projectID uuid.UUID) error {
	return self.base.DB.Initiative.UpdateOneID(id).AddProjectIDs(projectID).Exec(ctx)
}

func (self *InitiativeRepository) RemoveProject(ctx context.Context, id, projectID uuid.UUID) error {
	return self.base.DB.Initiative.UpdateOneID(id).RemoveProjectIDs(projectID).Exec(ctx)
}

func (self *InitiativeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Initiative.DeleteOneID(id).Exec(ctx)
}
