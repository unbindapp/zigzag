package labels_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	label_service "github.com/unbindapp/zigzag-api/internal/services/label"
)

type ListInput struct {
	server.BaseAuthInput
	label_service.ListLabelsInput
}

type LabelListResponse struct {
	Body struct {
		Data []*models.LabelTreeResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*LabelListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	labels, err := self.srv.LabelService.List(ctx, user.ID, &input.ListLabelsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &LabelListResponse{}
	resp.Body.Data = labels
	return resp, nil
}

type LabelResponse struct {
	Body struct {
		Data *models.LabelResponse `json:"data" nullable:"false"`
	}
}

func labelResponse(label *models.LabelResponse) *LabelResponse {
	resp := &LabelResponse{}
	resp.Body.Data = label
	return resp
}

type CreateInput struct {
	server.BaseAuthInput
	Body *label_service.CreateLabelInput `required:"true"`
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*LabelResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	label, err := self.srv.LabelService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return labelResponse(label), nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *label_service.UpdateLabelInput `required:"true"`
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*LabelResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	label, err := self.srv.LabelService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return labelResponse(label), nil
}

type RefInput struct {
	server.BaseAuthInput
	Body *label_service.LabelRefInput `required:"true"`
}

func (self *HandlerGroup) Archive(ctx context.Context, input *RefInput) (*LabelResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	label, err := self.srv.LabelService.Archive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return labelResponse(label), nil
}

func (self *HandlerGroup) Unarchive(ctx context.Context, input *RefInput) (*LabelResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	label, err := self.srv.LabelService.Unarchive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return labelResponse(label), nil
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
	if err := self.srv.LabelService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.LabelID, Deleted: true}
	return resp, nil
}
