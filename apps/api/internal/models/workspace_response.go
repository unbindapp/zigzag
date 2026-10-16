package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

func workspaceIconURL(entity *ent.Workspace) *string {
	if entity.IconUploadID == nil {
		return nil
	}
	url := UploadURL(*entity.IconUploadID)
	return &url
}

type WorkspaceResponse struct {
	ID            uuid.UUID            `json:"id" format:"uuid"`
	Name          string               `json:"name"`
	Slug          string               `json:"slug" doc:"URL key of the workspace"`
	Icon          string               `json:"icon"`
	IconURL       *string              `json:"icon_url" nullable:"true" doc:"Logo image, shown instead of icon when set"`
	Color         string               `json:"color"`
	Role          schema.WorkspaceRole `json:"role" doc:"The caller's effective role in this workspace"`
	MemberCount   int                  `json:"member_count"`
	TeamCount     int                  `json:"team_count"`
	DefaultTeamID *uuid.UUID           `json:"default_team_id,omitempty" required:"false" format:"uuid" doc:"The team created together with the workspace. Only returned when the workspace is created."`
	CreatedAt     time.Time            `json:"created_at"`
}

func TransformWorkspaceEntity(entity *ent.Workspace, role schema.WorkspaceRole, memberCount, teamCount int) *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:          entity.ID,
		Name:        entity.Name,
		Slug:        entity.Slug,
		Icon:        entity.Icon,
		IconURL:     workspaceIconURL(entity),
		Color:       entity.Color,
		Role:        role,
		MemberCount: memberCount,
		TeamCount:   teamCount,
		CreatedAt:   entity.CreatedAt,
	}
}

type WorkspaceMembership struct {
	ID      uuid.UUID            `json:"id" format:"uuid"`
	Name    string               `json:"name"`
	Slug    string               `json:"slug"`
	Icon    string               `json:"icon"`
	IconURL *string              `json:"icon_url" nullable:"true" doc:"Logo image, shown instead of icon when set"`
	Color   string               `json:"color"`
	Role    schema.WorkspaceRole `json:"role"`
}

func TransformWorkspaceMembership(entity *ent.Workspace, role schema.WorkspaceRole) *WorkspaceMembership {
	return &WorkspaceMembership{
		ID:      entity.ID,
		Name:    entity.Name,
		Slug:    entity.Slug,
		Icon:    entity.Icon,
		IconURL: workspaceIconURL(entity),
		Color:   entity.Color,
		Role:    role,
	}
}

type WorkspaceMemberResponse struct {
	User     MemberUser           `json:"user"`
	Role     schema.WorkspaceRole `json:"role"`
	JoinedAt time.Time            `json:"joined_at"`
	TeamIDs  []uuid.UUID          `json:"team_ids" format:"uuid" nullable:"false"`
}

func TransformWorkspaceMember(entity *ent.WorkspaceMember, teamIDs []uuid.UUID) *WorkspaceMemberResponse {
	if teamIDs == nil {
		teamIDs = []uuid.UUID{}
	}
	return &WorkspaceMemberResponse{
		User:     TransformMemberUser(entity.Edges.User),
		Role:     entity.Role,
		JoinedAt: entity.JoinedAt,
		TeamIDs:  teamIDs,
	}
}

type InvitationResponse struct {
	ID          uuid.UUID            `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID            `json:"workspace_id" format:"uuid"`
	Email       string               `json:"email"`
	Role        schema.WorkspaceRole `json:"role"`
	TeamIDs     []uuid.UUID          `json:"team_ids" format:"uuid" nullable:"false" doc:"Teams joined on accept"`
	ExpiresAt   time.Time            `json:"expires_at"`
	CreatedAt   time.Time            `json:"created_at"`
	Link        string               `json:"link,omitempty" required:"false" doc:"The invitation link. Only returned when the invitation is created or resent."`
	Emailed     bool                 `json:"emailed" doc:"Whether the link was also emailed. False when mail is not configured."`
}

func TransformInvitationEntity(entity *ent.Invitation) *InvitationResponse {
	teamIDs := entity.TeamIds
	if teamIDs == nil {
		teamIDs = []uuid.UUID{}
	}
	return &InvitationResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Email:       entity.Email,
		Role:        entity.Role,
		TeamIDs:     teamIDs,
		ExpiresAt:   entity.ExpiresAt,
		CreatedAt:   entity.CreatedAt,
	}
}

func TransformInvitationEntities(entities []*ent.Invitation) []*InvitationResponse {
	responses := make([]*InvitationResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformInvitationEntity(entity)
	}
	return responses
}

type AddWorkspaceMemberResponse struct {
	Member     *WorkspaceMemberResponse `json:"member,omitempty" required:"false" doc:"Set when the address already had an account and was added at once"`
	Invitation *InvitationResponse      `json:"invitation,omitempty" required:"false" doc:"Set when the address had no account and an invitation was created"`
}

type InvitationPreviewResponse struct {
	Email         string               `json:"email"`
	WorkspaceName string               `json:"workspace_name"`
	Role          schema.WorkspaceRole `json:"role"`
	ExpiresAt     time.Time            `json:"expires_at"`
	AccountExists bool                 `json:"account_exists" doc:"True when an active account with this email exists, so accepting needs no password"`
}

type SessionWorkspaceResponse struct {
	User      *UserResponse      `json:"user"`
	Workspace *WorkspaceResponse `json:"workspace"`
}
