package oauthserver_service

import (
	"context"
	"errors"
)

func (self *OAuthServerService) Cleanup(ctx context.Context) error {
	var failures []error
	now := self.now()
	repo := self.repo.OAuthServer()
	if _, err := repo.DeleteCodesExpiredBefore(ctx, now); err != nil {
		failures = append(failures, err)
	}
	if _, err := repo.DeleteTokensExpiredOrUsedBefore(ctx, now.Add(-retiredTokenKeep)); err != nil {
		failures = append(failures, err)
	}
	if _, err := repo.DeleteOrphanClientsBefore(ctx, now.Add(-orphanClientKeep)); err != nil {
		failures = append(failures, err)
	}
	if err := self.repo.Oauth().CleanTokenStore(ctx); err != nil {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}
