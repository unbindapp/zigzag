package issue_repo

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
	team_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/team"
)

type IssueRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo      *IssueRepository
	ada       *ent.User
	bob       *ent.User
	workspace *ent.Workspace
	eng       *ent.Team
	ops       *ent.Team
	states    map[string]*ent.WorkflowState
	number    int
}

func (suite *IssueRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewIssueRepository(suite.DB)
	suite.ada = suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.bob = suite.DB.User.Create().SetEmail("bob@example.com").SetName("Bob").SetDisplayName("bob").SaveX(suite.Ctx)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	teams := team_repo.NewTeamRepository(suite.DB)
	var err error
	suite.eng, err = teams.Create(suite.Ctx, nil, &team_repo.CreateTeamInput{WorkspaceID: suite.workspace.ID, Name: "Engineering", Key: "ENG", CreatorID: suite.ada.ID})
	suite.Require().NoError(err)
	suite.ops, err = teams.Create(suite.Ctx, nil, &team_repo.CreateTeamInput{WorkspaceID: suite.workspace.ID, Name: "Operations", Key: "OPS", CreatorID: suite.ada.ID})
	suite.Require().NoError(err)
	suite.states = map[string]*ent.WorkflowState{}
	for _, state := range suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(suite.eng.ID)).AllX(suite.Ctx) {
		suite.states[state.Name] = state
	}
	suite.number = 0
}

func (suite *IssueRepositorySuite) create(title string, edit func(*CreateIssueInput)) *ent.Issue {
	suite.number++
	input := &CreateIssueInput{
		WorkspaceID: suite.workspace.ID,
		TeamID:      suite.eng.ID,
		Number:      suite.number,
		Title:       title,
		StateID:     suite.states["Todo"].ID,
		CreatorID:   &suite.ada.ID,
		SortOrder:   float64(suite.number),
	}
	if edit != nil {
		edit(input)
	}
	created, err := suite.repo.Create(suite.Ctx, nil, input)
	suite.Require().NoError(err)
	return created
}

func (suite *IssueRepositorySuite) list(edit func(*ListFilter)) ([]string, int) {
	filter := &ListFilter{WorkspaceID: suite.workspace.ID, TeamIDs: []uuid.UUID{suite.eng.ID, suite.ops.ID}, Limit: 100}
	if edit != nil {
		edit(filter)
	}
	issues, total, err := suite.repo.List(suite.Ctx, filter)
	suite.Require().NoError(err)
	titles := make([]string, len(issues))
	for i, issue := range issues {
		titles[i] = issue.Title
		suite.NotNil(issue.Edges.Team, "list loads the embedded shapes")
		suite.NotNil(issue.Edges.State)
	}
	return titles, total
}

func (suite *IssueRepositorySuite) walk(base *ListFilter, pageSize int) []string {
	base.WorkspaceID = suite.workspace.ID
	base.TeamIDs = []uuid.UUID{suite.eng.ID, suite.ops.ID}
	base.Limit = pageSize + 1
	seen := []string{}
	for range 20 {
		issues, _, err := suite.repo.List(suite.Ctx, base)
		suite.Require().NoError(err)
		more := len(issues) > pageSize
		if more {
			issues = issues[:pageSize]
		}
		for _, issue := range issues {
			seen = append(seen, issue.Title)
		}
		if !more {
			return seen
		}
		base.After, err = DecodeCursor(NextCursor(base, issues[len(issues)-1]), base)
		suite.Require().NoError(err)
	}
	suite.FailNow("paging never ended")
	return nil
}

func (suite *IssueRepositorySuite) label(name string, parent *ent.Label) *ent.Label {
	create := suite.DB.Label.Create().SetWorkspaceID(suite.workspace.ID).SetName(name).SetColor("#ffffff")
	if parent != nil {
		create.SetParentID(parent.ID)
	}
	return create.SaveX(suite.Ctx)
}

