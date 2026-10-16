package github_sync_service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/crypto"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
	mocks_integrations_github "github.com/zigzagapp-org/zigzag-api/mocks/integrations/github"
)

const (
	token    = "github_pat_11ABCDEFG0123456789abcdefghijklmnop"
	fullName = "unbindapp/unbind"
)

type GithubSyncServiceSuite struct {
	services.ServiceTestSuite
	service  *GithubSyncService
	client   *mocks_integrations_github.GithubClientMock
	now      time.Time
	link     *ent.GithubIssueLink
	repoSync *ent.GithubRepoSync
	issue    *ent.Issue
	states   map[string]*ent.WorkflowState
	results  []string
	jobs     int
}

func (suite *GithubSyncServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	encryptor, err := crypto.NewEncryptor(bytes.Repeat([]byte{5}, 32))
	suite.Require().NoError(err)
	suite.client = mocks_integrations_github.NewGithubClientMock(suite.T())
	suite.service = NewGithubSyncService(suite.MockRepo, encryptor, suite.client)
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.jobs, suite.results = 0, nil
	suite.service.run = func(job func()) {
		suite.jobs++
		job()
	}

	sealed, err := encryptor.Encrypt([]byte(token))
	suite.Require().NoError(err)
	integration := &ent.GithubIntegration{ID: uuid.New(), TokenEncrypted: sealed}
	suite.repoSync = &ent.GithubRepoSync{ID: uuid.New(), FullName: fullName}
	suite.repoSync.Edges.Integration = integration
	suite.issue = &ent.Issue{ID: uuid.New(), Title: "Logs stop streaming", Description: "Seen on **staging**"}
	suite.link = &ent.GithubIssueLink{ID: uuid.New(), IssueID: suite.issue.ID, Number: 12}
	suite.link.Edges.RepoSync = suite.repoSync
	suite.states = map[string]*ent.WorkflowState{}
	for name, kind := range map[string]schema.StateType{"Todo": schema.StateTypeUnstarted, "In Progress": schema.StateTypeStarted, "Done": schema.StateTypeCompleted, "Canceled": schema.StateTypeCanceled} {
		state := &ent.WorkflowState{ID: uuid.New(), Name: name, Type: kind}
		suite.states[name] = state
		suite.MockTeamRepo.EXPECT().GetState(mock.Anything, state.ID).Return(state, nil).Maybe()
	}

	suite.MockGithubRepo.EXPECT().IssueLinkByIssue(mock.Anything, suite.issue.ID).RunAndReturn(func(_ context.Context, _ uuid.UUID) (*ent.GithubIssueLink, error) {
		if suite.link == nil {
			return nil, &ent.NotFoundError{}
		}
		return suite.link, nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, nil, suite.issue.ID).Return(suite.issue, nil).Maybe()
	suite.MockGithubRepo.EXPECT().RecordSyncResult(mock.Anything, suite.repoSync.ID, mock.Anything, suite.now).RunAndReturn(func(_ context.Context, _ uuid.UUID, message string, _ time.Time) error {
		suite.results = append(suite.results, message)
		return nil
	}).Maybe()
}

func (suite *GithubSyncServiceSuite) expectOutbound(githubUpdatedAt *time.Time) {
	suite.MockGithubRepo.EXPECT().RecordOutbound(mock.Anything, suite.link.ID, githubUpdatedAt, suite.now).Return(nil).Once()
}

func (suite *GithubSyncServiceSuite) stateChange(from, to string) schema.FieldChange {
	return schema.FieldChange{Field: "state", From: suite.states[from].ID.String(), To: suite.states[to].ID.String()}
}

func (suite *GithubSyncServiceSuite) TestChangesFromGithubAreNotSentBack() {
	ctx := github.WithInbound(suite.Ctx, "Closed on GitHub", "https://github.com/unbindapp/unbind/issues/12")
	suite.service.AfterIssueUpdate(ctx, suite.issue.ID, []schema.FieldChange{suite.stateChange("Todo", "Done")})
	suite.service.AfterCommentCreate(ctx, uuid.New())
	suite.service.AfterCommentUpdate(ctx, uuid.New())
	suite.service.AfterCommentDelete(ctx, uuid.New())
	suite.Zero(suite.jobs)
}

func (suite *GithubSyncServiceSuite) TestIssuesWithoutALinkAreLeftAlone() {
	suite.link = nil
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{{Field: "title"}})
	suite.Equal(1, suite.jobs)
	suite.Empty(suite.results)
}

