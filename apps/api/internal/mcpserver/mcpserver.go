package mcpserver

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	entSchema "github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/middleware"
	"github.com/unbindapp/zigzag-api/internal/oauthserver"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

const (
	maxRequestBodyBytes = 1 << 20
	requestsPerMinute   = 600

	instructions = `Zigzag is a self-hosted issue tracker modeled on Linear.

Call whoami first. It lists every workspace you can see with your role. When api_key is present, this connection is limited to that role and those workspaces; anything outside them answers "not found" or "forbidden", so do not retry. Roles nest: viewer reads, member also creates and edits issues, comments, projects and documents, admin also deletes and manages teams, labels, workflow states, cycles, templates, webhooks and imports. A viewer credential is offered read tools only.

Resources nest as workspace > team > issue. Each team has its own workflow states (list-workflow-states, typed triage, backlog, unstarted, started, completed or canceled) and cycles (list-cycles). Projects, labels, initiatives, documents, views and templates belong to a workspace and are shared by its teams; a project can span several teams and has milestones. Labels are workspace labels or team labels, and a label group cannot be put on an issue. Find ids with list-workspaces, list-teams, list-projects and list-labels, or search across issues, projects and documents with search.

Issue identifiers look like KEY-123, where KEY is the team key. Most issue tools accept either issue_id or identifier.

list-issues is filtered (team, state, assignee, label, project, cycle, priority, dates, text and more) and paginated: pass the opaque next_cursor from one page as cursor to get the next; total counts every match. list-comments, list-issue-history and list-notifications page the same way; other lists are not paginated. Archived and trashed issues are hidden unless you ask for them with include_archived or include_trashed.

Creating an issue needs team_id and title. Everything else defaults: the state is the team's default state and the priority is none (0). Priorities are 0 none, 1 urgent, 2 high, 3 medium, 4 low. On update tools omitted fields keep their value, and the nil uuid 00000000-0000-0000-0000-000000000000 clears an id field such as assignee_id. Prefer trash-issue over delete-issue: trash can be undone for 30 days.

Descriptions, comments, documents and updates accept markdown. Mention a user with @[Name](user:<user_id>); mentioned users are notified and subscribed. Your own notifications are in list-notifications.

import-linear takes a Linear export and is a dry run unless dry_run is false: read its report (counts, warnings, user mapping) before importing for real. Running it again updates what an earlier run imported instead of duplicating it. list-imports shows past runs.`
)

type Options struct {
	API     huma.API
	Router  http.Handler
	OAuth   AccessTokenVerifier
	APIKeys APIKeyStore
	Issuer  string
	Version string
}

type Server struct {
	handler http.Handler
	tools   []*tool
}

func New(opts Options) (*Server, error) {
	tools, err := buildTools(opts.API)
	if err != nil {
		return nil, err
	}

	servers := &serverSet{
		version:  opts.Version,
		tools:    tools,
		dispatch: &dispatcher{router: opts.Router},
		servers:  map[string]*mcp.Server{},
	}

	streamable := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		caller, ok := callerOf(auth.TokenInfoFromContext(r.Context()))
		if !ok {
			return servers.serverFor(everything)
		}
		return servers.serverFor(caller.Access)
	}, &mcp.StreamableHTTPOptions{
		Stateless:           true,
		JSONResponse:        true,
		MaxRequestBodyBytes: maxRequestBodyBytes,
	})

	verifier := newVerifier(opts.OAuth, opts.APIKeys, time.Now)
	handler := auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL:    oauthserver.ResourceMetadataURL(opts.Issuer),
		AllowMissingExpiration: true,
	})(streamable)

	return &Server{handler: handler, tools: tools}, nil
}

func (self *Server) Mount(r chi.Router, limiter *middleware.RateLimiter) {
	r.With(limiter.Limit("mcp", requestsPerMinute, time.Minute)).Handle("/mcp", self.handler)
}

var everything = permissions_repo.APIKeyAccess{
	Role:       entSchema.WorkspaceRoleAdmin,
	FullAccess: true,
	Privileges: entSchema.AllPrivileges(),
}

type serverSet struct {
	version  string
	tools    []*tool
	dispatch *dispatcher
	mu       sync.Mutex
	servers  map[string]*mcp.Server
}

func (self *serverSet) serverFor(access permissions_repo.APIKeyAccess) *mcp.Server {
	offered := make([]*tool, 0, len(self.tools))
	var key strings.Builder
	for i, t := range self.tools {
		if !t.offeredTo(access) {
			continue
		}
		offered = append(offered, t)
		key.WriteString(strconv.Itoa(i))
		key.WriteByte(',')
	}

	self.mu.Lock()
	defer self.mu.Unlock()
	if server, ok := self.servers[key.String()]; ok {
		return server
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "zigzag", Title: "Zigzag", Version: self.version}, &mcp.ServerOptions{
		Instructions: instructions,
	})
	for _, t := range offered {
		server.AddTool(t.definition, self.dispatch.handlerFor(t))
	}
	self.servers[key.String()] = server
	return server
}
