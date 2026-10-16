package github_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/githubrepository"
)

type CreateIntegrationInput struct {
	WorkspaceID     uuid.UUID
	CreatedByID     uuid.UUID
	SecretEncrypted []byte
	SecretPrefix    string
}

func (self *GithubRepository) Create(ctx context.Context, input *CreateIntegrationInput) (*ent.GithubIntegration, error) {
	_, err := self.base.DB.GithubIntegration.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetCreatedByID(input.CreatedByID).
		SetSecretEncrypted(input.SecretEncrypted).
		SetSecretPrefix(input.SecretPrefix).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.GetByWorkspace(ctx, input.WorkspaceID)
}

func (self *GithubRepository) UpdateSecret(ctx context.Context, id uuid.UUID, secretEncrypted []byte, secretPrefix string) (*ent.GithubIntegration, error) {
	updated, err := self.base.DB.GithubIntegration.UpdateOneID(id).
		SetSecretEncrypted(secretEncrypted).
		SetSecretPrefix(secretPrefix).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.GetByWorkspace(ctx, updated.WorkspaceID)
}

func (self *GithubRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.GithubIntegration.DeleteOneID(id).Exec(ctx)
}

func (self *GithubRepository) RecordDelivery(ctx context.Context, integrationID uuid.UUID, fullName, event string, at time.Time) error {
	if err := self.base.DB.GithubIntegration.UpdateOneID(integrationID).SetLastDeliveryAt(at).Exec(ctx); err != nil {
		return err
	}
	if fullName == "" {
		return nil
	}
	return self.base.DB.GithubRepository.Create().
		SetIntegrationID(integrationID).
		SetFullName(fullName).
		SetLastEvent(event).
		SetLastEventAt(at).
		OnConflictColumns(githubrepository.FieldIntegrationID, githubrepository.FieldFullName).
		Update(func(upsert *ent.GithubRepositoryUpsert) {
			upsert.UpdateLastEvent().UpdateLastEventAt().UpdateUpdatedAt()
		}).
		Exec(ctx)
}