func (suite *GithubSyncServiceSuite) TestStateChangesCloseOrReopen() {
	updated := time.Date(2026, 10, 16, 11, 59, 0, 0, time.UTC)
	for _, tc := range []struct {
		from, to      string
		state, reason string
	}{
		{"In Progress", "Done", "closed", "completed"},
		{"Todo", "Canceled", "closed", "not_planned"},
		{"Done", "Canceled", "closed", "not_planned"},
		{"Done", "Todo", "open", ""},
	} {
		suite.client.EXPECT().UpdateIssue(mock.Anything, token, fullName, 12, &github.IssuePatch{State: tc.state, StateReason: tc.reason}).Return(&github.Issue{UpdatedAt: updated}, nil).Once()
		suite.expectOutbound(&updated)
		suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{suite.stateChange(tc.from, tc.to)})
	}

	suite.expectOutbound(nil)
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{suite.stateChange("Todo", "In Progress")})
	suite.Equal([]string{"", "", "", "", ""}, suite.results, "moving between open states sends nothing")
}

func (suite *GithubSyncServiceSuite) TestTitleAndDescriptionArePatchedTogether() {
	updated := time.Date(2026, 10, 16, 11, 59, 0, 0, time.UTC)
	suite.client.EXPECT().UpdateIssue(mock.Anything, token, fullName, 12, &github.IssuePatch{Title: &suite.issue.Title, Body: &suite.issue.Description}).Return(&github.Issue{UpdatedAt: updated}, nil).Once()
	suite.expectOutbound(&updated)
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{{Field: "title"}, {Field: "description"}, {Field: "priority", To: 1}})
}

func (suite *GithubSyncServiceSuite) TestLabelsAreCreatedOnGithubWhenMissing() {
	bug := &ent.Label{ID: uuid.New(), Name: "bug", Color: "#D73A4A"}
	ui := &ent.Label{ID: uuid.New(), Name: "ui", Color: "#0E8A16"}
	suite.issue.Edges.Labels = []*ent.Label{bug, ui}
	suite.client.EXPECT().EnsureLabel(mock.Anything, token, fullName, "bug", "d73a4a").Return(nil).Once()
	suite.client.EXPECT().EnsureLabel(mock.Anything, token, fullName, "ui", "0e8a16").Return(nil).Once()
	suite.client.EXPECT().AddLabels(mock.Anything, token, fullName, 12, []string{"bug", "ui"}).Return(nil).Once()
	removed := &ent.Label{ID: uuid.New(), Name: "needs triage"}
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{removed.ID}).Return([]*ent.Label{removed}, nil).Once()
	suite.client.EXPECT().RemoveLabel(mock.Anything, token, fullName, 12, "needs triage").Return(nil).Once()
	suite.expectOutbound(nil)
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{
		{Field: "labels_added", To: []string{bug.ID.String(), ui.ID.String()}},
		{Field: "labels_removed", From: []any{removed.ID.String()}},
	})
	suite.Equal([]string{""}, suite.results)
}

func (suite *GithubSyncServiceSuite) TestAssigneesNeedAGithubUsername() {
	eve := &ent.User{ID: uuid.New(), GithubLogin: new("eve-gh")}
	sam := &ent.User{ID: uuid.New()}
	suite.MockUserRepo.EXPECT().GetByID(mock.Anything, eve.ID).Return(eve, nil)
	suite.MockUserRepo.EXPECT().GetByID(mock.Anything, sam.ID).Return(sam, nil)

	suite.client.EXPECT().AddAssignees(mock.Anything, token, fullName, 12, []string{"eve-gh"}).Return(nil).Once()
	suite.expectOutbound(nil)
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{{Field: "assignee", From: sam.ID.String(), To: eve.ID.String()}})

	suite.client.EXPECT().RemoveAssignees(mock.Anything, token, fullName, 12, []string{"eve-gh"}).Return(nil).Once()
	suite.expectOutbound(nil)
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{{Field: "assignee", From: eve.ID.String(), To: nil}})
}

