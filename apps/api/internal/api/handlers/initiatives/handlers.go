package initiatives_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	initiative_service "github.com/zigzagapp-org/zigzag-api/internal/services/initiative"
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
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

type InitiativeUpdateListResponse struct {
	Body struct {
		Data []*models.InitiativeUpdateResponse `json:"data" nullable:"false"`
	}
}

type InitiativeUpdateResponse struct {
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
	Body *initiative_service.CreateInitiativeInput `required:"true"`
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
	Body *initiative_service.UpdateInitiativeInput `required:"true"`
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
	Body *initiative_service.InitiativeRefInput `required:"true"`
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
	resp.Body.Data = models.DeletedResponse{ID: input.Body.InitiativeID, Deleted: true}
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
	Body *initiative_service.InitiativeProjectInput `required:"true"`
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

func (self *HandlerGroup) ListUpdates(ctx context.Context, input *ListUpdatesInput) (*InitiativeUpdateListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.ListUpdates(ctx, user.ID, &input.InitiativeQueryInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeUpdateListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateUpdateInput struct {
	server.BaseAuthInput
	Body *initiative_service.CreateInitiativeUpdateInput `required:"true"`
}

func (self *HandlerGroup) CreateUpdate(ctx context.Context, input *CreateUpdateInput) (*InitiativeUpdateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.InitiativeService.CreateUpdate(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InitiativeUpdateResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DeleteUpdateInput struct {
	server.BaseAuthInput
	Body *initiative_service.InitiativeUpdateRefInput `required:"true"`
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
	resp.Body.Data = models.DeletedResponse{ID: input.Body.UpdateID, Deleted: true}
	return resp, nil
}
