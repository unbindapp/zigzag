package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
)

type GithubRepositoryResponse struct {
	ID          uuid.UUID `json:"id" format:"uuid"`
	FullName    string    `json:"full_name" doc:"owner/name on GitHub"`
	LastEvent   string    `json:"last_event" doc:"Last event GitHub delivered for it, like pull_request.opened or push"`
	LastEventAt time.Time `json:"last_event_at"`
}

type GithubIntegrationResponse struct {
	ID             uuid.UUID                   `json:"id" format:"uuid"`
	WorkspaceID    uuid.UUID                   `json:"workspace_id" format:"uuid"`
	WebhookURL     string                      `json:"webhook_url" doc:"Payload URL for the GitHub webhook"`
	SecretPrefix   string                      `json:"secret_prefix" doc:"First characters of the webhook secret, to tell secrets apart"`
	CreatedBy      *UserSummary                `json:"created_by,omitempty" required:"false"`
	CreatedAt      time.Time                   `json:"created_at"`
	LastDeliveryAt *time.Time                  `json:"last_delivery_at" nullable:"true" doc:"When GitHub last delivered an event with a valid signature"`
	Repositories   []*GithubRepositoryResponse `json:"repositories" nullable:"false" doc:"Repositories that delivered events, most recent first"`
}

func TransformGithubIntegration(entity *ent.GithubIntegration, webhookURL string) *GithubIntegrationResponse {
	repositories := make([]*GithubRepositoryResponse, len(entity.Edges.Repositories))
	for i, repository := range entity.Edges.Repositories {
		repositories[i] = &GithubRepositoryResponse{
			ID:          repository.ID,
			FullName:    repository.FullName,
			LastEvent:   repository.LastEvent,
			LastEventAt: repository.LastEventAt,
		}
	}
	return &GithubIntegrationResponse{
		ID:             entity.ID,
		WorkspaceID:    entity.WorkspaceID,
		WebhookURL:     webhookURL,
		SecretPrefix:   entity.SecretPrefix,
		CreatedBy:      TransformUserSummary(entity.Edges.CreatedBy),
		CreatedAt:      entity.CreatedAt,
		LastDeliveryAt: entity.LastDeliveryAt,
		Repositories:   repositories,
	}
}

type GithubIntegrationStatusResponse struct {
	Connected   bool                       `json:"connected" doc:"Whether the workspace has a GitHub webhook secret"`
	WebhookURL  string                     `json:"webhook_url" doc:"Payload URL for the GitHub webhook"`
	Integration *GithubIntegrationResponse `json:"integration,omitempty" required:"false" doc:"Present when connected"`
}

type GithubIntegrationSecretResponse struct {
	GithubIntegrationResponse
	Secret string `json:"secret" doc:"Webhook secret to paste into GitHub. Returned only here, store it now"`
}

type GithubDeliveryResponse struct {
	Event   string `json:"event" doc:"The X-GitHub-Event header"`
	Handled bool   `json:"handled" doc:"False when Zigzag ignores this event or action"`
}