func (suite *IssueRepositorySuite) TestFilters() {
	bug := suite.label("Bug", nil)
	area := suite.DB.Label.Create().SetWorkspaceID(suite.workspace.ID).SetName("Area").SetColor("#000000").SetIsGroup(true).SaveX(suite.Ctx)
	api := suite.label("API", area)
	due := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	parent := suite.create("Parent", func(in *CreateIssueInput) {
		in.AssigneeID = &suite.bob.ID
		in.LabelIDs = []uuid.UUID{bug.ID}
		in.Priority = 1
	})
	suite.create("Child", func(in *CreateIssueInput) {
		in.ParentID = &parent.ID
		in.StateID = suite.states["In Progress"].ID
		in.LabelIDs = []uuid.UUID{api.ID}
		in.DueDate = &due
	})
	suite.create("Done thing", func(in *CreateIssueInput) {
		in.StateID = suite.states["Done"].ID
		in.SubscriberIDs = []uuid.UUID{suite.bob.ID}
	})
	suite.create("Ops work", func(in *CreateIssueInput) {
		in.TeamID = suite.ops.ID
		in.Priority = 3
		in.StateID = suite.ops.Edges.States[1].ID
	})

	titles, total := suite.list(nil)
	suite.Len(titles, 4)
	suite.Equal(4, total)

	titles, _ = suite.list(func(f *ListFilter) { f.TeamIDs = []uuid.UUID{suite.eng.ID} })
	suite.ElementsMatch([]string{"Parent", "Child", "Done thing"}, titles)
	titles, _ = suite.list(func(f *ListFilter) {
		f.StateTypes = []schema.StateType{schema.StateTypeStarted, schema.StateTypeCompleted}
	})
	suite.ElementsMatch([]string{"Child", "Done thing"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.StateIDs = []uuid.UUID{suite.states["Todo"].ID} })
	suite.Equal([]string{"Parent"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.AssigneeIDs = []uuid.UUID{suite.bob.ID} })
	suite.Equal([]string{"Parent"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.Unassigned = true; f.TeamIDs = []uuid.UUID{suite.eng.ID} })
	suite.ElementsMatch([]string{"Child", "Done thing"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.Unassigned = true; f.AssigneeIDs = []uuid.UUID{suite.bob.ID} })
	suite.Len(titles, 4, "null and a user combine with or")
	titles, _ = suite.list(func(f *ListFilter) { f.LabelIDs = []uuid.UUID{area.ID} })
	suite.Equal([]string{"Child"}, titles, "a label group matches its labels")
	titles, _ = suite.list(func(f *ListFilter) { f.LabelIDs = []uuid.UUID{bug.ID, api.ID} })
	suite.ElementsMatch([]string{"Parent", "Child"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.ParentID = &parent.ID })
	suite.Equal([]string{"Child"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.HasParent = new(false) })
	suite.Len(titles, 3)
	titles, _ = suite.list(func(f *ListFilter) { f.Priorities = []int{1, 3} })
	suite.ElementsMatch([]string{"Parent", "Ops work"}, titles)
	before, after := due.AddDate(0, 0, 1), due.AddDate(0, 0, -1)
	titles, _ = suite.list(func(f *ListFilter) { f.DueBefore = &before; f.DueAfter = &after })
	suite.Equal([]string{"Child"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.CreatorIDs = []uuid.UUID{suite.bob.ID} })
	suite.Empty(titles)
	titles, _ = suite.list(func(f *ListFilter) { f.SubscriberID = &suite.bob.ID })
	suite.Equal([]string{"Done thing"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.Query = "THING" })
	suite.Equal([]string{"Done thing"}, titles, "text search ignores case")
	titles, _ = suite.list(func(f *ListFilter) { f.Query = "ops-1"; f.QueryIdentifier = &IdentifierRef{Key: "OPS", Number: 4} })
	suite.Equal([]string{"Ops work"}, titles, "an identifier finds its issue")
	future := time.Now().Add(time.Hour)
	titles, _ = suite.list(func(f *ListFilter) { f.CreatedAfter = &future })
	suite.Empty(titles)
	titles, _ = suite.list(func(f *ListFilter) { f.TeamIDs = []uuid.UUID{} })
	suite.Empty(titles, "no visible team means no issue")
}

