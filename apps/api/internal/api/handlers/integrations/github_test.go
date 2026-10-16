package integrations_handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/attachment"
	"github.com/zigzagapp-org/zigzag-api/ent/enttest"
	"github.com/zigzagapp-org/zigzag-api/ent/issuehistory"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
	"github.com/zigzagapp-org/zigzag-api/internal/api/middleware"
	"github.com/zigzagapp-org/zigzag-api/internal/api/router"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/crypto"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
	team_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/team"
	workspace_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/workspace"
	comment_service "github.com/zigzagapp-org/zigzag-api/internal/services/comment"
	github_service "github.com/zigzagapp-org/zigzag-api/internal/services/github"
	github_sync_service "github.com/zigzagapp-org/zigzag-api/internal/services/githubsync"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
	label_service "github.com/zigzagapp-org/zigzag-api/internal/services/label"
	notification_service "github.com/zigzagapp-org/zigzag-api/internal/services/notification"
	project_service "github.com/zigzagapp-org/zigzag-api/internal/services/project"
	webhook_service "github.com/zigzagapp-org/zigzag-api/internal/services/webhook"
)

type GithubWebhookSuite struct {
	repositorytest.RepositoryBaseSuite
	url       string
	secret    string
	workspace *ent.Workspace
	team      *ent.Team
	issues    map[int]*ent.Issue
	user      *ent.User
	api       *githubStub
	srv       *server.Server
}

func (suite *GithubWebhookSuite) SetupTest() {
	suite.Ctx = context.Background()
	dsn := "file:" + filepath.Join(suite.T().TempDir(), "github.db") + "?_pragma=foreign_keys(1)&_pragma=journal_mode(wal)&_pragma=busy_timeout(10000)"
	suite.DB = enttest.Open(suite.T(), "sqlite3", dsn)
	suite.Require().NoError(suite.DB.Schema.Create(suite.Ctx))
	cfg := &config.Config{ExternalUIUrl: "http://zigzag.test"}
	repo := repositories.NewRepositories(suite.DB)
	encryptor, err := crypto.NewEncryptor(bytes.Repeat([]byte{3}, 32))
	suite.Require().NoError(err)
	webhooks := webhook_service.NewWebhookService(cfg, repo)
	notifications := notification_service.NewNotificationService(cfg, repo, nil)
	client := github.NewClient(suite.githubAPI())
	githubSync := github_sync_service.NewGithubSyncService(repo, encryptor, client)
	issues := issue_service.NewIssueService(cfg, repo, notifications, webhooks, githubSync)
	comments := comment_service.NewCommentService(cfg, repo, issues, notifications, webhooks, githubSync)
	githubService := github_service.NewGithubService(cfg, repo, encryptor, github_service.Dependencies{
		Issues: issues, Comments: comments, Labels: label_service.NewLabelService(cfg, repo), Projects: project_service.NewProjectService(cfg, repo, notifications, webhooks), Client: client,
	})
	srv := &server.Server{Cfg: cfg, Repository: repo, IssueService: issues, CommentService: comments, WebhookService: webhooks, GithubService: githubService}
	suite.srv = srv

	root := chi.NewRouter()
	api := humachi.New(root, router.NewHumaConfig("Zigzag API", "test", cfg.CookieSecure))
	router.RegisterRoutes(api, srv, middleware.NewMiddleware(cfg, repo, api, nil, nil))
	httpServer := httptest.NewServer(root)
	suite.T().Cleanup(httpServer.Close)
	suite.url = httpServer.URL

	user := suite.DB.User.Create().SetEmail("eve@example.com").SetName("Eve").SetDisplayName("eve").SaveX(suite.Ctx)
	suite.user = user
	suite.workspace, err = workspace_repo.NewWorkspaceRepository(suite.DB).Create(suite.Ctx, nil, &workspace_repo.CreateWorkspaceInput{Name: "Unbind", Slug: "unbind", CreatorID: user.ID})
	suite.Require().NoError(err)
	suite.team, err = team_repo.NewTeamRepository(suite.DB).Create(suite.Ctx, nil, &team_repo.CreateTeamInput{WorkspaceID: suite.workspace.ID, Name: "Unbind", Key: "UNB", CreatorID: user.ID})
	suite.Require().NoError(err)
	backlog := suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(suite.team.ID), workflowstate.Name("Backlog")).OnlyX(suite.Ctx)
	suite.issues = map[int]*ent.Issue{}
	for _, number := range []int{455, 456} {
		created, err := issue_repo.NewIssueRepository(suite.DB).Create(suite.Ctx, nil, &issue_repo.CreateIssueInput{
			WorkspaceID: suite.workspace.ID, TeamID: suite.team.ID, Number: number, Title: "Issue", StateID: backlog.ID, CreatorID: &user.ID,
		})
		suite.Require().NoError(err)
		suite.issues[number] = created
	}

	connected, err := srv.GithubService.Create(suite.Ctx, user.ID, &github_service.GithubIntegrationRefInput{WorkspaceID: suite.workspace.ID})
	suite.Require().NoError(err)
	suite.secret = connected.Secret
}

