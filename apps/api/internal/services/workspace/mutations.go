package workspace_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	team_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/team"
	workspace_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/workspace"
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
	var team *ent.Team
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
		if err != nil {
			return err
		}
		team, err = self.repo.Team().Create(ctx, tx, team_repo.DefaultTeamInput(created, userID))
		return err
	})
	if err != nil {
		return nil, err
	}
	resp, err := self.respond(ctx, userID, created)
	if err != nil {
		return nil, err
	}
	resp.DefaultTeamID = &team.ID
	return resp, nil
}

type UpdateWorkspaceInput struct {
	WorkspaceID  uuid.UUID  `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace to change"`
	Name         *string    `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Slug         *string    `json:"slug,omitempty" required:"false" maxLength:"32" doc:"New URL key; old links stop working"`
	Icon         *string    `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	Color        *string    `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color, empty to clear"`
	IconUploadID *uuid.UUID `json:"icon_upload_id,omitempty" required:"false" format:"uuid" doc:"Logo image uploaded to this workspace with create-upload. Send 00000000-0000-0000-0000-000000000000 to remove the logo."`
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
	if input.IconUploadID != nil {
		if err := self.applyIconUpload(ctx, input.WorkspaceID, *input.IconUploadID, update); err != nil {
			return nil, err
		}
	}
	updated, err := self.repo.Workspace().Update(ctx, input.WorkspaceID, update)
	if err != nil {
		return nil, err
	}
	return self.respond(ctx, userID, updated)
}

func (self *WorkspaceService) applyIconUpload(ctx context.Context, workspaceID, uploadID uuid.UUID, update *workspace_repo.UpdateWorkspaceInput) error {
	if uploadID == uuid.Nil {
		update.ClearIconUpload = true
		return nil
	}
	upload, err := self.repo.Upload().GetByID(ctx, uploadID)
	if ent.IsNotFound(err) || (err == nil && upload.WorkspaceID != workspaceID) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The logo must be an image uploaded to this workspace")
	}
	if err != nil {
		return err
	}
	if !strings.HasPrefix(upload.ContentType, "image/") || upload.Width == nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The logo must be a png, jpeg, gif or webp image")
	}
	update.IconUploadID = &uploadID
	return nil
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
