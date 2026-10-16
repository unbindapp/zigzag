package views_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	view_service "github.com/zigzagapp-org/zigzag-api/internal/services/view"
)

type ListInput struct {
	server.BaseAuthInput
	view_service.ListViewsInput
}

type ViewListResponse struct {
	Body struct {
		Data []*models.ViewResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*ViewListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	views, err := self.srv.ViewService.List(ctx, user.ID, &input.ListViewsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ViewListResponse{}
	resp.Body.Data = views
	return resp, nil
}

type ViewResponse struct {
	Body struct {
		Data *models.ViewResponse `json:"data" nullable:"false"`
	}
}

func viewResponse(view *models.ViewResponse) *ViewResponse {
	resp := &ViewResponse{}
	resp.Body.Data = view
	return resp
}

type GetInput struct {
	server.BaseAuthInput
	view_service.GetViewInput
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*ViewResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	view, err := self.srv.ViewService.Get(ctx, user.ID, &input.GetViewInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return viewResponse(view), nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *view_service.CreateViewInput `required:"true"`
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*ViewResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	view, err := self.srv.ViewService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return viewResponse(view), nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *view_service.UpdateViewInput `required:"true"`
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*ViewResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	view, err := self.srv.ViewService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return viewResponse(view), nil
}

type DeleteInput struct {
	server.BaseAuthInput
	Body *view_service.DeleteViewInput `required:"true"`
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
	if err := self.srv.ViewService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.ViewID, Deleted: true}
	return resp, nil
}
