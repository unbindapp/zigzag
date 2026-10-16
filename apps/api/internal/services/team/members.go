package team_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
)

type TeamRefInput struct {
	TeamID uuid.UUID `query:"team_id" required:"true" format:"uuid" doc:"The team"`
}

func (self *TeamService) ListMembers(ctx context.Context, userID, teamID uuid.UUID) ([]*models.TeamMemberResponse, error) {
	team, err := self.visibleTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	members, err := self.repo.Team().ListMembers(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	return models.TransformTeamMembers(members), nil
}

type AddTeamMemberInput struct {
	TeamID uuid.UUID `json:"team_id" required:"true" format:"uuid" doc:"The team"`
	UserID uuid.UUID `json:"user_id" required:"true" format:"uuid" doc:"A member of the team's workspace; your own id to join"`
	Owner  bool      `json:"owner,omitempty" required:"false" doc:"Mark as team owner; admins only"`
}

func (self *TeamService) AddMember(ctx context.Context, userID uuid.UUID, input *AddTeamMemberInput) (*models.TeamMemberResponse, error) {
	team, err := self.visibleTeam(ctx, userID, input.TeamID)
	if err != nil {
		return nil, err
	}
	joiningSelf := input.UserID == userID && !team.Private && !input.Owner
	required := schema.WorkspaceRoleAdmin
	if joiningSelf {
		required = schema.WorkspaceRoleMember
	}
	if err := self.check(ctx, userID, required, team.WorkspaceID); err != nil {
		return nil, err
	}
	if _, err := self.repo.Workspace().GetMember(ctx, nil, team.WorkspaceID, input.UserID); err != nil {
		if ent.IsNotFound(err) {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The user must be a member of the workspace first")
		}
		return nil, err
	}
	if err := self.repo.Team().AddMember(ctx, nil, team.ID, input.UserID, input.Owner); err != nil {
		return nil, err
	}
	member, err := self.repo.Team().GetMember(ctx, team.ID, input.UserID)
	if err != nil {
		return nil, err
	}
	return models.TransformTeamMember(member), nil
}

type RemoveTeamMemberInput struct {
	TeamID uuid.UUID `json:"team_id" required:"true" format:"uuid" doc:"The team"`
	UserID uuid.UUID `json:"user_id" required:"true" format:"uuid" doc:"The member to remove; your own id to leave"`
}

func (self *TeamService) RemoveMember(ctx context.Context, userID uuid.UUID, input *RemoveTeamMemberInput) error {
	team, err := self.visibleTeam(ctx, userID, input.TeamID)
	if err != nil {
		return err
	}
	required := schema.WorkspaceRoleAdmin
	if input.UserID == userID {
		required = schema.WorkspaceRoleViewer
	}
	if err := self.check(ctx, userID, required, team.WorkspaceID); err != nil {
		return err
	}
	return self.repo.Team().RemoveMember(ctx, team.ID, input.UserID)
}
