package linear

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
)

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func loadFixture(t *testing.T) *LinearExport {
	t.Helper()
	raw, err := os.ReadFile("testdata/unbind-export.json")
	require.NoError(t, err)
	export := &LinearExport{}
	require.NoError(t, json.Unmarshal(raw, export))
	return export
}

func projectStatuses() []*ent.ProjectStatus {
	types := []schema.ProjectStatusType{schema.ProjectStatusBacklog, schema.ProjectStatusPlanned, schema.ProjectStatusStarted, schema.ProjectStatusPaused, schema.ProjectStatusCompleted, schema.ProjectStatusCanceled}
	names := []string{"Backlog", "Planned", "In Progress", "Paused", "Completed", "Canceled"}
	statuses := make([]*ent.ProjectStatus, len(types))
	for i := range types {
		statuses[i] = &ent.ProjectStatus{ID: uuid.New(), Name: names[i], Type: types[i], Position: float64(i)}
	}
	return statuses
}

func defaultOptions() Options {
	return Options{ActorID: uuid.New(), CreateMissingUsers: true, Now: now}
}

func issueByRef(plan *Plan, export *LinearExport, identifier string) issue_repo.ImportedIssue {
	for _, issue := range export.Issues {
		if issue.Identifier != identifier {
			continue
		}
		for _, row := range plan.Issues {
			if row.ExternalID == issue.ID {
				return row
			}
		}
	}
	return issue_repo.ImportedIssue{}
}

func TestBuildMapsTheWholeUnbindExport(t *testing.T) {
	export := loadFixture(t)
	plan := Build(export, &Snapshot{ProjectStatuses: projectStatuses()}, defaultOptions())

	assert.Equal(t, schema.ImportEntityCount{Created: 456}, plan.Counts[EntityIssues])
	assert.Equal(t, schema.ImportEntityCount{Created: 21}, plan.Counts[EntityComments])
	assert.Equal(t, schema.ImportEntityCount{Created: 28}, plan.Counts[EntityAttachments])
	assert.Equal(t, schema.ImportEntityCount{Created: 9}, plan.Counts[EntityLabels])
	assert.Equal(t, schema.ImportEntityCount{Created: 7}, plan.Counts[EntityWorkflowStates])
	assert.Equal(t, schema.ImportEntityCount{Created: 1}, plan.Counts[EntityProjects])
	assert.Equal(t, schema.ImportEntityCount{Created: 1}, plan.Counts[EntityTeams])
	assert.Equal(t, schema.ImportEntityCount{Created: 3, Skipped: 1}, plan.Counts[EntityUsers])
	assert.Len(t, plan.Issues, 456)
	assert.Len(t, plan.Comments, 21)
	assert.Len(t, plan.Attachments, 28)
	assert.Len(t, plan.Labels, 9)
	assert.Len(t, plan.States, 7)
	assert.Len(t, plan.Projects, 1)

	archived := 0
	for _, row := range plan.Issues {
		if row.ArchivedAt != nil {
			archived++
		}
	}
	assert.Equal(t, 192, archived)

	spans := 0
	for _, issue := range export.Issues {
		spans += max(len(issue.StateHistory), 1)
	}
	assert.Equal(t, spans, plan.Counts[EntityHistory].Created)
	assert.Len(t, plan.History, spans)
	for _, entry := range plan.History {
		assert.NotEmpty(t, entry.Changes)
	}

	require.Len(t, plan.Teams, 1)
	team := plan.Teams[0]
	assert.Equal(t, "UNB", team.Key)
	assert.Equal(t, 458, plan.Counters[team.ID])
	require.NotNil(t, team.DefaultStateID)
	for _, state := range plan.States {
		if *team.DefaultStateID == state.ID {
			assert.Equal(t, "Todo", state.Name)
		}
		if state.Name == "Duplicate" {
			assert.Equal(t, schema.StateTypeCanceled, state.Type)
		}
	}

	for _, mapping := range plan.UserMapping {
		if IsBot(mapping.Email) {
			assert.Equal(t, "skipped", mapping.Action)
			assert.Nil(t, mapping.UserID)
			continue
		}
		assert.Equal(t, "created", mapping.Action)
		assert.NotNil(t, mapping.UserID)
	}
	assert.Len(t, plan.Members, 3)
}

