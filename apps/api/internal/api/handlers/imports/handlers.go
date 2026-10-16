package imports_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	import_service "github.com/zigzagapp-org/zigzag-api/internal/services/import"
)

type ImportLinearInput struct {
	server.BaseAuthInput
	Body *import_service.ImportLinearInput `required:"true"`
}

type ImportReportResponse struct {
	Body struct {
		Data *schema.ImportReport `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ImportLinear(ctx context.Context, input *ImportLinearInput) (*ImportReportResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	report, err := self.srv.ImportService.ImportLinear(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ImportReportResponse{}
	resp.Body.Data = report
	return resp, nil
}

type ListInput struct {
	server.BaseAuthInput
	import_service.ListImportsInput
}

type ImportListResponse struct {
	Body struct {
		Data []*models.ImportResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*ImportListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	imports, err := self.srv.ImportService.List(ctx, user.ID, &input.ListImportsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ImportListResponse{}
	resp.Body.Data = imports
	return resp, nil
}
