package projects_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	project_service "github.com/unbindapp/zigzag-api/internal/services/project"
)

type ProjectListResponse struct {
	Body struct {
		Data []*models.ProjectResponse `json:"data" nullable:"false"`
	}
}

type ProjectResponse struct {
	Body struct {
		Data *models.ProjectResponse `json:"data" nullable:"false"`
	}
}

type DeletedResponse struct {
	Body struct {
		Data server.DeletedResponse `json:"data" nullable:"false"`
	}
}

type StatusListResponse struct {
	Body struct {
		Data []*models.ProjectStatusResponse `json:"data" nullable:"false"`
	}
}

type StatusResponse struct {
	Body struct {
		Data *models.ProjectStatusResponse `json:"data" nullable:"false"`
	}
}

type MilestoneListResponse struct {
	Body struct {
		Data []*models.MilestoneResponse `json:"data" nullable:"false"`
	}
}

type MilestoneResponse struct {
	Body struct {
		Data *models.MilestoneResponse `json:"data" nullable:"false"`
	}
}

type UpdateListResponse struct {
	Body struct {
		Data []*models.ProjectUpdateResponse `json:"data" nullable:"false"`
	}
}

type UpdateResponse struct {
	Body struct {
		Data *models.ProjectUpdateResponse `json:"data" nullable:"false"`
	}
}

type LinkResponse struct {
	Body struct {
		Data *models.ProjectLinkResponse `json:"data" nullable:"false"`
	}
}

type ListInput struct {
	server.BaseAuthInput
	project_service.ListProjectsInput
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*ProjectListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.List(ctx, user.ID, &input.ListProjectsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type GetInput struct {
	server.BaseAuthInput
	project_service.GetProjectInput
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Get(ctx, user.ID, &input.GetProjectInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *project_service.CreateProjectInput
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *project_service.UpdateProjectInput
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

type ProjectRefBodyInput struct {
	server.BaseAuthInput
	Body *project_service.ProjectRefInput
}

func (self *HandlerGroup) Delete(ctx context.Context, input *ProjectRefBodyInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.ProjectService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.ProjectID.String(), Deleted: true}
	return resp, nil
}

func (self *HandlerGroup) Archive(ctx context.Context, input *ProjectRefBodyInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Archive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Unarchive(ctx context.Context, input *ProjectRefBodyInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Unarchive(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Trash(ctx context.Context, input *ProjectRefBodyInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Trash(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

func (self *HandlerGroup) Restore(ctx context.Context, input *ProjectRefBodyInput) (*ProjectResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.Restore(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ProjectResponse{}
	resp.Body.Data = data
	return resp, nil
}

type ListStatusesInput struct {
	server.BaseAuthInput
	project_service.ListProjectStatusesInput
}

func (self *HandlerGroup) ListStatuses(ctx context.Context, input *ListStatusesInput) (*StatusListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.ListStatuses(ctx, user.ID, &input.ListProjectStatusesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &StatusListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateStatusInput struct {
	server.BaseAuthInput
	Body *project_service.CreateProjectStatusInput
}

func (self *HandlerGroup) CreateStatus(ctx context.Context, input *CreateStatusInput) (*StatusResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.CreateStatus(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &StatusResponse{}
	resp.Body.Data = data
	return resp, nil
}

type UpdateStatusInput struct {
	server.BaseAuthInput
	Body *project_service.UpdateProjectStatusInput
}

func (self *HandlerGroup) UpdateStatus(ctx context.Context, input *UpdateStatusInput) (*StatusResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.UpdateStatus(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &StatusResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DeleteStatusInput struct {
	server.BaseAuthInput
	Body *project_service.DeleteProjectStatusInput
}

func (self *HandlerGroup) DeleteStatus(ctx context.Context, input *DeleteStatusInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.ProjectService.DeleteStatus(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.StatusID.String(), Deleted: true}
	return resp, nil
}

type ProjectQueryInput struct {
	server.BaseAuthInput
	project_service.ProjectQueryInput
}

func (self *HandlerGroup) ListMilestones(ctx context.Context, input *ProjectQueryInput) (*MilestoneListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.ListMilestones(ctx, user.ID, &input.ProjectQueryInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &MilestoneListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateMilestoneInput struct {
	server.BaseAuthInput
	Body *project_service.CreateMilestoneInput
}

func (self *HandlerGroup) CreateMilestone(ctx context.Context, input *CreateMilestoneInput) (*MilestoneResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.CreateMilestone(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &MilestoneResponse{}
	resp.Body.Data = data
	return resp, nil
}

type UpdateMilestoneInput struct {
	server.BaseAuthInput
	Body *project_service.UpdateMilestoneInput
}

func (self *HandlerGroup) UpdateMilestone(ctx context.Context, input *UpdateMilestoneInput) (*MilestoneResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.UpdateMilestone(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &MilestoneResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DeleteMilestoneInput struct {
	server.BaseAuthInput
	Body *project_service.MilestoneRefInput
}

func (self *HandlerGroup) DeleteMilestone(ctx context.Context, input *DeleteMilestoneInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.ProjectService.DeleteMilestone(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.MilestoneID.String(), Deleted: true}
	return resp, nil
}

func (self *HandlerGroup) ListUpdates(ctx context.Context, input *ProjectQueryInput) (*UpdateListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.ListUpdates(ctx, user.ID, &input.ProjectQueryInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UpdateListResponse{}
	resp.Body.Data = data
	return resp, nil
}

type CreateUpdateInput struct {
	server.BaseAuthInput
	Body *project_service.CreateProjectUpdateInput
}

func (self *HandlerGroup) CreateUpdate(ctx context.Context, input *CreateUpdateInput) (*UpdateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.CreateUpdate(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UpdateResponse{}
	resp.Body.Data = data
	return resp, nil
}

type EditUpdateInput struct {
	server.BaseAuthInput
	Body *project_service.EditProjectUpdateInput
}

func (self *HandlerGroup) EditUpdate(ctx context.Context, input *EditUpdateInput) (*UpdateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.EditUpdate(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &UpdateResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DeleteUpdateInput struct {
	server.BaseAuthInput
	Body *project_service.ProjectUpdateRefInput
}

func (self *HandlerGroup) DeleteUpdate(ctx context.Context, input *DeleteUpdateInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.ProjectService.DeleteUpdate(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.UpdateID.String(), Deleted: true}
	return resp, nil
}

type CreateLinkInput struct {
	server.BaseAuthInput
	Body *project_service.CreateProjectLinkInput
}

func (self *HandlerGroup) CreateLink(ctx context.Context, input *CreateLinkInput) (*LinkResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.CreateLink(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &LinkResponse{}
	resp.Body.Data = data
	return resp, nil
}

type UpdateLinkInput struct {
	server.BaseAuthInput
	Body *project_service.UpdateProjectLinkInput
}

func (self *HandlerGroup) UpdateLink(ctx context.Context, input *UpdateLinkInput) (*LinkResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	data, err := self.srv.ProjectService.UpdateLink(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &LinkResponse{}
	resp.Body.Data = data
	return resp, nil
}

type DeleteLinkInput struct {
	server.BaseAuthInput
	Body *project_service.ProjectLinkRefInput
}

func (self *HandlerGroup) DeleteLink(ctx context.Context, input *DeleteLinkInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.ProjectService.DeleteLink(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.LinkID.String(), Deleted: true}
	return resp, nil
}
