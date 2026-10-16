package webhook_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/issue"
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

func (self *WebhookRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]*ent.Webhook, error) {
	return self.base.DB.Webhook.Query().
		Where(webhook.WorkspaceID(workspaceID)).
		Order(ent.Asc(webhook.FieldCreatedAt)).
		All(ctx)
}

func (self *WebhookRepository) ListEnabled(ctx context.Context, workspaceID uuid.UUID) ([]*ent.Webhook, error) {
	return self.base.DB.Webhook.Query().
		Where(webhook.WorkspaceID(workspaceID), webhook.Enabled(true)).
		All(ctx)
}

type UpdateWebhookInput struct {
	URL       *string
	Label     *string
	Events    []string
	Enabled   *bool
	TeamID    *uuid.UUID
	ClearTeam bool
}

func (self *WebhookRepository) Update(ctx context.Context, id uuid.UUID, input *UpdateWebhookInput) (*ent.Webhook, error) {
	update := self.base.DB.Webhook.UpdateOneID(id).
		SetNillableURL(input.URL).
		SetNillableLabel(input.Label).
		SetNillableEnabled(input.Enabled).
		SetNillableTeamID(input.TeamID)
	if input.Events != nil {
		update.SetEvents(input.Events)
	}
	if input.ClearTeam {
		update.ClearTeamID()
	}
	return update.Save(ctx)
}

func (self *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Webhook.DeleteOneID(id).Exec(ctx)
}

func (self *WebhookRepository) TeamOfIssue(ctx context.Context, issueID uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Issue.Query().Where(issue.ID(issueID)).Select(issue.FieldTeamID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.TeamID, nil
}
