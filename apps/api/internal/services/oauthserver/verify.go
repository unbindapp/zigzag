package oauthserver_service

import (
	"context"
	"strings"

	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/auth"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

func (self *OAuthServerService) VerifyAccessToken(ctx context.Context, token string) (*ent.OAuthGrantToken, error) {
	if !strings.HasPrefix(token, auth.OAuthAccessTokenPrefix) {
		return nil, ErrInvalidAccessToken
	}
	now := self.now()
	stored, err := self.repo.OAuthServer().GetTokenByHash(ctx, auth.HashAPIKey(token))
	if ent.IsNotFound(err) {
		return nil, ErrInvalidAccessToken
	}
	if err != nil {
		return nil, err
	}
	grant := stored.Edges.Grant
	if stored.Kind != schema.OAuthTokenKindAccess || grant == nil || grant.Edges.User == nil {
		return nil, ErrInvalidAccessToken
	}
	if !stored.ExpiresAt.After(now) || grant.RevokedAt != nil || grant.Resource != self.resource {
		return nil, ErrInvalidAccessToken
	}
	if err := self.repo.OAuthServer().TouchGrantLastUsed(ctx, grant.ID, now, lastUsedInterval); err != nil {
		log.Warnf("oauth: record grant use: %v", err)
	}
	return stored, nil
}

func AccessOf(grant *ent.OAuthGrant) permissions_repo.APIKeyAccess {
	return permissions_repo.APIKeyAccess{Role: grant.Role, FullAccess: grant.FullAccess, WorkspaceIDs: grant.WorkspaceIds, Privileges: grant.Privileges}
}
