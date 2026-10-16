package notification_service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/mail"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	notification_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/notification"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type NotificationServiceSuite struct {
	services.ServiceTestSuite
	service     *NotificationService
	now         time.Time
	userID      uuid.UUID
	actorID     uuid.UUID
	workspaceID uuid.UUID
}

func (suite *NotificationServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewNotificationService(&config.Config{ExternalUIUrl: "https://zigzag.example.com"}, suite.MockRepo, suite.MockEmail)
	suite.now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID = uuid.New()
	suite.actorID = uuid.New()
	suite.workspaceID = uuid.New()
}

func (suite *NotificationServiceSuite) viewer(workspaceID uuid.UUID, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: workspaceID}}).Return(result).Maybe()
}

func (suite *NotificationServiceSuite) TestNotifySkipsTheActorDuplicatesAndOutsiders() {
	outsider := uuid.New()
	suite.MockNotificationRepo.EXPECT().ActiveMembers(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.userID, outsider}, mock.Anything).Return([]uuid.UUID{suite.userID}, nil).Once()
	suite.MockNotificationRepo.EXPECT().Settings(suite.Ctx, []uuid.UUID{suite.userID}).Return(map[uuid.UUID]*ent.NotificationSetting{}, nil).Once()
	var created []*notification_repo.CreateNotificationInput
	suite.MockNotificationRepo.EXPECT().CreateMany(suite.Ctx, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, inputs []*notification_repo.CreateNotificationInput) ([]*ent.Notification, error) {
		created = inputs
		return nil, nil
	}).Once()

	issueID := uuid.New()
	err := suite.service.Notify(suite.Ctx, models.NotificationEvent{
		Type:         schema.NotificationIssueAssigned,
		WorkspaceID:  suite.workspaceID,
		ActorID:      &suite.actorID,
		RecipientIDs: []uuid.UUID{suite.actorID, suite.userID, suite.userID, uuid.Nil, outsider},
		IssueID:      &issueID,
		Title:        "Fix login",
		Body:         strings.Repeat("a", 400),
	})
	suite.Require().NoError(err)
	suite.Require().Len(created, 1)
	suite.Equal(suite.userID, created[0].UserID)
	suite.Equal(&suite.actorID, created[0].ActorID)
	suite.Equal(&issueID, created[0].IssueID)
	suite.Equal("Fix login", created[0].Title)
	suite.Len([]rune(created[0].Body), bodyMaxRunes)
}

func (suite *NotificationServiceSuite) TestNotifyOnlyTheActorWritesNothing() {
	err := suite.service.Notify(suite.Ctx, models.NotificationEvent{
		Type:         schema.NotificationIssueCommented,
		WorkspaceID:  suite.workspaceID,
		ActorID:      &suite.actorID,
		RecipientIDs: []uuid.UUID{suite.actorID},
		Title:        "Fix login",
	})
	suite.NoError(err)
}

func (suite *NotificationServiceSuite) TestNotifyHonorsSettings() {
	muted := uuid.New()
	suite.MockNotificationRepo.EXPECT().ActiveMembers(suite.Ctx, suite.workspaceID, mock.Anything, mock.Anything).Return([]uuid.UUID{suite.userID, muted}, nil).Once()
	suite.MockNotificationRepo.EXPECT().Settings(suite.Ctx, mock.Anything).Return(map[uuid.UUID]*ent.NotificationSetting{
		suite.userID: {UserID: suite.userID, Mentioned: true},
		muted:        {UserID: muted, Mentioned: false, Assigned: true},
	}, nil).Once()
	suite.MockNotificationRepo.EXPECT().CreateMany(suite.Ctx, mock.Anything, mock.MatchedBy(func(inputs []*notification_repo.CreateNotificationInput) bool {
		return len(inputs) == 1 && inputs[0].UserID == suite.userID
	})).Return(nil, nil).Once()

	err := suite.service.Notify(suite.Ctx, models.NotificationEvent{
		Type:         schema.NotificationCommentMentioned,
		WorkspaceID:  suite.workspaceID,
		RecipientIDs: []uuid.UUID{suite.userID, muted},
		Title:        "Fix login",
	})
	suite.NoError(err)
}

