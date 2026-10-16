package models

import (
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type ConnectedAppClientInput struct {
	ClientID    string `query:"client_id" required:"true" maxLength:"512" doc:"OAuth client_id from the authorize request"`
	RedirectURI string `query:"redirect_uri" required:"true" maxLength:"512" doc:"redirect_uri from the authorize request"`
}

type ConnectedAppApproveInput struct {
	ClientID      string                `json:"client_id" required:"true" maxLength:"512"`
	RedirectURI   string                `json:"redirect_uri" required:"true" maxLength:"512"`
	State         string                `json:"state,omitempty" required:"false" maxLength:"1024"`
	CodeChallenge string                `json:"code_challenge" required:"true" minLength:"43" maxLength:"128"`
	Resource      string                `json:"resource,omitempty" required:"false" maxLength:"512"`
	Scope         string                `json:"scope,omitempty" required:"false" maxLength:"512"`
	Role          schema.WorkspaceRole  `json:"role" required:"true" doc:"Strongest role the app acts with. It never exceeds your own role in a workspace."`
	FullAccess    bool                  `json:"full_access" required:"true" doc:"Reach every workspace you can, capped at role. workspace_ids must be empty."`
	WorkspaceIDs  []uuid.UUID           `json:"workspace_ids" required:"true" nullable:"false" doc:"Workspaces the app is limited to. Required unless full_access."`
	Privileges    []schema.KeyPrivilege `json:"privileges,omitempty" required:"false" nullable:"false" doc:"What the app may see beyond its role. None exist yet."`
}

type ConnectedAppDenyInput struct {
	ClientID    string `json:"client_id" required:"true" maxLength:"512"`
	RedirectURI string `json:"redirect_uri" required:"true" maxLength:"512"`
	State       string `json:"state,omitempty" required:"false" maxLength:"1024"`
}

type ConnectedAppRevokeInput struct {
	ID uuid.UUID `json:"id" format:"uuid" required:"true" doc:"The connected app to revoke"`
}
