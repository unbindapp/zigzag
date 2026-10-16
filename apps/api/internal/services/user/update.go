package user_service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/user"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
)

type UpdateMeInput struct {
	Name           *string    `json:"name,omitempty" required:"false" maxLength:"100" doc:"Full name"`
	DisplayName    *string    `json:"display_name,omitempty" required:"false" maxLength:"33" doc:"Short handle used in mentions, lowercase and unique"`
	Timezone       *string    `json:"timezone,omitempty" required:"false" maxLength:"64" doc:"IANA time zone such as Europe/Berlin"`
	AvatarUploadID *uuid.UUID `json:"avatar_upload_id,omitempty" required:"false" format:"uuid" doc:"An image you uploaded. Send 00000000-0000-0000-0000-000000000000 to remove the avatar."`
	GithubLogin    *string    `json:"github_login,omitempty" required:"false" maxLength:"40" doc:"Your GitHub username, so GitHub Issues sync credits and assigns you; empty clears it"`
}

func (self *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, input *UpdateMeInput) (*models.UserResponse, error) {
	update := &user_repo.UpdateUserInput{}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.DisplayName != nil {
		handle, err := self.freeDisplayName(ctx, *input.DisplayName, &userID)
		if err != nil {
			return nil, err
		}
		update.DisplayName = &handle
	}
	if input.Timezone != nil {
		if _, err := time.LoadLocation(*input.Timezone); err != nil || *input.Timezone == "" {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Timezone must be an IANA time zone such as Europe/Berlin")
		}
		update.Timezone = input.Timezone
	}
	if input.AvatarUploadID != nil {
		if err := self.applyAvatar(ctx, userID, *input.AvatarUploadID, update); err != nil {
			return nil, err
		}
	}
	if input.GithubLogin != nil {
		login, err := self.freeGithubLogin(ctx, *input.GithubLogin, userID)
		if err != nil {
			return nil, err
		}
		update.GithubLogin = &login
	}

	updated, err := self.repo.User().Update(ctx, nil, userID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformUserEntity(updated), nil
}

func (self *UserService) applyAvatar(ctx context.Context, userID, uploadID uuid.UUID, update *user_repo.UpdateUserInput) error {
	if uploadID == uuid.Nil {
		update.ClearAvatar = true
		return nil
	}
	upload, err := self.repo.Upload().GetByID(ctx, uploadID)
	if ent.IsNotFound(err) || (err == nil && (upload.UploaderID == nil || *upload.UploaderID != userID)) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Avatar must be an image you uploaded")
	}
	if err != nil {
		return err
	}
	if !strings.HasPrefix(upload.ContentType, "image/") || upload.Width == nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Avatar must be a png, jpeg, gif or webp image")
	}
	update.AvatarUploadID = &uploadID
	return nil
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" required:"true" minLength:"1" doc:"Your current password"`
	NewPassword     string `json:"new_password" required:"true" minLength:"8" maxLength:"72" doc:"The new password, at least 8 characters"`
}

func (self *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, input *ChangePasswordInput) error {
	if len(input.NewPassword) < MinPasswordLength {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Password must be at least 8 characters")
	}
	if err := self.repo.User().VerifyPassword(ctx, userID, input.CurrentPassword); err != nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Current password is wrong")
	}
	return self.repo.User().UpdatePassword(ctx, nil, userID, input.NewPassword)
}

type UpdateNotificationSettingsInput struct {
	EmailEnabled   *bool `json:"email_enabled,omitempty" required:"false" doc:"Send email for notifications at all"`
	Assigned       *bool `json:"assigned,omitempty" required:"false" doc:"Issues assigned to or unassigned from you"`
	Mentioned      *bool `json:"mentioned,omitempty" required:"false" doc:"Mentions in issues and comments"`
	Commented      *bool `json:"commented,omitempty" required:"false" doc:"Comments on issues you subscribe to"`
	StateChanged   *bool `json:"state_changed,omitempty" required:"false" doc:"State changes on issues you subscribe to"`
	DueSoon        *bool `json:"due_soon,omitempty" required:"false" doc:"Issues due soon"`
	ProjectUpdates *bool `json:"project_updates,omitempty" required:"false" doc:"Status updates on projects you follow"`
}

func (self *UserService) UpdateNotificationSettings(ctx context.Context, userID uuid.UUID, input *UpdateNotificationSettingsInput) (*models.NotificationSettingsResponse, error) {
	settings, err := self.repo.User().UpdateNotificationSettings(ctx, userID, &user_repo.UpdateNotificationSettingsInput{
		EmailEnabled:   input.EmailEnabled,
		Assigned:       input.Assigned,
		Mentioned:      input.Mentioned,
		Commented:      input.Commented,
		StateChanged:   input.StateChanged,
		DueSoon:        input.DueSoon,
		ProjectUpdates: input.ProjectUpdates,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformNotificationSettings(settings), nil
}

func (self *UserService) freeGithubLogin(ctx context.Context, raw string, owner uuid.UUID) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	login, ok := github.NormalizeLogin(raw)
	if !ok {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "GitHub username must be letters, digits and single hyphens, at most 39 characters")
	}
	existing, err := self.repo.User().GetByGithubLogin(ctx, login)
	if ent.IsNotFound(err) {
		return login, nil
	}
	if err != nil {
		return "", err
	}
	if existing.ID == owner {
		return login, nil
	}
	return "", errdefs.NewCustomError(errdefs.ErrTypeConflict, "The GitHub username "+login+" belongs to another Zigzag user")
}

func (self *UserService) freeDisplayName(ctx context.Context, raw string, owner *uuid.UUID) (string, error) {
	return self.freeDisplayNameTx(ctx, nil, raw, owner)
}

func (self *UserService) freeDisplayNameTx(ctx context.Context, tx repository.TxInterface, raw string, owner *uuid.UUID) (string, error) {
	handle, err := names.DisplayName(raw)
	if err != nil {
		return "", err
	}
	existing, err := self.repo.User().GetByDisplayName(ctx, tx, handle)
	if ent.IsNotFound(err) {
		return handle, nil
	}
	if err != nil {
		return "", err
	}
	if owner != nil && existing.ID == *owner {
		return handle, nil
	}
	return "", errdefs.NewCustomError(errdefs.ErrTypeConflict, "The display name @"+handle+" is taken")
}

func (self *UserService) lastActiveSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	others, err := self.repo.User().CountSystemAdmins(ctx, user.ID(userID))
	return others == 0, err
}
