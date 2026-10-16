package issue_service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

var identifierPattern = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9]{0,6})-([0-9]+)$`)

func ParseIdentifier(raw string) (string, int, bool) {
	match := identifierPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		return "", 0, false
	}
	number, err := strconv.Atoi(match[2])
	if err != nil || number <= 0 {
		return "", 0, false
	}
	return strings.ToUpper(match[1]), number, true
}

type cursor struct {
	Offset int `json:"o"`
}

func EncodeCursor(offset int) *string {
	raw, _ := json.Marshal(cursor{Offset: offset})
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	return &encoded
}

func DecodeCursor(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	var parsed cursor
	if err == nil {
		err = json.Unmarshal(decoded, &parsed)
	}
	if err != nil || parsed.Offset < 0 {
		return 0, invalid("cursor is not a next_cursor from a previous page")
	}
	return parsed.Offset, nil
}

func nextCursor(offset, limit, returned int) *string {
	if returned <= limit {
		return nil
	}
	return EncodeCursor(offset + limit)
}

type ListIssuesInput struct {
	WorkspaceID     uuid.UUID          `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace to list issues of"`
	TeamID          uuid.UUID          `query:"team_id" required:"false" format:"uuid" doc:"Only issues of this team"`
	TeamIDs         []uuid.UUID        `query:"team_ids" required:"false" doc:"Only issues of these teams"`
	StateIDs        []uuid.UUID        `query:"state_ids" required:"false" doc:"Only issues in these workflow states"`
	StateTypes      []schema.StateType `query:"state_types" required:"false" doc:"Only issues whose state has one of these types: triage, backlog, unstarted, started, completed, canceled"`
	AssigneeIDs     []string           `query:"assignee_ids" required:"false" doc:"Only issues assigned to these users; include the word null for unassigned issues"`
	CreatorIDs      []uuid.UUID        `query:"creator_ids" required:"false" doc:"Only issues created by these users"`
	LabelIDs        []uuid.UUID        `query:"label_ids" required:"false" doc:"Only issues carrying any of these labels; a label group matches all of its labels"`
	ProjectIDs      []uuid.UUID        `query:"project_ids" required:"false" doc:"Only issues in these projects"`
	MilestoneIDs    []uuid.UUID        `query:"milestone_ids" required:"false" doc:"Only issues in these project milestones"`
	CycleIDs        []uuid.UUID        `query:"cycle_ids" required:"false" doc:"Only issues in these cycles"`
	ParentID        uuid.UUID          `query:"parent_id" required:"false" format:"uuid" doc:"Only sub-issues of this issue"`
	HasParent       string             `query:"has_parent" required:"false" enum:"true,false" doc:"true for sub-issues only, false for top-level issues only"`
	Priorities      []int              `query:"priorities" required:"false" doc:"Only these priorities: 0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	DueBefore       string             `query:"due_before" required:"false" format:"date" doc:"Only issues due on or before this date, YYYY-MM-DD"`
	DueAfter        string             `query:"due_after" required:"false" format:"date" doc:"Only issues due on or after this date, YYYY-MM-DD"`
	CreatedAfter    time.Time          `query:"created_after" required:"false" doc:"Only issues created at or after this RFC 3339 time"`
	UpdatedAfter    time.Time          `query:"updated_after" required:"false" doc:"Only issues updated at or after this RFC 3339 time"`
	Query           string             `query:"query" required:"false" maxLength:"200" doc:"Text in the title or description, or an identifier like ENG-123"`
	SubscriberID    uuid.UUID          `query:"subscriber_id" required:"false" format:"uuid" doc:"Only issues this user subscribes to"`
	IncludeArchived bool               `query:"include_archived" required:"false" doc:"Include archived issues"`
	IncludeTrashed  bool               `query:"include_trashed" required:"false" doc:"Include issues in the trash"`
	OrderBy         string             `query:"order_by" required:"false" enum:"manual,priority,created_at,updated_at,due_date,title" default:"manual" doc:"Sort key; manual is the drag and drop order"`
	Order           string             `query:"order" required:"false" enum:"asc,desc" doc:"Sort direction; defaults to desc for created_at and updated_at, asc otherwise"`
	GroupBy         string             `query:"group_by" required:"false" enum:"none,state,assignee,priority,project,cycle,label,parent" default:"none" doc:"Sort by this grouping first so pages arrive group by group; label groups by the alphabetically first label"`
	models.PaginationInput
}

type IssueListResult struct {
	Issues     []*models.IssueResponse
	NextCursor *string
	Total      int
}

func (self *IssueService) List(ctx context.Context, userID uuid.UUID, input *ListIssuesInput) (*IssueListResult, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	filter, err := self.listFilter(input)
	if err != nil {
		return nil, err
	}
	visible, err := self.VisibleTeams(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	filter.TeamIDs = visible.IDs()
	requested := slices.Clone(input.TeamIDs)
	if input.TeamID != uuid.Nil {
		requested = append(requested, input.TeamID)
	}
	if len(requested) > 0 {
		filter.TeamIDs = slices.DeleteFunc(requested, func(id uuid.UUID) bool { return !visible.Has(id) })
	}

	issues, total, err := self.repo.Issue().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	next := nextCursor(filter.Offset, input.Limit, len(issues))
	if len(issues) > input.Limit {
		issues = issues[:input.Limit]
	}
	responses, err := self.respond(ctx, userID, issues)
	if err != nil {
		return nil, err
	}
	return &IssueListResult{Issues: responses, NextCursor: next, Total: total}, nil
}

func (self *IssueService) listFilter(input *ListIssuesInput) (*issue_repo.ListFilter, error) {
	offset, err := DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	input.Limit = limit
	filter := &issue_repo.ListFilter{
		WorkspaceID:     input.WorkspaceID,
		StateIDs:        input.StateIDs,
		StateTypes:      input.StateTypes,
		CreatorIDs:      input.CreatorIDs,
		LabelIDs:        input.LabelIDs,
		ProjectIDs:      input.ProjectIDs,
		MilestoneIDs:    input.MilestoneIDs,
		CycleIDs:        input.CycleIDs,
		Priorities:      input.Priorities,
		Query:           strings.TrimSpace(input.Query),
		IncludeArchived: input.IncludeArchived,
		IncludeTrashed:  input.IncludeTrashed,
		OrderBy:         issue_repo.OrderBy(input.OrderBy),
		GroupBy:         issue_repo.GroupBy(input.GroupBy),
		Offset:          offset,
		Limit:           limit + 1,
	}
	if filter.OrderBy == "" {
		filter.OrderBy = issue_repo.OrderManual
	}
	switch input.Order {
	case "desc":
		filter.Descending = true
	case "":
		filter.Descending = filter.OrderBy == issue_repo.OrderCreatedAt || filter.OrderBy == issue_repo.OrderUpdatedAt
	}
	for _, stateType := range input.StateTypes {
		if !slices.Contains(schema.AllStateTypes(), stateType) {
			return nil, invalid("state_types takes triage, backlog, unstarted, started, completed or canceled")
		}
	}
	for _, priority := range input.Priorities {
		if priority < 0 || priority > 4 {
			return nil, invalid("priorities take 0 to 4")
		}
	}
	for _, raw := range input.AssigneeIDs {
		if strings.EqualFold(raw, "null") {
			filter.Unassigned = true
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, invalid("assignee_ids takes user ids or null")
		}
		filter.AssigneeIDs = append(filter.AssigneeIDs, id)
	}
	if input.ParentID != uuid.Nil {
		filter.ParentID = &input.ParentID
	}
	if input.SubscriberID != uuid.Nil {
		filter.SubscriberID = &input.SubscriberID
	}
	switch input.HasParent {
	case "true":
		filter.HasParent = new(true)
	case "false":
		filter.HasParent = new(false)
	}
	if filter.DueBefore, err = models.ParseDate("due_before", input.DueBefore); err != nil {
		return nil, err
	}
	if filter.DueAfter, err = models.ParseDate("due_after", input.DueAfter); err != nil {
		return nil, err
	}
	if !input.CreatedAfter.IsZero() {
		filter.CreatedAfter = &input.CreatedAfter
	}
	if !input.UpdatedAfter.IsZero() {
		filter.UpdatedAfter = &input.UpdatedAfter
	}
	if key, number, ok := ParseIdentifier(filter.Query); ok {
		filter.QueryIdentifier = &issue_repo.IdentifierRef{Key: key, Number: number}
	}
	return filter, nil
}

type GetIssueInput struct {
	IssueID     uuid.UUID `query:"issue_id" required:"false" format:"uuid" doc:"Issue id; give this or identifier"`
	Identifier  string    `query:"identifier" required:"false" maxLength:"32" doc:"Identifier like ENG-123; identifiers from before a team move still resolve"`
	WorkspaceID uuid.UUID `query:"workspace_id" required:"false" format:"uuid" doc:"Workspace to look the identifier up in; needed only when several of your workspaces have a team with that key"`
}

func (self *IssueService) Get(ctx context.Context, userID uuid.UUID, input *GetIssueInput) (*models.IssueDetailResponse, error) {
	issueID := input.IssueID
	if issueID == uuid.Nil {
		resolved, err := self.resolveIdentifier(ctx, userID, input.Identifier, input.WorkspaceID)
		if err != nil {
			return nil, err
		}
		issueID = resolved
	}
	issue, err := self.ReadableIssue(ctx, userID, issueID)
	if err != nil {
		return nil, err
	}
	return self.respondDetail(ctx, userID, issue)
}

func (self *IssueService) resolveIdentifier(ctx context.Context, userID uuid.UUID, identifier string, workspaceID uuid.UUID) (uuid.UUID, error) {
	if identifier == "" {
		return uuid.Nil, invalid("Give issue_id or identifier")
	}
	key, number, ok := ParseIdentifier(identifier)
	if !ok {
		return uuid.Nil, invalid("identifier looks like ENG-123")
	}
	workspaces := []uuid.UUID{workspaceID}
	if workspaceID == uuid.Nil {
		set, err := self.repo.Permissions().GetUserPermissionSet(ctx, userID)
		if err != nil {
			return uuid.Nil, err
		}
		workspaces = set.VisibleWorkspaceIDs()
	}
	ids, err := self.repo.Issue().FindByIdentifier(ctx, workspaces, key, number)
	if err != nil {
		return uuid.Nil, err
	}
	switch len(ids) {
	case 0:
		return uuid.Nil, issueNotFound()
	case 1:
		return ids[0], nil
	default:
		return uuid.Nil, invalid("Several of your workspaces have an issue " + key + "-" + strconv.Itoa(number) + "; give workspace_id")
	}
}

type ListIssueHistoryInput struct {
	IssueID uuid.UUID `query:"issue_id" required:"true" format:"uuid" doc:"The issue"`
	models.PaginationInput
}

func (self *IssueService) ListHistory(ctx context.Context, userID uuid.UUID, input *ListIssueHistoryInput) ([]*models.HistoryResponse, *string, error) {
	issue, err := self.ReadableIssue(ctx, userID, input.IssueID)
	if err != nil {
		return nil, nil, err
	}
	offset, err := DecodeCursor(input.Cursor)
	if err != nil {
		return nil, nil, err
	}
	limit := max(input.Limit, 1)
	entries, err := self.repo.Issue().ListHistory(ctx, issue.ID, offset, limit+1)
	if err != nil {
		return nil, nil, err
	}
	next := nextCursor(offset, limit, len(entries))
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return models.TransformHistoryEntities(entries), next, nil
}

type SearchIssuesInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace to search"`
	Query       string    `query:"query" required:"true" minLength:"1" maxLength:"200" doc:"Words to look for in titles and descriptions, or an identifier like ENG-123"`
	Limit       int       `query:"limit" required:"false" minimum:"1" maximum:"50" default:"20" doc:"Most results to return, 1 to 50"`
}

func (self *IssueService) Search(ctx context.Context, userID uuid.UUID, input *SearchIssuesInput) ([]*models.IssueResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	visible, err := self.VisibleTeams(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	limit := max(input.Limit, 1)
	ids := []uuid.UUID{}
	if key, number, ok := ParseIdentifier(input.Query); ok {
		found, err := self.repo.Issue().FindByIdentifier(ctx, []uuid.UUID{input.WorkspaceID}, key, number)
		if err != nil {
			return nil, err
		}
		ids = append(ids, found...)
	}
	ranked, err := self.repo.Issue().Search(ctx, input.WorkspaceID, visible.IDs(), input.Query, limit)
	if err != nil {
		return nil, err
	}
	for _, id := range ranked {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	issues, err := self.repo.Issue().GetManyWithEdges(ctx, nil, ids)
	if err != nil {
		return nil, err
	}
	issues = slices.DeleteFunc(issues, func(issue *ent.Issue) bool { return !visible.Has(issue.TeamID) || issue.TrashedAt != nil })
	if len(issues) > limit {
		issues = issues[:limit]
	}
	return self.respond(ctx, userID, issues)
}
