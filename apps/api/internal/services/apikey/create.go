package apikey_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/internal/auth"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	apikey_repo "github.com/unbindapp/zigzag-api/internal/repositories/apikey"
	"github.com/unbindapp/zigzag-api/internal/services/keyaccess"
)

func (self *APIKeyService) Create(ctx context.Context, requesterUserID uuid.UUID, input *models.APIKeyCreateInput) (*models.APIKeyCreatedResponse, error) {
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now()) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "expires_at must be in the future")
	}
	spec := keyaccess.Spec{Role: input.Role, FullAccess: input.FullAccess, WorkspaceIDs: input.WorkspaceIDs, Privileges: input.Privileges}
	if err := keyaccess.Validate(spec); err != nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, err.Error())
	}
	if err := keyaccess.RequesterHolds(ctx, self.repo.Permissions(), requesterUserID, spec); err != nil {
		return nil, err
	}

	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	takenNames, err := self.repo.APIKey().GetNamesByUser(ctx, nil, requesterUserID)
	if err != nil {
		return nil, err
	}
	if err := names.EnsureFree(name, takenNames, "API key", ""); err != nil {
		return nil, err
	}

	generated, err := auth.NewAPIKey()
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to generate the API key")
	}

	key, err := self.repo.APIKey().Create(ctx, &apikey_repo.CreateAPIKeyInput{
		UserID:       requesterUserID,
		Name:         name,
		TokenPrefix:  generated.Prefix,
		TokenHash:    generated.Hash,
		Role:         input.Role,
		FullAccess:   input.FullAccess,
		WorkspaceIDs: keyaccess.WorkspaceIDs(spec),
		Privileges:   keyaccess.Privileges(spec),
		ExpiresAt:    input.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &models.APIKeyCreatedResponse{
		APIKeyResponse: *models.TransformAPIKeyEntity(key),
		Token:          generated.Token,
	}, nil
}
