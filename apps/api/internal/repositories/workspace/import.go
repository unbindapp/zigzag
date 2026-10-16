package workspace_repo

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/workspacemember"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type ImportedMember struct {
	UserID uuid.UUID
	Role   schema.WorkspaceRole
}

// AddMembers keeps the role of users who are members already.
func (self *WorkspaceRepository) AddMembers(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, members []ImportedMember) error {
	db := self.base.Client(tx)
	builders := make([]*ent.WorkspaceMemberCreate, len(members))
	for i, member := range members {
		builders[i] = db.WorkspaceMember.Create().SetWorkspaceID(workspaceID).SetUserID(member.UserID).SetRole(member.Role)
	}
	for batch := range slices.Chunk(builders, 500) {
		err := db.WorkspaceMember.CreateBulk(batch...).
			OnConflictColumns(workspacemember.FieldWorkspaceID, workspacemember.FieldUserID).
			DoNothing().
			Exec(ctx)
		if err := repository.IgnoreExisting(err); err != nil {
			return err
		}
	}
	return nil
}
