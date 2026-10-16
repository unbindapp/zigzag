package keyaccess

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

type Spec struct {
	Role         schema.WorkspaceRole
	FullAccess   bool
	WorkspaceIDs []uuid.UUID
	Privileges   []schema.KeyPrivilege
}

func Validate(spec Spec) error {
	if !permissions_repo.IsValidRole(spec.Role) {
		return fmt.Errorf("unknown role %q, use admin, member or viewer", spec.Role)
	}
	if spec.FullAccess && len(spec.WorkspaceIDs) > 0 {
		return errors.New("workspace_ids must be empty when full_access is true")
	}
	if !spec.FullAccess && len(spec.WorkspaceIDs) == 0 {
		return errors.New("at least one workspace is required unless full_access is true")
	}
	seen := map[uuid.UUID]struct{}{}
	for i, id := range spec.WorkspaceIDs {
		if id == uuid.Nil {
			return fmt.Errorf("workspace_ids[%d]: workspace id is required", i)
		}
		if _, dup := seen[id]; dup {
			return fmt.Errorf("workspace_ids[%d]: duplicate workspace", i)
		}
		seen[id] = struct{}{}
	}
	return validatePrivileges(spec)
}

func validatePrivileges(spec Spec) error {
	seen := map[schema.KeyPrivilege]struct{}{}
	for i, privilege := range spec.Privileges {
		if !slices.Contains(schema.AllPrivileges(), privilege) {
			return fmt.Errorf("privileges[%d]: unknown privilege %q", i, privilege)
		}
		if _, dup := seen[privilege]; dup {
			return fmt.Errorf("privileges[%d]: duplicate privilege", i)
		}
		seen[privilege] = struct{}{}
	}
	return nil
}

func Privileges(spec Spec) []schema.KeyPrivilege {
	if spec.Privileges == nil {
		return []schema.KeyPrivilege{}
	}
	return spec.Privileges
}

func WorkspaceIDs(spec Spec) []uuid.UUID {
	if spec.WorkspaceIDs == nil {
		return []uuid.UUID{}
	}
	return spec.WorkspaceIDs
}

// RequesterHolds checks listed workspaces only; full access is intersected with the owner's roles on every request.
func RequesterHolds(ctx context.Context, perms permissions_repo.PermissionsRepositoryInterface, requesterUserID uuid.UUID, spec Spec) error {
	for _, workspaceID := range spec.WorkspaceIDs {
		err := perms.Check(ctx, requesterUserID, []permissions_repo.PermissionCheck{{Role: spec.Role, WorkspaceID: workspaceID}})
		if errors.Is(err, errdefs.ErrUnauthorized) {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, fmt.Sprintf("you do not have %s access to workspace %s", spec.Role, workspaceID))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
