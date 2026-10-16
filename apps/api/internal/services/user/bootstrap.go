package user_service

import (
	"context"

	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

type BootstrapInput struct {
	Email         string `json:"email" required:"true" format:"email" maxLength:"254" doc:"Email of the first user, who becomes system admin"`
	Password      string `json:"password" required:"true" minLength:"8" maxLength:"72" doc:"Password, at least 8 characters"`
	Name          string `json:"name" required:"true" minLength:"1" maxLength:"100" doc:"Full name"`
	DisplayName   string `json:"display_name" required:"true" minLength:"1" maxLength:"33" doc:"Short handle used in mentions"`
	WorkspaceName string `json:"workspace_name" required:"true" minLength:"1" maxLength:"64" doc:"Name of the first workspace"`
	WorkspaceSlug string `json:"workspace_slug" required:"true" minLength:"2" maxLength:"32" doc:"URL key of the first workspace, lowercase letters, digits and dashes"`
}

func (self *UserService) Bootstrap(ctx context.Context, input *BootstrapInput) (user *ent.User, workspace *ent.Workspace, err error) {
	email, err := names.Email(input.Email)
	if err != nil {
		return nil, nil, err
	}
	if len(input.Password) < MinPasswordLength {
		return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Password must be at least 8 characters")
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, nil, err
	}
	workspaceName, err := names.Clean(input.WorkspaceName)
	if err != nil {
		return nil, nil, err
	}
	slug, err := names.Slug(input.WorkspaceSlug)
	if err != nil {
		return nil, nil, err
	}

	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.System().LockSettings(ctx, tx); err != nil {
			return err
		}
		count, err := self.repo.User().Count(ctx, tx)
		if err != nil {
			return err
		}
		if count > 0 {
			return errdefs.ErrAlreadyBootstrapped
		}
		handle, err := self.freeDisplayNameTx(ctx, tx, input.DisplayName, nil)
		if err != nil {
			return err
		}
		user, err = self.repo.User().Create(ctx, tx, &user_repo.CreateUserInput{
			Email:       email,
			Password:    input.Password,
			Name:        name,
			DisplayName: handle,
			SystemAdmin: true,
			Active:      true,
		})
		if err != nil {
			return err
		}
		workspace, err = self.repo.Workspace().Create(ctx, tx, &workspace_repo.CreateWorkspaceInput{
			Name:      workspaceName,
			Slug:      slug,
			CreatorID: user.ID,
		})
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return user, workspace, nil
}
