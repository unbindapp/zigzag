package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/zigzagapp-org/zigzag-api/config"
	entmigrate "github.com/zigzagapp-org/zigzag-api/ent/migrate"
	oauthserver_handler "github.com/zigzagapp-org/zigzag-api/internal/api/handlers/oauthserver"
	"github.com/zigzagapp-org/zigzag-api/internal/api/middleware"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/router"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/cache"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/crypto"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/database"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/mail"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/storage"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	"github.com/zigzagapp-org/zigzag-api/internal/mcpserver"
	"github.com/zigzagapp-org/zigzag-api/internal/oauthserver"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	"github.com/zigzagapp-org/zigzag-api/internal/scheduler"
	apikey_service "github.com/zigzagapp-org/zigzag-api/internal/services/apikey"
	comment_service "github.com/zigzagapp-org/zigzag-api/internal/services/comment"
	cycle_service "github.com/zigzagapp-org/zigzag-api/internal/services/cycle"
	document_service "github.com/zigzagapp-org/zigzag-api/internal/services/document"
	email_service "github.com/zigzagapp-org/zigzag-api/internal/services/email"
	favorite_service "github.com/zigzagapp-org/zigzag-api/internal/services/favorite"
	github_service "github.com/zigzagapp-org/zigzag-api/internal/services/github"
	github_sync_service "github.com/zigzagapp-org/zigzag-api/internal/services/githubsync"
	import_service "github.com/zigzagapp-org/zigzag-api/internal/services/import"
	initiative_service "github.com/zigzagapp-org/zigzag-api/internal/services/initiative"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
	label_service "github.com/zigzagapp-org/zigzag-api/internal/services/label"
	notification_service "github.com/zigzagapp-org/zigzag-api/internal/services/notification"
	oauthserver_service "github.com/zigzagapp-org/zigzag-api/internal/services/oauthserver"
	project_service "github.com/zigzagapp-org/zigzag-api/internal/services/project"
	search_service "github.com/zigzagapp-org/zigzag-api/internal/services/search"
	system_service "github.com/zigzagapp-org/zigzag-api/internal/services/system"
	team_service "github.com/zigzagapp-org/zigzag-api/internal/services/team"
	template_service "github.com/zigzagapp-org/zigzag-api/internal/services/template"
	upload_service "github.com/zigzagapp-org/zigzag-api/internal/services/upload"
	user_service "github.com/zigzagapp-org/zigzag-api/internal/services/user"
	view_service "github.com/zigzagapp-org/zigzag-api/internal/services/view"
	webhook_service "github.com/zigzagapp-org/zigzag-api/internal/services/webhook"
	workspace_service "github.com/zigzagapp-org/zigzag-api/internal/services/workspace"
	"github.com/zigzagapp-org/zigzag-api/internal/web"
	_ "go.uber.org/automaxprocs"

	// the distroless image ships no zoneinfo, and time zones are validated against it
	_ "time/tzdata"
)

var Version = "development"

const docsPage = `<!doctype html>
<html>
	<head>
		<title>Zigzag API Reference</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script id="api-reference" data-url="openapi-public.json"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.29.3"></script>
	</body>
</html>`

