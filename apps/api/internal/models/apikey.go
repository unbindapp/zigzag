package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type APIKeyCreateInput struct {
	Name         string                `json:"name" required:"true" minLength:"1" maxLength:"100" doc:"Name to recognize the key by, unique per user"`
	ExpiresAt    *time.Time            `json:"expires_at,omitempty" nullable:"true" required:"false" doc:"When the key stops working. Omit for a key that never expires."`
	Role         schema.WorkspaceRole  `json:"role" required:"true" doc:"Strongest role the key acts with. It never exceeds your own role in a workspace."`
	FullAccess   bool                  `json:"full_access" required:"true" doc:"Reach every workspace you can, capped at role. workspace_ids must be empty."`
	WorkspaceIDs []uuid.UUID           `json:"workspace_ids" format:"uuid" required:"true" nullable:"false" doc:"Workspaces the key is limited to. Required unless full_access."`
	Privileges   []schema.KeyPrivilege `json:"privileges,omitempty" required:"false" nullable:"false" doc:"What the key may see beyond its role. None exist yet."`
}

type APIKeyDeleteInput struct {
	ID uuid.UUID `json:"id" format:"uuid" required:"true" doc:"The API key to revoke"`
}

type APIKeyResponse struct {
	ID           uuid.UUID             `json:"id" format:"uuid"`
	UserID       uuid.UUID             `json:"user_id" format:"uuid"`
	Name         string                `json:"name"`
	TokenPrefix  string                `json:"token_prefix" doc:"First characters of the token, for recognizing the key. Never the full token."`
	Role         schema.WorkspaceRole  `json:"role"`
	FullAccess   bool                  `json:"full_access"`
	WorkspaceIDs []uuid.UUID           `json:"workspace_ids" format:"uuid" nullable:"false"`
	Privileges   []schema.KeyPrivilege `json:"privileges" nullable:"false"`
	ExpiresAt    *time.Time            `json:"expires_at,omitempty" nullable:"true" required:"false"`
	LastUsedAt   *time.Time            `json:"last_used_at,omitempty" nullable:"true" required:"false"`
	CreatedAt    time.Time             `json:"created_at"`
}

type APIKeyCreatedResponse struct {
	APIKeyResponse
	Token string `json:"token" doc:"The full API key. Shown once, store it now."`
}

func uuidsOf(ids []uuid.UUID) []uuid.UUID {
	if ids == nil {
		return []uuid.UUID{}
	}
	return ids
}

func privilegesOf(privileges []schema.KeyPrivilege) []schema.KeyPrivilege {
	if privileges == nil {
		return []schema.KeyPrivilege{}
	}
	return privileges
}

func TransformAPIKeyEntity(entity *ent.APIKey) *APIKeyResponse {
	return &APIKeyResponse{
		ID:           entity.ID,
		UserID:       entity.UserID,
		Name:         entity.Name,
		TokenPrefix:  entity.TokenPrefix,
		Role:         entity.Role,
		FullAccess:   entity.FullAccess,
		WorkspaceIDs: uuidsOf(entity.WorkspaceIds),
		Privileges:   privilegesOf(entity.Privileges),
		ExpiresAt:    entity.ExpiresAt,
		LastUsedAt:   entity.LastUsedAt,
		CreatedAt:    entity.CreatedAt,
	}
}

func TransformAPIKeyEntities(entities []*ent.APIKey) []*APIKeyResponse {
	responses := make([]*APIKeyResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformAPIKeyEntity(entity)
	}
	return responses
}

type APIKeyListInput struct {
	UserID uuid.UUID `query:"user_id" format:"uuid" required:"false" doc:"List another user's keys. Requires system admin. Defaults to your own."`
}
