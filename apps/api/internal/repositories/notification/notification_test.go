package notification_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type NotificationRepositorySuite struct {
	repositorytest.RepositoryBaseSuite
	repo      *NotificationRepository
	now       time.Time
	user      *ent.User
	other     *ent.User
	workspace *ent.Workspace
	team      *ent.Team
	todo      *ent.WorkflowState
	done      *ent.WorkflowState
}

func (suite *NotificationRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewNotificationRepository(suite.DB)
	suite.now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	suite.user = suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.other = suite.DB.User.Create().SetEmail("bob@example.com").SetName("Bob").SetDisplayName("bob").SaveX(suite.Ctx)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.DB.WorkspaceMember.Create().SetWorkspaceID(suite.workspace.ID).SetUserID(suite.user.ID).SetRole(schema.WorkspaceRoleMember).SaveX(suite.Ctx)
	suite.team = suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Engineering").SetKey("ENG").SaveX(suite.Ctx)
	suite.todo = suite.DB.WorkflowState.Create().SetTeamID(suite.team.ID).SetName("Todo").SetType(schema.StateTypeUnstarted).SetColor("#e2e2e2").SetPosition(0).SaveX(suite.Ctx)
	suite.done = suite.DB.WorkflowState.Create().SetTeamID(suite.team.ID).SetName("Done").SetType(schema.StateTypeCompleted).SetColor("#5e6ad2").SetPosition(0).SaveX(suite.Ctx)
}

func (suite *NotificationRepositorySuite) notification(userID uuid.UUID, age time.Duration) *ent.Notification {
	return suite.DB.Notification.Create().
		SetUserID(userID).
		SetWorkspaceID(suite.workspace.ID).
		SetType(schema.NotificationIssueAssigned).
		SetTitle("Fix login").
		SetCreatedAt(suite.now.Add(-age)).
		SaveX(suite.Ctx)
}

func (suite *NotificationRepositorySuite) list(filter Filter, limit int, cursor *repository.TimeCursor) []*ent.Notification {
	rows, err := suite.repo.List(suite.Ctx, &ListInput{UserID: suite.user.ID, WorkspaceID: suite.workspace.ID, Filter: filter, Limit: limit, Cursor: cursor, Now: suite.now})
	suite.Require().NoError(err)
	return rows
}

func ids(rows []*ent.Notification) []uuid.UUID {
	out := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		out[i] = row.ID
	}
	return out
}

func (suite *NotificationRepositorySuite) TestFiltersSplitTheInbox() {
	unread := suite.notification(suite.user.ID, time.Minute)
	read := suite.notification(suite.user.ID, 2*time.Minute)
	archived := suite.notification(suite.user.ID, 3*time.Minute)
	snoozed := suite.notification(suite.user.ID, 4*time.Minute)
	woken := suite.notification(suite.user.ID, 5*time.Minute)
	suite.notification(suite.other.ID, time.Minute)

	suite.Require().NoError(suite.repo.SetRead(suite.Ctx, suite.user.ID, []uuid.UUID{read.ID}, &suite.now))
	suite.Require().NoError(suite.repo.SetArchived(suite.Ctx, suite.user.ID, []uuid.UUID{archived.ID}, &suite.now))
	suite.Require().NoError(suite.repo.Snooze(suite.Ctx, suite.user.ID, []uuid.UUID{snoozed.ID}, suite.now.Add(time.Hour)))
	suite.Require().NoError(suite.repo.Snooze(suite.Ctx, suite.user.ID, []uuid.UUID{woken.ID}, suite.now.Add(-time.Minute)))

	suite.Equal([]uuid.UUID{unread.ID, read.ID, woken.ID}, ids(suite.list(FilterInbox, 50, nil)))
	suite.Equal([]uuid.UUID{unread.ID, woken.ID}, ids(suite.list(FilterUnread, 50, nil)))
	suite.Equal([]uuid.UUID{read.ID}, ids(suite.list(FilterRead, 50, nil)))
	suite.Equal([]uuid.UUID{archived.ID}, ids(suite.list(FilterArchived, 50, nil)))
	suite.Equal([]uuid.UUID{snoozed.ID}, ids(suite.list(FilterSnoozed, 50, nil)))
	suite.Len(suite.list(FilterAll, 50, nil), 5)

	count, err := suite.repo.CountUnread(suite.Ctx, suite.user.ID, suite.workspace.ID, suite.now)
	suite.Require().NoError(err)
	suite.Equal(2, count)

	archivedRow := suite.DB.Notification.GetX(suite.Ctx, archived.ID)
	suite.NotNil(archivedRow.ReadAt, "archiving reads the notification")
}