func (suite *IssueRepositorySuite) TestArchivedAndTrashedAreHidden() {
	now := time.Now()
	suite.create("Visible", nil)
	archived := suite.create("Archived", nil)
	trashed := suite.create("Trashed", nil)
	suite.DB.Issue.UpdateOneID(archived.ID).SetArchivedAt(now).ExecX(suite.Ctx)
	suite.DB.Issue.UpdateOneID(trashed.ID).SetTrashedAt(now).ExecX(suite.Ctx)

	titles, total := suite.list(nil)
	suite.Equal([]string{"Visible"}, titles)
	suite.Equal(1, total)
	titles, _ = suite.list(func(f *ListFilter) { f.IncludeArchived = true })
	suite.ElementsMatch([]string{"Visible", "Archived"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.IncludeArchived = true; f.IncludeTrashed = true })
	suite.Len(titles, 3)
}

func (suite *IssueRepositorySuite) TestPagesWalkTheWholeOrderedList() {
	for _, title := range []string{"a", "b", "c", "d", "e"} {
		suite.create(title, nil)
	}
	suite.Equal([]string{"a", "b", "c", "d", "e"}, suite.walk(&ListFilter{OrderBy: OrderManual}, 2), "manual order ascends by sort_order")

	titles, _ := suite.list(func(f *ListFilter) { f.OrderBy = OrderManual; f.Descending = true; f.Limit = 2 })
	suite.Equal([]string{"e", "d"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.OrderBy = OrderTitle; f.Descending = true; f.Limit = 1 })
	suite.Equal([]string{"e"}, titles)
}

func (suite *IssueRepositorySuite) TestOrderingAndGrouping() {
	due := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	suite.create("none", nil)
	suite.create("low", func(in *CreateIssueInput) { in.Priority = 4; in.StateID = suite.states["Backlog"].ID })
	suite.create("urgent", func(in *CreateIssueInput) { in.Priority = 1; in.StateID = suite.states["Done"].ID; in.DueDate = &due })
	suite.create("high", func(in *CreateIssueInput) {
		in.Priority = 2
		in.StateID = suite.states["In Progress"].ID
		in.AssigneeID = &suite.ada.ID
	})

	titles, _ := suite.list(func(f *ListFilter) { f.OrderBy = OrderPriority })
	suite.Equal([]string{"urgent", "high", "low", "none"}, titles, "no priority sorts last")
	titles, _ = suite.list(func(f *ListFilter) { f.OrderBy = OrderDueDate })
	suite.Equal("urgent", titles[0], "missing due dates sort last")
	titles, _ = suite.list(func(f *ListFilter) { f.GroupBy = GroupState })
	suite.Equal([]string{"low", "none", "high", "urgent"}, titles, "states group in workflow order")
	titles, _ = suite.list(func(f *ListFilter) { f.GroupBy = GroupPriority; f.OrderBy = OrderTitle })
	suite.Equal([]string{"urgent", "high", "low", "none"}, titles)
	titles, _ = suite.list(func(f *ListFilter) { f.GroupBy = GroupAssignee })
	suite.Equal("high", titles[0], "unassigned issues group last")
	for _, group := range []GroupBy{GroupProject, GroupCycle, GroupLabel, GroupParent, GroupNone} {
		titles, _ = suite.list(func(f *ListFilter) { f.GroupBy = group })
		suite.Len(titles, 4, group)
	}
}

func (suite *IssueRepositorySuite) TestKeysetPagesMatchTheFullListInEveryOrdering() {
	due := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	bug := suite.label("Bug", nil)
	chore := suite.label("Chore", nil)
	parent := suite.create("parent", nil)
	suite.create("a", func(in *CreateIssueInput) {
		in.Priority = 2
		in.AssigneeID = &suite.ada.ID
		in.LabelIDs = []uuid.UUID{bug.ID}
	})
	suite.create("b", func(in *CreateIssueInput) { in.Priority = 2; in.DueDate = &due; in.ParentID = &parent.ID })
	suite.create("c", func(in *CreateIssueInput) { in.StateID = suite.states["Backlog"].ID; in.AssigneeID = &suite.bob.ID })
	suite.create("d", func(in *CreateIssueInput) { in.Priority = 1; in.LabelIDs = []uuid.UUID{chore.ID}; in.SortOrder = 2 })
	suite.create("e", func(in *CreateIssueInput) { in.StateID = suite.states["Done"].ID; in.DueDate = &due; in.SortOrder = 2 })
	suite.create("f", func(in *CreateIssueInput) { in.ParentID = &parent.ID; in.AssigneeID = &suite.ada.ID })

	groups := []GroupBy{GroupNone, GroupState, GroupAssignee, GroupPriority, GroupProject, GroupCycle, GroupLabel, GroupParent}
	orders := []OrderBy{OrderManual, OrderPriority, OrderCreatedAt, OrderUpdatedAt, OrderDueDate, OrderTitle}
	for _, group := range groups {
		for _, order := range orders {
			for _, descending := range []bool{false, true} {
				full, _ := suite.list(func(f *ListFilter) { f.GroupBy = group; f.OrderBy = order; f.Descending = descending })
				paged := suite.walk(&ListFilter{GroupBy: group, OrderBy: order, Descending: descending}, 2)
				suite.Equal(full, paged, "%s %s desc=%t", group, order, descending)
			}
		}
	}
}

