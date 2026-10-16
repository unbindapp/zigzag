package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
)

type WebhookResponse struct {
	ID          uuid.UUID  `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID  `json:"workspace_id" format:"uuid"`
	TeamID      *uuid.UUID `json:"team_id" nullable:"true" format:"uuid" doc:"Only events of this team are sent; null for the whole workspace"`
	URL         string     `json:"url"`
	Label       string     `json:"label"`
	Events      []string   `json:"events" nullable:"false"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func TransformWebhook(entity *ent.Webhook) *WebhookResponse {
	if entity == nil {
		return nil
	}
	events := entity.Events
	if events == nil {
		events = []string{}
	}
	return &WebhookResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		TeamID:      entity.TeamID,
		URL:         entity.URL,
		Label:       entity.Label,
		Events:      events,
		Enabled:     entity.Enabled,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

func TransformWebhooks(entities []*ent.Webhook) []*WebhookResponse {
	responses := make([]*WebhookResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformWebhook(entity)
	}
	return responses
}

type WebhookCreatedResponse struct {
	WebhookResponse
	Secret string `json:"secret" doc:"Signing secret for X-Zigzag-Signature. Returned only here, store it now"`
}

type WebhookTestResponse struct {
	Delivered  bool   `json:"delivered"`
	StatusCode int    `json:"status_code" doc:"HTTP status the endpoint answered, 0 when it could not be reached"`
	Error      string `json:"error" doc:"Why the delivery failed, empty on success"`
}
