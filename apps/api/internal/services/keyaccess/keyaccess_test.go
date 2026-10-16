package keyaccess

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

func TestValidate(t *testing.T) {
	ws := uuid.New()
	valid := []Spec{
		{Role: schema.WorkspaceRoleViewer, FullAccess: true},
		{Role: schema.WorkspaceRoleAdmin, WorkspaceIDs: []uuid.UUID{ws}},
	}
	for _, spec := range valid {
		assert.NoError(t, Validate(spec))
	}
	invalid := map[string]Spec{
		"unknown role":                  {Role: "owner", FullAccess: true},
		"full access with a workspace":  {Role: schema.WorkspaceRoleViewer, FullAccess: true, WorkspaceIDs: []uuid.UUID{ws}},
		"scoped without workspaces":     {Role: schema.WorkspaceRoleViewer},
		"nil workspace":                 {Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{uuid.Nil}},
		"duplicate workspace":           {Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{ws, ws}},
		"privilege that does not exist": {Role: schema.WorkspaceRoleViewer, FullAccess: true, Privileges: []schema.KeyPrivilege{"read_everything"}},
	}
	for name, spec := range invalid {
		assert.Error(t, Validate(spec), name)
	}
}

func TestListsNeverNil(t *testing.T) {
	assert.NotNil(t, Privileges(Spec{}))
	assert.NotNil(t, WorkspaceIDs(Spec{}))
}
