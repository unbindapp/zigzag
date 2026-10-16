package cycles_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	cycle_service "github.com/unbindapp/zigzag-api/internal/services/cycle"
)

type ListInput struct {
	server.BaseAuthInput
	cycle_service.ListCyclesInput
}

type CycleListResponse struct {
	Body struct {
		Data []*models.CycleResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*CycleListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	cycles, err := self.srv.CycleService.List(ctx, user.ID, &input.ListCyclesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &CycleListResponse{}
	resp.Body.Data = cycles
	return resp, nil
}

type CycleResponse struct {
	Body struct {
		Data *models.CycleResponse `json:"data" nullable:"false"`
	}
}

func cycleResponse(cycle *models.CycleResponse) *CycleResponse {
	resp := &CycleResponse{}
	resp.Body.Data = cycle
	return resp
}

type GetInput struct {
	server.BaseAuthInput
	cycle_service.GetCycleInput
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*CycleResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	cycle, err := self.srv.CycleService.Get(ctx, user.ID, &input.GetCycleInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return cycleResponse(cycle), nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *cycle_service.CreateCycleInput `required:"true"`
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*CycleResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	cycle, err := self.srv.CycleService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return cycleResponse(cycle), nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *cycle_service.UpdateCycleInput `required:"true"`
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*CycleResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	cycle, err := self.srv.CycleService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return cycleResponse(cycle), nil
}

type RefInput struct {
	server.BaseAuthInput
	Body *cycle_service.CycleRefInput `required:"true"`
}

func (self *HandlerGroup) Archive(ctx context.Context, input *RefInput) (*CycleResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	cycle, err := self.srv.CycleService.Archive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return cycleResponse(cycle), nil
}

type DeletedResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Delete(ctx context.Context, input *RefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.CycleService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.CycleID, Deleted: true}
	return resp, nil
}
