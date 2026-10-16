package workspace_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/workspace"
	"github.com/unbindapp/zigzag-api/ent/workspacemember"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

func (self *WorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Workspace, error) {
	return self.base.DB.Workspace.Get(ctx, id)
}

func (self *WorkspaceRepository) GetBySlug(ctx context.Context, slug string) (*ent.Workspace, error) {
	return self.base.DB.Workspace.Query().Where(workspace.Slug(slug)).Only(ctx)
}

func (self *WorkspaceRepository) SlugExists(ctx context.Context, tx repository.TxInterface, slug string, except *uuid.UUID) (bool, error) {
	q := self.base.Client(tx).Workspace.Query().Where(workspace.Slug(slug))
	if except != nil {
		q = q.Where(workspace.IDNEQ(*except))
	}
	return q.Exist(ctx)
}

func (self *WorkspaceRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.Workspace, error) {
	return self.base.DB.Workspace.Query().
		Where(workspace.IDIn(ids...), workspace.ArchivedAtIsNil()).
		Order(ent.Asc(workspace.FieldName)).
		All(ctx)
}

func (self *WorkspaceRepository) AllIDs(ctx context.Context) ([]uuid.UUID, error) {
	return self.base.DB.Workspace.Query().Where(workspace.ArchivedAtIsNil()).IDs(ctx)
}

type WorkspaceCounts struct {
	Members int
	Teams   int
}

func (self *WorkspaceRepository) Counts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]WorkspaceCounts, error) {
	counts := make(map[uuid.UUID]WorkspaceCounts, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}
	var members []struct {
		WorkspaceID uuid.UUID `json:"workspace_id"`
		Count       int       `json:"count"`
	}
	if err := self.base.DB.WorkspaceMember.Query().
		Where(workspacemember.WorkspaceIDIn(ids...)).
		GroupBy(workspacemember.FieldWorkspaceID).
		Aggregate(ent.Count()).
		Scan(ctx, &members); err != nil {
		return nil, err
	}
	var teams []struct {
		WorkspaceID uuid.UUID `json:"workspace_id"`
		Count       int       `json:"count"`
	}
	if err := self.base.DB.Team.Query().
		Where(team.WorkspaceIDIn(ids...), team.ArchivedAtIsNil()).
		GroupBy(team.FieldWorkspaceID).
		Aggregate(ent.Count()).
		Scan(ctx, &teams); err != nil {
		return nil, err
	}
	for _, row := range members {
		c := counts[row.WorkspaceID]
		c.Members = row.Count
		counts[row.WorkspaceID] = c
	}
	for _, row := range teams {
		c := counts[row.WorkspaceID]
		c.Teams = row.Count
		counts[row.WorkspaceID] = c
	}
	return counts, nil
}

func orderedByJoin(s *sql.Selector) {
	s.OrderBy(sql.Asc(s.C(workspacemember.FieldJoinedAt)))
}
