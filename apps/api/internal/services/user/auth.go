package user_service

import (
	"context"
	"errors"

	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/mail"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

func (self *UserService) Login(ctx context.Context, email, password string) (*ent.User, error) {
	user, err := self.repo.User().Authenticate(ctx, email, password)
	if errors.Is(err, user_repo.ErrInvalidUserInput) || errors.Is(err, user_repo.ErrUserNotFound) || errors.Is(err, user_repo.ErrInvalidPassword) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !user.Active {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// RequestPasswordReset answers the same whether or not the account exists.
func (self *UserService) RequestPasswordReset(ctx context.Context, rawEmail string) error {
	email, err := names.Email(rawEmail)
	if err != nil {
		return nil
	}
	user, err := self.repo.User().GetByEmail(ctx, email)
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !user.Active || (user.PasswordHash == "" && user.GoogleSubject == nil) {
		return nil
	}
	configured, err := self.email.Configured(ctx)
	if err != nil || !configured {
		return err
	}

	token, err := utils.RandomToken(32)
	if err != nil {
		return errdefs.NewInternalError(err, "Failed to generate a reset token")
	}
	if _, err := self.repo.User().CreatePasswordResetToken(ctx, user.ID, utils.HashToken(token), self.now().Add(passwordResetTTL)); err != nil {
		return err
	}
	link, err := utils.JoinURLPaths(self.cfg.ExternalUIUrl, "reset-password")
	if err != nil {
		return err
	}
	link += "?token=" + token
	self.email.SendAsync(ctx, user.Email, mail.Content{
		Subject:     "Reset your Zigzag password",
		Heading:     "Reset your password",
		Paragraphs:  []string{"Someone asked to reset the password for " + user.Email + ". Choose a new one with the link below."},
		ButtonLabel: "Choose a new password",
		ButtonURL:   link,
		Note:        "The link works once and expires in one hour. If you did not ask for it, ignore this email and your password stays the same.",
	})
	return nil
}

type ResetPasswordInput struct {
	Token    string `json:"token" required:"true" minLength:"1" maxLength:"200" doc:"Token from the reset link"`
	Password string `json:"password" required:"true" minLength:"8" maxLength:"72" doc:"The new password, at least 8 characters"`
}

func (self *UserService) ResetPassword(ctx context.Context, input *ResetPasswordInput) error {
	invalid := errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "This reset link is no longer valid")
	if len(input.Password) < MinPasswordLength {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Password must be at least 8 characters")
	}
	stored, err := self.repo.User().GetPasswordResetToken(ctx, utils.HashToken(input.Token))
	if ent.IsNotFound(err) {
		return invalid
	}
	if err != nil {
		return err
	}
	if stored.UsedAt != nil || !stored.ExpiresAt.After(self.now()) || stored.Edges.User == nil || !stored.Edges.User.Active {
		return invalid
	}

	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		claimed, err := self.repo.User().ClaimPasswordResetToken(ctx, tx, stored.ID, self.now())
		if err != nil {
			return err
		}
		if !claimed {
			return invalid
		}
		return self.repo.User().UpdatePassword(ctx, tx, stored.UserID, input.Password)
	})
	if err != nil {
		return err
	}
	if err := self.repo.Oauth().RevokeAllForUser(ctx, stored.UserID); err != nil {
		log.Warn("Failed to revoke sessions after a password reset", "user_id", stored.UserID, "err", err)
	}
	return nil
}

func (self *UserService) CleanupPasswordResetTokens(ctx context.Context) error {
	_, err := self.repo.User().DeletePasswordResetTokensBefore(ctx, self.now())
	return err
}

func (self *UserService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return self.repo.Oauth().RevokeRefreshToken(ctx, refreshToken)
}
