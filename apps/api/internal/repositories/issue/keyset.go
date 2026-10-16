package issue_repo

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/cycle"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	"github.com/zigzagapp-org/zigzag-api/ent/label"
	"github.com/zigzagapp-org/zigzag-api/ent/project"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/user"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
)

var ErrInvalidCursor = errors.New("invalid cursor")

type Cursor struct {
	Signature string            `json:"s"`
	Values    []json.RawMessage `json:"k"`
}

type termKind int

const (
	kindInt termKind = iota
	kindFloat
	kindText
	kindTime
	kindUUID
)

type sortTerm struct {
	expr  string
	kind  termKind
	desc  bool
	value func(*ent.Issue) any
}

func signature(filter *ListFilter) string {
	return fmt.Sprintf("%s|%s|%t", filter.GroupBy, filter.OrderBy, filter.Descending)
}

func DecodeCursor(raw string, filter *ListFilter) (*Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	var parsed Cursor
	if err := json.Unmarshal(decoded, &parsed); err != nil || parsed.Signature != signature(filter) {
		return nil, ErrInvalidCursor
	}
	return &parsed, nil
}

func NextCursor(filter *ListFilter, last *ent.Issue) string {
	terms := sortTerms(sql.Select().From(sql.Table(issue.Table)), filter)
	values := make([]json.RawMessage, len(terms))
	for i, term := range terms {
		values[i], _ = json.Marshal(term.value(last))
	}
	raw, _ := json.Marshal(Cursor{Signature: signature(filter), Values: values})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func keysetOrder(filter *ListFilter) issue.OrderOption {
	return func(s *sql.Selector) {
		terms := sortTerms(s, filter)
		for _, term := range terms {
			if term.desc {
				s.OrderExpr(sql.Expr(term.expr + " DESC"))
				continue
			}
			s.OrderExpr(sql.Expr(term.expr))
		}
		if filter.After == nil {
			return
		}
		after, err := afterPredicate(terms, filter.After.Values)
		if err != nil {
			s.AddError(err)
			return
		}
		s.Where(after)
	}
}

func afterPredicate(terms []sortTerm, raw []json.RawMessage) (*sql.Predicate, error) {
	if len(raw) != len(terms) {
		return nil, ErrInvalidCursor
	}
	values := make([]any, len(terms))
	for i, term := range terms {
		value, err := decodeValue(term.kind, raw[i])
		if err != nil {
			return nil, ErrInvalidCursor
		}
		values[i] = value
	}
	var branches []*sql.Predicate
	for i, term := range terms {
		if values[i] == nil {
			continue
		}
		op := sql.OpGT
		if term.desc {
			op = sql.OpLT
		}
		branches = append(branches, sql.P(func(b *sql.Builder) {
			for j := range i {
				writeEqual(b, terms[j].expr, values[j])
				b.WriteString(" AND ")
			}
			b.WriteString(term.expr).WriteOp(op).Arg(values[i])
		}))
	}
	return sql.Or(branches...), nil
}

func writeEqual(b *sql.Builder, expr string, value any) {
	if value == nil {
		b.WriteString(expr + " IS NULL")
		return
	}
	b.WriteString(expr).WriteOp(sql.OpEQ).Arg(value)
}

func decodeValue(kind termKind, raw json.RawMessage) (any, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	switch kind {
	case kindInt:
		var value int
		err := json.Unmarshal(raw, &value)
		return value, err
	case kindFloat:
		var value float64
		err := json.Unmarshal(raw, &value)
		return value, err
	case kindText:
		var value string
		err := json.Unmarshal(raw, &value)
		return value, err
	case kindTime:
		var value time.Time
		err := json.Unmarshal(raw, &value)
		return value, err
	case kindUUID:
		var value uuid.UUID
		err := json.Unmarshal(raw, &value)
		return value, err
	default:
		return nil, ErrInvalidCursor
	}
}

func isNullTerm(expr string, isNull func(*ent.Issue) bool) sortTerm {
	return sortTerm{
		expr: "CASE WHEN " + expr + " IS NULL THEN 1 ELSE 0 END",
		kind: kindInt,
		value: func(entity *ent.Issue) any {
			if isNull(entity) {
				return 1
			}
			return 0
		},
	}
}

func optionalUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func sortTerms(s *sql.Selector, filter *ListFilter) []sortTerm {
	terms := groupTerms(s, filter.GroupBy)
	terms = append(terms, orderTerms(s, filter.OrderBy, filter.Descending)...)
	return append(terms, sortTerm{expr: s.C(issue.FieldID), kind: kindUUID, value: func(entity *ent.Issue) any { return entity.ID }})
}

func priorityRankExpr(s *sql.Selector) string {
	return "CASE WHEN " + s.C(issue.FieldPriority) + " = 0 THEN 5 ELSE " + s.C(issue.FieldPriority) + " END"
}

func priorityRank(entity *ent.Issue) any {
	if entity.Priority == 0 {
		return 5
	}
	return entity.Priority
}

func groupTerms(s *sql.Selector, group GroupBy) []sortTerm {
	switch group {
	case GroupState:
		return stateTerms(s)
	case GroupAssignee:
		return assigneeTerms(s)
	case GroupPriority:
		return []sortTerm{{expr: priorityRankExpr(s), kind: kindInt, value: priorityRank}}
	case GroupProject:
		return projectTerms(s)
	case GroupCycle:
		return cycleTerms(s)
	case GroupLabel:
		return labelTerms(s)
	case GroupParent:
		return parentTerms(s)
	default:
		return nil
	}
}

func stateTerms(s *sql.Selector) []sortTerm {
	states := sql.Table(workflowstate.Table).As("group_state")
	s.LeftJoin(states).On(s.C(issue.FieldStateID), states.C(workflowstate.FieldID))
	rank := "CASE " + states.C(workflowstate.FieldType)
	for i, stateType := range schema.AllStateTypes() {
		rank += " WHEN '" + string(stateType) + "' THEN " + strconv.Itoa(i)
	}
	return []sortTerm{
		{expr: rank + " END", kind: kindInt, value: func(entity *ent.Issue) any {
			return slices.Index(schema.AllStateTypes(), entity.Edges.State.Type)
		}},
		{expr: states.C(workflowstate.FieldPosition), kind: kindFloat, value: func(entity *ent.Issue) any { return entity.Edges.State.Position }},
		{expr: s.C(issue.FieldStateID), kind: kindUUID, value: func(entity *ent.Issue) any { return entity.StateID }},
	}
}

func assigneeTerms(s *sql.Selector) []sortTerm {
	assignees := sql.Table(user.Table).As("group_assignee")
	s.LeftJoin(assignees).On(s.C(issue.FieldAssigneeID), assignees.C(user.FieldID))
	return []sortTerm{
		isNullTerm(s.C(issue.FieldAssigneeID), func(entity *ent.Issue) bool { return entity.AssigneeID == nil }),
		{expr: assignees.C(user.FieldName), kind: kindText, value: func(entity *ent.Issue) any {
			if entity.Edges.Assignee == nil {
				return nil
			}
			return entity.Edges.Assignee.Name
		}},
		{expr: s.C(issue.FieldAssigneeID), kind: kindUUID, value: func(entity *ent.Issue) any { return optionalUUID(entity.AssigneeID) }},
	}
}

func projectTerms(s *sql.Selector) []sortTerm {
	projects := sql.Table(project.Table).As("group_project")
	s.LeftJoin(projects).On(s.C(issue.FieldProjectID), projects.C(project.FieldID))
	return []sortTerm{
		isNullTerm(s.C(issue.FieldProjectID), func(entity *ent.Issue) bool { return entity.ProjectID == nil }),
		{expr: projects.C(project.FieldName), kind: kindText, value: func(entity *ent.Issue) any {
			if entity.Edges.Project == nil {
				return nil
			}
			return entity.Edges.Project.Name
		}},
		{expr: s.C(issue.FieldProjectID), kind: kindUUID, value: func(entity *ent.Issue) any { return optionalUUID(entity.ProjectID) }},
	}
}

func cycleTerms(s *sql.Selector) []sortTerm {
	cycles := sql.Table(cycle.Table).As("group_cycle")
	s.LeftJoin(cycles).On(s.C(issue.FieldCycleID), cycles.C(cycle.FieldID))
	return []sortTerm{
		isNullTerm(s.C(issue.FieldCycleID), func(entity *ent.Issue) bool { return entity.CycleID == nil }),
		{expr: cycles.C(cycle.FieldNumber), kind: kindInt, value: func(entity *ent.Issue) any {
			if entity.Edges.Cycle == nil {
				return nil
			}
			return entity.Edges.Cycle.Number
		}},
		{expr: s.C(issue.FieldCycleID), kind: kindUUID, value: func(entity *ent.Issue) any { return optionalUUID(entity.CycleID) }},
	}
}

// The first loaded label matches MIN(name) because withIssueEdges orders labels by name in the database collation.
func labelTerms(s *sql.Selector) []sortTerm {
	first := "(SELECT MIN(l." + label.FieldName + ") FROM issue_labels il JOIN " + label.Table + " l ON l.id = il.label_id WHERE il.issue_id = " + s.C(issue.FieldID) + ")"
	return []sortTerm{
		isNullTerm(first, func(entity *ent.Issue) bool { return len(entity.Edges.Labels) == 0 }),
		{expr: first, kind: kindText, value: func(entity *ent.Issue) any {
			if len(entity.Edges.Labels) == 0 {
				return nil
			}
			return entity.Edges.Labels[0].Name
		}},
	}
}

func parentTerms(s *sql.Selector) []sortTerm {
	parents := sql.Table(issue.Table).As("group_parent")
	s.LeftJoin(parents).On(s.C(issue.FieldParentID), parents.C(issue.FieldID))
	return []sortTerm{
		isNullTerm(s.C(issue.FieldParentID), func(entity *ent.Issue) bool { return entity.ParentID == nil }),
		{expr: parents.C(issue.FieldNumber), kind: kindInt, value: func(entity *ent.Issue) any {
			if entity.Edges.Parent == nil {
				return nil
			}
			return entity.Edges.Parent.Number
		}},
		{expr: s.C(issue.FieldParentID), kind: kindUUID, value: func(entity *ent.Issue) any { return optionalUUID(entity.ParentID) }},
	}
}

func sortOrderTerm(s *sql.Selector, descending bool) sortTerm {
	return sortTerm{expr: s.C(issue.FieldSortOrder), kind: kindFloat, desc: descending, value: func(entity *ent.Issue) any { return entity.SortOrder }}
}

func createdAtTerm(s *sql.Selector, descending bool) sortTerm {
	return sortTerm{expr: s.C(issue.FieldCreatedAt), kind: kindTime, desc: descending, value: func(entity *ent.Issue) any { return entity.CreatedAt }}
}

func orderTerms(s *sql.Selector, order OrderBy, descending bool) []sortTerm {
	switch order {
	case OrderPriority:
		return []sortTerm{{expr: priorityRankExpr(s), kind: kindInt, desc: descending, value: priorityRank}, sortOrderTerm(s, false)}
	case OrderCreatedAt:
		return []sortTerm{createdAtTerm(s, descending)}
	case OrderUpdatedAt:
		return []sortTerm{{expr: s.C(issue.FieldUpdatedAt), kind: kindTime, desc: descending, value: func(entity *ent.Issue) any { return entity.UpdatedAt }}}
	case OrderDueDate:
		return []sortTerm{
			isNullTerm(s.C(issue.FieldDueDate), func(entity *ent.Issue) bool { return entity.DueDate == nil }),
			{expr: s.C(issue.FieldDueDate), kind: kindTime, desc: descending, value: func(entity *ent.Issue) any {
				if entity.DueDate == nil {
					return nil
				}
				return *entity.DueDate
			}},
			sortOrderTerm(s, false),
		}
	case OrderTitle:
		return []sortTerm{{expr: s.C(issue.FieldTitle), kind: kindText, desc: descending, value: func(entity *ent.Issue) any { return entity.Title }}}
	default:
		return []sortTerm{sortOrderTerm(s, descending), createdAtTerm(s, true)}
	}
}
