package workspace_service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/mail"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

const minPasswordLength = 8

type CreateInvitationInput struct {
	WorkspaceID uuid.UUID            `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Email       string               `json:"email" required:"true" format:"email" maxLength:"254" doc:"Email to invite"`
	Role        schema.WorkspaceRole `json:"role" required:"true" doc:"admin, member or viewer"`
	TeamIDs     []uuid.UUID          `json:"team_ids,omitempty" required:"false" doc:"Teams of this workspace joined on accept"`
}

func (self *WorkspaceService) CreateInvitation(ctx context.Context, userID uuid.UUID, input *CreateInvitationInput) (*models.InvitationResponse, error) {
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
	if existing, err := self.repo.User().GetByEmail(ctx, email); err == nil {
		if _, err := self.repo.Workspace().GetMember(ctx, nil, input.WorkspaceID, existing.ID); err == nil {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, email+" is already a member of this workspace")
		}
	}

	token, err := utils.RandomToken(32)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to generate the invitation token")
	}
	invitation, err := self.repo.Workspace().CreateInvitation(ctx, &workspace_repo.CreateInvitationInput{
		WorkspaceID: input.WorkspaceID,
		Email:       email,
		Role:        input.Role,
		TeamIDs:     input.TeamIDs,
		TokenHash:   utils.HashToken(token),
		ExpiresAt:   self.now().Add(invitationTTL),
		InvitedBy:   userID,
	})
	if err != nil {
		return nil, err
	}
	return self.deliverInvitation(ctx, userID, invitation, token)
}

type ListInvitationsInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
}

func (self *WorkspaceService) ListInvitations(ctx context.Context, userID, workspaceID uuid.UUID) ([]*models.InvitationResponse, error) {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, workspaceID); err != nil {
		return nil, err
	}
	invitations, err := self.repo.Workspace().ListPendingInvitations(ctx, workspaceID, self.now())
	if err != nil {
		return nil, err
	}
	return models.TransformInvitationEntities(invitations), nil
}

type InvitationRefInput struct {
	WorkspaceID  uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	InvitationID uuid.UUID `json:"invitation_id" required:"true" format:"uuid" doc:"The invitation"`
}

func (self *WorkspaceService) RevokeInvitation(ctx context.Context, userID uuid.UUID, input *InvitationRefInput) error {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return err
	}
	return self.repo.Workspace().DeleteInvitation(ctx, input.WorkspaceID, input.InvitationID)
}

func (self *WorkspaceService) ResendInvitation(ctx context.Context, userID uuid.UUID, input *InvitationRefInput) (*models.InvitationResponse, error) {
	if err := self.checkVisible(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return nil, err
	}
	invitation, err := self.repo.Workspace().GetInvitation(ctx, input.WorkspaceID, input.InvitationID)
	if err != nil {
		return nil, err
	}
	if invitation.AcceptedAt != nil {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "This invitation was already accepted")
	}
	token, err := utils.RandomToken(32)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to generate the invitation token")
	}
	rotated, err := self.repo.Workspace().RotateInvitationToken(ctx, invitation.ID, utils.HashToken(token), self.now().Add(invitationTTL))
	if err != nil {
		return nil, err
	}
	return self.deliverInvitation(ctx, userID, rotated, token)
}

func (self *WorkspaceService) deliverInvitation(ctx context.Context, inviterID uuid.UUID, invitation *ent.Invitation, token string) (*models.InvitationResponse, error) {
	inviter, workspace, err := self.inviterAndWorkspace(ctx, inviterID, invitation.WorkspaceID)
	if err != nil {
		return nil, err
	}
	link, err := utils.JoinURLPaths(self.cfg.ExternalUIUrl, "invite", token)
	if err != nil {
		return nil, err
	}
	resp := models.TransformInvitationEntity(invitation)
	resp.Link = link
	resp.Emailed = self.email.SendAsync(ctx, invitation.Email, mail.Content{
		Subject:     fmt.Sprintf("%s invited you to %s on Zigzag", displayOf(inviter), workspace.Name),
		Heading:     fmt.Sprintf("Join %s", workspace.Name),
		Paragraphs:  []string{fmt.Sprintf("%s invited you to the workspace %s as %s.", displayOf(inviter), workspace.Name, invitation.Role), "Accept the invitation to get started."},
		ButtonLabel: "Accept invitation",
		ButtonURL:   link,
		Note:        "This invitation expires in 7 days. If you were not expecting it, you can ignore this email.",
	})
	return resp, nil
}

func (self *WorkspaceService) GetInvitation(ctx context.Context, token string) (*models.InvitationPreviewResponse, error) {
	invitation, err := self.pendingInvitation(ctx, token)
	if err != nil {
		return nil, err
	}
	existing, err := self.repo.User().GetByEmail(ctx, invitation.Email)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	return &models.InvitationPreviewResponse{
		Email:         invitation.Email,
		WorkspaceName: invitation.Edges.Workspace.Name,
		Role:          invitation.Role,
		ExpiresAt:     invitation.ExpiresAt,
		AccountExists: hasUsableAccount(existing),
	}, nil
}

type AcceptInvitationInput struct {
	Token       string  `json:"token" required:"true" minLength:"1" maxLength:"200" doc:"Token from the invitation link"`
	Password    *string `json:"password,omitempty" required:"false" maxLength:"72" doc:"The password of the existing account, or a new one of at least 8 characters when the account is created"`
	Name        *string `json:"name,omitempty" required:"false" maxLength:"100" doc:"Full name for a new account"`
	DisplayName *string `json:"display_name,omitempty" required:"false" maxLength:"33" doc:"Mention handle for a new account; derived from the email when omitted"`
}

// An existing account must prove its password: the invitation link is also handed to the admin who created it.
func (self *WorkspaceService) AcceptInvitation(ctx context.Context, input *AcceptInvitationInput) (*ent.User, *models.WorkspaceResponse, error) {
	invitation, err := self.pendingInvitation(ctx, input.Token)
	if err != nil {
		return nil, nil, err
	}
	existing, err := self.repo.User().GetByEmail(ctx, invitation.Email)
	if err != nil && !ent.IsNotFound(err) {
		return nil, nil, err
	}
	password := ""
	if input.Password != nil {
		password = *input.Password
	}
	if err := self.verifyAccepter(ctx, existing, invitation.Email, password); err != nil {
		return nil, nil, err
	}

	var user *ent.User
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		claimed, err := self.repo.Workspace().ClaimInvitation(ctx, tx, invitation.ID, self.now())
		if err != nil {
			return err
		}
		if !claimed {
			return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "This invitation is no longer valid")
		}
		user, err = self.accountForInvitation(ctx, tx, invitation, existing, password, input)
		if err != nil {
			return err
		}
		if _, err := self.repo.Workspace().AddMember(ctx, tx, invitation.WorkspaceID, user.ID, invitation.Role); err != nil {
			return err
		}
		for _, teamID := range invitation.TeamIds {
			err := self.repo.Team().AddMember(ctx, tx, teamID, user.ID, false)
			if err != nil && !ent.IsConstraintError(err) {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	joined, err := self.respond(ctx, user.ID, invitation.Edges.Workspace)
	if err != nil {
		return nil, nil, err
	}
	return user, joined, nil
}

func (self *WorkspaceService) verifyAccepter(ctx context.Context, existing *ent.User, email, password string) error {
	if !hasUsableAccount(existing) {
		if len(password) < minPasswordLength {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Choose a password of at least 8 characters")
		}
		return nil
	}
	if _, err := self.repo.User().Authenticate(ctx, email, password); err != nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Enter the password of your existing account to accept")
	}
	return nil
}

func (self *WorkspaceService) accountForInvitation(ctx context.Context, tx repository.TxInterface, invitation *ent.Invitation, existing *ent.User, password string, input *AcceptInvitationInput) (*ent.User, error) {
	if hasUsableAccount(existing) {
		return existing, nil
	}
	name := strings.TrimSpace(derefOr(input.Name, ""))
	if existing != nil {
		if err := self.repo.User().UpdatePassword(ctx, tx, existing.ID, password); err != nil {
			return nil, err
		}
		update := &user_repo.UpdateUserInput{Active: utils.ToPtr(true)}
		if name != "" {
			update.Name = &name
		}
		return self.repo.User().Update(ctx, tx, existing.ID, update)
	}
	handle, err := self.uniqueDisplayName(ctx, tx, derefOr(input.DisplayName, ""), invitation.Email)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = handle
	}
	return self.repo.User().Create(ctx, tx, &user_repo.CreateUserInput{
		Email:       invitation.Email,
		Password:    password,
		Name:        name,
		DisplayName: handle,
		Active:      true,
	})
}

func (self *WorkspaceService) uniqueDisplayName(ctx context.Context, tx repository.TxInterface, requested, email string) (string, error) {
	if requested != "" {
		handle, err := names.DisplayName(requested)
		if err != nil {
			return "", err
		}
		_, err = self.repo.User().GetByDisplayName(ctx, tx, handle)
		if ent.IsNotFound(err) {
			return handle, nil
		}
		if err != nil {
			return "", err
		}
		return "", errdefs.NewCustomError(errdefs.ErrTypeConflict, "The display name @"+handle+" is taken")
	}
	base := names.SuggestDisplayName(email)
	for attempt := 1; attempt <= 50; attempt++ {
		candidate := base
		if attempt > 1 {
			candidate = base + strconv.Itoa(attempt)
		}
		_, err := self.repo.User().GetByDisplayName(ctx, tx, candidate)
		if ent.IsNotFound(err) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
	suffix, err := utils.GenerateRandomSimpleID(6)
	if err != nil {
		return "", err
	}
	return base + "-" + suffix, nil
}

func (self *WorkspaceService) pendingInvitation(ctx context.Context, token string) (*ent.Invitation, error) {
	invalid := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "This invitation is no longer valid")
	invitation, err := self.repo.Workspace().GetInvitationByTokenHash(ctx, utils.HashToken(token))
	if ent.IsNotFound(err) {
		return nil, invalid
	}
	if err != nil {
		return nil, err
	}
	if invitation.AcceptedAt != nil || !invitation.ExpiresAt.After(self.now()) || invitation.Edges.Workspace == nil {
		return nil, invalid
	}
	return invitation, nil
}

func (self *WorkspaceService) CleanupExpiredInvitations(ctx context.Context) error {
	_, err := self.repo.Workspace().DeleteExpiredInvitations(ctx, self.now())
	return err
}

func (self *WorkspaceService) inviterAndWorkspace(ctx context.Context, inviterID, workspaceID uuid.UUID) (*ent.User, *ent.Workspace, error) {
	inviter, err := self.repo.User().GetByID(ctx, inviterID)
	if err != nil {
		return nil, nil, err
	}
	workspace, err := self.repo.Workspace().GetByID(ctx, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return inviter, workspace, nil
}

func (self *WorkspaceService) workspaceLink(workspace *ent.Workspace) (string, error) {
	return utils.JoinURLPaths(self.cfg.ExternalUIUrl, workspace.Slug)
}

func hasUsableAccount(user *ent.User) bool {
	return user != nil && user.Active && user.PasswordHash != ""
}

func displayOf(user *ent.User) string {
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}

func derefOr(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}
