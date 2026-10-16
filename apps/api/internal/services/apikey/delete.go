package apikey_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
)

func (self *APIKeyService) Delete(ctx context.Context, requesterUserID uuid.UUID, id uuid.UUID) error {
	key, err := self.repo.APIKey().GetByID(ctx, id)
	if ent.IsNotFound(err) {
		return notFound()
	}
	if err != nil {
		return err
	}

	if key.UserID != requesterUserID {
		if err := self.requireSystemAdmin(ctx, requesterUserID); err != nil {
			return errdefs.MaskAsNotFound(err, "API key not found")
		}
	}

	return self.repo.APIKey().Delete(ctx, id)
}
