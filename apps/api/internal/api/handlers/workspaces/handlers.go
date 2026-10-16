package workspaces_handler

import (
	"context"

	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/models"
	workspace_service "github.com/unbindapp/zigzag-api/internal/services/workspace"
)

type WorkspaceListResponse struct {
	Body struct {
		Data []*models.WorkspaceResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, _ *server.BaseAuthInput) (*WorkspaceListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	workspaces, err := self.srv.WorkspaceService.List(ctx, user.ID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WorkspaceListResponse{}
	resp.Body.Data = workspaces
	return resp, nil
}

type GetInput struct {
	server.BaseAuthInput
	workspace_service.GetWorkspaceInput
}

type WorkspaceResponse struct {
	Body struct {
		Data *models.WorkspaceResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*WorkspaceResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	workspace, err := self.srv.WorkspaceService.Get(ctx, user.ID, &input.GetWorkspaceInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WorkspaceResponse{}
	resp.Body.Data = workspace
	return resp, nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *workspace_service.CreateWorkspaceInput
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*WorkspaceResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	workspace, err := self.srv.WorkspaceService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WorkspaceResponse{}
	resp.Body.Data = workspace
	return resp, nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *workspace_service.UpdateWorkspaceInput
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*WorkspaceResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	workspace, err := self.srv.WorkspaceService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &WorkspaceResponse{}
	resp.Body.Data = workspace
	return resp, nil
}

type DeleteInput struct {
	server.BaseAuthInput
	Body *workspace_service.DeleteWorkspaceInput
}

type DeletedResponse struct {
	Body struct {
		Data server.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Delete(ctx context.Context, input *DeleteInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.WorkspaceService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.WorkspaceID.String(), Deleted: true}
	return resp, nil
}

type ListMembersInput struct {
	server.BaseAuthInput
	workspace_service.ListMembersInput
}

type MemberListResponse struct {
	Body struct {
		Data []*models.WorkspaceMemberResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ListMembers(ctx context.Context, input *ListMembersInput) (*MemberListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	members, err := self.srv.WorkspaceService.ListMembers(ctx, user.ID, input.WorkspaceID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &MemberListResponse{}
	resp.Body.Data = members
	return resp, nil
}

type AddMemberInput struct {
	server.BaseAuthInput
	Body *workspace_service.AddMemberInput
}

type AddMemberResponse struct {
	Body struct {
		Data *models.AddWorkspaceMemberResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) AddMember(ctx context.Context, input *AddMemberInput) (*AddMemberResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	added, err := self.srv.WorkspaceService.AddMember(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &AddMemberResponse{}
	resp.Body.Data = added
	return resp, nil
}

type UpdateMemberInput struct {
	server.BaseAuthInput
	Body *workspace_service.UpdateMemberInput
}

type MemberResponse struct {
	Body struct {
		Data *models.WorkspaceMemberResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) UpdateMember(ctx context.Context, input *UpdateMemberInput) (*MemberResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	member, err := self.srv.WorkspaceService.UpdateMember(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &MemberResponse{}
	resp.Body.Data = member
	return resp, nil
}

type RemoveMemberInput struct {
	server.BaseAuthInput
	Body *workspace_service.RemoveMemberInput
}

type SuccessResponse struct {
	Body struct {
		Data server.SuccessResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) RemoveMember(ctx context.Context, input *RemoveMemberInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.WorkspaceService.RemoveMember(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}

type ListInvitationsInput struct {
	server.BaseAuthInput
	workspace_service.ListInvitationsInput
}

type InvitationListResponse struct {
	Body struct {
		Data []*models.InvitationResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ListInvitations(ctx context.Context, input *ListInvitationsInput) (*InvitationListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	invitations, err := self.srv.WorkspaceService.ListInvitations(ctx, user.ID, input.WorkspaceID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InvitationListResponse{}
	resp.Body.Data = invitations
	return resp, nil
}

type CreateInvitationInput struct {
	server.BaseAuthInput
	Body *workspace_service.CreateInvitationInput
}

type InvitationResponse struct {
	Body struct {
		Data *models.InvitationResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) CreateInvitation(ctx context.Context, input *CreateInvitationInput) (*InvitationResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	invitation, err := self.srv.WorkspaceService.CreateInvitation(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InvitationResponse{}
	resp.Body.Data = invitation
	return resp, nil
}

type InvitationRefInput struct {
	server.BaseAuthInput
	Body *workspace_service.InvitationRefInput
}

func (self *HandlerGroup) RevokeInvitation(ctx context.Context, input *InvitationRefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.WorkspaceService.RevokeInvitation(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = server.DeletedResponse{ID: input.Body.InvitationID.String(), Deleted: true}
	return resp, nil
}

func (self *HandlerGroup) ResendInvitation(ctx context.Context, input *InvitationRefInput) (*InvitationResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	invitation, err := self.srv.WorkspaceService.ResendInvitation(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &InvitationResponse{}
	resp.Body.Data = invitation
	return resp, nil
}
