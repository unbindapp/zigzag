package oauthserver_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/oauthauthorizationcode"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type CreateCodeInput struct {
	UserID        uuid.UUID
	CodeHash      string
	ClientID      string
	ClientName    string
	ClientKind    schema.OAuthClientKind
	ClientURI     string
	RedirectURI   string
	CodeChallenge string
	Resource      string
	Scope         string
	Role          schema.WorkspaceRole
	FullAccess    bool
	WorkspaceIDs  []uuid.UUID
	Privileges    []schema.KeyPrivilege
	ExpiresAt     time.Time
}

func (self *OAuthServerRepository) CreateCode(ctx context.Context, input *CreateCodeInput) (*ent.OAuthAuthorizationCode, error) {
	return self.base.DB.OAuthAuthorizationCode.Create().
		SetUserID(input.UserID).
		SetCodeHash(input.CodeHash).
		SetClientID(input.ClientID).
		SetClientName(input.ClientName).
		SetClientKind(input.ClientKind).
		SetClientURI(input.ClientURI).
		SetRedirectURI(input.RedirectURI).
		SetCodeChallenge(input.CodeChallenge).
		SetResource(input.Resource).
		SetScope(input.Scope).
		SetRole(input.Role).
		SetFullAccess(input.FullAccess).
		SetWorkspaceIds(input.WorkspaceIDs).
		SetPrivileges(input.Privileges).
		SetExpiresAt(input.ExpiresAt).
		Save(ctx)
}

func (self *OAuthServerRepository) GetCodeByHash(ctx context.Context, codeHash string) (*ent.OAuthAuthorizationCode, error) {
	return self.base.DB.OAuthAuthorizationCode.Query().
		Where(oauthauthorizationcode.CodeHash(codeHash)).
		WithUser().
		Only(ctx)
}

func (self *OAuthServerRepository) MarkCodeUsed(ctx context.Context, id uuid.UUID, now time.Time, grantID uuid.UUID) (bool, error) {
	updated, err := self.base.DB.OAuthAuthorizationCode.Update().
		Where(oauthauthorizationcode.ID(id), oauthauthorizationcode.UsedAtIsNil()).
		SetUsedAt(now).
		SetGrantID(grantID).
		Save(ctx)
	return updated == 1, err
}

func (self *OAuthServerRepository) DeleteCodesExpiredBefore(ctx context.Context, before time.Time) (int, error) {
	return self.base.DB.OAuthAuthorizationCode.Delete().
		Where(oauthauthorizationcode.ExpiresAtLT(before)).
		Exec(ctx)
}
