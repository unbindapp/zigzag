package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/enttest"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/middleware"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/router"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/auth"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/cache"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/crypto"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/mail"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/storage"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/models"
	"github.com/unbindapp/zigzag-api/internal/oauthserver"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
	apikey_service "github.com/unbindapp/zigzag-api/internal/services/apikey"
	comment_service "github.com/unbindapp/zigzag-api/internal/services/comment"
	cycle_service "github.com/unbindapp/zigzag-api/internal/services/cycle"
	document_service "github.com/unbindapp/zigzag-api/internal/services/document"
	email_service "github.com/unbindapp/zigzag-api/internal/services/email"
	favorite_service "github.com/unbindapp/zigzag-api/internal/services/favorite"
	github_sync_service "github.com/unbindapp/zigzag-api/internal/services/githubsync"
	import_service "github.com/unbindapp/zigzag-api/internal/services/import"
	initiative_service "github.com/unbindapp/zigzag-api/internal/services/initiative"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	label_service "github.com/unbindapp/zigzag-api/internal/services/label"
	notification_service "github.com/unbindapp/zigzag-api/internal/services/notification"
	oauthserver_service "github.com/unbindapp/zigzag-api/internal/services/oauthserver"
	project_service "github.com/unbindapp/zigzag-api/internal/services/project"
	search_service "github.com/unbindapp/zigzag-api/internal/services/search"
	system_service "github.com/unbindapp/zigzag-api/internal/services/system"
	team_service "github.com/unbindapp/zigzag-api/internal/services/team"
	template_service "github.com/unbindapp/zigzag-api/internal/services/template"
	upload_service "github.com/unbindapp/zigzag-api/internal/services/upload"
	user_service "github.com/unbindapp/zigzag-api/internal/services/user"
	view_service "github.com/unbindapp/zigzag-api/internal/services/view"
	webhook_service "github.com/unbindapp/zigzag-api/internal/services/webhook"
	workspace_service "github.com/unbindapp/zigzag-api/internal/services/workspace"
)

type EndToEndSuite struct {
	repositorytest.RepositoryBaseSuite
	url       string
	api       huma.API
	user      *ent.User
	workspace *ent.Workspace
	team      *ent.Team
	admin     string
	viewer    string
	nextID    int
}

