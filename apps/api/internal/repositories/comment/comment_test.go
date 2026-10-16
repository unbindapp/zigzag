package comment_repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type CommentRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo  *CommentRepository
	user  *ent.User
	issue *ent.Issue
}

func (suite *CommentRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewCommentRepository(suite.DB)
	suite.user = suite.DB.User.Create().SetEmail("ada@example.com").SetDisplayName("ada").SaveX(suite.Ctx)
	workspace := suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	team := suite.DB.Team.Create().SetWorkspaceID(workspace.ID).SetName("Eng").SetKey("ENG").SaveX(suite.Ctx)
	state := suite.DB.WorkflowState.Create().SetTeamID(team.ID).SetName("Todo").SetType("unstarted").SetColor("#ffffff").SaveX(suite.Ctx)
	suite.issue = suite.DB.Issue.Create().SetWorkspaceID(workspace.ID).SetTeamID(team.ID).SetNumber(1).SetTitle("t").SetStateID(state.ID).SaveX(suite.Ctx)
}

func (suite *CommentRepositorySuite) comment(body string, parent *ent.Comment) *ent.Comment {
	input := &CreateCommentInput{IssueID: suite.issue.ID, AuthorID: &suite.user.ID, Body: body}
	if parent != nil {
		input.ParentID = &parent.ID
	}
	created, err := suite.repo.Create(suite.Ctx, nil, input)
	suite.Require().NoError(err)
	return created
}

func (suite *CommentRepositorySuite) TestThreadsPageByRootWithTheirReplies() {
	first := suite.comment("first", nil)
	suite.comment("second", nil)
	suite.comment("reply", first)

	page, err := suite.repo.ListThreads(suite.Ctx, suite.issue.ID, 0, 1)
	suite.Require().NoError(err)
	bodies := []string{}
	for _, comment := range page {
		bodies = append(bodies, comment.Body)
		suite.NotNil(comment.Edges.Author)
	}
	suite.Equal([]string{"first", "reply"}, bodies, "a page holds roots followed by their replies")

	page, err = suite.repo.ListThreads(suite.Ctx, suite.issue.ID, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Len(page, 1)
	suite.Equal("second", page[0].Body)
}

func (suite *CommentRepositorySuite) TestEditResolveAndDelete() {
	root := suite.comment("root", nil)
	reply := suite.comment("reply", root)
	now := time.Now()

	suite.NoError(suite.repo.UpdateBody(suite.Ctx, root.ID, "edited", now))
	suite.NoError(suite.repo.SetResolved(suite.Ctx, root.ID, &suite.user.ID, &now))
	loaded, err := suite.repo.GetWithEdges(suite.Ctx, root.ID)
	suite.Require().NoError(err)
	suite.Equal("edited", loaded.Body)
	suite.NotNil(loaded.EditedAt)
	suite.NotNil(loaded.Edges.ResolvedBy)
	suite.Len(loaded.Edges.Replies, 1)

	suite.NoError(suite.repo.SetResolved(suite.Ctx, root.ID, nil, nil))
	loaded, _ = suite.repo.GetByID(suite.Ctx, root.ID)
	suite.Nil(loaded.ResolvedAt)
	suite.Nil(loaded.ResolvedByID)

	suite.NoError(suite.repo.Delete(suite.Ctx, root.ID))
	_, err = suite.repo.GetByID(suite.Ctx, reply.ID)
	suite.True(ent.IsNotFound(err), "replies go with their thread")
	workspaceID, err := suite.repo.ResolveWorkspaceOfComment(suite.Ctx, suite.comment("x", nil).ID)
	suite.NoError(err)
	suite.Equal(suite.issue.WorkspaceID, workspaceID)
}

func TestCommentRepositorySuite(t *testing.T) {
	suite.Run(t, new(CommentRepositorySuite))
}
