package label_repo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	"github.com/zigzagapp-org/zigzag-api/ent/label"
	"github.com/zigzagapp-org/zigzag-api/ent/predicate"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

func (self *LabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Label, error) {
	return self.base.DB.Label.Get(ctx, id)
}

func (self *LabelRepository) ResolveWorkspaceOfLabel(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Label.Query().Where(label.ID(id)).Select(label.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *LabelRepository) List(ctx context.Context, workspaceID uuid.UUID, teamIDs []uuid.UUID, includeArchived bool) ([]*ent.Label, error) {
	scope := []predicate.Label{label.TeamIDIsNil()}
	if len(teamIDs) > 0 {
		scope = append(scope, label.TeamIDIn(teamIDs...))
	}
	q := self.base.DB.Label.Query().Where(label.WorkspaceID(workspaceID), label.Or(scope...))
	if !includeArchived {
		q = q.Where(label.ArchivedAtIsNil())
	}
	return q.Order(ent.Asc(label.FieldSortOrder), ent.Asc(label.FieldName)).All(ctx)
}

func (self *LabelRepository) GetMany(ctx context.Context, ids []uuid.UUID) ([]*ent.Label, error) {
	if len(ids) == 0 {
		return []*ent.Label{}, nil
	}
	return self.base.DB.Label.Query().Where(label.IDIn(ids...)).All(ctx)
}

func (self *LabelRepository) Children(ctx context.Context, parentID uuid.UUID) ([]*ent.Label, error) {
	return self.base.DB.Label.Query().Where(label.ParentID(parentID)).All(ctx)
}

func siblingScope(workspaceID uuid.UUID, teamID, parentID *uuid.UUID) []predicate.Label {
	where := []predicate.Label{label.WorkspaceID(workspaceID)}
	if teamID == nil {
		where = append(where, label.TeamIDIsNil())
	} else {
		where = append(where, label.TeamID(*teamID))
	}
	if parentID == nil {
		where = append(where, label.ParentIDIsNil())
	} else {
		where = append(where, label.ParentID(*parentID))
	}
	return where
}

func (self *LabelRepository) NameTaken(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, teamID, parentID *uuid.UUID, name string, except *uuid.UUID) (bool, error) {
	q := self.base.Client(tx).Label.Query().Where(siblingScope(workspaceID, teamID, parentID)...)
	if except != nil {
		q = q.Where(label.IDNEQ(*except))
	}
	names, err := q.Select(label.FieldName).Strings(ctx)
	if err != nil {
		return false, err
	}
	for _, existing := range names {
		if strings.EqualFold(existing, name) {
			return true, nil
		}
	}
	return false, nil
}

func (self *LabelRepository) NextSortOrder(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, teamID, parentID *uuid.UUID) (float64, error) {
	last, err := self.base.Client(tx).Label.Query().
		Where(siblingScope(workspaceID, teamID, parentID)...).
		Order(ent.Desc(label.FieldSortOrder)).
		First(ctx)
	if ent.IsNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder + 1, nil
}

func (self *LabelRepository) IssuesWithLabels(ctx context.Context, tx repository.TxInterface, labelIDs []uuid.UUID) (map[uuid.UUID][]*ent.Label, error) {
	issues, err := self.base.Client(tx).Issue.Query().
		Where(issue.HasLabelsWith(label.IDIn(labelIDs...))).
		WithLabels(func(q *ent.LabelQuery) { q.Where(label.IDIn(labelIDs...)) }).
		All(ctx)
	if err != nil {
		return nil, err
	}
	byIssue := make(map[uuid.UUID][]*ent.Label, len(issues))
	for _, found := range issues {
		byIssue[found.ID] = found.Edges.Labels
	}
	return byIssue, nil
}
