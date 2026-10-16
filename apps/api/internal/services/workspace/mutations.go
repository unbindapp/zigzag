package workspace_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

type CreateWorkspaceInput struct {
	Name  string `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"Workspace name"`
	Slug  string `json:"slug" required:"true" minLength:"2" maxLength:"32" doc:"URL key, lowercase letters, digits and dashes, unique across the instance"`
	Icon  string `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name"`
	Color string `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color such as #5e6ad2"`
}

func (self *WorkspaceService) Create(ctx context.Context, userID uuid.UUID, input *CreateWorkspaceInput) (*models.WorkspaceResponse, error) {
	if access, narrowed := permissions_repo.APIKeyAccessFromContext(ctx); narrowed && (!access.FullAccess || !access.AllowsWrites()) {
		return nil, errdefs.ErrUnauthorized
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	slug, err := names.Slug(input.Slug)
	if err != nil {
		return nil, err
	}
	color, err := names.OptionalColor(input.Color)
	if err != nil {
		return nil, err
	}

	var created *ent.Workspace
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		taken, err := self.repo.Workspace().SlugExists(ctx, tx, slug, nil)
		if err != nil {
			return err
		}
		if taken {
			return errdefs.NewCustomError(errdefs.ErrTypeConflict, "The workspace URL "+slug+" is taken")
		}
		created, err = self.repo.Workspace().Create(ctx, tx, &workspace_repo.CreateWorkspaceInput{
			Name: name, Slug: slug, Icon: input.Icon, Color: color, CreatorID: userID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return self.respond(ctx, userID, created)
}

type UpdateWorkspaceInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace to change"`
	Name        *string   `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Slug        *string   `json:"slug,omitempty" required:"false" maxLength:"32" doc:"New URL key; old links stop working"`
	Icon        *string   `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	Color       *string   `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color, empty to clear"`
}

func (self *WorkspaceService) Update(ctx context.Context, userID uuid.UUID, input *UpdateWorkspaceInput) (*models.WorkspaceResponse, error) {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return nil, err
	}
	update := &workspace_repo.UpdateWorkspaceInput{Icon: input.Icon}
	if input.Color != nil {
		color, err := names.OptionalColor(*input.Color)
		if err != nil {
			return nil, err
		}
		update.Color = &color
	}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Slug != nil {
		slug, err := names.Slug(*input.Slug)
		if err != nil {
			return nil, err
		}
		taken, err := self.repo.Workspace().SlugExists(ctx, nil, slug, &input.WorkspaceID)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, "The workspace URL "+slug+" is taken")
		}
		update.Slug = &slug
	}
	updated, err := self.repo.Workspace().Update(ctx, input.WorkspaceID, update)
	if err != nil {
		return nil, err
	}
	return self.respond(ctx, userID, updated)
}

type DeleteWorkspaceInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace to delete with all its teams, issues, projects and documents"`
	ConfirmSlug string    `json:"confirm_slug" required:"true" minLength:"1" maxLength:"32" doc:"The workspace slug, repeated to confirm"`
}

func (self *WorkspaceService) Delete(ctx context.Context, userID uuid.UUID, input *DeleteWorkspaceInput) error {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return err
	}
	workspace, err := self.repo.Workspace().GetByID(ctx, input.WorkspaceID)
	if err != nil {
		return err
	}
	if input.ConfirmSlug != workspace.Slug {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "confirm_slug does not match the workspace slug")
	}
	return self.repo.Workspace().Delete(ctx, workspace.ID)
}