// Handlers read outside their transactions, so the base suite's per-connection memory database will not do.
func (suite *EndToEndSuite) SetupTest() {
	suite.Ctx = context.Background()
	dsn := "file:" + filepath.Join(suite.T().TempDir(), "e2e.db") + "?_pragma=foreign_keys(1)&_pragma=journal_mode(wal)&_pragma=busy_timeout(10000)"
	suite.DB = enttest.Open(suite.T(), "sqlite3", dsn)
	suite.Require().NoError(suite.DB.Schema.Create(suite.Ctx))
	cfg := &config.Config{ExternalUIUrl: "http://zigzag.test", TokenAudience: "zigzag-api", DataDir: suite.T().TempDir(), MaxUploadBytes: 1 << 20}
	repo := repositories.NewRepositories(suite.DB)
	_, err := repo.System().GetSystemSettings(suite.Ctx, nil)
	suite.Require().NoError(err)

	issuer := oauthserver.Issuer(cfg.ExternalUIUrl)
	key, _, err := repo.Oauth().GetOrGenerateJWTPrivateKey(suite.Ctx)
	suite.Require().NoError(err)
	tokens := auth.NewTokenManager(key, issuer, cfg.TokenAudience)
	encryptor, err := crypto.NewEncryptor(bytes.Repeat([]byte{1}, 32))
	suite.Require().NoError(err)
	store, err := storage.NewLocalStore(cfg.DataDir)
	suite.Require().NoError(err)

	emails := email_service.NewEmailService(cfg, repo, encryptor, mail.SMTPSender{})
	oauthService := oauthserver_service.NewOAuthServerService(repo, oauthserver.NewMetadataFetcher(oauthserver.DefaultIPPolicy, cache.NewMemoryCache[oauthserver.ClientMetadata]()), issuer)
	notifications := notification_service.NewNotificationService(cfg, repo, emails)
	webhooks := webhook_service.NewWebhookService(cfg, repo)
	githubSync := github_sync_service.NewGithubSyncService(repo, nil, github.NewClient(""))
	cycles := cycle_service.NewCycleService(cfg, repo, webhooks)
	srv := &server.Server{
		Cfg:                 cfg,
		Repository:          repo,
		TokenManager:        tokens,
		Storage:             store,
		UserService:         user_service.NewUserService(cfg, repo, emails),
		WorkspaceService:    workspace_service.NewWorkspaceService(cfg, repo, emails),
		TeamService:         team_service.NewTeamService(repo, cycles),
		EmailService:        emails,
		APIKeyService:       apikey_service.NewAPIKeyService(repo),
		OAuthServerService:  oauthService,
		SystemService:       system_service.NewSystemService(repo),
		UploadService:       upload_service.NewUploadService(cfg, repo, store),
		IssueService:        issue_service.NewIssueService(cfg, repo, notifications, webhooks, githubSync),
		ProjectService:      project_service.NewProjectService(cfg, repo, notifications, webhooks),
		LabelService:        label_service.NewLabelService(cfg, repo),
		CycleService:        cycles,
		InitiativeService:   initiative_service.NewInitiativeService(cfg, repo),
		NotificationService: notifications,
		ViewService:         view_service.NewViewService(cfg, repo),
		TemplateService:     template_service.NewTemplateService(cfg, repo),
		FavoriteService:     favorite_service.NewFavoriteService(cfg, repo),
		WebhookService:      webhooks,
		ImportService:       import_service.NewImportService(cfg, repo),
	}
	srv.SearchService = search_service.NewSearchService(cfg, repo, srv.IssueService)
	srv.DocumentService = document_service.NewDocumentService(cfg, repo, srv.IssueService)
	srv.CommentService = comment_service.NewCommentService(cfg, repo, srv.IssueService, notifications, webhooks, githubSync)

	root := chi.NewRouter()
	limiter := middleware.NewRateLimiter()
	root.Group(func(r chi.Router) {
		suite.api = humachi.New(r, router.NewHumaConfig("Zigzag API", "test", cfg.CookieSecure))
		mw := middleware.NewMiddleware(cfg, repo, suite.api, tokens, nil)
		suite.api.UseMiddleware(mw.Recoverer)
		router.RegisterRoutes(suite.api, srv, mw)
		mcp, err := New(Options{API: suite.api, Router: root, OAuth: oauthService, APIKeys: repo.APIKey(), Issuer: issuer, Version: "test"})
		suite.Require().NoError(err)
		mcp.Mount(r, limiter)
	})
	httpServer := httptest.NewServer(root)
	suite.T().Cleanup(httpServer.Close)
	suite.url = httpServer.URL

	suite.user = suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada Lovelace").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.workspace, err = workspace_repo.NewWorkspaceRepository(suite.DB).Create(suite.Ctx, nil, &workspace_repo.CreateWorkspaceInput{Name: "Acme", Slug: "acme", CreatorID: suite.user.ID})
	suite.Require().NoError(err)
	suite.team, err = team_repo.NewTeamRepository(suite.DB).Create(suite.Ctx, nil, &team_repo.CreateTeamInput{WorkspaceID: suite.workspace.ID, Name: "Engineering", Key: "ENG", CreatorID: suite.user.ID})
	suite.Require().NoError(err)
	suite.admin = suite.apiKey("agent", schema.WorkspaceRoleAdmin)
	suite.viewer = suite.apiKey("reader", schema.WorkspaceRoleViewer)
}

func (suite *EndToEndSuite) apiKey(name string, role schema.WorkspaceRole) string {
	created, err := apikey_service.NewAPIKeyService(repositories.NewRepositories(suite.DB)).Create(suite.Ctx, suite.user.ID, &models.APIKeyCreateInput{
		Name: name, Role: role, WorkspaceIDs: []uuid.UUID{suite.workspace.ID},
	})
	suite.Require().NoError(err)
	return created.Token
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (suite *EndToEndSuite) rpc(token, method string, params any) json.RawMessage {
	suite.T().Helper()
	suite.nextID++
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": suite.nextID, "method": method, "params": params})
	suite.Require().NoError(err)
	req, err := http.NewRequest(http.MethodPost, suite.url+"/mcp", bytes.NewReader(body))
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	suite.Require().Equal(http.StatusOK, resp.StatusCode, method)
	decoded := rpcResponse{}
	suite.Require().NoError(json.NewDecoder(resp.Body).Decode(&decoded))
	if decoded.Error != nil {
		suite.FailNow(fmt.Sprintf("%s answered error %d: %s", method, decoded.Error.Code, decoded.Error.Message))
	}
	return decoded.Result
}

type toolResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

func (suite *EndToEndSuite) call(token, tool string, args map[string]any) map[string]any {
	suite.T().Helper()
	result := toolResult{}
	suite.Require().NoError(json.Unmarshal(suite.rpc(token, "tools/call", map[string]any{"name": tool, "arguments": args}), &result))
	suite.Require().Len(result.Content, 1, tool)
	suite.Require().False(result.IsError, "%s failed: %s", tool, result.Content[0].Text)
	decoded := map[string]any{}
	suite.Require().NoError(json.Unmarshal([]byte(result.Content[0].Text), &decoded), tool)
	return decoded
}

func (suite *EndToEndSuite) tools(token string) []string {
	listed := struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}{}
	suite.Require().NoError(json.Unmarshal(suite.rpc(token, "tools/list", map[string]any{}), &listed))
	names := make([]string, len(listed.Tools))
	for i, tool := range listed.Tools {
		names[i] = tool.Name
	}
	return names
}

