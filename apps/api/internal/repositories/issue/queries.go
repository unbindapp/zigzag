package issue_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/attachment"
	"github.com/unbindapp/zigzag-api/ent/comment"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/issuemove"
	"github.com/unbindapp/zigzag-api/ent/label"
	"github.com/unbindapp/zigzag-api/ent/projectmilestone"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

const maxParentDepth = 64

func withIssueEdges(query *ent.IssueQuery) *ent.IssueQuery {
	return query.
		WithTeam().
		WithWorkspace().
		WithState().
		WithAssignee().
		WithCreator().
		WithLabels(func(labels *ent.LabelQuery) { labels.Order(ent.Asc(label.FieldName)) }).
		WithProject().
		WithMilestone().
		WithCycle().
		WithParent(func(parent *ent.IssueQuery) { parent.WithTeam().WithState() })
}

func (self *IssueRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error) {
	return self.base.DB.Issue.Get(ctx, id)
}

func (self *IssueRepository) GetWithEdges(ctx context.Context, tx repository.TxInterface, id uuid.UUID) (*ent.Issue, error) {
	return withIssueEdges(self.base.Client(tx).Issue.Query().Where(issue.ID(id))).Only(ctx)
}

func (self *IssueRepository) LockWithEdges(ctx context.Context, tx repository.TxInterface, id uuid.UUID) (*ent.Issue, error) {
	return withIssueEdges(self.base.Client(tx).Issue.Query().Where(issue.ID(id), repository.LockRowsOnPostgres(issue.Table)).Unique(false)).Only(ctx)
}

// GetManyWithEdges keeps the order of ids and skips ids that do not exist.
func (self *IssueRepository) GetManyWithEdges(ctx context.Context, tx repository.TxInterface, ids []uuid.UUID) ([]*ent.Issue, error) {
	found, err := withIssueEdges(self.base.Client(tx).Issue.Query().Where(issue.IDIn(ids...))).All(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*ent.Issue, len(found))
	for _, entity := range found {
		byID[entity.ID] = entity
	}
	ordered := make([]*ent.Issue, 0, len(found))
	for _, id := range ids {
		if entity, ok := byID[id]; ok {
			ordered = append(ordered, entity)
			delete(byID, id)
		}
	}
	return ordered, nil
}

