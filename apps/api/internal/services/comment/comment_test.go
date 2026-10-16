package comment_service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	comment_repo "github.com/unbindapp/zigzag-api/internal/repositories/comment"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	mocks_services_githubsync "github.com/unbindapp/zigzag-api/mocks/services/githubsync"
	mocks_services_notification "github.com/unbindapp/zigzag-api/mocks/services/notification"
	mocks_services_webhook "github.com/unbindapp/zigzag-api/mocks/services/webhook"
)

type CommentServiceSuite struct {
	services.ServiceTestSuite
	service     *CommentService
	githubSync  *mocks_services_githubsync.GithubSyncServiceMock
	userID      uuid.UUID
	otherID     uuid.UUID
	workspaceID uuid.UUID
	issue       *ent.Issue
	events      []models.NotificationEvent
	emitted     []string
}

func (suite *CommentServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	notifications := mocks_services_notification.NewNotificationServiceMock(suite.T())
	webhooks := mocks_services_webhook.NewWebhookServiceMock(suite.T())
	cfg := &config.Config{ExternalUIUrl: "https://zz.example.com"}
	suite.githubSync = mocks_services_githubsync.NewGithubSyncServiceMock(suite.T())
	suite.githubSync.EXPECT().AfterCommentCreate(mock.Anything, mock.Anything).Maybe()
	suite.githubSync.EXPECT().AfterCommentUpdate(mock.Anything, mock.Anything).Maybe()
	suite.githubSync.EXPECT().AfterCommentDelete(mock.Anything, mock.Anything).Maybe()
	issues := issue_service.NewIssueService(cfg, suite.MockRepo, notifications, webhooks, suite.githubSync)
	suite.service = NewCommentService(cfg, suite.MockRepo, issues, notifications, webhooks, suite.githubSync)
	suite.service.now = func() time.Time { return time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC) }
	suite.userID, suite.otherID, suite.workspaceID = uuid.New(), uuid.New(), uuid.New()
	team := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "ENG"}
	suite.issue = &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: team.ID, Number: 4, Title: "Fix login", Edges: ent.IssueEdges{Team: team}}
	suite.events, suite.emitted = nil, nil
	suite.ExpectTx()
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, nil, suite.issue.ID).Return(suite.issue, nil).Maybe()
	suite.MockIssueRepo.EXPECT().ReactionsForComments(mock.Anything, mock.Anything).Return(map[uuid.UUID][]*ent.Reaction{}, nil).Maybe()
	notifications.EXPECT().Notify(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, event models.NotificationEvent) error {
		suite.events = append(suite.events, event)
		return nil
	}).Maybe()
	webhooks.EXPECT().Emit(mock.Anything, suite.workspaceID, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ uuid.UUID, event string, _ any) error {
		suite.emitted = append(suite.emitted, event)
		return nil
	}).Maybe()
}

func (suite *CommentServiceSuite) roles(held schema.WorkspaceRole) {
	for _, role := range []schema.WorkspaceRole{schema.WorkspaceRoleViewer, schema.WorkspaceRoleMember, schema.WorkspaceRoleAdmin} {
		var result error
		if !permissions_repo.RoleImplies(held, role) {
			result = errdefs.ErrUnauthorized
		}
		suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
	}
}

func (suite *CommentServiceSuite) stored(comment *ent.Comment) {
	suite.MockCommentRepo.EXPECT().GetByID(mock.Anything, comment.ID).Return(comment, nil).Maybe()
	suite.MockCommentRepo.EXPECT().GetWithEdges(mock.Anything, comment.ID).Return(comment, nil).Maybe()
}

func (suite *CommentServiceSuite) TestCreateNeedsMemberAndABody() {
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateCommentInput{IssueID: suite.issue.ID, Body: "hi"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateCommentInput{IssueID: suite.issue.ID, Body: "  "})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.SetupTest()
	suite.roles("")
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateCommentInput{IssueID: suite.issue.ID, Body: "hi"})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *CommentServiceSuite) TestCreateNotifiesMentionsThenSubscribers() {
	suite.roles(schema.WorkspaceRoleMember)
	mentioned, subscriber := uuid.New(), uuid.New()
	suite.MockWorkspaceRepo.EXPECT().GetMember(mock.Anything, nil, suite.workspaceID, mentioned).Return(&ent.WorkspaceMember{}, nil)
	created := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, Body: "x"}
	suite.stored(created)
	suite.MockCommentRepo.EXPECT().Create(mock.Anything, mock.Anything, &comment_repo.CreateCommentInput{IssueID: suite.issue.ID, AuthorID: &suite.userID, Body: "cc @[Cat](user:" + mentioned.String() + ")"}).Return(created, nil).Once()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID, suite.userID, mentioned).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, suite.issue.ID).Return([]uuid.UUID{suite.userID, mentioned, subscriber}, nil)

	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateCommentInput{IssueID: suite.issue.ID, Body: " cc @[Cat](user:" + mentioned.String() + ") "})
	suite.Require().NoError(err)
	suite.Require().Len(suite.events, 2)
	suite.Equal(schema.NotificationCommentMentioned, suite.events[0].Type)
	suite.Equal([]uuid.UUID{mentioned}, suite.events[0].RecipientIDs)
	suite.Equal(schema.NotificationIssueCommented, suite.events[1].Type)
	suite.Equal([]uuid.UUID{subscriber}, suite.events[1].RecipientIDs, "mentioned people and the author are not told twice")
	suite.Equal("ENG-4 Fix login", suite.events[1].Title)
	suite.Equal([]string{"comment.created"}, suite.emitted)
}

