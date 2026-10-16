package integrations_handler

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	github_service "github.com/unbindapp/zigzag-api/internal/services/github"
)

type GetGithubInput struct {
	server.BaseAuthInput
	github_service.GetGithubIntegrationInput
}

type GetGithubResponse struct {
	Body struct {
		Data *models.GithubIntegrationStatusResponse `json:"data"`
	}
}

func (self *HandlerGroup) GetGithub(ctx context.Context, input *GetGithubInput) (*GetGithubResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	status, err := self.srv.GithubService.Get(ctx, user.ID, &input.GetGithubIntegrationInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GetGithubResponse{}
	resp.Body.Data = status
	return resp, nil
}

type GithubRefInput struct {
	server.BaseAuthInput
	Body *github_service.GithubIntegrationRefInput `required:"true"`
}

type GithubSecretResponse struct {
	Body struct {
		Data *models.GithubIntegrationSecretResponse `json:"data"`
	}
}

func (self *HandlerGroup) CreateGithub(ctx context.Context, input *GithubRefInput) (*GithubSecretResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	created, err := self.srv.GithubService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubSecretResponse{}
	resp.Body.Data = created
	return resp, nil
}

func (self *HandlerGroup) RegenerateGithubSecret(ctx context.Context, input *GithubRefInput) (*GithubSecretResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := self.srv.GithubService.RegenerateSecret(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubSecretResponse{}
	resp.Body.Data = updated
	return resp, nil
}

type DeleteGithubResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data"`
	}
}

func (self *HandlerGroup) DeleteGithub(ctx context.Context, input *GithubRefInput) (*DeleteGithubResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	id, err := self.srv.GithubService.Delete(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeleteGithubResponse{}
	resp.Body.Data = models.DeletedResponse{ID: id, Deleted: true}
	return resp, nil
}

type GithubWebhookInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace the delivery is for, part of the webhook URL shown in Zigzag"`
	Event       string    `header:"X-GitHub-Event" doc:"GitHub event name, like ping, pull_request or push"`
	Signature   string    `header:"X-Hub-Signature-256" doc:"sha256= followed by the hex HMAC-SHA256 of the raw body"`
	RawBody     []byte
}

type GithubWebhookResponse struct {
	Body struct {
		Data models.GithubDeliveryResponse `json:"data"`
	}
}

func (self *HandlerGroup) GithubWebhook(ctx context.Context, input *GithubWebhookInput) (*GithubWebhookResponse, error) {
	handled, err := self.srv.GithubService.HandleWebhook(ctx, &github_service.Delivery{
		WorkspaceID: input.WorkspaceID,
		Event:       input.Event,
		Signature:   input.Signature,
		Body:        input.RawBody,
	})
	if errors.Is(err, github_service.ErrInvalidSignature) {
		return nil, huma.Error401Unauthorized("Invalid signature")
	}
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &GithubWebhookResponse{}
	resp.Body.Data = models.GithubDeliveryResponse{Event: input.Event, Handled: handled}
	return resp, nil
}
