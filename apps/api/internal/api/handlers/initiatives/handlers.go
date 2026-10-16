package initiatives_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	initiative_service "github.com/unbindapp/zigzag-api/internal/services/initiative"
)

type InitiativeListResponse struct {
	Body struct {
		Data []*models.InitiativeResponse `json:"data" nullable:"false"`
	}
}

type InitiativeResponse struct {
	Body struct {
		Data *models.InitiativeResponse `json:"data" nullable:"false"`
	}
}

type DeletedResponse struct {
	Body struct {
		Data server.DeletedResponse `json:"data" nullable:"false"`
	}
}

type UpdateListResponse struct {
	Body struct {
		Data []*models.InitiativeUpdateResponse `json:"data" nullable:"false"`
	}
}

type UpdateResponse struct {
	Body struct {
		Data *models.InitiativeUpdateResponse `json:"data" nullable:"false"`
	}
}

type ListInput struct {
	server.BaseAuthInput
	initiative_service.ListInitiativesInput
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*InitiativeListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.List(ctx, user.ID, &input.ListInitiativesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type GetInput struct {
	server.BaseAuthInput
	initiative_service.GetInitiativeInput
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.Get(ctx, user.ID, &input.GetInitiativeInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *initiative_service.CreateInitiativeInput
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *initiative_service.UpdateInitiativeInput
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

type InitiativeRefBodyInput struct {
	server.BaseAuthInput
	Body *initiative_service.InitiativeRefInput
}

func (self *HandlerGroup) Delete(ctx context.Context, input *InitiativeRefBodyInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.InitiativeService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.InitiativeID.String(), Deleted: true}
	return resp, nil
}

func (self *HandlerGroup) Archive(ctx context.Context, input *InitiativeRefBodyInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.Archive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Unarchive(ctx context.Context, input *InitiativeRefBodyInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.Unarchive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

type InitiativeProjectBodyInput struct {
	server.BaseAuthInput
	Body *initiative_service.InitiativeProjectInput
}

func (self *HandlerGroup) AddProject(ctx context.Context, input *InitiativeProjectBodyInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.AddProject(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) RemoveProject(ctx context.Context, input *InitiativeProjectBodyInput) (*InitiativeResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.RemoveProject(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeResponse{}
	resp.Body.Data = data
	return resp, nil
}

type ListUpdatesInput struct {
	server.BaseAuthInput
	initiative_service.InitiativeQueryInput
}

func (self *HandlerGroup) ListUpdates(ctx context.Context, input *ListUpdatesInput) (*UpdateListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.ListUpdates(ctx, user.ID, &input.InitiativeQueryInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UpdateListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateUpdateInput struct {
	server.BaseAuthInput
	Body *initiative_service.CreateInitiativeUpdateInput
}

func (self *HandlerGroup) CreateUpdate(ctx context.Context, input *CreateUpdateInput) (*UpdateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.CreateUpdate(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UpdateResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DeleteUpdateInput struct {
	server.BaseAuthInput
	Body *initiative_service.InitiativeUpdateRefInput
}

func (self *HandlerGroup) DeleteUpdate(ctx context.Context, input *DeleteUpdateInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.InitiativeService.DeleteUpdate(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.UpdateID.String(), Deleted: true}
	return resp, nil
}
