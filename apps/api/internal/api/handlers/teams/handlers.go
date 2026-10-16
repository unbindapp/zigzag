package teams_handler

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	team_service "github.com/zigzagapp-org/zigzag-api/internal/services/team"
)

type ListInput struct {
	server.BaseAuthInput
	team_service.ListTeamsInput
}

type TeamListResponse struct {
	Body struct {
		Data []*models.TeamResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*TeamListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	teams, err := self.srv.TeamService.List(ctx, user.ID, &input.ListTeamsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TeamListResponse{}
	resp.Body.Data = teams
	return resp, nil
}

type GetInput struct {
	server.BaseAuthInput
	team_service.GetTeamInput
}

type TeamResponse struct {
	Body struct {
		Data *models.TeamResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*TeamResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	team, err := self.srv.TeamService.Get(ctx, user.ID, &input.GetTeamInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TeamResponse{}
	resp.Body.Data = team
	return resp, nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *team_service.CreateTeamInput `required:"true"`
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*TeamResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	team, err := self.srv.TeamService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TeamResponse{}
	resp.Body.Data = team
	return resp, nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *team_service.UpdateTeamInput `required:"true"`
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*TeamResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	team, err := self.srv.TeamService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TeamResponse{}
	resp.Body.Data = team
	return resp, nil
}

type DeleteInput struct {
	server.BaseAuthInput
	Body *team_service.DeleteTeamInput `required:"true"`
}

type DeletedResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Delete(ctx context.Context, input *DeleteInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.TeamService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.TeamID, Deleted: true}
	return resp, nil
}

type TeamRefInput struct {
	server.BaseAuthInput
	team_service.TeamRefInput
}

type TeamMemberListResponse struct {
	Body struct {
		Data []*models.TeamMemberResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ListMembers(ctx context.Context, input *TeamRefInput) (*TeamMemberListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	members, err := self.srv.TeamService.ListMembers(ctx, user.ID, input.TeamID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TeamMemberListResponse{}
	resp.Body.Data = members
	return resp, nil
}

type AddMemberInput struct {
	server.BaseAuthInput
	Body *team_service.AddTeamMemberInput `required:"true"`
}

type TeamMemberResponse struct {
	Body struct {
		Data *models.TeamMemberResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) AddMember(ctx context.Context, input *AddMemberInput) (*TeamMemberResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	member, err := self.srv.TeamService.AddMember(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &TeamMemberResponse{}
	resp.Body.Data = member
	return resp, nil
}

type RemoveMemberInput struct {
	server.BaseAuthInput
	Body *team_service.RemoveTeamMemberInput `required:"true"`
}

type SuccessResponse struct {
	Body struct {
		Data models.SuccessResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) RemoveMember(ctx context.Context, input *RemoveMemberInput) (*SuccessResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.TeamService.RemoveMember(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}

type StateListResponse struct {
	Body struct {
		Data []*models.WorkflowStateResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) ListStates(ctx context.Context, input *TeamRefInput) (*StateListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	states, err := self.srv.TeamService.ListStates(ctx, user.ID, input.TeamID)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &StateListResponse{}
	resp.Body.Data = states
	return resp, nil
}

type CreateStateInput struct {
	server.BaseAuthInput
	Body *team_service.CreateStateInput `required:"true"`
}

type StateResponse struct {
	Body struct {
		Data *models.WorkflowStateResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) CreateState(ctx context.Context, input *CreateStateInput) (*StateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	state, err := self.srv.TeamService.CreateState(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &StateResponse{}
	resp.Body.Data = state
	return resp, nil
}

type UpdateStateInput struct {
	server.BaseAuthInput
	Body *team_service.UpdateStateInput `required:"true"`
}

func (self *HandlerGroup) UpdateState(ctx context.Context, input *UpdateStateInput) (*StateResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	state, err := self.srv.TeamService.UpdateState(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &StateResponse{}
	resp.Body.Data = state
	return resp, nil
}

type DeleteStateInput struct {
	server.BaseAuthInput
	Body *team_service.DeleteStateInput `required:"true"`
}

func (self *HandlerGroup) DeleteState(ctx context.Context, input *DeleteStateInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.TeamService.DeleteState(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: input.Body.StateID, Deleted: true}
	return resp, nil
}