func (suite *CommentServiceSuite) TestRepliesJoinTheRootThread() {
	suite.roles(schema.WorkspaceRoleMember)
	root := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID}
	reply := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, ParentID: &root.ID}
	elsewhere := &ent.Comment{ID: uuid.New(), IssueID: uuid.New()}
	suite.stored(root)
	suite.stored(reply)
	suite.stored(elsewhere)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateCommentInput{IssueID: suite.issue.ID, Body: "x", ParentID: &elsewhere.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	created := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, ParentID: &root.ID}
	suite.stored(created)
	suite.MockCommentRepo.EXPECT().Create(mock.Anything, mock.Anything, mock.MatchedBy(func(input *comment_repo.CreateCommentInput) bool {
		return *input.ParentID == root.ID
	})).Return(created, nil).Once()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID, suite.userID).Return(nil)
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, suite.issue.ID).Return([]uuid.UUID{}, nil)
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateCommentInput{IssueID: suite.issue.ID, Body: "x", ParentID: &reply.ID})
	suite.NoError(err)
}

func (suite *CommentServiceSuite) TestOnlyTheAuthorOrAnAdminEdits() {
	theirs := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, AuthorID: &suite.otherID}
	mine := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, AuthorID: &suite.userID, Body: "old"}
	suite.stored(theirs)
	suite.stored(mine)
	suite.roles(schema.WorkspaceRoleMember)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateCommentInput{CommentID: theirs.ID, Body: "x"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.service.Delete(suite.Ctx, suite.userID, &CommentRefInput{CommentID: theirs.ID}), errdefs.ErrUnauthorized)

	suite.MockCommentRepo.EXPECT().UpdateBody(mock.Anything, mock.Anything, mine.ID, "new", mock.Anything).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil)
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateCommentInput{CommentID: mine.ID, Body: "new"})
	suite.NoError(err)

	suite.SetupTest()
	theirs.IssueID = suite.issue.ID
	suite.stored(theirs)
	suite.roles(schema.WorkspaceRoleAdmin)
	suite.MockCommentRepo.EXPECT().Delete(mock.Anything, theirs.ID).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &CommentRefInput{CommentID: theirs.ID}))
}

func (suite *CommentServiceSuite) TestResolveNeedsMember() {
	comment := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID}
	suite.stored(comment)
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Resolve(suite.Ctx, suite.userID, &CommentRefInput{CommentID: comment.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)

	suite.SetupTest()
	comment.IssueID = suite.issue.ID
	suite.stored(comment)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockCommentRepo.EXPECT().SetResolved(mock.Anything, comment.ID, &suite.userID, mock.Anything).Return(nil).Once()
	_, err = suite.service.Resolve(suite.Ctx, suite.userID, &CommentRefInput{CommentID: comment.ID})
	suite.NoError(err)
	_, err = suite.service.Unresolve(suite.Ctx, suite.userID, &CommentRefInput{CommentID: comment.ID})
	suite.NoError(err, "unresolving an open comment changes nothing")
}

func (suite *CommentServiceSuite) TestMissingCommentsReadAsNotFound() {
	missing := uuid.New()
	suite.MockCommentRepo.EXPECT().GetByID(mock.Anything, missing).Return(nil, &ent.NotFoundError{})
	_, err := suite.service.Resolve(suite.Ctx, suite.userID, &CommentRefInput{CommentID: missing})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	hidden := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID}
	suite.stored(hidden)
	suite.roles("")
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateCommentInput{CommentID: hidden.ID, Body: "x"})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *CommentServiceSuite) TestListThreadsWithCursor() {
	suite.roles(schema.WorkspaceRoleViewer)
	now := time.Now().UTC()
	first, second := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, CreatedAt: now}, &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, CreatedAt: now.Add(time.Second)}
	reply := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, ParentID: &first.ID}
	suite.MockCommentRepo.EXPECT().ListThreads(mock.Anything, suite.issue.ID, (*repository.TimeCursor)(nil), 2).Return([]*ent.Comment{first, second, reply}, nil).Once()
	threads, next, err := suite.service.List(suite.Ctx, suite.userID, &ListCommentsInput{IssueID: suite.issue.ID, PaginationInput: models.PaginationInput{Limit: 1}})
	suite.Require().NoError(err)
	suite.Require().Len(threads, 1)
	suite.Len(threads[0].Replies, 1)
	suite.Require().NotNil(next)

	after := &repository.TimeCursor{CreatedAt: first.CreatedAt, ID: first.ID}
	suite.MockCommentRepo.EXPECT().ListThreads(mock.Anything, suite.issue.ID, after, 2).Return([]*ent.Comment{second}, nil).Once()
	threads, next, err = suite.service.List(suite.Ctx, suite.userID, &ListCommentsInput{IssueID: suite.issue.ID, PaginationInput: models.PaginationInput{Limit: 1, Cursor: *next}})
	suite.Require().NoError(err)
	suite.Len(threads, 1)
	suite.Nil(next)

	_, _, err = suite.service.List(suite.Ctx, suite.userID, &ListCommentsInput{IssueID: suite.issue.ID, PaginationInput: models.PaginationInput{Limit: 1, Cursor: "garbage"}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func TestCommentServiceSuite(t *testing.T) {
	suite.Run(t, new(CommentServiceSuite))
}
