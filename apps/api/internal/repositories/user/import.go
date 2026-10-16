package user_repo

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/predicate"
	"github.com/zigzagapp-org/zigzag-api/ent/user"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

const importBatchSize = 500

type ImportedUser struct {
	ID          uuid.UUID
	Email       string
	Name        string
	DisplayName string
}

func (self *UserRepository) GetByEmails(ctx context.Context, emails []string) ([]*ent.User, error) {
	if len(emails) == 0 {
		return []*ent.User{}, nil
	}
	lowered := make([]string, len(emails))
	for i, email := range emails {
		lowered[i] = strings.ToLower(strings.TrimSpace(email))
	}
	return self.base.DB.User.Query().Where(user.EmailIn(lowered...)).All(ctx)
}

func (self *UserRepository) DisplayNamesWithPrefixes(ctx context.Context, prefixes []string) ([]string, error) {
	if len(prefixes) == 0 {
		return []string{}, nil
	}
	matches := make([]predicate.User, len(prefixes))
	for i, prefix := range prefixes {
		matches[i] = user.DisplayNameHasPrefix(prefix)
	}
	return self.base.DB.User.Query().Where(user.Or(matches...)).Select(user.FieldDisplayName).Strings(ctx)
}

func (self *UserRepository) CreatePlaceholders(ctx context.Context, tx repository.TxInterface, users []ImportedUser) error {
	db := self.base.Client(tx)
	builders := make([]*ent.UserCreate, len(users))
	for i, u := range users {
		builders[i] = db.User.Create().
			SetID(u.ID).
			SetEmail(strings.ToLower(u.Email)).
			SetName(u.Name).
			SetDisplayName(u.DisplayName).
			SetActive(false)
	}
	for batch := range slices.Chunk(builders, importBatchSize) {
		if err := db.User.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
