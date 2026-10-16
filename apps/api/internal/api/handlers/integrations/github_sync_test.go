package integrations_handler_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/comment"
	"github.com/zigzagapp-org/zigzag-api/ent/githubissuelink"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	"github.com/zigzagapp-org/zigzag-api/ent/issuehistory"
	"github.com/zigzagapp-org/zigzag-api/ent/label"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	comment_service "github.com/zigzagapp-org/zigzag-api/internal/services/comment"
	github_service "github.com/zigzagapp-org/zigzag-api/internal/services/github"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
)

const syncedIssueURL = "https://github.com/unbindapp/unbind/issues/1347"

func (suite *GithubWebhookSuite) syncRepository(token string) *models.GithubRepoSyncResponse {
	if token != "" {
		_, err := suite.srv.GithubService.UpdateIntegration(suite.Ctx, suite.user.ID, &github_service.UpdateGithubIntegrationInput{WorkspaceID: suite.workspace.ID, Token: token})
		suite.Require().NoError(err)
	}
	added, err := suite.srv.GithubService.AddRepoSync(suite.Ctx, suite.user.ID, &github_service.AddGithubRepoSyncInput{
		WorkspaceID: suite.workspace.ID, FullName: "UnbindApp/Unbind", TeamID: suite.team.ID,
	})
	suite.Require().NoError(err)
	return added
}

func (suite *GithubWebhookSuite) linkEve() {
	suite.DB.User.UpdateOneID(suite.user.ID).SetGithubLogin("eve-gh").ExecX(suite.Ctx)
}

func (suite *GithubWebhookSuite) syncedIssue() *ent.Issue {
	link := suite.DB.GithubIssueLink.Query().Where(githubissuelink.Number(1347)).OnlyX(suite.Ctx)
	return suite.DB.Issue.Query().Where(issue.ID(link.IssueID)).WithLabels().WithState().OnlyX(suite.Ctx)
}

func (suite *GithubWebhookSuite) githubLabels(issueID uuid.UUID) []string {
	labels := []string{}
	for _, row := range suite.DB.IssueHistory.Query().Where(issuehistory.IssueID(issueID)).Order(ent.Asc(issuehistory.FieldCreatedAt)).AllX(suite.Ctx) {
		suite.Equal("system", string(row.Source))
		for _, change := range row.Changes {
			if change.Field == "github" {
				labels = append(labels, change.ToLabel)
			}
		}
	}
	return labels
}

func (suite *GithubWebhookSuite) fixture(name string) []byte {
	body, err := os.ReadFile(filepath.Join("testdata", name))
	suite.Require().NoError(err)
	return body
}

func (suite *GithubWebhookSuite) handled(event, fixture string) bool {
	result := suite.deliver(event, fixture)
	suite.Require().Equal(http.StatusOK, result.status, result.body)
	return result.body["data"].(map[string]any)["handled"].(bool)
}

func (suite *GithubWebhookSuite) TestIssueEventsOfUnsyncedRepositoriesAreIgnored() {
	suite.False(suite.handled(github.EventIssues, "issues_opened.json"))
	suite.False(suite.handled(github.EventIssueComment, "issue_comment_created.json"))
	suite.Zero(suite.DB.GithubIssueLink.Query().CountX(suite.Ctx))
}

