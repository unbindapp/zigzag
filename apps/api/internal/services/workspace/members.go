package workspace_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/mail"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

type ListMembersInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
}

func (self *WorkspaceService) ListMembers(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error) {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleViewer, workspaceID); err != nil {
		return nil, err
	}
	members, err := self.repo.Workspace().ListMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	teamIDs, err := self.repo.Workspace().MemberTeamIDs(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.WorkspaceMemberResponse, len(members))
	for i, member := range members {
		responses[i] = models.TransformWorkspaceMember(member, teamIDs[member.UserID])
	}
	return responses, nil
}

type AddMemberInput struct {
	WorkspaceID uuid.UUID            `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Email       string               `json:"email" required:"true" format:"email" maxLength:"254" doc:"Email of the person to add"`
	Role        schema.WorkspaceRole `json:"role" required:"true" doc:"admin, member or viewer"`
	TeamIDs     []uuid.UUID          `json:"team_ids,omitempty" required:"false" doc:"Teams of this workspace to join"`
}

func (self *WorkspaceService) AddMember(ctx context.Context, userID uuid.UUID, input *AddMemberInput) (*models.AddWorkspaceMemberResponse, error) {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return nil, err
	}
	if !permissions_repo.IsValidRole(input.Role) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Role must be admin, member or viewer")
	}
	email, err := names.Email(input.Email)
	if err != nil {
		return nil, err
	}
	if err := self.validateTeams(ctx, input.WorkspaceID, input.TeamIDs); err != nil {
		return nil, err
	}

	existing, err := self.repo.User().GetByEmail(ctx, email)
	if ent.IsNotFound(err) || (err == nil && !existing.Active) {
		invitation, err := self.CreateInvitation(ctx, userID, &CreateInvitationInput{WorkspaceID: input.WorkspaceID, Email: email, Role: input.Role, TeamIDs: input.TeamIDs})
		if err != nil {
			return nil, err
		}
		return &models.AddWorkspaceMemberResponse{Invitation: invitation}, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = self.repo.Workspace().GetMember(ctx, nil, input.WorkspaceID, existing.ID)
	if err == nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, email+" is already a member of this workspace")
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	var member *ent.WorkspaceMember
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		member, err = self.repo.Workspace().AddMember(ctx, tx, input.WorkspaceID, existing.ID, input.Role)
		if err != nil {
			return err
		}
		for _, teamID := range input.TeamIDs {
			if err := self.repo.Team().AddMember(ctx, tx, teamID, existing.ID, false); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	self.notifyAdded(ctx, userID, input.WorkspaceID, existing.Email, input.Role)
	return &models.AddWorkspaceMemberResponse{Member: models.TransformWorkspaceMember(member, input.TeamIDs)}, nil
}

type UpdateMemberInput struct {
	WorkspaceID uuid.UUID            `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	UserID      uuid.UUID            `json:"user_id" required:"true" format:"uuid" doc:"The member"`
	Role        schema.WorkspaceRole `json:"role" required:"true" doc:"admin, member or viewer"`
}

func (self *WorkspaceService) UpdateMember(ctx context.Context, userID uuid.UUID, input *UpdateMemberInput) (*models.WorkspaceMemberResponse, error) {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return nil, err
	}
	if !permissions_repo.IsValidRole(input.Role) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Role must be admin, member or viewer")
	}
	current, err := self.repo.Workspace().GetMember(ctx, nil, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Member not found")
	}
	if current.Role == schema.WorkspaceRoleAdmin && input.Role != schema.WorkspaceRoleAdmin {
		if err := self.ensureAnotherAdmin(ctx, input.WorkspaceID); err != nil {
			return nil, err
		}
	}
	updated, err := self.repo.Workspace().UpdateMemberRole(ctx, input.WorkspaceID, input.UserID, input.Role)
	if err != nil {
		return nil, err
	}
	teamIDs, err := self.repo.Team().TeamIDsOfUser(ctx, input.UserID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return models.TransformWorkspaceMember(updated, teamIDs), nil
}

type RemoveMemberInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	UserID      uuid.UUID `json:"user_id" required:"true" format:"uuid" doc:"The member to remove; your own id to leave"`
}

func (self *WorkspaceService) RemoveMember(ctx context.Context, userID uuid.UUID, input *RemoveMemberInput) error {
	required := schema.WorkspaceRoleAdmin
	if input.UserID == userID {
		required = schema.WorkspaceRoleViewer
	}
	if err := self.checkVisible(ctx, userID, required, input.WorkspaceID); err != nil {
		return err
	}
	current, err := self.repo.Workspace().GetMember(ctx, nil, input.WorkspaceID, input.UserID)
	if err != nil {
		return errdefs.MaskAsNotFound(err, "Member not found")
	}
	if current.Role == schema.WorkspaceRoleAdmin {
		if err := self.ensureAnotherAdmin(ctx, input.WorkspaceID); err != nil {
			return err
		}
	}
	return self.repo.Workspace().RemoveMember(ctx, input.WorkspaceID, input.UserID)
}

func (self *WorkspaceService) ensureAnotherAdmin(ctx context.Context, workspaceID uuid.UUID) error {
	admins, err := self.repo.Workspace().CountAdmins(ctx, workspaceID)
	if err != nil {
		return err
	}
	if admins <= 1 {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A workspace needs at least one admin")
	}
	return nil
}

func (self *WorkspaceService) validateTeams(ctx context.Context, workspaceID uuid.UUID, teamIDs []uuid.UUID) error {
	if len(teamIDs) == 0 {
		return nil
	}
	teams, err := self.repo.Team().TeamsInWorkspace(ctx, workspaceID, teamIDs)
	if err != nil {
		return err
	}
	if len(teams) != len(uniqueIDs(teamIDs)) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "team_ids must be teams of this workspace")
	}
	return nil
}

func uniqueIDs(ids []uuid.UUID) map[uuid.UUID]struct{} {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

func (self *WorkspaceService) notifyAdded(ctx context.Context, inviterID, workspaceID uuid.UUID, to string, role schema.WorkspaceRole) {
	inviter, workspace, err := self.inviterAndWorkspace(ctx, inviterID, workspaceID)
	if err != nil {
		return
	}
	link, _ := self.workspaceLink(workspace)
	self.email.SendAsync(ctx, to, mail.Content{
		Subject:     fmt.Sprintf("You were added to %s on Zigzag", workspace.Name),
		Heading:     fmt.Sprintf("You are now in %s", workspace.Name),
		Paragraphs:  []string{fmt.Sprintf("%s added you to the workspace %s as %s.", displayOf(inviter), workspace.Name, role)},
		ButtonLabel: "Open the workspace",
		ButtonURL:   link,
	})
}
