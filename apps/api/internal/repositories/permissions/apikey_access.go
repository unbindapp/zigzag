package permissions_repo

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type APIKeyAccess struct {
	Role         schema.WorkspaceRole
	FullAccess   bool
	WorkspaceIDs []uuid.UUID
	Privileges   []schema.KeyPrivilege
}

type apiKeyAccessKey struct{}

func APIKeyAccessOf(key *ent.APIKey) APIKeyAccess {
	return APIKeyAccess{Role: key.Role, FullAccess: key.FullAccess, WorkspaceIDs: key.WorkspaceIds, Privileges: key.Privileges}
}

func WithAPIKeyAccess(ctx context.Context, access APIKeyAccess) context.Context {
	return context.WithValue(ctx, apiKeyAccessKey{}, access)
}

func APIKeyAccessFromContext(ctx context.Context) (APIKeyAccess, bool) {
	access, ok := ctx.Value(apiKeyAccessKey{}).(APIKeyAccess)
	return access, ok
}

func (a APIKeyAccess) AllowsWrites() bool {
	return RoleImplies(a.Role, schema.WorkspaceRoleMember)
}

func (a APIKeyAccess) Has(privilege schema.KeyPrivilege) bool {
	return slices.Contains(a.Privileges, privilege)
}

func (a APIKeyAccess) Reaches(workspaceID uuid.UUID) bool {
	return a.FullAccess || slices.Contains(a.WorkspaceIDs, workspaceID)
}

func (a APIKeyAccess) allows(required schema.WorkspaceRole, workspaceID uuid.UUID) bool {
	return RoleImplies(a.Role, required) && a.Reaches(workspaceID)
}

func HasPrivilege(ctx context.Context, privilege schema.KeyPrivilege) bool {
	access, limited := APIKeyAccessFromContext(ctx)
	if !limited {
		return true
	}
	return access.Has(privilege)
}
