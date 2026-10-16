package models

import (
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type WhoamiAPIKey struct {
	Role         schema.WorkspaceRole  `json:"role" doc:"Strongest role this connection acts with"`
	FullAccess   bool                  `json:"full_access" doc:"True when it reaches every workspace the user can"`
	WorkspaceIDs []uuid.UUID           `json:"workspace_ids" nullable:"false" doc:"Workspaces this connection is limited to when not full access"`
	Privileges   []schema.KeyPrivilege `json:"privileges" nullable:"false"`
}

type WhoamiResponse struct {
	User        *UserResponse          `json:"user"`
	SystemAdmin bool                   `json:"system_admin" doc:"Effective system admin, false for narrowed credentials unless they hold full admin access"`
	Workspaces  []*WorkspaceMembership `json:"workspaces" nullable:"false" doc:"Workspaces this connection can see, with the effective role in each"`
	APIKey      *WhoamiAPIKey          `json:"api_key,omitempty" required:"false" doc:"Present for API keys and connected apps: the limit the credential puts on this user"`
}