func (suite *GithubWebhookSuite) TestOpenedIssueBecomesALinkedIssueOnce() {
	added := suite.syncRepository("")
	suite.Equal("unbindapp/unbind", added.FullName)
	suite.Require().NotNil(added.Project)
	suite.Equal("unbindapp/unbind", added.Project.Name)
	suite.linkEve()

	suite.True(suite.handled(github.EventIssues, "issues_opened.json"))
	suite.True(suite.handled(github.EventIssues, "issues_opened.json"))

	suite.Equal(1, suite.DB.GithubIssueLink.Query().CountX(suite.Ctx))
	synced := suite.syncedIssue()
	suite.Equal("Deploy logs stop streaming after 10 minutes", synced.Title)
	suite.Equal("The log stream closes with **1006** after about ten minutes.", synced.Description)
	suite.Equal(suite.team.ID, synced.TeamID)
	suite.Require().NotNil(synced.ProjectID)
	suite.Equal(added.Project.ID, *synced.ProjectID)
	suite.Nil(synced.CreatorID)
	suite.Require().NotNil(synced.AssigneeID)
	suite.Equal(suite.user.ID, *synced.AssigneeID)
	suite.Equal(schema.StateTypeUnstarted, synced.Edges.State.Type)
	suite.Require().Len(synced.Edges.Labels, 1)
	suite.Equal("bug", synced.Edges.Labels[0].Name)
	suite.Equal("#d73a4a", synced.Edges.Labels[0].Color)
	suite.Nil(synced.Edges.Labels[0].TeamID)
	suite.Equal([]string{"Synced from GitHub #1347 by @octocat"}, suite.githubLabels(synced.ID))

	detail, err := suite.srv.IssueService.Get(suite.Ctx, suite.user.ID, &issue_service.GetIssueInput{IssueID: synced.ID})
	suite.Require().NoError(err)
	suite.Require().NotNil(detail.GithubIssue)
	suite.Equal(models.GithubIssueLinkResponse{Repository: "unbindapp/unbind", Number: 1347, HTMLURL: syncedIssueURL}, *detail.GithubIssue)
}

func (suite *GithubWebhookSuite) TestClosedReopenedLabeledAndStaleEdits() {
	suite.syncRepository("")
	suite.linkEve()
	suite.True(suite.handled(github.EventIssues, "issues_opened.json"))

	suite.True(suite.handled(github.EventIssues, "issues_closed_not_planned.json"))
	suite.Equal("Canceled", suite.syncedIssue().Edges.State.Name)

	suite.True(suite.handled(github.EventIssues, "issues_edited_stale.json"))
	suite.Equal("Deploy logs stop streaming after 10 minutes", suite.syncedIssue().Title, "an edit older than the last applied state is dropped")

	suite.True(suite.handled(github.EventIssues, "issues_labeled.json"))
	names := []string{}
	for _, applied := range suite.syncedIssue().Edges.Labels {
		names = append(names, applied.Name)
	}
	suite.ElementsMatch([]string{"bug", "websocket"}, names)
	suite.Equal("#0e8a16", suite.DB.Label.Query().Where(label.Name("websocket")).OnlyX(suite.Ctx).Color)

	suite.True(suite.handled(github.EventIssues, "issues_reopened.json"))
	reopened := suite.syncedIssue()
	suite.Equal("Todo", reopened.Edges.State.Name, "reopening returns to the team default state")

	suite.Equal([]string{"Synced from GitHub #1347 by @octocat", "Closed on GitHub", "Labeled on GitHub by @octocat", "Reopened on GitHub"}, suite.githubLabels(reopened.ID))
	closedBy := suite.DB.IssueHistory.Query().Where(issuehistory.IssueID(reopened.ID)).Order(ent.Asc(issuehistory.FieldCreatedAt)).AllX(suite.Ctx)[1]
	suite.Require().NotNil(closedBy.ActorID)
	suite.Equal(suite.user.ID, *closedBy.ActorID, "a sender with a linked GitHub username is the actor")
}

func (suite *GithubWebhookSuite) TestCommentsSyncInWithTheirAuthor() {
	suite.syncRepository("")
	suite.True(suite.handled(github.EventIssueComment, "issue_comment_created.json"))
	suite.True(suite.handled(github.EventIssueComment, "issue_comment_created.json"))

	synced := suite.syncedIssue()
	comments := suite.DB.Comment.Query().Where(comment.IssueID(synced.ID)).AllX(suite.Ctx)
	suite.Require().Len(comments, 1, "a comment on an unknown issue imports the issue first, once")
	suite.Nil(comments[0].AuthorID)
	suite.Equal("**@octocat on GitHub**\n\nSame here on the staging cluster.", comments[0].Body)
	suite.True(comments[0].CreatedAt.Equal(time.Date(2026, 10, 16, 10, 10, 0, 0, time.UTC)))

	suite.True(suite.handled(github.EventIssueComment, "issue_comment_edited.json"))
	suite.Equal("**@octocat on GitHub**\n\nSame here on the staging and production clusters.", suite.DB.Comment.GetX(suite.Ctx, comments[0].ID).Body)

	suite.True(suite.handled(github.EventIssueComment, "issue_comment_deleted.json"))
	suite.Zero(suite.DB.Comment.Query().Where(comment.IssueID(synced.ID)).CountX(suite.Ctx))
	suite.Zero(suite.DB.GithubCommentLink.Query().CountX(suite.Ctx))
}

