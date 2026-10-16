package user_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
)

type CreateUserInput struct {
	Email       string `json:"email" required:"true" format:"email" maxLength:"254" doc:"Email address, used to sign in"`
	Password    string `json:"password" required:"true" minLength:"8" maxLength:"72" doc:"Initial password, at least 8 characters"`
	Name        string `json:"name" required:"true" minLength:"1" maxLength:"100" doc:"Full name"`
	DisplayName string `json:"display_name" required:"true" minLength:"1" maxLength:"33" doc:"Short handle used in mentions, lowercase and unique"`
	SystemAdmin bool   `json:"system_admin" required:"false" doc:"Grant instance administration"`
}

func (self *UserService) CreateUser(ctx context.Context, requesterUserID uuid.UUID, input *CreateUserInput) (*models.UserResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return nil, err
	}
	created, err := self.createAccount(ctx, input)
	if err != nil {
		return nil, err
	}
	return models.TransformUserEntity(created), nil
}

func (self *UserService) createAccount(ctx context.Context, input *CreateUserInput) (*ent.User, error) {
	email, err := names.Email(input.Email)
	if err != nil {
		return nil, err
	}
	if len(input.Password) < MinPasswordLength {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Password must be at least 8 characters")
	}
	_, err = self.repo.User().GetByEmail(ctx, email)
	if err == nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, "An account with this email already exists")
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	handle, err := self.freeDisplayName(ctx, input.DisplayName, nil)
	if err != nil {
		return nil, err
	}
	return self.repo.User().Create(ctx, nil, &user_repo.CreateUserInput{
		Email:       email,
		Password:    input.Password,
		Name:        name,
		DisplayName: handle,
		SystemAdmin: input.SystemAdmin,
		Active:      true,
	})
}

type UpdateUserInput struct {
	UserID      uuid.UUID `json:"user_id" required:"true" format:"uuid" doc:"The user to change"`
	Name        *string   `json:"name,omitempty" required:"false" maxLength:"100" doc:"Full name"`
	DisplayName *string   `json:"display_name,omitempty" required:"false" maxLength:"33" doc:"Short handle used in mentions"`
	SystemAdmin *bool     `json:"system_admin,omitempty" required:"false" doc:"Grant or revoke instance administration"`
	Active      *bool     `json:"active,omitempty" required:"false" doc:"Inactive users cannot sign in"`
	Password    *string   `json:"password,omitempty" required:"false" minLength:"8" maxLength:"72" doc:"Set a new password; also activates placeholder accounts when active is sent"`
	GithubLogin *string   `json:"github_login,omitempty" required:"false" maxLength:"40" doc:"GitHub username used by GitHub Issues sync; empty clears it"`
}

func (self *UserService) UpdateUser(ctx context.Context, requesterUserID uuid.UUID, input *UpdateUserInput) (*models.UserResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return nil, err
	}
	target, err := self.repo.User().GetByID(ctx, input.UserID)
	if err != nil {
		return nil, errdefs.MaskAsNotFound(err, "User not found")
	}

	demoting := (input.SystemAdmin != nil && !*input.SystemAdmin) || (input.Active != nil && !*input.Active)
	if demoting && target.SystemAdmin {
		last, err := self.lastActiveSystemAdmin(ctx, target.ID)
		if err != nil {
			return nil, err
		}
		if last {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The last system admin cannot be demoted or deactivated")
		}
	}

	update := &user_repo.UpdateUserInput{SystemAdmin: input.SystemAdmin, Active: input.Active}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.DisplayName != nil {
		handle, err := self.freeDisplayName(ctx, *input.DisplayName, &target.ID)
		if err != nil {
			return nil, err
		}
		update.DisplayName = &handle
	}
	if input.GithubLogin != nil {
		login, err := self.freeGithubLogin(ctx, *input.GithubLogin, target.ID)
		if err != nil {
			return nil, err
		}
		update.GithubLogin = &login
	}
	if input.Password != nil {
		if len(*input.Password) < MinPasswordLength {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Password must be at least 8 characters")
		}
		if err := self.repo.User().UpdatePassword(ctx, nil, target.ID, *input.Password); err != nil {
			return nil, err
		}
	}

	updated, err := self.repo.User().Update(ctx, nil, target.ID, update)
	if err != nil {
		return nil, err
	}
	if input.Active != nil && !*input.Active {
		if err := self.repo.Oauth().RevokeAllForUser(ctx, target.ID); err != nil {
			return nil, err
		}
	}
	return models.TransformUserEntity(updated), nil
}

type DeleteUserInput struct {
	UserID uuid.UUID `json:"user_id" required:"true" format:"uuid" doc:"The user to delete. Their issues and comments stay, without an author."`
}

func (self *UserService) DeleteUser(ctx context.Context, requesterUserID uuid.UUID, input *DeleteUserInput) error {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return err
	}
	if input.UserID == requesterUserID {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "You cannot delete your own account")
	}
	target, err := self.repo.User().GetByID(ctx, input.UserID)
	if err != nil {
		return errdefs.MaskAsNotFound(err, "User not found")
	}
	if target.SystemAdmin {
		last, err := self.lastActiveSystemAdmin(ctx, target.ID)
		if err != nil {
			return err
		}
		if last {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The last system admin cannot be deleted")
		}
	}
	return self.repo.User().Delete(ctx, target.ID)
}
