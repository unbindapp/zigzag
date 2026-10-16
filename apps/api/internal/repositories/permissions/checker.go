package permissions_repo

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/workspace"
	"github.com/unbindapp/zigzag-api/ent/workspacemember"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
)

type PermissionCheck struct {
	Role        schema.WorkspaceRole
	WorkspaceID uuid.UUID
}

// Check passes when any one check passes.
func (self *PermissionsRepository) Check(ctx context.Context, userID uuid.UUID, checks []PermissionCheck) error {
	if IsSystemCaller(ctx) {
		return nil
	}
	if len(checks) == 0 {
		return errdefs.ErrUnauthorized
	}

	caller, err := self.base.DB.User.Get(ctx, userID)
	if ent.IsNotFound(err) {
		return errdefs.ErrUnauthorized
	}
	if err != nil {
		return fmt.Errorf("error fetching user: %w", err)
	}
	if !caller.Active {
		return errdefs.ErrUnauthorized
	}

	key, narrowed := APIKeyAccessFromContext(ctx)
	ids := make([]uuid.UUID, 0, len(checks))
	for _, check := range checks {
		ids = append(ids, check.WorkspaceID)
	}
	memberships, err := self.base.DB.WorkspaceMember.Query().
		Where(workspacemember.UserID(userID), workspacemember.WorkspaceIDIn(ids...)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("error fetching memberships: %w", err)
	}
	roles := make(map[uuid.UUID]schema.WorkspaceRole, len(memberships))
	for _, membership := range memberships {
		roles[membership.WorkspaceID] = membership.Role
	}

	for _, check := range checks {
		if check.WorkspaceID == uuid.Nil || !IsValidRole(check.Role) {
			continue
		}
		if narrowed && !key.allows(check.Role, check.WorkspaceID) {
			continue
		}
		if caller.SystemAdmin {
			exists, err := self.base.DB.Workspace.Query().Where(workspace.ID(check.WorkspaceID)).Exist(ctx)
			if err != nil {
				return fmt.Errorf("error checking workspace: %w", err)
			}
			if exists {
				return nil
			}
			continue
		}
		if role, member := roles[check.WorkspaceID]; member && RoleImplies(role, check.Role) {
			return nil
		}
	}
	return errdefs.ErrUnauthorized
}

// CheckSystemAdmin refuses narrowed credentials unless they carry full admin access.
func (self *PermissionsRepository) CheckSystemAdmin(ctx context.Context, userID uuid.UUID) error {
	if IsSystemCaller(ctx) {
		return nil
	}
	if key, narrowed := APIKeyAccessFromContext(ctx); narrowed && (!key.FullAccess || key.Role != schema.WorkspaceRoleAdmin) {
		return errdefs.ErrUnauthorized
	}
	caller, err := self.base.DB.User.Get(ctx, userID)
	if ent.IsNotFound(err) {
		return errdefs.ErrUnauthorized
	}
	if err != nil {
		return fmt.Errorf("error fetching user: %w", err)
	}
	if !caller.SystemAdmin || !caller.Active {
		return errdefs.ErrUnauthorized
	}
	return nil
}

type PermissionSet struct {
	IsSystemAdmin bool
	roles         map[uuid.UUID]schema.WorkspaceRole
}

func (self *PermissionSet) WorkspaceRole(workspaceID uuid.UUID) (schema.WorkspaceRole, bool) {
	role, ok := self.roles[workspaceID]
	return role, ok
}

func (self *PermissionSet) Can(role schema.WorkspaceRole, workspaceID uuid.UUID) bool {
	held, ok := self.roles[workspaceID]
	return ok && RoleImplies(held, role)
}

func (self *PermissionSet) VisibleWorkspaceIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(self.roles))
	for id := range self.roles {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	return ids
}

func (self *PermissionsRepository) GetUserPermissionSet(ctx context.Context, userID uuid.UUID) (*PermissionSet, error) {
	caller, err := self.base.DB.User.Get(ctx, userID)
	if ent.IsNotFound(err) {
		return &PermissionSet{roles: map[uuid.UUID]schema.WorkspaceRole{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	set := &PermissionSet{IsSystemAdmin: caller.SystemAdmin && caller.Active, roles: map[uuid.UUID]schema.WorkspaceRole{}}
	if !caller.Active {
		return set, nil
	}

	set.roles, err = self.heldRoles(ctx, userID, set.IsSystemAdmin)
	if err != nil {
		return nil, err
	}

	key, narrowed := APIKeyAccessFromContext(ctx)
	if !narrowed {
		return set, nil
	}
	if !key.FullAccess || key.Role != schema.WorkspaceRoleAdmin {
		set.IsSystemAdmin = false
	}
	for id, role := range set.roles {
		if !key.Reaches(id) || !IsValidRole(key.Role) {
			delete(set.roles, id)
			continue
		}
		set.roles[id] = WeakerRole(role, key.Role)
	}
	return set, nil
}

func (self *PermissionsRepository) heldRoles(ctx context.Context, userID uuid.UUID, systemAdmin bool) (map[uuid.UUID]schema.WorkspaceRole, error) {
	roles := map[uuid.UUID]schema.WorkspaceRole{}
	if systemAdmin {
		ids, err := self.base.DB.Workspace.Query().IDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing workspaces: %w", err)
		}
		for _, id := range ids {
			roles[id] = schema.WorkspaceRoleAdmin
		}
		return roles, nil
	}
	memberships, err := self.base.DB.WorkspaceMember.Query().Where(workspacemember.UserID(userID)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching memberships: %w", err)
	}
	for _, membership := range memberships {
		roles[membership.WorkspaceID] = membership.Role
	}
	return roles, nil
}

func NewPermissionSet(isSystemAdmin bool, roles map[uuid.UUID]schema.WorkspaceRole) *PermissionSet {
	if roles == nil {
		roles = map[uuid.UUID]schema.WorkspaceRole{}
	}
	return &PermissionSet{IsSystemAdmin: isSystemAdmin, roles: roles}
}
