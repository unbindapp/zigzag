package apikey_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/apikey"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type CreateAPIKeyInput struct {
	UserID       uuid.UUID
	Name         string
	TokenPrefix  string
	TokenHash    string
	Role         schema.WorkspaceRole
	FullAccess   bool
	WorkspaceIDs []uuid.UUID
	Privileges   []schema.KeyPrivilege
	ExpiresAt    *time.Time
}

func (self *APIKeyRepository) Create(ctx context.Context, input *CreateAPIKeyInput) (*ent.APIKey, error) {
	return self.base.DB.APIKey.Create().
		SetUserID(input.UserID).
		SetName(input.Name).
		SetTokenPrefix(input.TokenPrefix).
		SetTokenHash(input.TokenHash).
		SetRole(input.Role).
		SetFullAccess(input.FullAccess).
		SetWorkspaceIds(input.WorkspaceIDs).
		SetPrivileges(input.Privileges).
		SetNillableExpiresAt(input.ExpiresAt).
		Save(ctx)
}

func (self *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.APIKey.DeleteOneID(id).Exec(ctx)
}

func (self *APIKeyRepository) TouchLastUsed(ctx context.Context, id uuid.UUID, now time.Time, minInterval time.Duration) error {
	_, err := self.base.DB.APIKey.Update().
		Where(
			apikey.ID(id),
			apikey.Or(
				apikey.LastUsedAtIsNil(),
				apikey.LastUsedAtLT(now.Add(-minInterval)),
			),
		).
		SetLastUsedAt(now).
		Save(ctx)
	return err
}
