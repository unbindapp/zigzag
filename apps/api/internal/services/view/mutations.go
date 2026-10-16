package view_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	view_repo "github.com/unbindapp/zigzag-api/internal/repositories/view"
)

type CreateViewInput struct {
	WorkspaceID uuid.UUID          `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID      *uuid.UUID         `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Make it a view of this team; omit for a workspace view"`
	Name        string             `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"View name"`
	Description string             `json:"description,omitempty" required:"false" maxLength:"500" doc:"What the view shows"`
	Icon        string             `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name"`
	Color       string             `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color such as #5e6ad2"`
	Kind        schema.ViewKind    `json:"kind" required:"true" doc:"issue for issue views; project views are reserved"`
	Shared      bool               `json:"shared,omitempty" required:"false" doc:"Share with everyone in the workspace; private views are visible to you only"`
	Filters     map[string]any     `json:"filters" required:"true" doc:"list-issues parameters to apply, for example {\"state_types\": [\"started\"], \"assignee_ids\": [\"<user id>\"]}"`
	Display     schema.ViewDisplay `json:"display" required:"true" doc:"How the issues show: layout, grouping, ordering"`
}

func (self *ViewService) Create(ctx context.Context, userID uuid.UUID, input *CreateViewInput) (*models.ViewResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleMember, input.WorkspaceID); err != nil {
		return nil, err
	}
	if input.TeamID != nil {
		if _, err := self.teamInWorkspace(ctx, userID, input.WorkspaceID, *input.TeamID); err != nil {
			return nil, err
		}
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	if !validKind(input.Kind) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "kind must be issue or project")
	}
	color, err := names.OptionalColor(input.Color)
	if err != nil {
		return nil, err
	}
	filters := input.Filters
	if filters == nil {
		filters = map[string]any{}
	}
	if err := validateFilters(filters); err != nil {
		return nil, err
	}
	if err := validateDisplay(input.Display); err != nil {
		return nil, err
	}
	created, err := self.repo.View().Create(ctx, nil, &view_repo.CreateViewInput{
		WorkspaceID: input.WorkspaceID,
		TeamID:      input.TeamID,
		OwnerID:     userID,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		Icon:        input.Icon,
		Color:       color,
		Kind:        input.Kind,
		Shared:      input.Shared,
		Filters:     filters,
		Display:     input.Display,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformViewEntity(created, userID), nil
}

type UpdateViewInput struct {
	ViewID      uuid.UUID           `json:"view_id" required:"true" format:"uuid" doc:"The view to change; you must own it, or be an admin for a shared view"`
	Name        *string             `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Description *string             `json:"description,omitempty" required:"false" maxLength:"500" doc:"Description, empty to clear"`
	Icon        *string             `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	Color       *string             `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color, empty to clear"`
	Shared      *bool               `json:"shared,omitempty" required:"false" doc:"Share with the workspace or make private"`
	Filters     map[string]any      `json:"filters,omitempty" required:"false" doc:"Replaces all filters; list-issues parameter names"`
	Display     *schema.ViewDisplay `json:"display,omitempty" required:"false" doc:"Replaces the display settings"`
}

func (self *ViewService) Update(ctx context.Context, userID uuid.UUID, input *UpdateViewInput) (*models.ViewResponse, error) {
	current, err := self.editableView(ctx, userID, input.ViewID)
	if err != nil {
		return nil, err
	}
	update := &view_repo.UpdateViewInput{Icon: input.Icon, Shared: input.Shared, Filters: input.Filters, Display: input.Display}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		update.Description = &description
	}
	if input.Color != nil {
		color, err := names.OptionalColor(*input.Color)
		if err != nil {
			return nil, err
		}
		update.Color = &color
	}
	if input.Shared != nil && !*input.Shared && current.OwnerID != userID {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Only the owner can make a shared view private")
	}
	if err := validateFilters(input.Filters); err != nil {
		return nil, err
	}
	if input.Display != nil {
		if err := validateDisplay(*input.Display); err != nil {
			return nil, err
		}
	}
	updated, err := self.repo.View().Update(ctx, current.ID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformViewEntity(updated, userID), nil
}

type DeleteViewInput struct {
	ViewID uuid.UUID `json:"view_id" required:"true" format:"uuid" doc:"The view to delete; you must own it, or be an admin for a shared view"`
}

func (self *ViewService) Delete(ctx context.Context, userID uuid.UUID, input *DeleteViewInput) error {
	current, err := self.editableView(ctx, userID, input.ViewID)
	if err != nil {
		return err
	}
	return self.repo.View().Delete(ctx, current.ID)
}
