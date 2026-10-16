package router

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/internal/api/middleware"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
	"github.com/unbindapp/zigzag-api/internal/auth"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"

	apikeys_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/apikeys"
	auth_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/auth"
	connectedapps_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/connectedapps"
	cycles_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/cycles"
	documents_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/documents"
	favorites_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/favorites"
	imports_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/imports"
	initiatives_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/initiatives"
	invitations_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/invitations"
	issues_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/issues"
	labels_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/labels"
	notifications_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/notifications"
	projects_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/projects"
	search_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/search"
	setup_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/setup"
	system_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/system"
	teams_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/teams"
	templates_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/templates"
	uploads_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/uploads"
	user_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/users"
	views_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/views"
	webhooks_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/webhooks"
	workspaces_handler "github.com/unbindapp/zigzag-api/internal/api/handlers/workspaces"
)

func NewHumaConfig(title, version string, cookieSecure bool) huma.Config {
	huma.NewError = errdefs.HumaErrorFunc

	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)

	return huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:   title,
				Version: version,
			},
			Components: &huma.Components{
				Schemas: registry,
				SecuritySchemes: map[string]*huma.SecurityScheme{
					"cookieAuth": {
						Type:        "apiKey",
						In:          "cookie",
						Name:        auth.AccessTokenCookieName(cookieSecure),
						Description: "Session cookie set by /auth/login. Sent automatically by browsers.",
					},
					"bearerAuth": {
						Type:         "http",
						Scheme:       "bearer",
						BearerFormat: "JWT",
						Description:  "Access token or API key (`zz_...`) passed as `Authorization: Bearer <token>`.",
					},
				},
			},
		},
		OpenAPIPath:   "/openapi",
		DocsPath:      "/docs",
		SchemasPath:   "/schemas",
		Formats:       huma.DefaultFormats,
		DefaultFormat: "application/json",
		Transformers:  []huma.Transformer{middleware.ErrorTransformer},
	}
}

// RegisterRoutes only wires routes and metadata, so it is safe with an empty *server.Server for the offline spec.
func RegisterRoutes(api huma.API, srvImpl *server.Server, mw *middleware.Middleware) {
	authSecurity := []map[string][]string{
		{"cookieAuth": {}},
		{"bearerAuth": {}},
	}
	registerGroup := func(prefix, tag string, authed, sessionOnly bool, fn func(*server.Server, *huma.Group)) {
		grp := huma.NewGroup(api, prefix)
		grp.UseMiddleware(mw.RateLimitOperations)
		if authed {
			grp.UseMiddleware(mw.Authenticate)
			grp.UseMiddleware(mw.CSRF)
		}
		grp.UseModifier(func(op *huma.Operation, next func(*huma.Operation)) {
			op.Tags = []string{tag}
			if authed {
				op.Security = authSecurity
			}
			if sessionOnly {
				oapi.MarkSessionOnly(op)
			}
			next(op)
		})
		fn(srvImpl, grp)
	}
	register := func(prefix, tag string, authed bool, fn func(*server.Server, *huma.Group)) {
		registerGroup(prefix, tag, authed, false, fn)
	}
	registerSessionOnly := func(prefix, tag string, fn func(*server.Server, *huma.Group)) {
		registerGroup(prefix, tag, true, true, fn)
	}

	register("/setup", "Setup", false, setup_handler.RegisterHandlers)
	register("/auth", "Auth", false, auth_handler.RegisterHandlers)
	register("/users", "Users", true, user_handler.RegisterHandlers)
	register("/workspaces", "Workspaces", true, workspaces_handler.RegisterHandlers)
	register("/invitations", "Invitations", false, invitations_handler.RegisterHandlers)
	register("/teams", "Teams", true, teams_handler.RegisterHandlers)
	register("/labels", "Labels", true, labels_handler.RegisterHandlers)
	register("/projects", "Projects", true, projects_handler.RegisterHandlers)
	register("/initiatives", "Initiatives", true, initiatives_handler.RegisterHandlers)
	register("/cycles", "Cycles", true, cycles_handler.RegisterHandlers)
	register("/issues", "Issues", true, issues_handler.RegisterHandlers)
	register("/documents", "Documents", true, documents_handler.RegisterHandlers)
	register("/views", "Views", true, views_handler.RegisterHandlers)
	register("/templates", "Templates", true, templates_handler.RegisterHandlers)
	register("/favorites", "Favorites", true, favorites_handler.RegisterHandlers)
	register("/notifications", "Notifications", true, notifications_handler.RegisterHandlers)
	registerSessionOnly("/uploads", "Uploads", uploads_handler.RegisterHandlers)
	registerSessionOnly("/api-keys", "API Keys", apikeys_handler.RegisterHandlers)
	registerSessionOnly("/connected-apps", "Connected Apps", connectedapps_handler.RegisterHandlers)
	registerSessionOnly("/system", "System", system_handler.RegisterHandlers)
	register("/webhooks", "Webhooks", true, webhooks_handler.RegisterHandlers)
	register("/imports", "Imports", true, imports_handler.RegisterHandlers)
	register("/search", "Search", true, search_handler.RegisterHandlers)
}