func (suite *IssueRepositorySuite) TestKeysetPagesDoNotShiftWhenRowsArriveAhead() {
	for _, title := range []string{"a", "b", "c", "d"} {
		suite.create(title, nil)
	}
	filter := &ListFilter{WorkspaceID: suite.workspace.ID, TeamIDs: []uuid.UUID{suite.eng.ID}, OrderBy: OrderManual, Limit: 2}
	first, _, err := suite.repo.List(suite.Ctx, filter)
	suite.Require().NoError(err)
	suite.create("new on top", func(in *CreateIssueInput) { in.SortOrder = -1 })

	filter.After, err = DecodeCursor(NextCursor(filter, first[1]), filter)
	suite.Require().NoError(err)
	second, _, err := suite.repo.List(suite.Ctx, filter)
	suite.Require().NoError(err)
	suite.Equal([]string{"c", "d"}, []string{second[0].Title, second[1].Title})

	_, err = DecodeCursor(NextCursor(filter, first[1]), &ListFilter{OrderBy: OrderTitle})
	suite.ErrorIs(err, ErrInvalidCursor)
	_, err = DecodeCursor("garbage", filter)
	suite.ErrorIs(err, ErrInvalidCursor)
}

func (suite *IssueRepositorySuite) TestIdentifiersFollowMoves() {
	moved := suite.create("Moved", nil)
	suite.Require().NoError(suite.repo.RecordMove(suite.Ctx, nil, moved.ID, suite.eng.ID, moved.Number))
	fields := FieldsOf(moved)
	fields.TeamID, fields.Number, fields.StateID = suite.ops.ID, 1, suite.ops.Edges.States[0].ID
	suite.Require().NoError(suite.repo.Update(suite.Ctx, nil, moved.ID, &fields))

	ids, err := suite.repo.FindByIdentifier(suite.Ctx, []uuid.UUID{suite.workspace.ID}, "OPS", 1)
	suite.NoError(err)
	suite.Equal([]uuid.UUID{moved.ID}, ids)
	ids, _ = suite.repo.FindByIdentifier(suite.Ctx, []uuid.UUID{suite.workspace.ID}, "ENG", moved.Number)
	suite.Equal([]uuid.UUID{moved.ID}, ids, "the old identifier still resolves")
	ids, _ = suite.repo.FindByIdentifier(suite.Ctx, []uuid.UUID{uuid.New()}, "OPS", 1)
	suite.Empty(ids, "other workspaces do not match")
}

func (suite *IssueRepositorySuite) TestUpdateWritesAndClearsEveryField() {
	estimate := 3.0
	issue := suite.create("Full", func(in *CreateIssueInput) { in.AssigneeID = &suite.bob.ID; in.Estimate = &estimate })
	fields := FieldsOf(issue)
	fields.Title, fields.AssigneeID, fields.Estimate, fields.Priority = "Renamed", nil, nil, 2
	suite.Require().NoError(suite.repo.Update(suite.Ctx, nil, issue.ID, &fields))
	updated, err := suite.repo.GetWithEdges(suite.Ctx, nil, issue.ID)
	suite.Require().NoError(err)
	suite.Equal("Renamed", updated.Title)
	suite.Nil(updated.AssigneeID)
	suite.Nil(updated.Estimate)
	suite.Equal(2, updated.Priority)
}

