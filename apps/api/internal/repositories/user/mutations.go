package user_repo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/user"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

const PasswordCost = 12

type CreateUserInput struct {
	Email       string
	Password    string
	Name        string
	DisplayName string
	SystemAdmin bool
	Active      bool
}

const maxPasswordBytes = 72

func HashPassword(password string) (string, error) {
	if len(password) > maxPasswordBytes {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A password is at most 72 bytes; fewer characters when it uses accents or emoji")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (self *UserRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateUserInput) (*ent.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || input.DisplayName == "" {
		return nil, ErrInvalidUserInput
	}
	hash := ""
	if input.Password != "" {
		hashed, err := HashPassword(input.Password)
		if err != nil {
			return nil, err
		}
		hash = hashed
	}
	return self.base.Client(tx).User.Create().
		SetEmail(email).
		SetPasswordHash(hash).
		SetName(input.Name).
		SetDisplayName(strings.ToLower(input.DisplayName)).
		SetSystemAdmin(input.SystemAdmin).
		SetActive(input.Active).
		Save(ctx)
}

func (self *UserRepository) UpdatePassword(ctx context.Context, tx repository.TxInterface, userID uuid.UUID, password string) error {
	if password == "" {
		return ErrInvalidUserInput
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return self.base.Client(tx).User.UpdateOneID(userID).SetPasswordHash(hash).Exec(ctx)
}

type UpdateUserInput struct {
	Name           *string
	DisplayName    *string
	Timezone       *string
	AvatarUploadID *uuid.UUID
	ClearAvatar    bool
	SystemAdmin    *bool
	Active         *bool
}

func (self *UserRepository) Update(ctx context.Context, tx repository.TxInterface, userID uuid.UUID, input *UpdateUserInput) (*ent.User, error) {
	m := self.base.Client(tx).User.UpdateOneID(userID)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.DisplayName != nil {
		m.SetDisplayName(strings.ToLower(*input.DisplayName))
	}
	if input.Timezone != nil {
		m.SetTimezone(*input.Timezone)
	}
	if input.AvatarUploadID != nil {
		m.SetAvatarUploadID(*input.AvatarUploadID)
	}
	if input.ClearAvatar {
		m.ClearAvatarUploadID()
	}
	if input.SystemAdmin != nil {
		m.SetSystemAdmin(*input.SystemAdmin)
	}
	if input.Active != nil {
		m.SetActive(*input.Active)
	}
	return m.Save(ctx)
}

func (self *UserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return self.base.DB.User.DeleteOneID(userID).Exec(ctx)
}

func (self *UserRepository) TouchLastSeen(ctx context.Context, userID uuid.UUID) error {
	return self.base.DB.User.Update().Where(user.ID(userID)).SetLastSeenAt(timeNow()).Exec(ctx)
}
