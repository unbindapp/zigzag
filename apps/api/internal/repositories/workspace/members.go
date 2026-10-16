package workspace_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/team"
	"github.com/zigzagapp-org/zigzag-api/ent/teammember"
	"github.com/zigzagapp-org/zigzag-api/ent/workspacemember"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

func (self *WorkspaceRepository) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*ent.WorkspaceMember, error) {
	return self.base.DB.WorkspaceMember.Query().
		Where(workspacemember.WorkspaceID(workspaceID)).
		WithUser().
		Order(orderedByJoin).
		All(ctx)
}

func (self *WorkspaceRepository) GetMember(ctx context.Context, tx repository.TxInterface, workspaceID, userID uuid.UUID) (*ent.WorkspaceMember, error) {
	return self.base.Client(tx).WorkspaceMember.Query().
		Where(workspacemember.WorkspaceID(workspaceID), workspacemember.UserID(userID)).
		WithUser().
		Only(ctx)
}

func (self *WorkspaceRepository) MembershipsOfUser(ctx context.Context, userID uuid.UUID) ([]*ent.WorkspaceMember, error) {
	return self.base.DB.WorkspaceMember.Query().Where(workspacemember.UserID(userID)).Order(orderedByJoin).All(ctx)
}

// AddMember keeps the existing role when the user is already a member.
func (self *WorkspaceRepository) AddMember(ctx context.Context, tx repository.TxInterface, workspaceID, userID uuid.UUID, role schema.WorkspaceRole) (*ent.WorkspaceMember, error) {
	db := self.base.Client(tx)
	err := db.WorkspaceMember.Create().
		SetWorkspaceID(workspaceID).
		SetUserID(userID).
		SetRole(role).
		OnConflictColumns(workspacemember.FieldWorkspaceID, workspacemember.FieldUserID).
		DoNothing().
		Exec(ctx)
	if err := repository.IgnoreExisting(err); err != nil {
		return nil, err
	}
	return self.GetMember(ctx, tx, workspaceID, userID)
}

func (self *WorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role schema.WorkspaceRole) (*ent.WorkspaceMember, error) {
	updated, err := self.base.DB.WorkspaceMember.Update().
		Where(workspacemember.WorkspaceID(workspaceID), workspacemember.UserID(userID)).
		SetRole(role).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if updated == 0 {
		return nil, &ent.NotFoundError{}
	}
	return self.GetMember(ctx, nil, workspaceID, userID)
}

// RemoveMember also takes the user off every team of the workspace.
func (self *WorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		db := tx.Client()
		deleted, err := db.WorkspaceMember.Delete().
			Where(workspacemember.WorkspaceID(workspaceID), workspacemember.UserID(userID)).
			Exec(ctx)
		if err != nil {
			return err
		}
		if deleted == 0 {
			return &ent.NotFoundError{}
		}
		_, err = db.TeamMember.Delete().
			Where(teammember.UserID(userID), teammember.HasTeamWith(team.WorkspaceID(workspaceID))).
			Exec(ctx)
		return err
	})
}

func (self *WorkspaceRepository) CountAdmins(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	return self.base.DB.WorkspaceMember.Query().
		Where(workspacemember.WorkspaceID(workspaceID), workspacemember.RoleEQ(schema.WorkspaceRoleAdmin)).
		Count(ctx)
}

func (self *WorkspaceRepository) MemberTeamIDs(ctx context.Context, workspaceID uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	memberships, err := self.base.DB.TeamMember.Query().
		Where(teammember.HasTeamWith(team.WorkspaceID(workspaceID))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	byUser := map[uuid.UUID][]uuid.UUID{}
	for _, membership := range memberships {
		byUser[membership.UserID] = append(byUser[membership.UserID], membership.TeamID)
	}
	return byUser, nil
}