type deliveryResult struct {
	status int
	body   map[string]any
}

func (suite *GithubWebhookSuite) post(event string, body []byte, signature string) deliveryResult {
	req, err := http.NewRequest(http.MethodPost, suite.url+"/integrations/github/webhook?workspace_id="+suite.workspace.ID.String(), bytes.NewReader(body))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GitHub-Hookshot/abc123")
	req.Header.Set("X-GitHub-Event", event)
	req.Header.Set("X-GitHub-Delivery", "72d3162e-cc78-11e3-81ab-4c9367dc0958")
	if signature != "" {
		req.Header.Set(github.SignatureHeader, signature)
	}
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	result := deliveryResult{status: resp.StatusCode}
	_ = json.NewDecoder(resp.Body).Decode(&result.body)
	return result
}

func (suite *GithubWebhookSuite) deliver(event, fixture string) deliveryResult {
	body, err := os.ReadFile(filepath.Join("testdata", fixture))
	suite.Require().NoError(err)
	return suite.post(event, body, github.Sign(suite.secret, body))
}

func (suite *GithubWebhookSuite) stateOf(number int) string {
	issue := suite.DB.Issue.GetX(suite.Ctx, suite.issues[number].ID)
	return suite.DB.WorkflowState.GetX(suite.Ctx, issue.StateID).Name
}

func (suite *GithubWebhookSuite) attachmentsOf(number int) []*ent.Attachment {
	return suite.DB.Attachment.Query().Where(attachment.IssueID(suite.issues[number].ID)).AllX(suite.Ctx)
}

func (suite *GithubWebhookSuite) TestPing() {
	result := suite.deliver(github.EventPing, "ping.json")
	suite.Equal(http.StatusOK, result.status)
	suite.Equal(map[string]any{"event": "ping", "handled": true}, result.body["data"])
	repos := suite.DB.GithubRepository.Query().AllX(suite.Ctx)
	suite.Require().Len(repos, 1)
	suite.Equal("unbindapp/unbind", repos[0].FullName)
	suite.Equal("ping", repos[0].LastEvent)
}

func (suite *GithubWebhookSuite) TestInvalidSignaturesAnswer401() {
	body, err := os.ReadFile(filepath.Join("testdata", "pull_request_opened.json"))
	suite.Require().NoError(err)
	for name, signature := range map[string]string{
		"wrong secret": github.Sign("not-the-secret", body),
		"missing":      "",
		"malformed":    "sha256=nothex",
	} {
		suite.Equal(http.StatusUnauthorized, suite.post(github.EventPullRequest, body, signature).status, name)
	}
	suite.Empty(suite.attachmentsOf(455))
	suite.Equal("Backlog", suite.stateOf(455))
	suite.Zero(suite.DB.GithubRepository.Query().CountX(suite.Ctx))
}

func (suite *GithubWebhookSuite) TestPullRequestOpenedThenMerged() {
	opened := suite.deliver(github.EventPullRequest, "pull_request_opened.json")
	suite.Equal(http.StatusOK, opened.status)
	suite.Equal(true, opened.body["data"].(map[string]any)["handled"])
	suite.Equal("In Progress", suite.stateOf(455))
	attachments := suite.attachmentsOf(455)
	suite.Require().Len(attachments, 1)
	suite.Equal("#42 Limit websocket message size", attachments[0].Title)
	suite.Equal("https://github.com/unbindapp/unbind/pull/42", attachments[0].URL)
	suite.Equal("github", attachments[0].SourceType)
	suite.Equal("open", attachments[0].Metadata["state"])
	suite.Equal("bbedward/unb-455-websocket-limit", attachments[0].Metadata["branch"])

	merged := suite.deliver(github.EventPullRequest, "pull_request_closed_merged.json")
	suite.Equal(http.StatusOK, merged.status)
	suite.Equal("Done", suite.stateOf(455))
	attachments = suite.attachmentsOf(455)
	suite.Require().Len(attachments, 1)
	suite.Equal("merged", attachments[0].Metadata["state"])

	history := suite.DB.IssueHistory.Query().AllX(suite.Ctx)
	labels := []string{}
	for _, row := range history {
		suite.Nil(row.ActorID)
		suite.Equal("system", string(row.Source))
		labels = append(labels, row.Changes[0].ToLabel)
	}
	suite.ElementsMatch([]string{"Pull request #42 opened", "Pull request #42 merged"}, labels)
}

