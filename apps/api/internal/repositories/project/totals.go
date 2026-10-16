package project_repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/predicate"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/internal/models"
)

func (self *ProjectRepository) IssueTotals(ctx context.Context, projectIDs []uuid.UUID) (map[uuid.UUID]*models.IssueTotals, error) {
	return self.issueTotals(ctx, issue.FieldProjectID, projectIDs)
}

func (self *ProjectRepository) MilestoneIssueTotals(ctx context.Context, milestoneIDs []uuid.UUID) (map[uuid.UUID]*models.IssueTotals, error) {
	return self.issueTotals(ctx, issue.FieldMilestoneID, milestoneIDs)
}

func (self *ProjectRepository) issueTotals(ctx context.Context, column string, ids []uuid.UUID) (map[uuid.UUID]*models.IssueTotals, error) {
	totals := map[uuid.UUID]*models.IssueTotals{}
	if len(ids) == 0 {
		return totals, nil
	}
	values := make([]any, len(ids))
	for i, id := range ids {
		values[i] = id
	}
	var rows []struct {
		GroupID uuid.UUID `json:"group_id"`
		Type    string    `json:"type"`
		Count   int       `json:"count"`
		Scope   float64   `json:"scope"`
	}
	err := self.base.DB.Issue.Query().
		Where(issue.TrashedAtIsNil(), predicate.Issue(func(s *sql.Selector) { s.Where(sql.In(s.C(column), values...)) })).
		Modify(func(s *sql.Selector) {
			states := sql.Table(workflowstate.Table)
			estimate := s.C(issue.FieldEstimate)
			s.Join(states).On(s.C(issue.FieldStateID), states.C(workflowstate.FieldID))
			s.Select(
				sql.As(s.C(column), "group_id"),
				sql.As(states.C(workflowstate.FieldType), "type"),
				sql.As(sql.Count("*"), "count"),
				sql.As("SUM(CASE WHEN "+estimate+" > 0 THEN "+estimate+" ELSE 1 END)", "scope"),
			)
			s.GroupBy(s.C(column), states.C(workflowstate.FieldType))
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		group, ok := totals[row.GroupID]
		if !ok {
			group = &models.IssueTotals{Counts: map[schema.StateType]int{}, Scopes: map[schema.StateType]float64{}}
			totals[row.GroupID] = group
		}
		group.Counts[schema.StateType(row.Type)] += row.Count
		group.Scopes[schema.StateType(row.Type)] += row.Scope
	}
	return totals, nil
}
