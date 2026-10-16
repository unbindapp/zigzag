package webhook_service

import (
	"context"
	"net"
	"net/url"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	webhook_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/webhook"
)

const (
	urlMaxLength   = 2048
	labelMaxLength = 100
	secretPrefix   = "whsec_"
)

const eventsDoc = "Events to send: issue.created, issue.updated, issue.deleted, comment.created, project.updated, cycle.updated"

type ListWebhooksInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose webhooks to list"`
}

type CreateWebhookInput struct {
	WorkspaceID uuid.UUID  `json:"workspace_id" required:"true" format:"uuid" doc:"Workspace the webhook belongs to"`
	TeamID      *uuid.UUID `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Only send events of this team; omit for the whole workspace"`
	URL         string     `json:"url" required:"true" maxLength:"2048" doc:"Public http or https endpoint that receives a JSON POST per event"`
	Label       string     `json:"label" required:"true" maxLength:"100" doc:"Name shown in settings, like Slack bridge"`
	Events      []string   `json:"events" required:"true" nullable:"false" minItems:"1" enum:"issue.created,issue.updated,issue.deleted,comment.created,project.updated,cycle.updated" doc:"Events to send: issue.created, issue.updated, issue.deleted, comment.created, project.updated, cycle.updated"`
	Enabled     bool       `json:"enabled" required:"true" doc:"Whether events are sent"`
}

type UpdateWebhookInput struct {
	WebhookID uuid.UUID  `json:"webhook_id" required:"true" format:"uuid" doc:"Webhook to change"`
	TeamID    *uuid.UUID `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Only send events of this team; the nil uuid 00000000-0000-0000-0000-000000000000 sends the whole workspace"`
	URL       *string    `json:"url,omitempty" required:"false" maxLength:"2048" doc:"New endpoint"`
	Label     *string    `json:"label,omitempty" required:"false" maxLength:"100" doc:"New name"`
	Events    []string   `json:"events,omitempty" required:"false" nullable:"false" minItems:"1" enum:"issue.created,issue.updated,issue.deleted,comment.created,project.updated,cycle.updated" doc:"Replaces the event list"`
	Enabled   *bool      `json:"enabled,omitempty" required:"false" doc:"Turn delivery on or off"`
}

type WebhookRefInput struct {
	WebhookID uuid.UUID `json:"webhook_id" required:"true" format:"uuid" doc:"Webhook id from list-webhooks"`
}

func (self *WebhookService) List(ctx context.Context, userID uuid.UUID, input *ListWebhooksInput) ([]*models.WebhookResponse, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	hooks, err := self.repo.Webhook().List(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return models.TransformWebhooks(hooks), nil
}

func (self *WebhookService) Create(ctx context.Context, userID uuid.UUID, input *CreateWebhookInput) (*models.WebhookCreatedResponse, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	target, err := self.validateURL(input.URL)
	if err != nil {
		return nil, err
	}
	label, err := validateLabel(input.Label)
	if err != nil {
		return nil, err
	}
	events, err := validateEvents(input.Events)
	if err != nil {
		return nil, err
	}
	if input.TeamID != nil {
		if err := self.ensureTeamInWorkspace(ctx, *input.TeamID, input.WorkspaceID); err != nil {
			return nil, err
		}
	}
	secret, err := utils.RandomHex(24)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to generate a webhook secret")
	}
	secret = secretPrefix + secret

	hook, err := self.repo.Webhook().Create(ctx, nil, &webhook_repo.CreateWebhookInput{
		WorkspaceID: input.WorkspaceID,
		TeamID:      input.TeamID,
		URL:         target,
		Secret:      secret,
		Enabled:     input.Enabled,
		Events:      events,
		Label:       label,
	})
	if err != nil {
		return nil, err
	}
	return &models.WebhookCreatedResponse{WebhookResponse: *models.TransformWebhook(hook), Secret: secret}, nil
}

func (self *WebhookService) Update(ctx context.Context, userID uuid.UUID, input *UpdateWebhookInput) (*models.WebhookResponse, error) {
	hook, err := self.adminWebhook(ctx, userID, input.WebhookID)
	if err != nil {
		return nil, err
	}
	update := &webhook_repo.UpdateWebhookInput{Enabled: input.Enabled}
	if input.URL != nil {
		target, err := self.validateURL(*input.URL)
		if err != nil {
			return nil, err
		}
		update.URL = &target
	}
	if input.Label != nil {
		label, err := validateLabel(*input.Label)
		if err != nil {
			return nil, err
		}
		update.Label = &label
	}
	if input.Events != nil {
		events, err := validateEvents(input.Events)
		if err != nil {
			return nil, err
		}
		update.Events = events
	}
	if input.TeamID != nil && *input.TeamID == uuid.Nil {
		update.ClearTeam = true
	}
	if input.TeamID != nil && *input.TeamID != uuid.Nil {
		if err := self.ensureTeamInWorkspace(ctx, *input.TeamID, hook.WorkspaceID); err != nil {
			return nil, err
		}
		update.TeamID = input.TeamID
	}

	updated, err := self.repo.Webhook().Update(ctx, hook.ID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformWebhook(updated), nil
}

func (self *WebhookService) Delete(ctx context.Context, userID uuid.UUID, input *WebhookRefInput) error {
	hook, err := self.adminWebhook(ctx, userID, input.WebhookID)
	if err != nil {
		return err
	}
	return self.repo.Webhook().Delete(ctx, hook.ID)
}

func (self *WebhookService) Test(ctx context.Context, userID uuid.UUID, input *WebhookRefInput) (*models.WebhookTestResponse, error) {
	hook, err := self.adminWebhook(ctx, userID, input.WebhookID)
	if err != nil {
		return nil, err
	}
	delivery, err := self.newDelivery(hook.WorkspaceID, "webhook.test", map[string]any{"webhook_id": hook.ID, "label": hook.Label})
	if err != nil {
		return nil, err
	}
	status, err := self.post(ctx, hook, delivery)
	if err != nil {
		return &models.WebhookTestResponse{Delivered: false, StatusCode: status, Error: err.Error()}, nil
	}
	return &models.WebhookTestResponse{Delivered: true, StatusCode: status}, nil
}

func (self *WebhookService) ensureTeamInWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID) error {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) {
		return notFound
	}
	if err != nil {
		return err
	}
	if team.WorkspaceID != workspaceID {
		return notFound
	}
	return nil
}

func (self *WebhookService) validateURL(raw string) (string, error) {
	invalid := func(message string) (string, error) {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, message)
	}
	target := strings.TrimSpace(raw)
	if target == "" || len(target) > urlMaxLength {
		return invalid("url is required and at most 2048 characters")
	}
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return invalid("url must be an absolute http or https URL")
	}
	if parsed.User != nil {
		return invalid("url must not carry credentials")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return invalid("url must point to a public address")
	}
	if ip := net.ParseIP(host); ip != nil && !self.policy(ip) {
		return invalid("url must point to a public address")
	}
	return target, nil
}

func validateLabel(raw string) (string, error) {
	label := strings.TrimSpace(raw)
	if len([]rune(label)) > labelMaxLength {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "label is at most 100 characters")
	}
	return label, nil
}

func validateEvents(events []string) ([]string, error) {
	if len(events) == 0 {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "events needs at least one event. "+eventsDoc)
	}
	picked := make([]string, 0, len(events))
	for _, event := range events {
		if !slices.Contains(Events, event) {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "unknown event "+event+". "+eventsDoc)
		}
		if !slices.Contains(picked, event) {
			picked = append(picked, event)
		}
	}
	return picked, nil
}
