package user_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/passwordresettoken"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

var timeNow = time.Now

func (self *UserRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.PasswordResetToken, error) {
	return self.base.DB.PasswordResetToken.Create().
		SetUserID(userID).
		SetTokenHash(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
}

func (self *UserRepository) GetPasswordResetToken(ctx context.Context, tokenHash string) (*ent.PasswordResetToken, error) {
	return self.base.DB.PasswordResetToken.Query().
		Where(passwordresettoken.TokenHash(tokenHash)).
		WithUser().
		Only(ctx)
}

func (self *UserRepository) ClaimPasswordResetToken(ctx context.Context, tx repository.TxInterface, id uuid.UUID, now time.Time) (bool, error) {
	updated, err := self.base.Client(tx).PasswordResetToken.Update().
		Where(passwordresettoken.ID(id), passwordresettoken.UsedAtIsNil(), passwordresettoken.ExpiresAtGT(now)).
		SetUsedAt(now).
		Save(ctx)
	return updated == 1, err
}

func (self *UserRepository) DeletePasswordResetTokensBefore(ctx context.Context, before time.Time) (int, error) {
	return self.base.DB.PasswordResetToken.Delete().
		Where(passwordresettoken.Or(passwordresettoken.ExpiresAtLT(before), passwordresettoken.UsedAtLT(before))).
		Exec(ctx)
}