func (suite *GithubWebhookSuite) TestPushLinksTheBranch() {
	result := suite.deliver(github.EventPush, "push.json")
	suite.Equal(http.StatusOK, result.status)
	attachments := suite.attachmentsOf(456)
	suite.Require().Len(attachments, 1)
	suite.Equal("bbedward/unb-456-retry-deploys", attachments[0].Title)
	suite.Equal("https://github.com/unbindapp/unbind/tree/bbedward/unb-456-retry-deploys", attachments[0].URL)
	suite.Equal("github_branch", attachments[0].SourceType)
	suite.Equal("In Progress", suite.stateOf(456))
	suite.Empty(suite.attachmentsOf(455))
}

func (suite *GithubWebhookSuite) addIssues(numbers ...int) {
	backlog := suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(suite.team.ID), workflowstate.Name("Backlog")).OnlyX(suite.Ctx)
	for _, number := range numbers {
		created, err := issue_repo.NewIssueRepository(suite.DB).Create(suite.Ctx, nil, &issue_repo.CreateIssueInput{
			WorkspaceID: suite.workspace.ID, TeamID: suite.team.ID, Number: number, Title: "Issue", StateID: backlog.ID, CreatorID: &suite.user.ID,
		})
		suite.Require().NoError(err)
		suite.issues[number] = created
	}
}

func (suite *GithubWebhookSuite) historyLabels(number int) []string {
	labels := []string{}
	for _, row := range suite.DB.IssueHistory.Query().Where(issuehistory.IssueID(suite.issues[number].ID)).AllX(suite.Ctx) {
		labels = append(labels, row.Changes[0].ToLabel)
	}
	return labels
}

func (suite *GithubWebhookSuite) TestCommitsPushedToTheDefaultBranchCloseWhatTheyFix() {
	suite.addIssues(12, 13, 14)
	result := suite.deliver(github.EventPush, "push_commits_master.json")
	suite.Equal(http.StatusOK, result.status)
	suite.Equal(map[string]any{"event": "push", "handled": true}, result.body["data"])

	suite.Equal("Done", suite.stateOf(12))
	suite.Equal("Done", suite.stateOf(13))
	suite.Equal("In Progress", suite.stateOf(14))
	attachments := suite.attachmentsOf(12)
	suite.Require().Len(attachments, 1)
	suite.Equal("3e1f0c9 Cap websocket frames at 1 MiB", attachments[0].Title)
	suite.Equal("https://github.com/unbindapp/unbind/commit/3e1f0c9a7b5d2e8f4a6c1b9d0e7f3a5c8b2d4e6f", attachments[0].URL)
	suite.Equal("github_commit", attachments[0].SourceType)
	suite.Equal(map[string]any{"repo": "unbindapp/unbind", "sha": "3e1f0c9a7b5d2e8f4a6c1b9d0e7f3a5c8b2d4e6f", "branch": "master", "author": "bbedward"}, attachments[0].Metadata)
	suite.Equal([]string{"Commit 3e1f0c9 on master closed the issue"}, suite.historyLabels(12))
	suite.Equal([]string{"Commit b84d2a6 pushed to master"}, suite.historyLabels(14))

	redelivered := suite.deliver(github.EventPush, "push_commits_master.json")
	suite.Equal(http.StatusOK, redelivered.status)
	for _, number := range []int{12, 13, 14} {
		suite.Len(suite.attachmentsOf(number), 1, number)
		suite.Len(suite.historyLabels(number), 1, number)
	}
}

func (suite *GithubWebhookSuite) TestCommitsPushedToAFeatureBranchOnlyStartIssues() {
	suite.addIssues(12, 13, 14)
	result := suite.deliver(github.EventPush, "push_commits_feature.json")
	suite.Equal(http.StatusOK, result.status)
	for _, number := range []int{12, 13, 14} {
		suite.Equal("In Progress", suite.stateOf(number), number)
		suite.Len(suite.attachmentsOf(number), 1, number)
	}
	suite.Equal([]string{"Commit 3e1f0c9 pushed to bbedward/websocket-frames"}, suite.historyLabels(12))
}

func (suite *GithubWebhookSuite) TestOtherEventsAreAcknowledged() {
	body := []byte(`{"action":"created","repository":{"full_name":"unbindapp/unbind"}}`)
	result := suite.post("star", body, github.Sign(suite.secret, body))
	suite.Equal(http.StatusOK, result.status)
	suite.Equal(map[string]any{"event": "star", "handled": false}, result.body["data"])
}

func (suite *GithubWebhookSuite) TestBodiesOverOneMebibyteAreRefused() {
	body := []byte(`{"zen":"` + strings.Repeat("x", 1<<20) + `"}`)
	result := suite.post(github.EventPing, body, github.Sign(suite.secret, body))
	suite.Equal(http.StatusRequestEntityTooLarge, result.status)
}

func TestGithubWebhookSuite(t *testing.T) {
	suite.Run(t, new(GithubWebhookSuite))
}