func (suite *NotificationRepositorySuite) TestCursorPagesWithoutGapsOrRepeats() {
	var created []uuid.UUID
	for i := range 5 {
		created = append(created, suite.notification(suite.user.ID, time.Duration(i)*time.Minute).ID)
	}
	same := suite.now.Add(-10 * time.Minute)
	twinA := suite.DB.Notification.Create().SetUserID(suite.user.ID).SetWorkspaceID(suite.workspace.ID).SetType(schema.NotificationReaction).SetTitle("a").SetCreatedAt(same).SaveX(suite.Ctx)
	twinB := suite.DB.Notification.Create().SetUserID(suite.user.ID).SetWorkspaceID(suite.workspace.ID).SetType(schema.NotificationReaction).SetTitle("b").SetCreatedAt(same).SaveX(suite.Ctx)

	var seen []uuid.UUID
	var cursor *repository.TimeCursor
	for page := 0; page < 10; page++ {
		rows := suite.list(FilterInbox, 2, cursor)
		more := len(rows) > 2
		if more {
			rows = rows[:2]
		}
		seen = append(seen, ids(rows)...)
		if !more {
			break
		}
		last := rows[len(rows)-1]
		encoded := repository.TimeCursor{CreatedAt: last.CreatedAt, ID: last.ID}.Encode()
		decoded, err := repository.DecodeTimeCursor(encoded)
		suite.Require().NoError(err)
		cursor = decoded
	}
	suite.Len(seen, 7)
	suite.Equal(created, seen[:5])
	suite.ElementsMatch([]uuid.UUID{twinA.ID, twinB.ID}, seen[5:])

	_, err := repository.DecodeTimeCursor("garbage")
	suite.ErrorIs(err, repository.ErrInvalidCursor)
}

func (suite *NotificationRepositorySuite) TestWritesTouchOnlyTheOwnersRows() {
	mine := suite.notification(suite.user.ID, time.Minute)
	theirs := suite.notification(suite.other.ID, time.Minute)

	owned, err := suite.repo.OwnedWorkspaces(suite.Ctx, suite.user.ID, []uuid.UUID{mine.ID, theirs.ID})
	suite.Require().NoError(err)
	suite.Equal(map[uuid.UUID]uuid.UUID{mine.ID: suite.workspace.ID}, owned)

	suite.Require().NoError(suite.repo.SetRead(suite.Ctx, suite.user.ID, []uuid.UUID{mine.ID, theirs.ID}, &suite.now))
	suite.Nil(suite.DB.Notification.GetX(suite.Ctx, theirs.ID).ReadAt)
	suite.NotNil(suite.DB.Notification.GetX(suite.Ctx, mine.ID).ReadAt)

	suite.Require().NoError(suite.repo.SetRead(suite.Ctx, suite.user.ID, []uuid.UUID{mine.ID}, nil))
	suite.Nil(suite.DB.Notification.GetX(suite.Ctx, mine.ID).ReadAt)

	suite.Require().NoError(suite.repo.MarkAllRead(suite.Ctx, suite.user.ID, suite.workspace.ID, suite.now))
	suite.NotNil(suite.DB.Notification.GetX(suite.Ctx, mine.ID).ReadAt)
	suite.Nil(suite.DB.Notification.GetX(suite.Ctx, theirs.ID).ReadAt)
}