func TestBuildKeepsNumbersLabelsProjectAndRelations(t *testing.T) {
	export := loadFixture(t)
	plan := Build(export, &Snapshot{ProjectStatuses: projectStatuses()}, defaultOptions())

	labelNames := map[uuid.UUID]string{}
	for _, label := range plan.Labels {
		labelNames[label.ID] = label.Name
	}
	recent := issueByRef(plan, export, "UNB-455")
	assert.Equal(t, 455, recent.Number)
	assert.Equal(t, 4, recent.Priority)
	assert.NotNil(t, recent.ProjectID)
	assert.Equal(t, plan.Projects[0].ID, *recent.ProjectID)
	names := []string{}
	for _, id := range recent.LabelIDs {
		names = append(names, labelNames[id])
	}
	assert.ElementsMatch(t, []string{"Bug", "sec-low"}, names)
	assert.Nil(t, recent.ArchivedAt)
	assert.Equal(t, "https://linear.app/unbind/issue/UNB-455/f83-terminal-websocket-reads-whole-messages-with-no-size-limit", recent.URL)

	related := issueByRef(plan, export, "UNB-289")
	hasRelation := false
	for _, relation := range plan.Relations {
		if relation.IssueID == related.ID || relation.RelatedIssueID == related.ID {
			hasRelation = true
		}
	}
	assert.True(t, hasRelation)

	pairs := map[[2]uuid.UUID]bool{}
	for _, relation := range plan.Relations {
		key := [2]uuid.UUID{relation.IssueID, relation.RelatedIssueID}
		reversed := [2]uuid.UUID{relation.RelatedIssueID, relation.IssueID}
		assert.False(t, pairs[key] || pairs[reversed], "each relation is planned once")
		pairs[key] = true
	}
	assert.Equal(t, len(plan.Relations), plan.Counts[EntityRelations].Created)

	project := plan.Projects[0]
	assert.Equal(t, "security-scan-2026-10-15-7b05243d", project.Slug)
	assert.Len(t, project.TeamIDs, 1)
}

func TestBuildRestoresArchivedIssuesWhenAsked(t *testing.T) {
	options := defaultOptions()
	options.RestoreArchived = true
	plan := Build(loadFixture(t), &Snapshot{ProjectStatuses: projectStatuses()}, options)
	for _, row := range plan.Issues {
		assert.Nil(t, row.ArchivedAt)
	}
}

func smallExport() *LinearExport {
	ptr := func(s string) *string { return &s }
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return &LinearExport{
		Users: []LinearUser{
			{ID: "u-ada", Email: "Ada@Example.com", Name: ptr("Ada"), DisplayName: ptr("ada"), IsAdmin: true},
			{ID: "u-new", Email: "new@example.com", Name: ptr("New Person"), DisplayName: ptr("ada")},
			{ID: "u-bot", Email: "linear-x@linear.linear.app", Name: ptr("Linear")},
		},
		Teams: []LinearTeam{{ID: "t-eng", Key: "ENG", Name: "Engineering"}},
		WorkflowStates: []LinearWorkflowState{
			{ID: "s-todo", TeamID: "t-eng", Name: "Todo", Type: "unstarted"},
			{ID: "s-dup", TeamID: "t-eng", Name: "Dupe", Type: "duplicate"},
		},
		Labels: []LinearLabel{
			{ID: "l-size", Name: "Size", IsGroup: true, GroupType: ptr("singleSelect"), TeamID: ptr("t-eng")},
			{ID: "l-small", Name: "Small", ParentID: ptr("l-size"), TeamID: ptr("t-eng")},
			{ID: "l-large", Name: "Large", ParentID: ptr("l-size"), TeamID: ptr("t-eng")},
			{ID: "l-bug", Name: "bug"},
		},
		Issues: []LinearIssue{
			{
				ID: "i-1", Identifier: "ENG-1", Number: 1, TeamID: "t-eng", Title: "First", CreatedAt: &created,
				StateID: ptr("s-todo"), CreatorID: ptr("u-ada"), AssigneeID: ptr("u-bot"),
				LabelIDs:  []string{"l-small", "l-large", "l-size", "l-bug"},
				Relations: LinearRelations{RelatedTo: []string{"i-2"}, Blocks: []string{"i-2"}},
				Comments: []LinearComment{
					{ID: "c-1", Body: "root", AuthorID: ptr("u-ada"), CreatedAt: &created},
					{ID: "c-3", Body: "nested reply", AuthorID: ptr("u-bot"), ParentID: ptr("c-2"), CreatedAt: &created},
					{ID: "c-2", Body: "reply", AuthorID: ptr("u-new"), ParentID: ptr("c-1"), CreatedAt: &created},
				},
			},
			{
				ID: "i-2", Identifier: "ENG-2", Number: 2, TeamID: "t-eng", Title: "Second", CreatedAt: &created,
				StateName: ptr("Dupe"), StateType: ptr("canceled"),
				Relations: LinearRelations{RelatedTo: []string{"i-1"}, BlockedBy: []string{"i-1"}, DuplicateOf: ptr("i-missing")},
			},
		},
	}
}

