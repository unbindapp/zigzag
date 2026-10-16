package connectedapps_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
)

type GetClientInput struct {
	server.BaseAuthInput
	models.ConnectedAppClientInput
}

type GetClientResponse struct {
	Body struct {
		Data *models.ConnectedAppClientResponse `json:"data"`
	}
}

func (self *HandlerGroup) GetClient(ctx context.Context, input *GetClientInput) (*GetClientResponse, error) {
	if _, err := self.srv.AuthenticatedUser(ctx); err != nil {
		return nil, err
	}

	client, err := self.srv.OAuthServerService.ClientInfo(ctx, &input.ConnectedAppClientInput)
	if err != nil {
		return nil, oapi.MapError(asInputError(err))
	}

	resp := &GetClientResponse{}
	resp.Body.Data = client
	return resp, nil
}