func object(value any) map[string]any {
	decoded, _ := value.(map[string]any)
	return decoded
}

func list(value any) []any {
	decoded, _ := value.([]any)
	return decoded
}

func (suite *EndToEndSuite) TestInitializeDescribesTheServer() {
	initialized := struct {
		ServerInfo   struct{ Name string } `json:"serverInfo"`
		Instructions string                `json:"instructions"`
	}{}
	params := map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "e2e", "version": "1"}}
	suite.Require().NoError(json.Unmarshal(suite.rpc(suite.admin, "initialize", params), &initialized))
	suite.Equal("zigzag", initialized.ServerInfo.Name)
	suite.Contains(initialized.Instructions, "whoami")
	suite.Contains(initialized.Instructions, "import-linear")
}

func (suite *EndToEndSuite) TestToolsListMatchesTheSpec() {
	operations := map[string]*huma.Operation{}
	expected := []string{}
	for _, op := range Operations(suite.api) {
		operations[op.OperationID] = op
		if choice, _ := oapi.MCPChoiceOf(op); choice.Exposed {
			expected = append(expected, op.OperationID)
		}
	}
	listed := suite.tools(suite.admin)
	suite.ElementsMatch(expected, listed)
	for _, name := range listed {
		suite.False(oapi.IsSessionOnly(operations[name]), "%s is session only", name)
		suite.NotEmpty(operations[name].Security, "%s is public", name)
	}

	for _, name := range suite.tools(suite.viewer) {
		action, _ := oapi.ActionOf(operations[name])
		suite.Equal(oapi.Read, action, "a viewer key is offered %s", name)
	}
}

