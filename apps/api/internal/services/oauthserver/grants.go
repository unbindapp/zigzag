package oauthserver_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
)

func (self *OAuthServerService) ListGrants(ctx context.Context, userID uuid.UUID) ([]*models.ConnectedAppResponse, error) {
	grants, err := self.repo.OAuthServer().ListGrantsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return models.TransformOAuthGrantEntities(grants), nil
}

func (self *OAuthServerService) RevokeGrant(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	grant, err := self.repo.OAuthServer().GetGrantByID(ctx, id)
	if ent.IsNotFound(err) || (err == nil && grant.UserID != userID) {
		return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Connected app not found")
	}
	if err != nil {
		return err
	}
	return self.repo.OAuthServer().RevokeGrant(ctx, id, self.now())
}