func startAPI(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalCh
		log.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	dbConnInfo, err := database.GetSqlDbConn(cfg)
	if err != nil {
		log.Fatalf("Failed to get database connection info: %v", err)
	}
	log.Infof("Using PostgreSQL database %s@%s:%d/%s", cfg.PostgresUser, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
	db, sqlDB, err := database.NewEntClient(dbConnInfo)
	if err != nil {
		log.Fatalf("Failed to create ent client: %v", err)
	}

	log.Info("Running migrations...")
	goose.SetLogger(log.GetLogger())
	goose.SetBaseFS(entmigrate.MigrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect error: %v", err)
	}
	if err := goose.Up(sqlDB, "migrations", goose.WithAllowMissing()); err != nil {
		log.Fatalf("goose up err: %v", err)
	}
	log.Info("Migrations applied successfully")

	repo := repositories.NewRepositories(db)
	if _, err := repo.System().GetSystemSettings(ctx, nil); err != nil {
		log.Fatalf("Failed to load system settings: %v", err)
	}

	encryptor, err := crypto.NewEncryptor(cfg.EncryptionKeyBytes())
	if err != nil {
		log.Fatalf("Invalid ENCRYPTION_KEY: %v", err)
	}
	store, err := storage.NewLocalStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to prepare the upload directory: %v", err)
	}

	issuer := oauthserver.Issuer(cfg.ExternalUIUrl)
	pkey, _, err := repo.Oauth().GetOrGenerateJWTPrivateKey(ctx)
	if err != nil {
		log.Fatalf("Failed to load JWT signing key: %v", err)
	}
	tokenManager := auth.NewTokenManager(pkey, issuer, cfg.TokenAudience)
	oidcHandler := auth.NewOIDCHandler(tokenManager)

	emailService := email_service.NewEmailService(cfg, repo, encryptor, mail.SMTPSender{})
	metadataFetcher := oauthserver.NewMetadataFetcher(oauthserver.DefaultIPPolicy, cache.NewMemoryCache[oauthserver.ClientMetadata]())
	oauthServerService := oauthserver_service.NewOAuthServerService(repo, metadataFetcher, issuer)
	notificationService := notification_service.NewNotificationService(cfg, repo, emailService)
	webhookService := webhook_service.NewWebhookService(cfg, repo)
	cycleService := cycle_service.NewCycleService(cfg, repo, webhookService)
	githubClient := github.NewClient(cfg.GithubAPIURL)
	githubSyncService := github_sync_service.NewGithubSyncService(repo, encryptor, githubClient)

	srvImpl := &server.Server{
		Cfg:                 cfg,
		Repository:          repo,
		TokenManager:        tokenManager,
		Storage:             store,
		UserService:         user_service.NewUserService(cfg, repo, emailService),
		WorkspaceService:    workspace_service.NewWorkspaceService(cfg, repo, emailService),
		TeamService:         team_service.NewTeamService(repo, cycleService),
		EmailService:        emailService,
		APIKeyService:       apikey_service.NewAPIKeyService(repo),
		OAuthServerService:  oauthServerService,
		SystemService:       system_service.NewSystemService(repo),
		UploadService:       upload_service.NewUploadService(cfg, repo, store),
		IssueService:        issue_service.NewIssueService(cfg, repo, notificationService, webhookService, githubSyncService),
		ProjectService:      project_service.NewProjectService(cfg, repo, notificationService, webhookService),
		LabelService:        label_service.NewLabelService(cfg, repo),
		CycleService:        cycleService,
		InitiativeService:   initiative_service.NewInitiativeService(cfg, repo),
		NotificationService: notificationService,
		ViewService:         view_service.NewViewService(cfg, repo),
		TemplateService:     template_service.NewTemplateService(cfg, repo),
		FavoriteService:     favorite_service.NewFavoriteService(cfg, repo),
		WebhookService:      webhookService,
		ImportService:       import_service.NewImportService(cfg, repo),
	}
	srvImpl.SearchService = search_service.NewSearchService(cfg, repo, srvImpl.IssueService)
	srvImpl.DocumentService = document_service.NewDocumentService(cfg, repo, srvImpl.IssueService)
	srvImpl.CommentService = comment_service.NewCommentService(cfg, repo, srvImpl.IssueService, notificationService, webhookService, githubSyncService)
	srvImpl.GithubService = github_service.NewGithubService(cfg, repo, encryptor, github_service.Dependencies{
		Issues: srvImpl.IssueService, Comments: srvImpl.CommentService, Labels: srvImpl.LabelService, Projects: srvImpl.ProjectService, Client: githubClient,
	})
	if err := srvImpl.GithubService.RecoverBackfills(ctx); err != nil {
		log.Error("Failed to mark interrupted GitHub imports", "err", err)
	}

	allowedOrigins := []string{cfg.ExternalUIUrl}
	if cfg.InjectDevOrigins {
		allowedOrigins = append(allowedOrigins, "http://localhost:5174", "http://127.0.0.1:5174", "http://localhost:8091")
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", middleware.RequestIDHeader},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version": "` + Version + `"}`))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	r.Get("/.well-known/openid-configuration", oidcHandler.HandleOpenIDConfiguration)
	r.Get("/.well-known/jwks.json", oidcHandler.HandleJWKS)
	r.Handle(cfg.APIPathPrefix()+"/*", underPrefix(cfg.APIPathPrefix(), r))

	rateLimiter := middleware.NewRateLimiter()
	root := r
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)

		humaConfig := router.NewHumaConfig("Zigzag API", Version, cfg.CookieSecure)
		humaConfig.DocsPath = ""
		humaConfig.Servers = []*huma.Server{{URL: cfg.ExternalAPIURL}}
		api := humachi.New(r, humaConfig)

		mw := middleware.NewMiddleware(cfg, repo, api, tokenManager, allowedOrigins)
		api.UseMiddleware(mw.Recoverer)

		r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(docsPage))
		})

		router.RegisterRoutes(api, srvImpl, mw)

		publicSpec, err := json.Marshal(oapi.APIKeySpec(api.OpenAPI()))
		if err != nil {
			log.Fatal("Failed to render the public OpenAPI spec", "err", err)
		}
		r.Get("/openapi-public.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(publicSpec)
		})

		oauthserver_handler.NewHandler(oauthServerService).Mount(r, rateLimiter)

		mcpServer, err := mcpserver.New(mcpserver.Options{
			API:     api,
			Router:  root,
			OAuth:   oauthServerService,
			APIKeys: repo.APIKey(),
			Issuer:  issuer,
			Version: Version,
		})
		if err != nil {
			log.Fatal("Failed to build the MCP server", "err", err)
		}
		mcpServer.Mount(r, rateLimiter)
	})

	r.NotFound(web.Handler().ServeHTTP)

	jobs, err := scheduler.Start(ctx, scheduler.Jobs(scheduler.Dependencies{
		Cycles:         srvImpl.CycleService,
		OAuth:          oauthServerService,
		Invitations:    srvImpl.WorkspaceService,
		PasswordResets: srvImpl.UserService,
		Trash:          srvImpl.IssueService,
		Automation:     srvImpl.IssueService,
		Notifications:  srvImpl.NotificationService,
	}))
	if err != nil {
		log.Fatal("Failed to start the scheduler", "err", err)
	}
	defer func() {
		if err := jobs.Shutdown(); err != nil {
			log.Error("Scheduler shutdown error", "err", err)
		}
	}()

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	httpServer := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           r,
		Protocols:         protocols,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Infof("Starting server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Info("Server gracefully stopped")
}

var rootOnlyPaths = []string{"/mcp", "/oauth/", "/.well-known/"}

// underPrefix serves the API again below prefix; chi's route state is reset so the inner pass routes from scratch.
func underPrefix(prefix string, next http.Handler) http.Handler {
	return http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, path := range rootOnlyPaths {
			if req.URL.Path == strings.TrimSuffix(path, "/") || strings.HasPrefix(req.URL.Path, path) {
				http.NotFound(w, req)
				return
			}
		}
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext())))
	}))
}

func main() {
	if err := godotenv.Overload(); err != nil && !os.IsNotExist(err) {
		log.Warn("Error loading .env file", "err", err)
	}
	cfg := config.NewConfig()
	log.SetLevel(cfg.LogLevel)
	log.Infof("Starting Zigzag version %s", Version)
	startAPI(cfg)
}