func (suite *GithubWebhookSuite) TestDeletedIssueIsTrashedAndUnlinked() {
	suite.syncRepository("")
	suite.True(suite.handled(github.EventIssues, "issues_opened.json"))
	synced := suite.syncedIssue()
	suite.True(suite.handled(github.EventIssues, "issues_deleted.json"))
	suite.NotNil(suite.DB.Issue.GetX(suite.Ctx, synced.ID).TrashedAt)
	suite.Zero(suite.DB.GithubIssueLink.Query().CountX(suite.Ctx))
}

func (suite *GithubWebhookSuite) waitForCalls(count int) []stubCall {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if calls := suite.api.recorded(); len(calls) >= count {
			return calls
		}
		time.Sleep(20 * time.Millisecond)
	}
	return suite.api.recorded()
}

func (suite *GithubWebhookSuite) TestZigzagChangesGoBackToGithubWithoutEchoes() {
	suite.syncRepository("github_pat_11AAAAAAA0123456789abcdefghijklmnop")
	suite.Equal("/repos/unbindapp/unbind", suite.waitForCalls(1)[0].Path)
	suite.True(suite.handled(github.EventIssues, "issues_opened.json"))
	suite.True(suite.handled(github.EventIssues, "issues_closed_not_planned.json"))
	time.Sleep(100 * time.Millisecond)
	suite.Len(suite.api.recorded(), 1, "changes that came from GitHub are not sent back")

	synced := suite.syncedIssue()
	done := suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(suite.team.ID), workflowstate.Name("Done")).OnlyX(suite.Ctx)
	_, err := suite.srv.IssueService.Update(suite.Ctx, suite.user.ID, &issue_service.UpdateIssueInput{IssueID: synced.ID, StateID: &done.ID})
	suite.Require().NoError(err)
	calls := suite.waitForCalls(2)
	suite.Require().Len(calls, 2)
	suite.Equal(http.MethodPatch, calls[1].Method)
	suite.Equal("/repos/unbindapp/unbind/issues/1347", calls[1].Path)
	suite.Equal("Bearer github_pat_11AAAAAAA0123456789abcdefghijklmnop", calls[1].Auth)
	suite.Equal(map[string]any{"state": "closed", "state_reason": "completed"}, calls[1].Body)

	_, err = suite.srv.CommentService.Create(suite.Ctx, suite.user.ID, &comment_service.CreateCommentInput{IssueID: synced.ID, Body: "Fixed by raising the idle timeout."})
	suite.Require().NoError(err)
	calls = suite.waitForCalls(3)
	suite.Require().Len(calls, 3)
	suite.Equal(http.MethodPost, calls[2].Method)
	suite.Equal("/repos/unbindapp/unbind/issues/1347/comments", calls[2].Path)
	suite.Equal("**Eve on Zigzag**\n\nFixed by raising the idle timeout.", calls[2].Body["body"])
	suite.Eventually(func() bool { return suite.DB.GithubCommentLink.Query().CountX(suite.Ctx) == 1 }, 5*time.Second, 20*time.Millisecond)

	echo := []byte(strings.Replace(string(suite.fixture("issue_comment_created.json")), "Same here on the staging cluster.", `**Eve on Zigzag**\n\nFixed by raising the idle timeout.`, 1))
	echo = []byte(strings.ReplaceAll(string(echo), "2400000001", "9001"))
	suite.Equal(http.StatusOK, suite.post(github.EventIssueComment, echo, github.Sign(suite.secret, echo)).status)
	suite.Equal(1, suite.DB.Comment.Query().Where(comment.IssueID(synced.ID)).CountX(suite.Ctx), "GitHub's copy of a Zigzag comment does not come back")

	repoSyncs, err := suite.srv.GithubService.ListRepoSyncs(suite.Ctx, suite.user.ID, &github_service.ListGithubRepoSyncsInput{WorkspaceID: suite.workspace.ID})
	suite.Require().NoError(err)
	suite.Require().Len(repoSyncs, 1)
	suite.Equal(1, repoSyncs[0].LinkedIssues)
	suite.Empty(repoSyncs[0].LastError)
}
