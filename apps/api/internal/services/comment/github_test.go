package comment_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	comment_repo "github.com/unbindapp/zigzag-api/internal/repositories/comment"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

func (suite *CommentServiceSuite) TestCreateFromGithubNeedsTheSystemCaller() {
	_, err := suite.service.CreateFromGithub(suite.Ctx, &GithubCommentInput{IssueID: suite.issue.ID, Body: "hi"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *CommentServiceSuite) TestCreateFromGithubKeepsTheGithubTimeAndLinksInTheTransaction() {
	ctx := permissions_repo.WithSystemCaller(suite.Ctx)
	written := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	created := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, Body: "**@octocat on GitHub**\n\nSeen it"}
	suite.stored(created)
	suite.MockCommentRepo.EXPECT().Create(mock.Anything, mock.Anything, &comment_repo.CreateCommentInput{IssueID: suite.issue.ID, Body: created.Body, CreatedAt: &written}).Return(created, nil).Once()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil).Once()
	subscriber := uuid.New()
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, suite.issue.ID).Return([]uuid.UUID{subscriber}, nil)

	linked := uuid.Nil
	_, err := suite.service.CreateFromGithub(ctx, &GithubCommentInput{
		IssueID: suite.issue.ID, Body: created.Body, CreatedAt: written,
		Link: func(_ context.Context, _ repository.TxInterface, commentID uuid.UUID) error {
			linked = commentID
			return nil
		},
	})
	suite.Require().NoError(err)
	suite.Equal(created.ID, linked)
	suite.Require().Len(suite.events, 1)
	suite.Equal(schema.NotificationIssueCommented, suite.events[0].Type)
	suite.Nil(suite.events[0].ActorID)
	suite.Equal([]uuid.UUID{subscriber}, suite.events[0].RecipientIDs)
	suite.Equal([]string{"comment.created"}, suite.emitted)
	suite.githubSync.AssertCalled(suite.T(), "AfterCommentCreate", mock.Anything, created.ID)
}

func (suite *CommentServiceSuite) TestEditsAndDeletesHandTheCommentToGithubSync() {
	suite.roles(schema.WorkspaceRoleMember)
	mine := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, AuthorID: &suite.userID, Body: "old"}
	suite.stored(mine)
	suite.MockCommentRepo.EXPECT().UpdateBody(mock.Anything, mock.Anything, mine.ID, "new", mock.Anything).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil).Once()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateCommentInput{CommentID: mine.ID, Body: "new"})
	suite.Require().NoError(err)
	suite.githubSync.AssertCalled(suite.T(), "AfterCommentUpdate", mock.Anything, mine.ID)

	suite.MockCommentRepo.EXPECT().Delete(mock.Anything, mine.ID).Return(nil).Once()
	suite.Require().NoError(suite.service.Delete(suite.Ctx, suite.userID, &CommentRefInput{CommentID: mine.ID}))
	suite.githubSync.AssertCalled(suite.T(), "AfterCommentDelete", mock.Anything, mine.ID)
}
