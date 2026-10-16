package workspace_repo

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/invitation"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type CreateInvitationInput struct {
	WorkspaceID uuid.UUID
	Email       string
	Role        schema.WorkspaceRole
	TeamIDs     []uuid.UUID
	TokenHash   string
	ExpiresAt   time.Time
	InvitedBy   uuid.UUID
}

// CreateInvitation replaces any earlier invitation for the same address in the workspace.
func (self *WorkspaceRepository) CreateInvitation(ctx context.Context, input *CreateInvitationInput) (created *ent.Invitation, err error) {
	teamIDs := input.TeamIDs
	if teamIDs == nil {
		teamIDs = []uuid.UUID{}
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	err = self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		if _, err := tx.Client().Invitation.Delete().
			Where(invitation.WorkspaceID(input.WorkspaceID), invitation.EmailEqualFold(email)).
			Exec(ctx); err != nil {
			return err
		}
		created, err = tx.Client().Invitation.Create().
			SetWorkspaceID(input.WorkspaceID).
			SetEmail(email).
			SetRole(input.Role).
			SetTeamIds(teamIDs).
			SetTokenHash(input.TokenHash).
			SetExpiresAt(input.ExpiresAt).
			SetInvitedBy(input.InvitedBy).
			Save(ctx)
		return err
	})
	return created, err
}

func (self *WorkspaceRepository) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*ent.Invitation, error) {
	return self.base.DB.Invitation.Query().
		Where(invitation.TokenHash(tokenHash)).
		WithWorkspace().
		Only(ctx)
}

func (self *WorkspaceRepository) GetInvitation(ctx context.Context, workspaceID, invitationID uuid.UUID) (*ent.Invitation, error) {
	return self.base.DB.Invitation.Query().
		Where(invitation.ID(invitationID), invitation.WorkspaceID(workspaceID)).
		Only(ctx)
}

func (self *WorkspaceRepository) ListPendingInvitations(ctx context.Context, workspaceID uuid.UUID, now time.Time) ([]*ent.Invitation, error) {
	return self.base.DB.Invitation.Query().
		Where(invitation.WorkspaceID(workspaceID), invitation.AcceptedAtIsNil(), invitation.ExpiresAtGT(now)).
		Order(ent.Asc(invitation.FieldEmail)).
		All(ctx)
}

func (self *WorkspaceRepository) ListPendingInvitationsForEmail(ctx context.Context, tx repository.TxInterface, email string, now time.Time) ([]*ent.Invitation, error) {
	return self.base.Client(tx).Invitation.Query().
		Where(invitation.EmailEqualFold(email), invitation.AcceptedAtIsNil(), invitation.ExpiresAtGT(now)).
		Order(ent.Asc(invitation.FieldCreatedAt)).
		All(ctx)
}

func (self *WorkspaceRepository) RotateInvitationToken(ctx context.Context, invitationID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.Invitation, error) {
	return self.base.DB.Invitation.UpdateOneID(invitationID).
		SetTokenHash(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
}

func (self *WorkspaceRepository) DeleteInvitation(ctx context.Context, workspaceID, invitationID uuid.UUID) error {
	deleted, err := self.base.DB.Invitation.Delete().
		Where(invitation.ID(invitationID), invitation.WorkspaceID(workspaceID)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// ClaimInvitation spends the invitation once; false means another accept got there first.
func (self *WorkspaceRepository) ClaimInvitation(ctx context.Context, tx repository.TxInterface, invitationID uuid.UUID, now time.Time) (bool, error) {
	updated, err := self.base.Client(tx).Invitation.Update().
		Where(invitation.ID(invitationID), invitation.AcceptedAtIsNil(), invitation.ExpiresAtGT(now)).
		SetAcceptedAt(now).
		Save(ctx)
	return updated == 1, err
}

func (self *WorkspaceRepository) DeleteExpiredInvitations(ctx context.Context, before time.Time) (int, error) {
	return self.base.DB.Invitation.Delete().
		Where(invitation.AcceptedAtIsNil(), invitation.ExpiresAtLT(before)).
		Exec(ctx)
}
