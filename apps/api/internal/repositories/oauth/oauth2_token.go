package oauth_repo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/oauth2token"
	"github.com/zigzagapp-org/zigzag-api/ent/user"
)

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (self *OauthRepository) CreateToken(ctx context.Context, accessToken, refreshToken, clientID, scope string, expiresAt time.Time, user *ent.User) (*ent.Oauth2Token, error) {
	created, err := self.base.DB.Oauth2Token.Create().SetAccessToken(hashToken(accessToken)).SetRefreshToken(hashToken(refreshToken)).SetClientID(clientID).SetExpiresAt(expiresAt).SetScope(scope).SetUser(user).Save(ctx)
	if err != nil {
		return nil, err
	}

	created.Edges.User = user

	return created, nil
}

func (self *OauthRepository) RevokeAccessToken(ctx context.Context, accessToken string) error {
	_, err := self.base.DB.Oauth2Token.Update().Where(oauth2token.AccessToken(hashToken(accessToken))).SetRevoked(true).Save(ctx)
	return err
}

func (self *OauthRepository) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := self.base.DB.Oauth2Token.Update().Where(oauth2token.RefreshToken(hashToken(refreshToken))).SetRevoked(true).Save(ctx)
	return err
}

func (self *OauthRepository) GetByAccessToken(ctx context.Context, accessToken string) (*ent.Oauth2Token, error) {
	return self.base.DB.Oauth2Token.Query().Where(
		oauth2token.AccessToken(hashToken(accessToken)),
	).WithUser().Only(ctx)
}

func (self *OauthRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*ent.Oauth2Token, error) {
	return self.base.DB.Oauth2Token.Query().Where(
		oauth2token.RefreshToken(hashToken(refreshToken)),
	).WithUser().Only(ctx)
}

func (self *OauthRepository) CleanTokenStore(ctx context.Context) error {
	_, err := self.base.DB.Oauth2Token.Delete().Where(
		oauth2token.Or(
			oauth2token.Revoked(true),
			oauth2token.ExpiresAtLT(time.Now()),
		),
	).Exec(ctx)
	return err
}

func (self *OauthRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := self.base.DB.Oauth2Token.Update().
		Where(oauth2token.HasUserWith(user.ID(userID)), oauth2token.Revoked(false)).
		SetRevoked(true).
		Save(ctx)
	return err
}