func (suite *NotificationRepositorySuite) TestIssuesOfPrivateTeamsReachOnlyTheirMembersAndAdmins() {
	admin := suite.DB.User.Create().SetEmail("root@example.com").SetName("Root").SetDisplayName("root").SaveX(suite.Ctx)
	suite.DB.WorkspaceMember.Create().SetWorkspaceID(suite.workspace.ID).SetUserID(admin.ID).SetRole(schema.WorkspaceRoleAdmin).SaveX(suite.Ctx)
	suite.DB.WorkspaceMember.Create().SetWorkspaceID(suite.workspace.ID).SetUserID(suite.other.ID).SetRole(schema.WorkspaceRoleMember).SaveX(suite.Ctx)
	private := suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Security").SetKey("SEC").SetPrivate(true).SaveX(suite.Ctx)
	state := suite.DB.WorkflowState.Create().SetTeamID(private.ID).SetName("Todo").SetType(schema.StateTypeUnstarted).SetColor("#e2e2e2").SetPosition(0).SaveX(suite.Ctx)
	suite.DB.TeamMember.Create().SetTeamID(private.ID).SetUserID(suite.other.ID).SaveX(suite.Ctx)
	secret := suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(private.ID).SetNumber(1).SetTitle("Secret").SetStateID(state.ID).SaveX(suite.Ctx)
	open := suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(suite.team.ID).SetNumber(1).SetTitle("Open").SetStateID(suite.todo.ID).SaveX(suite.Ctx)
	everyone := []uuid.UUID{suite.user.ID, suite.other.ID, admin.ID}

	members, err := suite.repo.ActiveMembers(suite.Ctx, suite.workspace.ID, everyone, &secret.ID)
	suite.Require().NoError(err)
	suite.ElementsMatch([]uuid.UUID{suite.other.ID, admin.ID}, members, "a member outside the private team hears nothing about its issues")

	members, err = suite.repo.ActiveMembers(suite.Ctx, suite.workspace.ID, everyone, &open.ID)
	suite.Require().NoError(err)
	suite.ElementsMatch(everyone, members)
}

func (suite *NotificationRepositorySuite) TestActiveMembersAndSettings() {
	inactive := suite.DB.User.Create().SetEmail("gone@example.com").SetName("Gone").SetDisplayName("gone").SetActive(false).SaveX(suite.Ctx)
	suite.DB.WorkspaceMember.Create().SetWorkspaceID(suite.workspace.ID).SetUserID(inactive.ID).SetRole(schema.WorkspaceRoleMember).SaveX(suite.Ctx)

	members, err := suite.repo.ActiveMembers(suite.Ctx, suite.workspace.ID, []uuid.UUID{suite.user.ID, suite.other.ID, inactive.ID}, nil)
	suite.Require().NoError(err)
	suite.Equal([]uuid.UUID{suite.user.ID}, members)

	suite.DB.NotificationSetting.Create().SetUserID(suite.user.ID).SetEmailEnabled(false).SaveX(suite.Ctx)
	settings, err := suite.repo.Settings(suite.Ctx, []uuid.UUID{suite.user.ID, suite.other.ID})
	suite.Require().NoError(err)
	suite.Len(settings, 1)
	suite.False(settings[suite.user.ID].EmailEnabled)
}

func (suite *NotificationRepositorySuite) TestPendingEmailAndMarkEmailed() {
	due := suite.notification(suite.user.ID, 10*time.Minute)
	suite.notification(suite.user.ID, time.Minute)
	suite.notification(suite.user.ID, 48*time.Hour)
	read := suite.notification(suite.user.ID, 10*time.Minute)
	suite.Require().NoError(suite.repo.SetRead(suite.Ctx, suite.user.ID, []uuid.UUID{read.ID}, &suite.now))

	pending, err := suite.repo.PendingEmail(suite.Ctx, suite.now.Add(-2*time.Minute), suite.now.Add(-24*time.Hour))
	suite.Require().NoError(err)
	suite.Equal([]uuid.UUID{due.ID}, ids(pending))
	suite.Equal(suite.user.Email, pending[0].Edges.User.Email)
	suite.Equal("acme", pending[0].Edges.Workspace.Slug)

	suite.Require().NoError(suite.repo.MarkEmailed(suite.Ctx, []uuid.UUID{due.ID}, suite.now))
	pending, err = suite.repo.PendingEmail(suite.Ctx, suite.now.Add(-2*time.Minute), suite.now.Add(-24*time.Hour))
	suite.Require().NoError(err)
	suite.Empty(pending)
}