func TestBuildMatchesTheWorkspaceItImportsInto(t *testing.T) {
	adaID, teamID, todoID, takenID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	bugID := uuid.New()
	snapshot := &Snapshot{
		Users:             []*ent.User{{ID: adaID, Email: "ada@example.com", DisplayName: "ada"}},
		TakenDisplayNames: []string{"ada"},
		Members:           []*ent.WorkspaceMember{{UserID: adaID, Role: schema.WorkspaceRoleMember}},
		Teams:             []*ent.Team{{ID: teamID, Key: "ENG", Name: "Engineering", IssueCounter: 1}},
		States:            []*ent.WorkflowState{{ID: todoID, TeamID: teamID, Name: "todo", Type: schema.StateTypeUnstarted}},
		Labels:            []*ent.Label{{ID: bugID, Name: "Bug"}},
		Issues:            []*ent.Issue{{ID: takenID, TeamID: teamID, Number: 1}},
	}
	plan := Build(smallExport(), snapshot, defaultOptions())

	assert.Equal(t, schema.ImportEntityCount{Created: 1, Updated: 1, Skipped: 1}, plan.Counts[EntityUsers])
	require.Len(t, plan.Users, 1)
	assert.NotEqual(t, "ada", plan.Users[0].DisplayName)
	assert.Len(t, plan.Members, 1, "ada is a member already and keeps her role")
	assert.Empty(t, plan.Teams)
	assert.Equal(t, schema.ImportEntityCount{Updated: 1}, plan.Counts[EntityTeams])
	assert.Equal(t, schema.ImportEntityCount{Created: 1, Skipped: 1}, plan.Counts[EntityWorkflowStates])
	require.Len(t, plan.States, 1)
	assert.Equal(t, schema.StateTypeCanceled, plan.States[0].Type)

	require.Len(t, plan.Issues, 2)
	first, second := plan.Issues[0], plan.Issues[1]
	assert.Equal(t, 3, first.Number, "ENG-1 is taken, so it moves past the highest number")
	assert.Equal(t, 2, second.Number)
	assert.Equal(t, 3, plan.Counters[teamID])
	assert.Equal(t, todoID, first.StateID)
	assert.Equal(t, plan.States[0].ID, second.StateID)
	assert.Equal(t, adaID, *first.CreatorID)
	assert.Nil(t, first.AssigneeID, "the Linear bot has no account")

	small := uuid.Nil
	for _, label := range plan.Labels {
		if label.Name == "Small" {
			small = label.ID
		}
	}
	assert.ElementsMatch(t, []uuid.UUID{small, bugID}, first.LabelIDs, "groups and a second single-select choice are dropped; bug matches Bug")

	require.Len(t, plan.Comments, 3)
	root := plan.Comments[0]
	assert.Nil(t, root.ParentID)
	for _, reply := range plan.Comments[1:] {
		require.NotNil(t, reply.ParentID)
		assert.Equal(t, root.ID, *reply.ParentID, "replies to replies join the root thread")
	}

	assert.Equal(t, schema.ImportEntityCount{Created: 2, Skipped: 1}, plan.Counts[EntityRelations])
	kinds := []schema.IssueRelationType{}
	for _, relation := range plan.Relations {
		kinds = append(kinds, relation.Type)
		assert.Equal(t, first.ID, relation.IssueID)
	}
	assert.ElementsMatch(t, []schema.IssueRelationType{schema.IssueRelationRelated, schema.IssueRelationBlocks}, kinds)
}