func (suite *IssueRepositorySuite) TestCountsAncestorsAndSortOrder() {
	parent := suite.create("Parent", nil)
	child := suite.create("Child", func(in *CreateIssueInput) { in.ParentID = &parent.ID; in.StateID = suite.states["Done"].ID })
	grandchild := suite.create("Grandchild", func(in *CreateIssueInput) { in.ParentID = &child.ID })
	suite.DB.Comment.Create().SetIssueID(parent.ID).SetBody("hi").ExecX(suite.Ctx)

	counts, err := suite.repo.Counts(suite.Ctx, []uuid.UUID{parent.ID, child.ID})
	suite.NoError(err)
	suite.Equal(1, counts[parent.ID].Children)
	suite.Equal(1, counts[parent.ID].CompletedChildren)
	suite.Equal(1, counts[parent.ID].Comments)
	suite.Equal(0, counts[child.ID].CompletedChildren)

	ancestors, err := suite.repo.AncestorIDs(suite.Ctx, nil, grandchild.ID)
	suite.NoError(err)
	suite.Equal([]uuid.UUID{child.ID, parent.ID}, ancestors)

	top, err := suite.repo.TopSortOrder(suite.Ctx, nil, suite.states["Todo"].ID)
	suite.NoError(err)
	suite.Equal(0.0, top, "one above the lowest sort_order in the state")

	suite.NoError(suite.repo.Delete(suite.Ctx, nil, child.ID))
	orphan, err := suite.repo.GetByID(suite.Ctx, grandchild.ID)
	suite.NoError(err)
	suite.Nil(orphan.ParentID, "children survive their parent")
}

func (suite *IssueRepositorySuite) TestSubscriptionsAndReactions() {
	issue := suite.create("Watched", func(in *CreateIssueInput) { in.SubscriberIDs = []uuid.UUID{suite.ada.ID} })
	suite.NoError(suite.repo.Subscribe(suite.Ctx, nil, issue.ID, suite.ada.ID, suite.bob.ID, suite.bob.ID))
	ids, err := suite.repo.SubscriberIDs(suite.Ctx, issue.ID)
	suite.NoError(err)
	suite.ElementsMatch([]uuid.UUID{suite.ada.ID, suite.bob.ID}, ids)
	suite.NoError(suite.repo.Unsubscribe(suite.Ctx, issue.ID, suite.bob.ID))
	ids, _ = suite.repo.SubscriberIDs(suite.Ctx, issue.ID)
	suite.Equal([]uuid.UUID{suite.ada.ID}, ids)

	target := ReactionTarget{IssueID: &issue.ID}
	added, err := suite.repo.AddReaction(suite.Ctx, suite.ada.ID, "👍", target)
	suite.NoError(err)
	suite.True(added)
	added, err = suite.repo.AddReaction(suite.Ctx, suite.ada.ID, "👍", target)
	suite.NoError(err)
	suite.False(added, "reacting twice is a no-op")
	reactions, err := suite.repo.Reactions(suite.Ctx, target)
	suite.NoError(err)
	suite.Len(reactions, 1)
	suite.NoError(suite.repo.RemoveReaction(suite.Ctx, suite.ada.ID, "👍", target))
	reactions, _ = suite.repo.Reactions(suite.Ctx, target)
	suite.Empty(reactions)
}

func (suite *IssueRepositorySuite) TestRelationsBothDirections() {
	blocker := suite.create("Blocker", nil)
	blocked := suite.create("Blocked", nil)
	_, err := suite.repo.CreateRelation(suite.Ctx, nil, blocker.ID, blocked.ID, schema.IssueRelationBlocks)
	suite.Require().NoError(err)
	exists, err := suite.repo.RelationExists(suite.Ctx, blocked.ID, blocker.ID, schema.IssueRelationBlocks)
	suite.NoError(err)
	suite.True(exists, "a relation is found from either side")

	outgoing, incoming, err := suite.repo.Relations(suite.Ctx, blocked.ID)
	suite.NoError(err)
	suite.Empty(outgoing)
	suite.Require().Len(incoming, 1)
	suite.Equal(blocker.ID, incoming[0].Edges.Issue.ID)
}