func (suite *NotificationRepositorySuite) issue(number int, state *ent.WorkflowState, due *time.Time, assignee *uuid.UUID) *ent.Issue {
	return suite.DB.Issue.Create().
		SetWorkspaceID(suite.workspace.ID).
		SetTeamID(suite.team.ID).
		SetNumber(number).
		SetTitle("Issue").
		SetStateID(state.ID).
		SetNillableDueDate(due).
		SetNillableAssigneeID(assignee).
		SaveX(suite.Ctx)
}

func (suite *NotificationRepositorySuite) TestDueBetweenAndHasRecent() {
	today := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	tomorrow := today.AddDate(0, 0, 1)
	later := today.AddDate(0, 0, 3)
	yesterday := today.AddDate(0, 0, -1)
	open := suite.issue(1, suite.todo, &tomorrow, &suite.user.ID)
	suite.issue(2, suite.done, &tomorrow, &suite.user.ID)
	suite.issue(3, suite.todo, &tomorrow, nil)
	suite.issue(4, suite.todo, &later, &suite.user.ID)
	suite.issue(5, suite.todo, &yesterday, &suite.user.ID)
	suite.issue(6, suite.todo, nil, &suite.user.ID)

	due, err := suite.repo.DueBetween(suite.Ctx, today, suite.now.Add(24*time.Hour))
	suite.Require().NoError(err)
	suite.Require().Len(due, 1)
	suite.Equal(open.ID, due[0].ID)
	suite.Equal("ENG", due[0].Edges.Team.Key)

	recent, err := suite.repo.HasRecent(suite.Ctx, suite.user.ID, open.ID, schema.NotificationIssueDueSoon, suite.now.Add(-48*time.Hour))
	suite.Require().NoError(err)
	suite.False(recent)
	suite.DB.Notification.Create().SetUserID(suite.user.ID).SetWorkspaceID(suite.workspace.ID).SetIssueID(open.ID).SetType(schema.NotificationIssueDueSoon).SetTitle("Issue").SetCreatedAt(suite.now.Add(-time.Hour)).SaveX(suite.Ctx)
	recent, err = suite.repo.HasRecent(suite.Ctx, suite.user.ID, open.ID, schema.NotificationIssueDueSoon, suite.now.Add(-48*time.Hour))
	suite.Require().NoError(err)
	suite.True(recent)
}

func (suite *NotificationRepositorySuite) TestCreateMany() {
	issue := suite.issue(1, suite.todo, nil, nil)
	created, err := suite.repo.CreateMany(suite.Ctx, nil, []*CreateNotificationInput{
		{UserID: suite.user.ID, WorkspaceID: suite.workspace.ID, Type: schema.NotificationIssueAssigned, IssueID: &issue.ID, Title: "Issue", ActorID: &suite.other.ID},
		{UserID: suite.other.ID, WorkspaceID: suite.workspace.ID, Type: schema.NotificationIssueAssigned, IssueID: &issue.ID, Title: "Issue"},
	})
	suite.Require().NoError(err)
	suite.Len(created, 2)

	rows := suite.list(FilterInbox, 50, nil)
	suite.Require().Len(rows, 1)
	suite.Equal("Bob", rows[0].Edges.Actor.Name)
	suite.Equal("ENG", rows[0].Edges.Issue.Edges.Team.Key)
	suite.Equal("Todo", rows[0].Edges.Issue.Edges.State.Name)
}

func TestNotificationRepositorySuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositorySuite))
}
