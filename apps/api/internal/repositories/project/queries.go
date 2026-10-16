package project_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/initiative"
	"github.com/zigzagapp-org/zigzag-api/ent/label"
	"github.com/zigzagapp-org/zigzag-api/ent/project"
	"github.com/zigzagapp-org/zigzag-api/ent/projectlink"
	"github.com/zigzagapp-org/zigzag-api/ent/projectmilestone"
	"github.com/zigzagapp-org/zigzag-api/ent/projectpost"
	"github.com/zigzagapp-org/zigzag-api/ent/projectstatus"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/team"
	"github.com/zigzagapp-org/zigzag-api/ent/user"
)

type ProjectOrder string

const (
	OrderBySortOrder  ProjectOrder = "sort_order"
	OrderByName       ProjectOrder = "name"
	OrderByPriority   ProjectOrder = "priority"
	OrderByTargetDate ProjectOrder = "target_date"
	OrderByCreatedAt  ProjectOrder = "created_at"
	OrderByUpdatedAt  ProjectOrder = "updated_at"
)

type ListProjectsFilter struct {
	WorkspaceID     uuid.UUID
	TeamID          *uuid.UUID
	StatusType      *schema.ProjectStatusType
	StatusID        *uuid.UUID
	LeadID          *uuid.UUID
	MemberID        *uuid.UUID
	InitiativeID    *uuid.UUID
	LabelID         *uuid.UUID
	Query           string
	IncludeArchived bool
	IncludeTrashed  bool
	OrderBy         ProjectOrder
}

func withProjectEdges(q *ent.ProjectQuery) *ent.ProjectQuery {
	return q.
		WithStatus().
		WithLead().
		WithTeams(func(q *ent.TeamQuery) { q.Order(team.ByName()) }).
		WithMembers(func(q *ent.UserQuery) { q.Order(user.ByName()) }).
		WithLabels(func(q *ent.LabelQuery) { q.Order(label.ByName()) }).
		WithInitiatives(func(q *ent.InitiativeQuery) { q.Select(initiative.FieldID) }).
		WithMilestones(func(q *ent.ProjectMilestoneQuery) {
			q.Where(projectmilestone.ArchivedAtIsNil()).Order(milestoneOrder...)
		}).
		WithLinks(func(q *ent.ProjectLinkQuery) { q.Order(projectlink.BySortOrder(), projectlink.ByCreatedAt()) })
}

var milestoneOrder = []projectmilestone.OrderOption{
	projectmilestone.BySortOrder(),
	projectmilestone.ByTargetDate(sql.OrderNullsLast()),
	projectmilestone.ByCreatedAt(),
}

func (self *ProjectRepository) List(ctx context.Context, filter *ListProjectsFilter) ([]*ent.Project, error) {
	q := withProjectEdges(self.base.DB.Project.Query().Where(project.WorkspaceID(filter.WorkspaceID)))
	if !filter.IncludeArchived {
		q.Where(project.ArchivedAtIsNil())
	}
	if !filter.IncludeTrashed {
		q.Where(project.TrashedAtIsNil())
	}
	if filter.TeamID != nil {
		q.Where(project.HasTeamsWith(team.ID(*filter.TeamID)))
	}
	if filter.StatusType != nil {
		q.Where(project.HasStatusWith(projectstatus.TypeEQ(*filter.StatusType)))
	}
	if filter.StatusID != nil {
		q.Where(project.StatusID(*filter.StatusID))
	}
	if filter.LeadID != nil {
		q.Where(project.LeadID(*filter.LeadID))
	}
	if filter.MemberID != nil {
		q.Where(project.HasMembersWith(user.ID(*filter.MemberID)))
	}
	if filter.InitiativeID != nil {
		q.Where(project.HasInitiativesWith(initiative.ID(*filter.InitiativeID)))
	}
	if filter.LabelID != nil {
		q.Where(project.HasLabelsWith(label.ID(*filter.LabelID)))
	}
	if filter.Query != "" {
		q.Where(project.Or(project.NameContainsFold(filter.Query), project.SummaryContainsFold(filter.Query)))
	}
	return q.Order(projectOrder(filter.OrderBy)...).All(ctx)
}

func projectOrder(order ProjectOrder) []project.OrderOption {
	switch order {
	case OrderByName:
		return []project.OrderOption{project.ByName(), project.ByCreatedAt()}
	case OrderByPriority:
		return []project.OrderOption{func(s *sql.Selector) {
			s.OrderExpr(sql.Expr("CASE WHEN " + s.C(project.FieldPriority) + " = 0 THEN 5 ELSE " + s.C(project.FieldPriority) + " END"))
		}, project.ByName()}
	case OrderByTargetDate:
		return []project.OrderOption{project.ByTargetDate(sql.OrderNullsLast()), project.ByName()}
	case OrderByCreatedAt:
		return []project.OrderOption{project.ByCreatedAt(sql.OrderDesc())}
	case OrderByUpdatedAt:
		return []project.OrderOption{project.ByUpdatedAt(sql.OrderDesc())}
	default:
		return []project.OrderOption{project.BySortOrder(), project.ByCreatedAt()}
	}
}

func (self *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Project, error) {
	return withProjectEdges(self.base.DB.Project.Query().Where(project.ID(id))).Only(ctx)
}

func (self *ProjectRepository) GetBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*ent.Project, error) {
	return withProjectEdges(self.base.DB.Project.Query().Where(project.WorkspaceID(workspaceID), project.Slug(slug))).Only(ctx)
}

func (self *ProjectRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*ent.Project, error) {
	return withProjectEdges(self.base.DB.Project.Query().
		Where(project.WorkspaceID(workspaceID), project.NameEqualFold(name), project.TrashedAtIsNil()).
		Order(ent.Asc(project.FieldCreatedAt))).
		First(ctx)
}

func (self *ProjectRepository) ResolveWorkspaceOfProject(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Project.Query().Where(project.ID(id)).Select(project.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *ProjectRepository) AudienceIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	found, err := self.base.DB.Project.Query().
		Where(project.ID(projectID)).
		WithMembers(func(q *ent.UserQuery) { q.Select(user.FieldID) }).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	ids := []uuid.UUID{}
	seen := map[uuid.UUID]bool{}
	add := func(id uuid.UUID) {
		if seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if found.LeadID != nil {
		add(*found.LeadID)
	}
	for _, member := range found.Edges.Members {
		add(member.ID)
	}
	return ids, nil
}

func (self *ProjectRepository) MaxSortOrder(ctx context.Context, workspaceID uuid.UUID) (float64, error) {
	last, err := self.base.DB.Project.Query().
		Where(project.WorkspaceID(workspaceID)).
		Order(project.BySortOrder(sql.OrderDesc())).
		Select(project.FieldSortOrder).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder, nil
}

func (self *ProjectRepository) LatestUpdate(ctx context.Context, projectID uuid.UUID) (*ent.ProjectPost, error) {
	found, err := withUpdateEdges(self.base.DB.ProjectPost.Query().Where(projectpost.ProjectID(projectID))).
		Order(projectpost.ByCreatedAt(sql.OrderDesc())).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return found, err
}