func (suite *NotificationServiceSuite) TestWantsMapsEveryTypeToItsSetting() {
	off := &ent.NotificationSetting{}
	for _, kind := range []schema.NotificationType{
		schema.NotificationIssueAssigned, schema.NotificationIssueUnassigned, schema.NotificationIssueMentioned,
		schema.NotificationCommentMentioned, schema.NotificationIssueCommented, schema.NotificationIssueStateChanged,
		schema.NotificationIssuePriorityChanged, schema.NotificationIssueBlocking, schema.NotificationIssueDueSoon,
		schema.NotificationProjectUpdate,
	} {
		suite.False(wants(off, kind), kind)
		suite.True(wants(nil, kind), kind)
	}
	for _, kind := range []schema.NotificationType{schema.NotificationProjectMember, schema.NotificationProjectLead, schema.NotificationReaction} {
		suite.True(wants(off, kind), kind)
	}
}

func (suite *NotificationServiceSuite) TestListMasksForeignWorkspaces() {
	suite.viewer(suite.workspaceID, errdefs.ErrUnauthorized)
	_, _, err := suite.service.List(suite.Ctx, suite.userID, &ListNotificationsInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *NotificationServiceSuite) TestListRejectsForeignCursors() {
	suite.viewer(suite.workspaceID, nil)
	input := &ListNotificationsInput{WorkspaceID: suite.workspaceID}
	input.Cursor = "not-a-cursor"
	_, _, err := suite.service.List(suite.Ctx, suite.userID, input)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *NotificationServiceSuite) TestListPaginates() {
	suite.viewer(suite.workspaceID, nil)
	rows := []*ent.Notification{
		{ID: uuid.New(), Type: schema.NotificationIssueAssigned, CreatedAt: suite.now},
		{ID: uuid.New(), Type: schema.NotificationIssueAssigned, CreatedAt: suite.now.Add(-time.Minute)},
		{ID: uuid.New(), Type: schema.NotificationIssueAssigned, CreatedAt: suite.now.Add(-2 * time.Minute)},
	}
	suite.MockNotificationRepo.EXPECT().List(suite.Ctx, mock.MatchedBy(func(input *notification_repo.ListInput) bool {
		return input.UserID == suite.userID && input.Filter == notification_repo.FilterInbox && input.Limit == 2 && input.Cursor == nil
	})).Return(rows, nil).Once()

	input := &ListNotificationsInput{WorkspaceID: suite.workspaceID}
	input.Limit = 2
	page, next, err := suite.service.List(suite.Ctx, suite.userID, input)
	suite.Require().NoError(err)
	suite.Len(page, 2)
	suite.Require().NotNil(next)
	cursor, err := repository.DecodeTimeCursor(*next)
	suite.Require().NoError(err)
	suite.Equal(rows[1].ID, cursor.ID)
	suite.True(rows[1].CreatedAt.Equal(cursor.CreatedAt))

	suite.MockNotificationRepo.EXPECT().List(suite.Ctx, mock.MatchedBy(func(input *notification_repo.ListInput) bool {
		return input.Filter == notification_repo.FilterArchived && input.Cursor != nil && input.Cursor.ID == rows[1].ID
	})).Return(rows[2:], nil).Once()
	input.Filter = "archived"
	input.Cursor = *next
	page, next, err = suite.service.List(suite.Ctx, suite.userID, input)
	suite.Require().NoError(err)
	suite.Len(page, 1)
	suite.Nil(next)
}

func (suite *NotificationServiceSuite) TestCounts() {
	suite.viewer(suite.workspaceID, nil)
	suite.MockNotificationRepo.EXPECT().CountUnread(suite.Ctx, suite.userID, suite.workspaceID, suite.now).Return(3, nil).Once()
	counts, err := suite.service.Counts(suite.Ctx, suite.userID, &WorkspaceInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Equal(3, counts.Unread)

	other := uuid.New()
	suite.viewer(other, errdefs.ErrUnauthorized)
	_, err = suite.service.Counts(suite.Ctx, suite.userID, &WorkspaceInput{WorkspaceID: other})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *NotificationServiceSuite) TestMarkReadRefusesSomeoneElsesNotification() {
	mine, theirs := uuid.New(), uuid.New()
	suite.MockNotificationRepo.EXPECT().OwnedWorkspaces(suite.Ctx, suite.userID, []uuid.UUID{mine, theirs}).Return(map[uuid.UUID]uuid.UUID{mine: suite.workspaceID}, nil).Once()
	err := suite.service.MarkRead(suite.Ctx, suite.userID, &NotificationIDsInput{NotificationIDs: []uuid.UUID{mine, theirs}})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *NotificationServiceSuite) TestMarkReadRefusesWorkspacesOutOfReach() {
	id := uuid.New()
	suite.MockNotificationRepo.EXPECT().OwnedWorkspaces(suite.Ctx, suite.userID, []uuid.UUID{id}).Return(map[uuid.UUID]uuid.UUID{id: suite.workspaceID}, nil).Once()
	suite.viewer(suite.workspaceID, errdefs.ErrUnauthorized)
	err := suite.service.MarkRead(suite.Ctx, suite.userID, &NotificationIDsInput{NotificationIDs: []uuid.UUID{id}})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *NotificationServiceSuite) expectOwned(ids ...uuid.UUID) {
	owned := map[uuid.UUID]uuid.UUID{}
	for _, id := range ids {
		owned[id] = suite.workspaceID
	}
	suite.MockNotificationRepo.EXPECT().OwnedWorkspaces(suite.Ctx, suite.userID, ids).Return(owned, nil).Once()
	suite.viewer(suite.workspaceID, nil)
}

func (suite *NotificationServiceSuite) TestReadUnreadArchiveUnarchive() {
	a, b := uuid.New(), uuid.New()
	ids := []uuid.UUID{a, b}
	input := &NotificationIDsInput{NotificationIDs: ids}

	suite.expectOwned(ids...)
	suite.MockNotificationRepo.EXPECT().SetRead(suite.Ctx, suite.userID, ids, &suite.now).Return(nil).Once()
	suite.NoError(suite.service.MarkRead(suite.Ctx, suite.userID, input))

	suite.expectOwned(ids...)
	suite.MockNotificationRepo.EXPECT().SetRead(suite.Ctx, suite.userID, ids, (*time.Time)(nil)).Return(nil).Once()
	suite.NoError(suite.service.MarkUnread(suite.Ctx, suite.userID, input))

	suite.expectOwned(ids...)
	suite.MockNotificationRepo.EXPECT().SetArchived(suite.Ctx, suite.userID, ids, &suite.now).Return(nil).Once()
	suite.NoError(suite.service.Archive(suite.Ctx, suite.userID, input))

	suite.expectOwned(ids...)
	suite.MockNotificationRepo.EXPECT().SetArchived(suite.Ctx, suite.userID, ids, (*time.Time)(nil)).Return(nil).Once()
	suite.NoError(suite.service.Unarchive(suite.Ctx, suite.userID, input))
}

func (suite *NotificationServiceSuite) TestEmptyIDsAreInvalid() {
	err := suite.service.Archive(suite.Ctx, suite.userID, &NotificationIDsInput{})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *NotificationServiceSuite) TestSnoozeNeedsAFutureTime() {
	id := uuid.New()
	err := suite.service.Snooze(suite.Ctx, suite.userID, &SnoozeNotificationsInput{NotificationIDs: []uuid.UUID{id}, Until: suite.now})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	until := suite.now.Add(time.Hour)
	suite.expectOwned(id)
	suite.MockNotificationRepo.EXPECT().Snooze(suite.Ctx, suite.userID, []uuid.UUID{id}, until).Return(nil).Once()
	suite.NoError(suite.service.Snooze(suite.Ctx, suite.userID, &SnoozeNotificationsInput{NotificationIDs: []uuid.UUID{id}, Until: until}))
}

func (suite *NotificationServiceSuite) TestMarkAllRead() {
	suite.viewer(suite.workspaceID, nil)
	suite.MockNotificationRepo.EXPECT().MarkAllRead(suite.Ctx, suite.userID, suite.workspaceID, suite.now).Return(nil).Once()
	suite.NoError(suite.service.MarkAllRead(suite.Ctx, suite.userID, &MarkAllReadInput{WorkspaceID: suite.workspaceID}))

	other := uuid.New()
	suite.viewer(other, errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.service.MarkAllRead(suite.Ctx, suite.userID, &MarkAllReadInput{WorkspaceID: other}), errdefs.ErrNotFound)
}

func (suite *NotificationServiceSuite) TestEmailBatchDoesNothingWithoutMail() {
	suite.MockEmail.EXPECT().Configured(suite.Ctx).Return(false, nil).Once()
	suite.NoError(suite.service.SendEmailBatch(suite.Ctx))
}

func (suite *NotificationServiceSuite) pending(user *ent.User, title string) *ent.Notification {
	issue := &ent.Issue{ID: uuid.New(), Number: 12, Title: title, Edges: ent.IssueEdges{Team: &ent.Team{Key: "ENG"}}}
	row := suite.row(user, issue, schema.NotificationIssueAssigned, "Assigned to you")
	row.Edges.Actor = &ent.User{ID: suite.actorID, Name: "Ada Lovelace"}
	return row
}

func (suite *NotificationServiceSuite) row(user *ent.User, issue *ent.Issue, kind schema.NotificationType, body string) *ent.Notification {
	row := &ent.Notification{ID: uuid.New(), UserID: user.ID, WorkspaceID: suite.workspaceID, Type: kind, Body: body, CreatedAt: suite.now.Add(-10 * time.Minute)}
	row.Edges.User = user
	row.Edges.Workspace = &ent.Workspace{ID: suite.workspaceID, Name: "Acme", Slug: "acme"}
	if issue != nil {
		row.IssueID = &issue.ID
		row.Title = "ENG-" + strconv.Itoa(issue.Number) + " " + issue.Title
		row.Edges.Issue = issue
	}
	return row
}

func (suite *NotificationServiceSuite) TestEmailBatchSendsOneEmailPerUser() {
	grace := &ent.User{ID: uuid.New(), Email: "grace@example.com", Active: true}
	quiet := &ent.User{ID: uuid.New(), Email: "quiet@example.com", Active: true}
	gone := &ent.User{ID: uuid.New(), Email: "gone@example.com", Active: false}
	first, second := suite.pending(grace, "Fix login"), suite.pending(grace, "Ship it")
	muted, inactive := suite.pending(quiet, "Muted"), suite.pending(gone, "Gone")

	suite.MockEmail.EXPECT().Configured(suite.Ctx).Return(true, nil).Once()
	suite.MockNotificationRepo.EXPECT().PendingEmail(suite.Ctx, suite.now.Add(-emailDelay), suite.now.Add(-emailMaxAge)).Return([]*ent.Notification{first, muted, second, inactive}, nil).Once()
	suite.MockNotificationRepo.EXPECT().Settings(suite.Ctx, []uuid.UUID{grace.ID, quiet.ID, gone.ID}).Return(map[uuid.UUID]*ent.NotificationSetting{
		quiet.ID: {UserID: quiet.ID, EmailEnabled: false},
	}, nil).Once()
	var sent mail.Content
	suite.MockEmail.EXPECT().Send(suite.Ctx, "grace@example.com", mock.Anything).RunAndReturn(func(_ context.Context, _ string, content mail.Content) (bool, error) {
		sent = content
		return true, nil
	}).Once()
	suite.MockNotificationRepo.EXPECT().MarkEmailed(suite.Ctx, []uuid.UUID{first.ID, second.ID}, suite.now).Return(nil).Once()
	suite.MockNotificationRepo.EXPECT().MarkEmailed(suite.Ctx, []uuid.UUID{muted.ID}, suite.now).Return(nil).Once()
	suite.MockNotificationRepo.EXPECT().MarkEmailed(suite.Ctx, []uuid.UUID{inactive.ID}, suite.now).Return(nil).Once()

	suite.Require().NoError(suite.service.SendEmailBatch(suite.Ctx))
	suite.Equal("2 unread notifications", sent.Subject)
	suite.Equal("You have 2 unread notifications in Acme.", sent.Heading)
	suite.Equal([]string{"Fix login", "Ship it"}, []string{sent.Items[0].Title, sent.Items[1].Title})
	suite.Equal(mail.Button{Label: "Open your inbox", URL: "https://zigzag.example.com/acme/inbox"}, sent.Button)
	suite.Equal("https://zigzag.example.com/acme/settings/account/notifications", sent.Footer.ManageURL)
	suite.Equal("Notification settings", sent.Footer.ManageLabel)
}

func (suite *NotificationServiceSuite) TestDigestGroupsNotificationsPerIssue() {
	grace := &ent.User{ID: uuid.New(), Email: "grace@example.com", Active: true}
	ada := &ent.User{ID: suite.actorID, Name: "Ada Lovelace"}
	handle := &ent.User{ID: uuid.New(), DisplayName: "ngyekta"}
	login := &ent.Issue{ID: uuid.New(), Number: 12, Title: "Fix login", Edges: ent.IssueEdges{Team: &ent.Team{Key: "ENG"}}}
	billing := &ent.Issue{ID: uuid.New(), Number: 7, Title: "Billing page", Edges: ent.IssueEdges{Team: &ent.Team{Key: "ENG"}}}
	project := &ent.Project{ID: uuid.New(), Name: "Launch", Slug: "launch-1a2b3c4d"}

	with := func(row *ent.Notification, actor *ent.User) *ent.Notification {
		row.Edges.Actor = actor
		return row
	}
	projectLead := with(suite.row(grace, nil, schema.NotificationProjectLead, ""), ada)
	projectLead.ProjectID = &project.ID
	projectLead.Title = "You were made lead of Launch"
	projectLead.Edges.Project = project
	batch := []*ent.Notification{
		with(suite.row(grace, login, schema.NotificationIssueAssigned, "Assigned to you"), handle),
		with(suite.row(grace, billing, schema.NotificationIssueMentioned, "See @grace"), ada),
		with(suite.row(grace, login, schema.NotificationIssueStateChanged, "Moved to Done"), ada),
		with(suite.row(grace, login, schema.NotificationIssuePriorityChanged, "Priority set to High"), ada),
		with(suite.row(grace, billing, schema.NotificationIssueCommented, "Looks good"), handle),
		suite.row(grace, billing, schema.NotificationIssueDueSoon, "Due today"),
		projectLead,
	}

	content := suite.service.digest(batch)

	suite.Equal("You have 7 unread notifications in Acme.", content.Heading)
	suite.Equal("ngyekta assigned the issue to you", content.Preheader)
	suite.Require().Len(content.Items, 3)
	suite.Equal(mail.Item{
		Identifier: "ENG-12",
		Title:      "Fix login",
		URL:        "https://zigzag.example.com/acme/issue/ENG-12",
		Lines: []mail.Line{
			mail.ActorLine(handle.ID.String(), "ngyekta", "assigned the issue to you"),
			mail.ActorLine(ada.ID.String(), "Ada Lovelace", "changed the status to Done"),
			mail.ActorLine(ada.ID.String(), "Ada Lovelace", "changed the priority to High"),
		},
	}, content.Items[0])
	suite.Equal([]mail.Line{
		mail.ActorLine(ada.ID.String(), "Ada Lovelace", "mentioned you"),
		mail.ActorLine(handle.ID.String(), "ngyekta", "commented"),
		{Text: "Due today"},
	}, content.Items[1].Lines)
	suite.Equal(mail.Item{
		Title: "Launch",
		URL:   "https://zigzag.example.com/acme/project/launch-1a2b3c4d",
		Lines: []mail.Line{mail.ActorLine(ada.ID.String(), "Ada Lovelace", "made you the project lead")},
	}, content.Items[2])
	suite.Empty(content.Note)
}

func (suite *NotificationServiceSuite) TestDigestCapsTheListAndNamesOneNotificationInTheSubject() {
	grace := &ent.User{ID: uuid.New(), Email: "grace@example.com", Active: true}
	single := suite.service.digest([]*ent.Notification{suite.pending(grace, "Fix login")})
	suite.Equal("ENG-12 Fix login: Ada Lovelace assigned the issue to you", single.Subject)
	suite.Equal("You have an unread notification in Acme.", single.Heading)

	batch := make([]*ent.Notification, emailMaxListItems+3)
	for i := range batch {
		batch[i] = suite.pending(grace, "Issue "+strconv.Itoa(i))
	}
	batch[0].WorkspaceID = uuid.New()
	capped := suite.service.digest(batch)
	suite.Len(capped.Items, emailMaxListItems)
	suite.Equal("And 3 more in your inbox.", capped.Note)
	suite.Equal("You have 23 unread notifications.", capped.Heading, "a batch across workspaces names none")
}

func (suite *NotificationServiceSuite) TestEmailBatchKeepsUnsentNotificationsPending() {
	grace := &ent.User{ID: uuid.New(), Email: "grace@example.com", Active: true}
	row := suite.pending(grace, "Fix login")
	suite.MockEmail.EXPECT().Configured(suite.Ctx).Return(true, nil).Once()
	suite.MockNotificationRepo.EXPECT().PendingEmail(suite.Ctx, mock.Anything, mock.Anything).Return([]*ent.Notification{row}, nil).Once()
	suite.MockNotificationRepo.EXPECT().Settings(suite.Ctx, mock.Anything).Return(map[uuid.UUID]*ent.NotificationSetting{}, nil).Once()
	suite.MockEmail.EXPECT().Send(suite.Ctx, "grace@example.com", mock.Anything).Return(false, errors.New("relay down")).Once()

	suite.Error(suite.service.SendEmailBatch(suite.Ctx))
}

func (suite *NotificationServiceSuite) TestDueSoonNotifiesTheAssigneeOnce() {
	today := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	tomorrow := today.AddDate(0, 0, 1)
	reminded := uuid.New()
	team := ent.IssueEdges{Team: &ent.Team{Key: "ENG"}}
	dueToday := &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, Number: 4, Title: "Ship", AssigneeID: &suite.userID, DueDate: &today, Edges: team}
	dueTomorrow := &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, Number: 5, Title: "Test", AssigneeID: &suite.userID, DueDate: &tomorrow, Edges: team}
	alreadyReminded := &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, Number: 6, Title: "Reminded", AssigneeID: &reminded, DueDate: &today, Edges: team}

	suite.MockNotificationRepo.EXPECT().DueBetween(suite.Ctx, today, suite.now.Add(dueSoonWindow)).Return([]*ent.Issue{dueToday, dueTomorrow, alreadyReminded}, nil).Once()
	since := suite.now.Add(-dueSoonRepeat)
	suite.MockNotificationRepo.EXPECT().HasRecent(suite.Ctx, suite.userID, dueToday.ID, schema.NotificationIssueDueSoon, since).Return(false, nil).Once()
	suite.MockNotificationRepo.EXPECT().HasRecent(suite.Ctx, suite.userID, dueTomorrow.ID, schema.NotificationIssueDueSoon, since).Return(false, nil).Once()
	suite.MockNotificationRepo.EXPECT().HasRecent(suite.Ctx, reminded, alreadyReminded.ID, schema.NotificationIssueDueSoon, since).Return(true, nil).Once()
	suite.MockNotificationRepo.EXPECT().ActiveMembers(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.userID}, mock.Anything).Return([]uuid.UUID{suite.userID}, nil).Twice()
	suite.MockNotificationRepo.EXPECT().Settings(suite.Ctx, []uuid.UUID{suite.userID}).Return(map[uuid.UUID]*ent.NotificationSetting{}, nil).Twice()
	var bodies, titles []string
	suite.MockNotificationRepo.EXPECT().CreateMany(suite.Ctx, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, inputs []*notification_repo.CreateNotificationInput) ([]*ent.Notification, error) {
		suite.Equal(schema.NotificationIssueDueSoon, inputs[0].Type)
		suite.Nil(inputs[0].ActorID)
		bodies = append(bodies, inputs[0].Body)
		titles = append(titles, inputs[0].Title)
		return nil, nil
	}).Twice()

	suite.Require().NoError(suite.service.NotifyDueSoon(suite.Ctx))
	suite.Equal([]string{"Due today", "Due tomorrow"}, bodies)
	suite.Equal([]string{"ENG-4 Ship", "ENG-5 Test"}, titles, "titles carry the identifier like every other issue notification")
}

func TestNotificationServiceSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceSuite))
}
