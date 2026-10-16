package webhooks_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	webhook_service "github.com/unbindapp/zigzag-api/internal/services/webhook"
)

type ListInput struct {
	server.BaseAuthInput
	webhook_service.ListWebhooksInput
}

type WebhookListResponse struct {
	Body struct {
		Data []*models.WebhookResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*WebhookListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	hooks, err := self.srv.WebhookService.List(ctx, user.ID, &input.ListWebhooksInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WebhookListResponse{}
	resp.Body.Data = hooks
	return resp, nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *webhook_service.CreateWebhookInput
}

type WebhookCreatedResult struct {
	Body struct {
		Data *models.WebhookCreatedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*WebhookCreatedResult, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	hook, err := self.srv.WebhookService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WebhookCreatedResult{}
	resp.Body.Data = hook
	return resp, nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *webhook_service.UpdateWebhookInput
}

type WebhookResult struct {
	Body struct {
		Data *models.WebhookResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*WebhookResult, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	hook, err := self.srv.WebhookService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WebhookResult{}
	resp.Body.Data = hook
	return resp, nil
}

type RefInput struct {
	server.BaseAuthInput
	Body *webhook_service.WebhookRefInput
}

type DeletedResponse struct {
	Body struct {
		Data server.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Delete(ctx context.Context, input *RefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.WebhookService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.WebhookID.String(), Deleted: true}
	return resp, nil
}

type WebhookTestResult struct {
	Body struct {
		Data *models.WebhookTestResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Test(ctx context.Context, input *RefInput) (*WebhookTestResult, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := self.srv.WebhookService.Test(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WebhookTestResult{}
	resp.Body.Data = result
	return resp, nil
}
