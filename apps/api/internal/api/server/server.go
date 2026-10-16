package server

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/api/apictx"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	apikey_service "github.com/zigzagapp-org/zigzag-api/internal/services/apikey"
	comment_service "github.com/zigzagapp-org/zigzag-api/internal/services/comment"
	cycle_service "github.com/zigzagapp-org/zigzag-api/internal/services/cycle"
	document_service "github.com/zigzagapp-org/zigzag-api/internal/services/document"
	email_service "github.com/zigzagapp-org/zigzag-api/internal/services/email"
	favorite_service "github.com/zigzagapp-org/zigzag-api/internal/services/favorite"
	github_service "github.com/zigzagapp-org/zigzag-api/internal/services/github"
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
)

type EmptyInput struct{}

type BaseAuthInput struct{}

type Server struct {
	Cfg          *config.Config
	Repository   repositories.RepositoriesInterface
	TokenManager *auth.TokenManager

	UserService         *user_service.UserService
	WorkspaceService    *workspace_service.WorkspaceService
	TeamService         *team_service.TeamService
	EmailService        *email_service.EmailService
	APIKeyService       *apikey_service.APIKeyService
	OAuthServerService  *oauthserver_service.OAuthServerService
	SystemService       *system_service.SystemService
	UploadService       *upload_service.UploadService
	IssueService        *issue_service.IssueService
	ProjectService      *project_service.ProjectService
	LabelService        *label_service.LabelService
	CycleService        *cycle_service.CycleService
	InitiativeService   *initiative_service.InitiativeService
	CommentService      *comment_service.CommentService
	NotificationService *notification_service.NotificationService
	DocumentService     *document_service.DocumentService
	ViewService         *view_service.ViewService
	TemplateService     *template_service.TemplateService
	FavoriteService     *favorite_service.FavoriteService
	WebhookService      *webhook_service.WebhookService
	ImportService       *import_service.ImportService
	SearchService       *search_service.SearchService
	GithubService       *github_service.GithubService
}

func (self *Server) GetUserFromContext(ctx context.Context) (user *ent.User, found bool) {
	user, found = ctx.Value(apictx.UserKey).(*ent.User)
	return user, found
}

func (self *Server) AuthenticatedUser(ctx context.Context) (*ent.User, error) {
	user, found := self.GetUserFromContext(ctx)
	if !found {
		return nil, huma.Error401Unauthorized("Unable to retrieve user")
	}
	return user, nil
}

func (self *Server) IssueSession(ctx context.Context, user *ent.User) ([]http.Cookie, error) {
	accessToken, accessExpiresAt, err := self.TokenManager.MintAccessToken(user)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to issue a session token")
	}
	refreshToken := auth.NewRefreshToken()
	_, err = self.Repository.Oauth().CreateToken(ctx, accessToken, refreshToken, self.Cfg.TokenAudience, auth.TokenScope, time.Now().Add(auth.RefreshTokenTTL), user)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to store the session")
	}
	csrfToken := auth.MintCSRFToken(self.TokenManager.CSRFSecret(), refreshToken)
	return auth.SessionCookies(accessToken, accessExpiresAt, refreshToken, csrfToken, self.Cfg.CookieSecure), nil
}
