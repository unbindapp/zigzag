package favorites_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	favorite_service "github.com/zigzagapp-org/zigzag-api/internal/services/favorite"
)

type ListInput struct {
	server.BaseAuthInput
	favorite_service.ListFavoritesInput
}

type FavoriteListResponse struct {
	Body struct {
		Data []*models.FavoriteResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*FavoriteListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	favorites, err := self.srv.FavoriteService.List(ctx, user.ID, &input.ListFavoritesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &FavoriteListResponse{}
	resp.Body.Data = favorites
	return resp, nil
}

type TargetInput struct {
	server.BaseAuthInput
	Body *favorite_service.FavoriteTargetInput `required:"true"`
}

type FavoriteResponse struct {
	Body struct {
		Data *models.FavoriteResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Add(ctx context.Context, input *TargetInput) (*FavoriteResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	favorite, err := self.srv.FavoriteService.Add(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &FavoriteResponse{}
	resp.Body.Data = favorite
	return resp, nil
}

type DeletedResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Remove(ctx context.Context, input *TargetInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.FavoriteService.Remove(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.EntityID, Deleted: true}
	return resp, nil
}

type ReorderInput struct {
	server.BaseAuthInput
	Body *favorite_service.ReorderFavoritesInput `required:"true"`
}

func (self *HandlerGroup) Reorder(ctx context.Context, input *ReorderInput) (*FavoriteListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	favorites, err := self.srv.FavoriteService.Reorder(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &FavoriteListResponse{}
	resp.Body.Data = favorites
	return resp, nil
}
