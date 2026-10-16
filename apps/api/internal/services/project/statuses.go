package project_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
)

type ListProjectStatusesInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
}

func (self *ProjectService) ListStatuses(ctx context.Context, userID uuid.UUID, input *ListProjectStatusesInput) ([]*models.ProjectStatusResponse, error) {
	if err := self.checkVisible(ctx, userID, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	statuses, err := self.repo.Project().ListStatuses(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return models.TransformProjectStatuses(statuses), nil
}

type CreateProjectStatusInput struct {
	WorkspaceID uuid.UUID                `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Name        string                   `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"Status name, unique in the workspace"`
	Type        schema.ProjectStatusType `json:"type" required:"true" doc:"backlog, planned, started, paused, completed or canceled"`
	Color       string                   `json:"color" required:"true" maxLength:"7" pattern:"^#[0-9a-fA-F]{6}$" doc:"Hex color such as #f2c94c"`
	Description string                   `json:"description,omitempty" required:"false" maxLength:"500" doc:"What the status means"`
	Position    *float64                 `json:"position,omitempty" required:"false" doc:"Order within its type; appended last when omitted"`
}

func (self *ProjectService) CreateStatus(ctx context.Context, userID uuid.UUID, input *CreateProjectStatusInput) (*models.ProjectStatusResponse, error) {
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	if !validStatusType(input.Type) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "type must be backlog, planned, started, paused, completed or canceled")
	}
	name, err := self.freeStatusName(ctx, input.WorkspaceID, input.Name, nil)
	if err != nil {
		return nil, err
	}
	color, err := names.Color(input.Color)
	if err != nil {
		return nil, err
	}
	position := input.Position
	if position == nil {
		next, err := self.repo.Project().NextStatusPosition(ctx, input.WorkspaceID, input.Type)
		if err != nil {
			return nil, err
		}
		position = &next
	}
	status, err := self.repo.Project().CreateStatus(ctx, &project_repo.CreateStatusInput{
		WorkspaceID: input.WorkspaceID, Name: name, Type: input.Type, Color: color,
		Description: strings.TrimSpace(input.Description), Position: *position,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformProjectStatus(status), nil
}

type UpdateProjectStatusInput struct {
	StatusID    uuid.UUID `json:"status_id" required:"true" format:"uuid" doc:"The project status to change"`
	Name        *string   `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Color       *string   `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^#[0-9a-fA-F]{6}$" doc:"Hex color"`
	Description *string   `json:"description,omitempty" required:"false" maxLength:"500" doc:"Description, empty to clear"`
	Position    *float64  `json:"position,omitempty" required:"false" doc:"Order within its type"`
}

func (self *ProjectService) UpdateStatus(ctx context.Context, userID uuid.UUID, input *UpdateProjectStatusInput) (*models.ProjectStatusResponse, error) {
	status, err := self.adminStatus(ctx, userID, input.StatusID)
	if err != nil {
		return nil, err
	}
	update := &project_repo.UpdateStatusInput{Position: input.Position}
	if input.Name != nil {
		name, err := self.freeStatusName(ctx, status.WorkspaceID, *input.Name, &status.ID)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Color != nil {
		color, err := names.Color(*input.Color)
		if err != nil {
			return nil, err
		}
		update.Color = &color
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		update.Description = &description
	}
	updated, err := self.repo.Project().UpdateStatus(ctx, status.ID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformProjectStatus(updated), nil
}

type DeleteProjectStatusInput struct {
	StatusID       uuid.UUID `json:"status_id" required:"true" format:"uuid" doc:"The project status to delete"`
	MoveToStatusID uuid.UUID `json:"move_to_status_id" required:"true" format:"uuid" doc:"Another status of the same workspace that receives the deleted status's projects"`
}

func (self *ProjectService) DeleteStatus(ctx context.Context, userID uuid.UUID, input *DeleteProjectStatusInput) error {
	status, err := self.adminStatus(ctx, userID, input.StatusID)
	if err != nil {
		return err
	}
	if input.MoveToStatusID == status.ID {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "move_to_status_id must be a different status")
	}
	if _, err := self.workspaceStatus(ctx, status.WorkspaceID, input.MoveToStatusID); err != nil {
		return err
	}
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		return self.repo.Project().DeleteStatus(ctx, tx, status.ID, input.MoveToStatusID)
	})
}

func (self *ProjectService) adminStatus(ctx context.Context, userID, statusID uuid.UUID) (*ent.ProjectStatus, error) {
	status, err := self.repo.Project().GetStatus(ctx, statusID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Project status not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleAdmin, status.WorkspaceID, "Project status not found"); err != nil {
		return nil, err
	}
	return status, nil
}

func (self *ProjectService) freeStatusName(ctx context.Context, workspaceID uuid.UUID, raw string, except *uuid.UUID) (string, error) {
	name, err := names.Clean(raw)
	if err != nil {
		return "", err
	}
	taken, err := self.repo.Project().StatusNameTaken(ctx, workspaceID, name, except)
	if err != nil {
		return "", err
	}
	if taken {
		return "", names.EnsureFree(name, []string{name}, "project status", "workspace")
	}
	return name, nil
}
