package workspace_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type defaultProjectStatus struct {
	Name  string
	Type  schema.ProjectStatusType
	Color string
}

var DefaultProjectStatuses = []defaultProjectStatus{
	{Name: "Backlog", Type: schema.ProjectStatusBacklog, Color: "#95a2b3"},
	{Name: "Planned", Type: schema.ProjectStatusPlanned, Color: "#e2e2e2"},
	{Name: "In Progress", Type: schema.ProjectStatusStarted, Color: "#f2c94c"},
	{Name: "Paused", Type: schema.ProjectStatusPaused, Color: "#f2994a"},
	{Name: "Completed", Type: schema.ProjectStatusCompleted, Color: "#5e6ad2"},
	{Name: "Canceled", Type: schema.ProjectStatusCanceled, Color: "#95a2b3"},
}

type CreateWorkspaceInput struct {
	Name      string
	Slug      string
	Icon      string
	Color     string
	CreatorID uuid.UUID
}

// Writes several tables: call it inside a transaction.
func (self *WorkspaceRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateWorkspaceInput) (*ent.Workspace, error) {
	db := self.base.Client(tx)
	created, err := db.Workspace.Create().
		SetName(input.Name).
		SetSlug(input.Slug).
		SetIcon(input.Icon).
		SetColor(input.Color).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]*ent.ProjectStatusCreate, len(DefaultProjectStatuses))
	for i, status := range DefaultProjectStatuses {
		statuses[i] = db.ProjectStatus.Create().
			SetWorkspaceID(created.ID).
			SetName(status.Name).
			SetType(status.Type).
			SetColor(status.Color).
			SetPosition(float64(i))
	}
	if _, err := db.ProjectStatus.CreateBulk(statuses...).Save(ctx); err != nil {
		return nil, err
	}

	if _, err := self.AddMember(ctx, tx, created.ID, input.CreatorID, schema.WorkspaceRoleAdmin); err != nil {
		return nil, err
	}
	return created, nil
}

type UpdateWorkspaceInput struct {
	Name            *string
	Slug            *string
	Icon            *string
	Color           *string
	IconUploadID    *uuid.UUID
	ClearIconUpload bool
}

func (self *WorkspaceRepository) Update(ctx context.Context, id uuid.UUID, input *UpdateWorkspaceInput) (*ent.Workspace, error) {
	m := self.base.DB.Workspace.UpdateOneID(id)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.Slug != nil {
		m.SetSlug(*input.Slug)
	}
	if input.Icon != nil {
		m.SetIcon(*input.Icon)
	}
	if input.Color != nil {
		m.SetColor(*input.Color)
	}
	if input.IconUploadID != nil {
		m.SetIconUploadID(*input.IconUploadID)
	}
	if input.ClearIconUpload {
		m.ClearIconUploadID()
	}
	return m.Save(ctx)
}

func (self *WorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Workspace.DeleteOneID(id).Exec(ctx)
}