func (suite *EndToEndSuite) TestViewerKeyCannotWrite() {
	suite.NotContains(suite.tools(suite.viewer), "create-issue")

	body, _ := json.Marshal(map[string]any{"team_id": suite.team.ID, "title": "Nope"})
	req, _ := http.NewRequest(http.MethodPost, suite.url+"/issues/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+suite.viewer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	_ = resp.Body.Close()
	suite.Equal(http.StatusForbidden, resp.StatusCode)
	suite.Zero(suite.DB.Issue.Query().CountX(suite.Ctx))
}

func (suite *EndToEndSuite) TestAnAgentWorksAnIssue() {
	whoami := object(suite.call(suite.admin, "whoami", nil)["data"])
	suite.Equal(suite.user.ID.String(), object(whoami["user"])["id"])
	suite.Equal("admin", object(whoami["api_key"])["role"])

	workspaces := list(suite.call(suite.admin, "list-workspaces", nil)["data"])
	suite.Require().Len(workspaces, 1)
	suite.Equal(suite.workspace.ID.String(), object(workspaces[0])["id"])
	suite.Equal("admin", object(workspaces[0])["role"])

	created := object(suite.call(suite.admin, "create-issue", map[string]any{"team_id": suite.team.ID, "title": "Wire the importer"})["data"])
	suite.Equal("ENG-1", created["identifier"])
	suite.Equal("Todo", object(created["state"])["name"])
	suite.EqualValues(0, created["priority"])

	fetched := object(suite.call(suite.admin, "get-issue", map[string]any{"identifier": "ENG-1"})["data"])
	suite.Equal(created["id"], fetched["id"])
	suite.NotNil(fetched["relations"])
	suite.NotNil(fetched["children"])

	updated := object(suite.call(suite.admin, "update-issue", map[string]any{"issue_id": created["id"], "priority": 2, "title": "Wire the Linear importer"})["data"])
	suite.EqualValues(2, updated["priority"])
	suite.Equal("Wire the Linear importer", updated["title"])

	for _, title := range []string{"Second", "Third"} {
		suite.call(suite.admin, "create-issue", map[string]any{"team_id": suite.team.ID, "title": title, "priority": 2})
	}
	suite.call(suite.admin, "create-issue", map[string]any{"team_id": suite.team.ID, "title": "Low", "priority": 4})
	filter := map[string]any{"workspace_id": suite.workspace.ID, "team_id": suite.team.ID, "priorities": []int{2}, "order_by": "created_at", "limit": 2}
	page := suite.call(suite.admin, "list-issues", filter)
	suite.Len(list(page["data"]), 2)
	suite.EqualValues(3, page["total"])
	cursor, _ := page["next_cursor"].(string)
	suite.Require().NotEmpty(cursor)
	filter["cursor"] = cursor
	last := suite.call(suite.admin, "list-issues", filter)
	suite.Require().Len(list(last["data"]), 1)
	suite.Nil(last["next_cursor"])
	suite.Equal("ENG-1", object(list(last["data"])[0])["identifier"])

	comment := object(suite.call(suite.admin, "create-comment", map[string]any{"issue_id": created["id"], "body": "Started on this"})["data"])
	suite.Equal("Started on this", comment["body"])
	suite.Equal(suite.user.ID.String(), object(comment["author"])["id"])

	label := object(suite.call(suite.admin, "create-label", map[string]any{"workspace_id": suite.workspace.ID, "name": "import", "color": "#5e6ad2"})["data"])
	suite.Equal("import", label["name"])
	suite.Nil(label["team_id"])

	project := object(suite.call(suite.admin, "create-project", map[string]any{"workspace_id": suite.workspace.ID, "name": "Linear import", "team_ids": []any{suite.team.ID}})["data"])
	suite.Equal("Linear import", project["name"])
	suite.True(strings.HasPrefix(project["slug"].(string), "linear-import-"))
	suite.Len(list(project["teams"]), 1)

	notifications := suite.call(suite.admin, "list-notifications", map[string]any{"workspace_id": suite.workspace.ID})
	suite.NotNil(notifications["data"])
	suite.Contains(notifications, "next_cursor")
}

func (suite *EndToEndSuite) TestImportLinearDryRun() {
	raw, err := os.ReadFile("testdata/linear-3-issues.json")
	suite.Require().NoError(err)
	export := map[string]any{}
	suite.Require().NoError(json.Unmarshal(raw, &export))

	report := object(suite.call(suite.admin, "import-linear", map[string]any{"workspace_id": suite.workspace.ID, "export": export})["data"])
	suite.Equal(true, report["dry_run"])
	suite.NotEmpty(report["import_id"])
	counts := object(report["counts"])
	suite.EqualValues(3, object(counts["issues"])["created"])
	suite.EqualValues(3, object(counts["attachments"])["created"])
	suite.EqualValues(1, object(counts["teams"])["created"], "UNB is a new team key in this workspace")
	suite.NotEmpty(list(report["warnings"]), "UNB-289 relates to an issue outside the fixture")
	mapping := list(report["user_mapping"])
	suite.Len(mapping, 4)
	suite.True(slices.ContainsFunc(mapping, func(entry any) bool { return object(entry)["action"] == "skipped" }), "the Linear bot is skipped")
	suite.Zero(suite.DB.Issue.Query().CountX(suite.Ctx), "a dry run keeps nothing")

	imports := list(suite.call(suite.admin, "list-imports", map[string]any{"workspace_id": suite.workspace.ID})["data"])
	suite.Require().Len(imports, 1)
	suite.Equal("completed", object(imports[0])["status"])
	suite.Equal(true, object(imports[0])["dry_run"])
}

func (suite *EndToEndSuite) TestImportLinearTakesExportsOverOneMiB() {
	raw, err := os.ReadFile("testdata/linear-3-issues.json")
	suite.Require().NoError(err)
	export := map[string]any{}
	suite.Require().NoError(json.Unmarshal(raw, &export))
	export["exported_by"] = strings.Repeat("x", 3<<20)

	report := object(suite.call(suite.admin, "import-linear", map[string]any{"workspace_id": suite.workspace.ID, "export": export})["data"])
	suite.EqualValues(3, object(object(report["counts"])["issues"])["created"], "a real export is over 1 MiB and must fit through /mcp")
}

func TestEndToEndSuite(t *testing.T) {
	suite.Run(t, new(EndToEndSuite))
}