func TestBuildSkipsMissingUsersWhenNotCreatingThem(t *testing.T) {
	options := defaultOptions()
	options.CreateMissingUsers = false
	plan := Build(smallExport(), &Snapshot{}, options)

	assert.Empty(t, plan.Users)
	assert.Equal(t, schema.ImportEntityCount{Skipped: 3}, plan.Counts[EntityUsers])
	for _, row := range plan.Issues {
		assert.Nil(t, row.CreatorID)
	}
}

func TestBuildUpdatesWhatAnEarlierImportCreated(t *testing.T) {
	export := smallExport()
	first := Build(export, &Snapshot{}, defaultOptions())

	snapshot := &Snapshot{}
	for _, u := range first.Users {
		snapshot.Users = append(snapshot.Users, &ent.User{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName})
		snapshot.Members = append(snapshot.Members, &ent.WorkspaceMember{UserID: u.ID})
	}
	for _, team := range first.Teams {
		snapshot.Teams = append(snapshot.Teams, &ent.Team{ID: team.ID, Key: team.Key, Name: team.Name, IssueCounter: first.Counters[team.ID]})
	}
	for _, state := range first.States {
		snapshot.States = append(snapshot.States, &ent.WorkflowState{ID: state.ID, TeamID: state.TeamID, Name: state.Name, Type: state.Type})
	}
	for _, label := range first.Labels {
		snapshot.Labels = append(snapshot.Labels, &ent.Label{ID: label.ID, TeamID: label.TeamID, ParentID: label.ParentID, Name: label.Name, IsGroup: label.IsGroup, GroupSingleSelect: label.GroupSingleSelect})
	}
	for _, issue := range first.Issues {
		snapshot.Issues = append(snapshot.Issues, &ent.Issue{ID: issue.ID, TeamID: issue.TeamID, Number: issue.Number, ImportSource: Source, ImportExternalID: issue.ExternalID})
		snapshot.HistoryIssueIDs = append(snapshot.HistoryIssueIDs, issue.ID)
	}
	for _, c := range first.Comments {
		snapshot.Comments = append(snapshot.Comments, &ent.Comment{ID: c.ID, IssueID: c.IssueID, ImportSource: Source, ImportExternalID: c.ExternalID})
	}
	for _, r := range first.Relations {
		snapshot.Relations = append(snapshot.Relations, &ent.IssueRelation{IssueID: r.IssueID, RelatedIssueID: r.RelatedIssueID, Type: r.Type})
	}

	second := Build(export, snapshot, defaultOptions())
	for _, entity := range Entities {
		assert.Zero(t, second.Counts[entity].Created, entity)
	}
	assert.Empty(t, second.Users)
	assert.Empty(t, second.Teams)
	assert.Empty(t, second.States)
	assert.Empty(t, second.Labels)
	assert.Empty(t, second.Relations)
	assert.Empty(t, second.History)
	assert.Empty(t, second.Counters)
	assert.Equal(t, 2, second.Counts[EntityIssues].Updated)
	assert.Equal(t, 3, second.Counts[EntityComments].Updated)
	for i, row := range second.Issues {
		assert.True(t, row.Existing)
		assert.Equal(t, first.Issues[i].ID, row.ID)
	}
}

func TestBuildRelatesToIssuesOfEarlierImports(t *testing.T) {
	earlierID, teamID := uuid.New(), uuid.New()
	export := smallExport()
	export.Issues[1].Relations = LinearRelations{DuplicateOf: strPtr("i-earlier")}
	snapshot := &Snapshot{
		Teams:  []*ent.Team{{ID: teamID, Key: "ENG", Name: "Engineering"}},
		Issues: []*ent.Issue{{ID: earlierID, TeamID: teamID, Number: 9, ImportSource: Source, ImportExternalID: "i-earlier"}},
	}
	plan := Build(export, snapshot, defaultOptions())

	duplicates := 0
	for _, relation := range plan.Relations {
		if relation.Type == schema.IssueRelationDuplicate {
			duplicates++
			assert.Equal(t, earlierID, relation.RelatedIssueID)
		}
	}
	assert.Equal(t, 1, duplicates)
	assert.Len(t, plan.Issues, 2, "issues outside the export are related to, not rewritten")
}

func strPtr(value string) *string { return &value }
