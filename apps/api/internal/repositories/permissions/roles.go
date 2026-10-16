package permissions_repo

import "github.com/zigzagapp-org/zigzag-api/ent/schema"

func roleRank(role schema.WorkspaceRole) int {
	switch role {
	case schema.WorkspaceRoleAdmin:
		return 3
	case schema.WorkspaceRoleMember:
		return 2
	case schema.WorkspaceRoleViewer:
		return 1
	default:
		return 0
	}
}

func RoleImplies(held, required schema.WorkspaceRole) bool {
	rank := roleRank(held)
	return rank > 0 && rank >= roleRank(required)
}

func WeakerRole(a, b schema.WorkspaceRole) schema.WorkspaceRole {
	if roleRank(a) <= roleRank(b) {
		return a
	}
	return b
}

func IsValidRole(role schema.WorkspaceRole) bool {
	return roleRank(role) > 0
}