func (self *IssueRepository) ResolveWorkspaceOfIssue(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Issue.Query().Where(issue.ID(id)).Select(issue.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *IssueRepository) FindByIdentifier(ctx context.Context, workspaceIDs []uuid.UUID, key string, number int) ([]uuid.UUID, error) {
	if len(workspaceIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	teamOf := team.And(team.WorkspaceIDIn(workspaceIDs...), team.Key(key))
	current, err := self.base.DB.Issue.Query().
		Where(issue.Number(number), issue.HasTeamWith(teamOf)).
		IDs(ctx)
	if err != nil {
		return nil, err
	}
	if len(current) > 0 {
		return current, nil
	}
	moves, err := self.base.DB.IssueMove.Query().
		Where(issuemove.OldNumber(number), issuemove.HasOldTeamWith(teamOf)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(moves))
	for i, move := range moves {
		ids[i] = move.IssueID
	}
	return ids, nil
}

type parentCount struct {
	ParentID uuid.UUID `json:"parent_id"`
	Count    int       `json:"count"`
}

type issueCount struct {
	IssueID uuid.UUID `json:"issue_id"`
	Count   int       `json:"count"`
}

func (self *IssueRepository) Counts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.IssueCounts, error) {
	counts := make(map[uuid.UUID]models.IssueCounts, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}
	var children, completed []parentCount
	var comments, attachments []issueCount
	err := self.base.DB.Issue.Query().
		Where(issue.ParentIDIn(ids...), issue.TrashedAtIsNil()).
		GroupBy(issue.FieldParentID).
		Aggregate(ent.As(ent.Count(), "count")).
		Scan(ctx, &children)
	if err != nil {
		return nil, err
	}
	err = self.base.DB.Issue.Query().
		Where(issue.ParentIDIn(ids...), issue.TrashedAtIsNil(), issue.HasStateWith(workflowstate.TypeIn(schema.StateTypeCompleted, schema.StateTypeCanceled))).
		GroupBy(issue.FieldParentID).
		Aggregate(ent.As(ent.Count(), "count")).
		Scan(ctx, &completed)
	if err != nil {
		return nil, err
	}
	err = self.base.DB.Comment.Query().
		Where(comment.IssueIDIn(ids...)).
		GroupBy(comment.FieldIssueID).
		Aggregate(ent.As(ent.Count(), "count")).
		Scan(ctx, &comments)
	if err != nil {
		return nil, err
	}
	err = self.base.DB.Attachment.Query().
		Where(attachment.IssueIDIn(ids...)).
		GroupBy(attachment.FieldIssueID).
		Aggregate(ent.As(ent.Count(), "count")).
		Scan(ctx, &attachments)
	if err != nil {
		return nil, err
	}
	for _, row := range children {
		entry := counts[row.ParentID]
		entry.Children = row.Count
		counts[row.ParentID] = entry
	}
	for _, row := range completed {
		entry := counts[row.ParentID]
		entry.CompletedChildren = row.Count
		counts[row.ParentID] = entry
	}
	for _, row := range comments {
		entry := counts[row.IssueID]
		entry.Comments = row.Count
		counts[row.IssueID] = entry
	}
	for _, row := range attachments {
		entry := counts[row.IssueID]
		entry.Attachments = row.Count
		counts[row.IssueID] = entry
	}
	return counts, nil
}

func (self *IssueRepository) Children(ctx context.Context, parentID uuid.UUID) ([]*ent.Issue, error) {
	return withIssueEdges(self.base.DB.Issue.Query().
		Where(issue.ParentID(parentID), issue.TrashedAtIsNil())).
		Order(issue.BySubIssueSortOrder(), issue.ByNumber()).
		All(ctx)
}

func (self *IssueRepository) ChildIDs(ctx context.Context, tx repository.TxInterface, parentID uuid.UUID) ([]uuid.UUID, error) {
	return self.base.Client(tx).Issue.Query().Where(issue.ParentID(parentID)).IDs(ctx)
}

func (self *IssueRepository) AncestorIDs(ctx context.Context, tx repository.TxInterface, id uuid.UUID) ([]uuid.UUID, error) {
	ancestors := []uuid.UUID{}
	current := id
	for range maxParentDepth {
		found, err := self.base.Client(tx).Issue.Query().Where(issue.ID(current)).Select(issue.FieldParentID).Only(ctx)
		if err != nil {
			return nil, err
		}
		if found.ParentID == nil {
			return ancestors, nil
		}
		ancestors = append(ancestors, *found.ParentID)
		current = *found.ParentID
	}
	return ancestors, nil
}

func (self *IssueRepository) TopSortOrder(ctx context.Context, tx repository.TxInterface, stateID uuid.UUID) (float64, error) {
	first, err := self.base.Client(tx).Issue.Query().
		Where(issue.StateID(stateID)).
		Order(issue.BySortOrder()).
		Select(issue.FieldSortOrder).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return first.SortOrder - 1, nil
}

func (self *IssueRepository) LabelsByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.Label, error) {
	if len(ids) == 0 {
		return []*ent.Label{}, nil
	}
	return self.base.DB.Label.Query().Where(label.IDIn(ids...)).WithParent().All(ctx)
}

func (self *IssueRepository) GetMilestone(ctx context.Context, id uuid.UUID) (*ent.ProjectMilestone, error) {
	return self.base.DB.ProjectMilestone.Query().Where(projectmilestone.ID(id)).Only(ctx)
}

func (self *IssueRepository) GetProject(ctx context.Context, id uuid.UUID) (*ent.Project, error) {
	return self.base.DB.Project.Get(ctx, id)
}