func (suite *IssueRepositorySuite) TestAutomationQueries() {
	old := time.Now().AddDate(0, -7, 0)
	stale := suite.create("Stale", nil)
	suite.create("Fresh", nil)
	done := suite.create("Done long ago", func(in *CreateIssueInput) { in.StateID = suite.states["Done"].ID; in.CompletedAt = &old })
	suite.DB.Issue.UpdateOneID(stale.ID).SetUpdatedAt(old).ExecX(suite.Ctx)
	cutoff := time.Now().AddDate(0, -6, 0)

	found, err := suite.repo.StaleOpenIssues(suite.Ctx, suite.eng.ID, cutoff)
	suite.NoError(err)
	suite.Require().Len(found, 1)
	suite.Equal(stale.ID, found[0].ID)
	closed, err := suite.repo.ClosedIssuesBefore(suite.Ctx, suite.eng.ID, cutoff)
	suite.NoError(err)
	suite.Require().Len(closed, 1)
	suite.Equal(done.ID, closed[0].ID)

	suite.DB.Issue.UpdateOneID(stale.ID).SetTrashedAt(old).ExecX(suite.Ctx)
	trashed, err := suite.repo.TrashedBefore(suite.Ctx, cutoff)
	suite.NoError(err)
	suite.Equal([]uuid.UUID{stale.ID}, trashed)
}

func TestIssueRepositorySuite(t *testing.T) {
	suite.Run(t, new(IssueRepositorySuite))
}

func (suite *IssueRepositorySuite) TestTitlesCountCharactersNotBytes() {
	suite.number++
	_, err := suite.repo.Create(suite.Ctx, nil, &CreateIssueInput{
		WorkspaceID: suite.workspace.ID, TeamID: suite.eng.ID, Number: suite.number, Title: strings.Repeat("é", 255), StateID: suite.states["Todo"].ID,
	})
	suite.NoError(err, "255 accented letters are 510 bytes but a valid title")
	suite.number++
	_, err = suite.repo.Create(suite.Ctx, nil, &CreateIssueInput{
		WorkspaceID: suite.workspace.ID, TeamID: suite.eng.ID, Number: suite.number, Title: strings.Repeat("é", 256), StateID: suite.states["Todo"].ID,
	})
	suite.True(ent.IsValidationError(err))
}

func (suite *IssueRepositorySuite) TestCrowdedSortOrdersAreRespacedInListOrder() {
	a := suite.create("a", func(in *CreateIssueInput) { in.SortOrder = 1 })
	b := suite.create("b", func(in *CreateIssueInput) { in.SortOrder = 2 })
	c := suite.create("c", func(in *CreateIssueInput) { in.SortOrder = 1.5 })
	suite.Require().NoError(suite.repo.RespaceIfCrowded(suite.Ctx, nil, suite.workspace.ID, c.ID, c.SortOrder))
	suite.Equal(1.5, suite.DB.Issue.GetX(suite.Ctx, c.ID).SortOrder, "a sort order with room around it is left alone")

	suite.DB.Issue.UpdateOneID(c.ID).SetSortOrder(1 + 1e-12).ExecX(suite.Ctx)
	suite.Require().NoError(suite.repo.RespaceIfCrowded(suite.Ctx, nil, suite.workspace.ID, c.ID, 1+1e-12))
	suite.Equal(1024.0, suite.DB.Issue.GetX(suite.Ctx, a.ID).SortOrder)
	suite.Equal(2048.0, suite.DB.Issue.GetX(suite.Ctx, c.ID).SortOrder, "the moved issue keeps its place between its neighbours")
	suite.Equal(3072.0, suite.DB.Issue.GetX(suite.Ctx, b.ID).SortOrder)
	titles, _ := suite.list(nil)
	suite.Equal([]string{"a", "c", "b"}, titles)

	parent := suite.create("parent", nil)
	first := suite.create("first", func(in *CreateIssueInput) { in.ParentID = &parent.ID; in.SubIssueSortOrder = 1 })
	second := suite.create("second", func(in *CreateIssueInput) { in.ParentID = &parent.ID; in.SubIssueSortOrder = 1 })
	suite.Require().NoError(suite.repo.RespaceSubIssuesIfCrowded(suite.Ctx, nil, parent.ID, second.ID, 1))
	suite.Equal(1024.0, suite.DB.Issue.GetX(suite.Ctx, first.ID).SubIssueSortOrder)
	suite.Equal(2048.0, suite.DB.Issue.GetX(suite.Ctx, second.ID).SubIssueSortOrder)
}
