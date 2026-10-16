package templates_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	template_service "github.com/zigzagapp-org/zigzag-api/internal/services/template"
)

type ListInput struct {
	server.BaseAuthInput
	template_service.ListTemplatesInput
}

type TemplateListResponse struct {
	Body struct {
		Data []*models.TemplateResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*TemplateListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	templates, err := self.srv.TemplateService.List(ctx, user.ID, &input.ListTemplatesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TemplateListResponse{}
	resp.Body.Data = templates
	return resp, nil
}

type TemplateResponse struct {
	Body struct {
		Data *models.TemplateResponse `json:"data" nullable:"false"`
	}
}

func templateResponse(template *models.TemplateResponse) *TemplateResponse {
	resp := &TemplateResponse{}
	resp.Body.Data = template
	return resp
}

type GetInput struct {
	server.BaseAuthInput
	template_service.GetTemplateInput
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*TemplateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	template, err := self.srv.TemplateService.Get(ctx, user.ID, &input.GetTemplateInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return templateResponse(template), nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *template_service.CreateTemplateInput `required:"true"`
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*TemplateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	template, err := self.srv.TemplateService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return templateResponse(template), nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *template_service.UpdateTemplateInput `required:"true"`
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*TemplateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	template, err := self.srv.TemplateService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return templateResponse(template), nil
}

type DeleteInput struct {
	server.BaseAuthInput
	Body *template_service.DeleteTemplateInput `required:"true"`
}

type DeletedResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Delete(ctx context.Context, input *DeleteInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.TemplateService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.TemplateID, Deleted: true}
	return resp, nil
}
