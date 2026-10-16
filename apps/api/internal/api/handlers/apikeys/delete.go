package apikeys_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
)

type DeleteAPIKeyInput struct {
	server.BaseAuthInput
	Body *models.APIKeyDeleteInput `required:"true"`
}

type DeleteAPIKeyResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data"`
	}
}

func (self *HandlerGroup) DeleteAPIKey(ctx context.Context, input *DeleteAPIKeyInput) (*DeleteAPIKeyResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := self.srv.APIKeyService.Delete(ctx, user.ID, input.Body.ID); err != nil {
		return nil, oapi.MapError(err)
	}

	resp := &DeleteAPIKeyResponse{}
	resp.Body.Data = models.DeletedResponse{
		ID:      input.Body.ID,
		Deleted: true,
	}
	return resp, nil
}
