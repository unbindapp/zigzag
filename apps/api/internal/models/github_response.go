package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
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
	HasToken       bool                        `json:"has_token" doc:"Whether a GitHub token is stored for Issues sync"`
	TokenPrefix    string                      `json:"token_prefix" doc:"First characters of the GitHub token, empty without one"`
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
		HasToken:       len(entity.TokenEncrypted) > 0,
		TokenPrefix:    entity.TokenPrefix,
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

type GithubRepoSyncResponse struct {
	ID             uuid.UUID                    `json:"id" format:"uuid"`
	WorkspaceID    uuid.UUID                    `json:"workspace_id" format:"uuid"`
	FullName       string                       `json:"full_name" doc:"owner/name on GitHub"`
	HTMLURL        string                       `json:"html_url" doc:"The repository on GitHub"`
	Team           *TeamSummary                 `json:"team,omitempty" required:"false" doc:"Team new issues from GitHub land in"`
	Project        *ProjectSummary              `json:"project,omitempty" required:"false" doc:"Project new issues from GitHub join; missing when it was deleted"`
	SyncClosed     bool                         `json:"sync_closed" doc:"Whether imports include closed issues"`
	LinkedIssues   int                          `json:"linked_issues" doc:"Zigzag issues linked to an issue of this repository"`
	LastError      string                       `json:"last_error" doc:"Last error talking to GitHub, empty after a success"`
	LastSyncedAt   *time.Time                   `json:"last_synced_at" nullable:"true" doc:"Last successful import or change sent to GitHub"`
	BackfillStatus schema.GithubBackfillStatus  `json:"backfill_status" doc:"none, running, completed or failed"`
	BackfillReport *schema.GithubBackfillReport `json:"backfill_report,omitempty" required:"false" doc:"Counts of the running or last import"`
	CreatedAt      time.Time                    `json:"created_at"`
}

func TransformGithubRepoSync(entity *ent.GithubRepoSync, linkedIssues int) *GithubRepoSyncResponse {
	response := &GithubRepoSyncResponse{
		ID:             entity.ID,
		FullName:       entity.FullName,
		HTMLURL:        "https://github.com/" + entity.FullName,
		Project:        TransformProjectSummary(entity.Edges.Project),
		SyncClosed:     entity.SyncClosed,
		LinkedIssues:   linkedIssues,
		LastError:      entity.LastError,
		LastSyncedAt:   entity.LastSyncedAt,
		BackfillStatus: entity.BackfillStatus,
		BackfillReport: entity.BackfillReport,
		CreatedAt:      entity.CreatedAt,
	}
	if integration := entity.Edges.Integration; integration != nil {
		response.WorkspaceID = integration.WorkspaceID
	}
	if team := entity.Edges.Team; team != nil {
		response.Team = TransformTeamSummaries([]*ent.Team{team})[0]
	}
	return response
}

type GithubIssueLinkResponse struct {
	Repository string `json:"repository" doc:"owner/name on GitHub"`
	Number     int    `json:"number" doc:"Issue number on GitHub"`
	HTMLURL    string `json:"html_url" doc:"The issue on GitHub"`
}

func TransformGithubIssueLink(entity *ent.GithubIssueLink) *GithubIssueLinkResponse {
	if entity == nil {
		return nil
	}
	response := &GithubIssueLinkResponse{Number: entity.Number, HTMLURL: entity.HTMLURL}
	if entity.Edges.RepoSync != nil {
		response.Repository = entity.Edges.RepoSync.FullName
	}
	return response
}