func (suite *GithubSyncServiceSuite) TestFailuresAndMissingTokensAreRecorded() {
	suite.client.EXPECT().UpdateIssue(mock.Anything, token, fullName, 12, mock.Anything).Return(nil, &github.APIError{Status: 403, Message: "Resource not accessible by personal access token"}).Once()
	suite.expectOutbound(nil)
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{{Field: "title"}})

	suite.repoSync.Edges.Integration.TokenEncrypted = nil
	suite.service.AfterIssueUpdate(suite.Ctx, suite.issue.ID, []schema.FieldChange{{Field: "title"}})
	suite.Equal([]string{"GitHub answered 403: Resource not accessible by personal access token", errNoToken.Error()}, suite.results)
}

func (suite *GithubSyncServiceSuite) comment(authorName string) *ent.Comment {
	comment := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, Body: "Fixed by raising the idle timeout."}
	if authorName != "" {
		comment.Edges.Author = &ent.User{Name: authorName}
	}
	suite.MockCommentRepo.EXPECT().GetWithEdges(mock.Anything, comment.ID).Return(comment, nil).Maybe()
	return comment
}

func (suite *GithubSyncServiceSuite) TestCommentsArePostedWithTheirAuthorAndLinked() {
	comment := suite.comment("Eve")
	suite.client.EXPECT().CreateComment(mock.Anything, token, fullName, 12, "**Eve on Zigzag**\n\nFixed by raising the idle timeout.").Return(&github.Comment{ID: 9001}, nil).Once()
	suite.MockGithubRepo.EXPECT().CreateCommentLink(mock.Anything, nil, comment.ID, suite.link.ID, int64(9001)).Return(nil).Once()
	suite.service.AfterCommentCreate(suite.Ctx, comment.ID)
	suite.Equal([]string{""}, suite.results)
}

func (suite *GithubSyncServiceSuite) TestCommentEditsAndDeletesFollowTheLink() {
	comment := suite.comment("Eve")
	link := &ent.GithubCommentLink{ID: uuid.New(), CommentID: comment.ID, IssueLinkID: suite.link.ID, GithubCommentID: 9001}
	link.Edges.IssueLink = suite.link
	suite.MockGithubRepo.EXPECT().CommentLinkByComment(mock.Anything, comment.ID).Return(link, nil).Twice()
	suite.client.EXPECT().UpdateComment(mock.Anything, token, fullName, int64(9001), "**Eve on Zigzag**\n\nFixed by raising the idle timeout.").Return(&github.Comment{ID: 9001}, nil).Once()
	suite.service.AfterCommentUpdate(suite.Ctx, comment.ID)

	suite.client.EXPECT().DeleteComment(mock.Anything, token, fullName, int64(9001)).Return(nil).Once()
	suite.MockGithubRepo.EXPECT().DeleteCommentLink(mock.Anything, link.ID).Return(nil).Once()
	suite.service.AfterCommentDelete(suite.Ctx, comment.ID)
	suite.Equal([]string{"", ""}, suite.results)
}

func (suite *GithubSyncServiceSuite) TestCommentsOnUnlinkedIssuesStayInZigzag() {
	suite.link = nil
	comment := suite.comment("Eve")
	suite.service.AfterCommentCreate(suite.Ctx, comment.ID)
	suite.MockGithubRepo.EXPECT().CommentLinkByComment(mock.Anything, comment.ID).Return(nil, &ent.NotFoundError{}).Once()
	suite.service.AfterCommentDelete(suite.Ctx, comment.ID)
	suite.Empty(suite.results)
}

func (suite *GithubSyncServiceSuite) TestDeletedCommentFailureIsRecorded() {
	comment := suite.comment("")
	link := &ent.GithubCommentLink{ID: uuid.New(), CommentID: comment.ID, GithubCommentID: 9001}
	link.Edges.IssueLink = suite.link
	suite.MockGithubRepo.EXPECT().CommentLinkByComment(mock.Anything, comment.ID).Return(link, nil).Once()
	suite.client.EXPECT().DeleteComment(mock.Anything, token, fullName, int64(9001)).Return(errors.New("connection reset")).Once()
	suite.service.AfterCommentDelete(suite.Ctx, comment.ID)
	suite.Equal([]string{"connection reset"}, suite.results)
}

func TestGithubSyncServiceSuite(t *testing.T) {
	suite.Run(t, new(GithubSyncServiceSuite))
}
