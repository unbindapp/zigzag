package issue_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	"github.com/zigzagapp-org/zigzag-api/ent/label"
	"github.com/zigzagapp-org/zigzag-api/ent/predicate"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/team"
	"github.com/zigzagapp-org/zigzag-api/ent/user"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
)

type OrderBy string

const (
	OrderManual    OrderBy = "manual"
	OrderPriority  OrderBy = "priority"
	OrderCreatedAt OrderBy = "created_at"
	OrderUpdatedAt OrderBy = "updated_at"
	OrderDueDate   OrderBy = "due_date"
	OrderTitle     OrderBy = "title"
)

type GroupBy string

const (
	GroupNone     GroupBy = "none"
	GroupState    GroupBy = "state"
	GroupAssignee GroupBy = "assignee"
	GroupPriority GroupBy = "priority"
	GroupProject  GroupBy = "project"
	GroupCycle    GroupBy = "cycle"
	GroupLabel    GroupBy = "label"
	GroupParent   GroupBy = "parent"
)

type IdentifierRef struct {
	Key    string
	Number int
}

type ListFilter struct {
	WorkspaceID     uuid.UUID
	TeamIDs         []uuid.UUID
	StateIDs        []uuid.UUID
	StateTypes      []schema.StateType
	AssigneeIDs     []uuid.UUID
	Unassigned      bool
	CreatorIDs      []uuid.UUID
	LabelIDs        []uuid.UUID
	ProjectIDs      []uuid.UUID
	MilestoneIDs    []uuid.UUID
	CycleIDs        []uuid.UUID
	ParentID        *uuid.UUID
	HasParent       *bool
	Priorities      []int
	DueBefore       *time.Time
	DueAfter        *time.Time
	CreatedAfter    *time.Time
	UpdatedAfter    *time.Time
	Query           string
	QueryIdentifier *IdentifierRef
	SubscriberID    *uuid.UUID
	IncludeArchived bool
	IncludeTrashed  bool
	OrderBy         OrderBy
	Descending      bool
	GroupBy         GroupBy
	After           *Cursor
	Limit           int
}

func (self *IssueRepository) predicates(filter *ListFilter) []predicate.Issue {
	where := []predicate.Issue{issue.WorkspaceID(filter.WorkspaceID), issue.TeamIDIn(filter.TeamIDs...)}
	if len(filter.StateIDs) > 0 {
		where = append(where, issue.StateIDIn(filter.StateIDs...))
	}
	if len(filter.StateTypes) > 0 {
		where = append(where, issue.HasStateWith(workflowstate.TypeIn(filter.StateTypes...)))
	}
	switch {
	case filter.Unassigned && len(filter.AssigneeIDs) > 0:
		where = append(where, issue.Or(issue.AssigneeIDIsNil(), issue.AssigneeIDIn(filter.AssigneeIDs...)))
	case filter.Unassigned:
		where = append(where, issue.AssigneeIDIsNil())
	case len(filter.AssigneeIDs) > 0:
		where = append(where, issue.AssigneeIDIn(filter.AssigneeIDs...))
	}
	if len(filter.CreatorIDs) > 0 {
		where = append(where, issue.CreatorIDIn(filter.CreatorIDs...))
	}
	if len(filter.LabelIDs) > 0 {
		where = append(where, issue.HasLabelsWith(label.Or(label.IDIn(filter.LabelIDs...), label.ParentIDIn(filter.LabelIDs...))))
	}
	if len(filter.ProjectIDs) > 0 {
		where = append(where, issue.ProjectIDIn(filter.ProjectIDs...))
	}
	if len(filter.MilestoneIDs) > 0 {
		where = append(where, issue.MilestoneIDIn(filter.MilestoneIDs...))
	}
	if len(filter.CycleIDs) > 0 {
		where = append(where, issue.CycleIDIn(filter.CycleIDs...))
	}
	if filter.ParentID != nil {
		where = append(where, issue.ParentID(*filter.ParentID))
	}
	if filter.HasParent != nil && *filter.HasParent {
		where = append(where, issue.ParentIDNotNil())
	}
	if filter.HasParent != nil && !*filter.HasParent {
		where = append(where, issue.ParentIDIsNil())
	}
	if len(filter.Priorities) > 0 {
		where = append(where, issue.PriorityIn(filter.Priorities...))
	}
	if filter.DueBefore != nil {
		where = append(where, issue.DueDateLTE(*filter.DueBefore))
	}
	if filter.DueAfter != nil {
		where = append(where, issue.DueDateGTE(*filter.DueAfter))
	}
	if filter.CreatedAfter != nil {
		where = append(where, issue.CreatedAtGTE(*filter.CreatedAfter))
	}
	if filter.UpdatedAfter != nil {
		where = append(where, issue.UpdatedAtGTE(*filter.UpdatedAfter))
	}
	if filter.Query != "" {
		matches := []predicate.Issue{issue.TitleContainsFold(filter.Query), issue.DescriptionTextContainsFold(filter.Query)}
		if ref := filter.QueryIdentifier; ref != nil {
			matches = append(matches, issue.And(issue.Number(ref.Number), issue.HasTeamWith(team.KeyEqualFold(ref.Key))))
		}
		where = append(where, issue.Or(matches...))
	}
	if filter.SubscriberID != nil {
		where = append(where, issue.HasSubscribersWith(user.ID(*filter.SubscriberID)))
	}
	if !filter.IncludeArchived {
		where = append(where, issue.ArchivedAtIsNil())
	}
	if !filter.IncludeTrashed {
		where = append(where, issue.TrashedAtIsNil())
	}
	return where
}

func (self *IssueRepository) List(ctx context.Context, filter *ListFilter) ([]*ent.Issue, int, error) {
	where := self.predicates(filter)
	total, err := self.base.DB.Issue.Query().Where(where...).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	query := withIssueEdges(self.base.DB.Issue.Query().Where(where...))
	issues, err := query.Order(keysetOrder(filter)).Limit(filter.Limit).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return issues, total, nil
}
