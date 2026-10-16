package github_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	"github.com/unbindapp/zigzag-api/internal/models"
	github_repo "github.com/unbindapp/zigzag-api/internal/repositories/github"
)

const (
	secretPrefix       = "zzgh_"
	secretBytes        = 24
	secretPrefixLength = 12
)

type GetGithubIntegrationInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose GitHub connection to read"`
}

type GithubIntegrationRefInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose GitHub connection to change"`
}

func notConnected() error {
	return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "GitHub is not connected to this workspace")
}

func (self *GithubService) Get(ctx context.Context, userID uuid.UUID, input *GetGithubIntegrationInput) (*models.GithubIntegrationStatusResponse, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	status := &models.GithubIntegrationStatusResponse{WebhookURL: self.webhookURL(input.WorkspaceID)}
	integration, err := self.repo.Github().GetByWorkspace(ctx, input.WorkspaceID)
	if ent.IsNotFound(err) {
		return status, nil
	}
	if err != nil {
		return nil, err
	}
	status.Connected = true
	status.Integration = models.TransformGithubIntegration(integration, status.WebhookURL)
	return status, nil
}

func (self *GithubService) Create(ctx context.Context, userID uuid.UUID, input *GithubIntegrationRefInput) (*models.GithubIntegrationSecretResponse, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	_, err := self.repo.Github().GetByWorkspace(ctx, input.WorkspaceID)
	if err == nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, "GitHub is already connected to this workspace; regenerate the secret instead")
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	secret, sealed, err := self.newSecret()
	if err != nil {
		return nil, err
	}
	created, err := self.repo.Github().Create(ctx, &github_repo.CreateIntegrationInput{
		WorkspaceID:     input.WorkspaceID,
		CreatedByID:     userID,
		SecretEncrypted: sealed,
		SecretPrefix:    secret[:secretPrefixLength],
	})
	if err != nil {
		return nil, err
	}
	return self.withSecret(created, secret), nil
}

func (self *GithubService) RegenerateSecret(ctx context.Context, userID uuid.UUID, input *GithubIntegrationRefInput) (*models.GithubIntegrationSecretResponse, error) {
	integration, err := self.adminIntegration(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	secret, sealed, err := self.newSecret()
	if err != nil {
		return nil, err
	}
	updated, err := self.repo.Github().UpdateSecret(ctx, integration.ID, sealed, secret[:secretPrefixLength])
	if err != nil {
		return nil, err
	}
	return self.withSecret(updated, secret), nil
}

func (self *GithubService) Delete(ctx context.Context, userID uuid.UUID, input *GithubIntegrationRefInput) (uuid.UUID, error) {
	integration, err := self.adminIntegration(ctx, userID, input.WorkspaceID)
	if err != nil {
		return uuid.Nil, err
	}
	return integration.ID, self.repo.Github().Delete(ctx, integration.ID)
}

func (self *GithubService) adminIntegration(ctx context.Context, userID, workspaceID uuid.UUID) (*ent.GithubIntegration, error) {
	if err := self.checkAdmin(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	integration, err := self.repo.Github().GetByWorkspace(ctx, workspaceID)
	if ent.IsNotFound(err) {
		return nil, notConnected()
	}
	return integration, err
}

func (self *GithubService) newSecret() (string, []byte, error) {
	random, err := utils.RandomHex(secretBytes)
	if err != nil {
		return "", nil, errdefs.NewInternalError(err, "Failed to generate a webhook secret")
	}
	secret := secretPrefix + random
	sealed, err := self.encryptor.Encrypt([]byte(secret))
	if err != nil {
		return "", nil, errdefs.NewInternalError(err, "Failed to encrypt the webhook secret")
	}
	return secret, sealed, nil
}

func (self *GithubService) withSecret(integration *ent.GithubIntegration, secret string) *models.GithubIntegrationSecretResponse {
	return &models.GithubIntegrationSecretResponse{
		GithubIntegrationResponse: *models.TransformGithubIntegration(integration, self.webhookURL(integration.WorkspaceID)),
		Secret:                    secret,
	}
}
