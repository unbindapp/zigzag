package initiative_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/initiative"
	"github.com/unbindapp/zigzag-api/ent/initiativepost"
	"github.com/unbindapp/zigzag-api/ent/project"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type ListInitiativesFilter struct {
	WorkspaceID     uuid.UUID
	Status          *schema.InitiativeStatus
	OwnerID         *uuid.UUID
	IncludeArchived bool
	IncludeTrashed  bool
}

func withInitiativeEdges(q *ent.InitiativeQuery) *ent.InitiativeQuery {
	return q.WithOwner().WithProjects(func(q *ent.ProjectQuery) {
		q.Where(project.TrashedAtIsNil()).Order(project.BySortOrder(), project.ByName())
	})
}

func (self *InitiativeRepository) List(ctx context.Context, filter *ListInitiativesFilter) ([]*ent.Initiative, error) {
	q := withInitiativeEdges(self.base.DB.Initiative.Query().Where(initiative.WorkspaceID(filter.WorkspaceID)))
	if !filter.IncludeArchived {
		q.Where(initiative.ArchivedAtIsNil())
	}
	if !filter.IncludeTrashed {
		q.Where(initiative.TrashedAtIsNil())
	}
	if filter.Status != nil {
		q.Where(initiative.StatusEQ(*filter.Status))
	}
	if filter.OwnerID != nil {
		q.Where(initiative.OwnerID(*filter.OwnerID))
	}
	return q.Order(initiative.BySortOrder(), initiative.ByCreatedAt()).All(ctx)
}

func (self *InitiativeRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Initiative, error) {
	return withInitiativeEdges(self.base.DB.Initiative.Query().Where(initiative.ID(id))).Only(ctx)
}

func (self *InitiativeRepository) GetBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*ent.Initiative, error) {
	return withInitiativeEdges(self.base.DB.Initiative.Query().Where(initiative.WorkspaceID(workspaceID), initiative.Slug(slug))).Only(ctx)
}

func (self *InitiativeRepository) ResolveWorkspaceOfInitiative(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Initiative.Query().Where(initiative.ID(id)).Select(initiative.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *InitiativeRepository) MaxSortOrder(ctx context.Context, workspaceID uuid.UUID) (float64, error) {
	last, err := self.base.DB.Initiative.Query().
		Where(initiative.WorkspaceID(workspaceID)).
		Order(initiative.BySortOrder(sql.OrderDesc())).
		Select(initiative.FieldSortOrder).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder, nil
}

func (self *InitiativeRepository) CountProjectsInWorkspace(ctx context.Context, workspaceID uuid.UUID, projectIDs []uuid.UUID) (int, error) {
	return self.base.DB.Project.Query().Where(project.WorkspaceID(workspaceID), project.IDIn(projectIDs...)).Count(ctx)
}

func (self *InitiativeRepository) LatestHealths(ctx context.Context, initiativeIDs []uuid.UUID) (map[uuid.UUID]schema.Health, error) {
	healths := map[uuid.UUID]schema.Health{}
	if len(initiativeIDs) == 0 {
		return healths, nil
	}
	posts, err := self.base.DB.InitiativePost.Query().
		Where(initiativepost.InitiativeIDIn(initiativeIDs...)).
		Order(initiativepost.ByCreatedAt(sql.OrderDesc())).
		Select(initiativepost.FieldInitiativeID, initiativepost.FieldHealth).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, post := range posts {
		if _, ok := healths[post.InitiativeID]; !ok {
			healths[post.InitiativeID] = post.Health
		}
	}
	return healths, nil
}
