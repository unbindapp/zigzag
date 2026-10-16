package webhook_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/webhook"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateWebhookInput struct {
	WorkspaceID uuid.UUID
	TeamID      *uuid.UUID
	URL         string
	Secret      string
	Enabled     bool
	Events      []string
	Label       string
}

func (self *WebhookRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateWebhookInput) (*ent.Webhook, error) {
	return self.base.Client(tx).Webhook.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetNillableTeamID(input.TeamID).
		SetURL(input.URL).
		SetSecret(input.Secret).
		SetEnabled(input.Enabled).
		SetEvents(input.Events).
		SetLabel(input.Label).
		Save(ctx)
}

func (self *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Webhook, error) {
	return self.base.DB.Webhook.Get(ctx, id)
}

func (self *WebhookRepository) ResolveWorkspaceOfWebhook(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Webhook.Query().Where(webhook.ID(id)).Select(webhook.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
