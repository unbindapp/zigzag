package documents_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	document_service "github.com/unbindapp/zigzag-api/internal/services/document"
)

type DocumentListResponse struct {
	Body struct {
		Data []*models.DocumentResponse `json:"data" nullable:"false"`
	}
}

type DocumentResponse struct {
	Body struct {
		Data *models.DocumentResponse `json:"data" nullable:"false"`
	}
}

type DeletedResponse struct {
	Body struct {
		Data server.DeletedResponse `json:"data" nullable:"false"`
	}
}

type ListInput struct {
	server.BaseAuthInput
	document_service.ListDocumentsInput
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*DocumentListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.List(ctx, user.ID, &input.ListDocumentsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type GetInput struct {
	server.BaseAuthInput
	document_service.GetDocumentInput
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Get(ctx, user.ID, &input.GetDocumentInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *document_service.CreateDocumentInput
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *document_service.UpdateDocumentInput
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DocumentRefBodyInput struct {
	server.BaseAuthInput
	Body *document_service.DocumentRefInput
}

func (self *HandlerGroup) Delete(ctx context.Context, input *DocumentRefBodyInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.DocumentService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.DocumentID.String(), Deleted: true}
	return resp, nil
}

func (self *HandlerGroup) Archive(ctx context.Context, input *DocumentRefBodyInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Archive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Unarchive(ctx context.Context, input *DocumentRefBodyInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Unarchive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Trash(ctx context.Context, input *DocumentRefBodyInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Trash(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Restore(ctx context.Context, input *DocumentRefBodyInput) (*DocumentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.DocumentService.Restore(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DocumentResponse{}
	resp.Body.Data = data
	return resp, nil
}
