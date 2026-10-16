package user_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/predicate"
	"github.com/unbindapp/zigzag-api/ent/user"
	"github.com/unbindapp/zigzag-api/ent/workspacemember"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidUserInput = errors.New("invalid user input")
)

var timingHash, _ = bcrypt.GenerateFromPassword([]byte("zigzag-timing-equalizer"), PasswordCost)

func (self *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return self.base.DB.User.Get(ctx, id)
}

func (self *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.User, error) {
	return self.base.DB.User.Query().Where(user.IDIn(ids...)).All(ctx)
}

func (self *UserRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	return self.base.DB.User.Query().Where(user.EmailEqualFold(strings.TrimSpace(email))).Only(ctx)
}

func (self *UserRepository) GetByDisplayName(ctx context.Context, tx repository.TxInterface, displayName string) (*ent.User, error) {
	return self.base.Client(tx).User.Query().Where(user.DisplayNameEqualFold(displayName)).Only(ctx)
}

func (self *UserRepository) Count(ctx context.Context, tx repository.TxInterface) (int, error) {
	return self.base.Client(tx).User.Query().Count(ctx)
}

func (self *UserRepository) Authenticate(ctx context.Context, email, password string) (*ent.User, error) {
	if email == "" || password == "" {
		return nil, ErrInvalidUserInput
	}
	found, err := self.GetByEmail(ctx, email)
	if ent.IsNotFound(err) {
		_ = bcrypt.CompareHashAndPassword(timingHash, []byte(password))
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	if found.PasswordHash == "" {
		_ = bcrypt.CompareHashAndPassword(timingHash, []byte(password))
		return nil, ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(found.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}
	return found, nil
}

func (self *UserRepository) VerifyPassword(ctx context.Context, userID uuid.UUID, password string) error {
	found, err := self.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if found.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(found.PasswordHash), []byte(password)) != nil {
		return ErrInvalidPassword
	}
	return nil
}

type ListUsersInput struct {
	Query string
	Limit int
	// nil lists every user; otherwise only members of these workspaces
	WorkspaceIDs []uuid.UUID
}

func (self *UserRepository) List(ctx context.Context, input *ListUsersInput) ([]*ent.User, error) {
	q := self.base.DB.User.Query()
	if input.WorkspaceIDs != nil {
		q = q.Where(user.HasWorkspaceMembershipsWith(workspacemember.WorkspaceIDIn(input.WorkspaceIDs...)))
	}
	if query := strings.TrimSpace(input.Query); query != "" {
		q = q.Where(user.Or(
			user.EmailContainsFold(query),
			user.NameContainsFold(query),
			user.DisplayNameContainsFold(query),
		))
	}
	if input.Limit > 0 {
		q = q.Limit(input.Limit)
	}
	return q.Order(ent.Asc(user.FieldName), ent.Asc(user.FieldEmail)).All(ctx)
}

func (self *UserRepository) CountSystemAdmins(ctx context.Context, exclude ...predicate.User) (int, error) {
	q := self.base.DB.User.Query().Where(user.SystemAdmin(true), user.Active(true))
	if len(exclude) > 0 {
		q = q.Where(user.Not(user.And(exclude...)))
	}
	return q.Count(ctx)
}
