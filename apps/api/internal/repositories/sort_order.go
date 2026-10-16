package repository

import (
	"context"
	"math"

	"github.com/google/uuid"
)

const (
	sortOrderStep      = "1024"
	crowdedRelativeGap = 1e-9
)

var (
	IssueSortOrder    = SortColumn{Table: "issues", Column: "sort_order", ScopeColumn: "workspace_id", OrderBy: "sort_order, created_at DESC, id"}
	SubIssueSortOrder = SortColumn{Table: "issues", Column: "sub_issue_sort_order", ScopeColumn: "parent_id", OrderBy: "sub_issue_sort_order, number, id"}
	LabelSortOrder    = SortColumn{Table: "labels", Column: "sort_order", ScopeColumn: "workspace_id", OrderBy: "sort_order, name, id"}
	ProjectSortOrder  = SortColumn{Table: "projects", Column: "sort_order", ScopeColumn: "workspace_id", OrderBy: "sort_order, created_at, id"}
	MilestoneOrder    = SortColumn{Table: "project_milestones", Column: "sort_order", ScopeColumn: "project_id", OrderBy: "sort_order, target_date NULLS LAST, created_at, id"}
	StatePosition     = SortColumn{Table: "workflow_states", Column: "position", ScopeColumn: "team_id", OrderBy: "position, id"}
	StatusPosition    = SortColumn{Table: "project_statuses", Column: "position", ScopeColumn: "workspace_id", OrderBy: "position, id"}
)

type SortColumn struct {
	Table       string
	Column      string
	ScopeColumn string
	OrderBy     string
}

func (r *BaseRepository) RespaceIfCrowded(ctx context.Context, tx TxInterface, sort SortColumn, scope uuid.UUID, id uuid.UUID, value float64) (bool, error) {
	client := r.Client(tx)
	tolerance := math.Max(math.Abs(value), 1) * crowdedRelativeGap
	rows, err := client.QueryContext(ctx,
		"SELECT COUNT(*) FROM "+sort.Table+" WHERE "+sort.ScopeColumn+" = $1 AND id <> $2 AND ABS("+sort.Column+" - $3) <= $4",
		scope, id, value, tolerance)
	if err != nil {
		return false, err
	}
	crowded := 0
	for rows.Next() {
		if err := rows.Scan(&crowded); err != nil {
			_ = rows.Close()
			return false, err
		}
	}
	if err := rows.Close(); err != nil {
		return false, err
	}
	if crowded == 0 {
		return false, nil
	}
	_, err = client.ExecContext(ctx,
		"UPDATE "+sort.Table+" SET "+sort.Column+" = ranked.position * "+sortOrderStep+" FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY "+sort.OrderBy+") AS position FROM "+sort.Table+" WHERE "+sort.ScopeColumn+" = $1) AS ranked WHERE "+sort.Table+".id = ranked.id",
		scope)
	return err == nil, err
}
