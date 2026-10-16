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

Call whoami first. It lists every workspace you can see with your role. When api_key is present, this connection is limited to that role and those workspaces; anything outside them answers "not found" or "forbidden", so do not retry.

Resources nest as workspace > team > issue. Projects, labels, initiatives, documents and views belong to a workspace and are shared by its teams; a project can span several teams. Find ids with list-workspaces, list-teams, list-projects and list-labels.

Issue identifiers look like KEY-123, where KEY is the team key. Most issue tools accept either issue_id or identifier.

list-issues is filtered and paginated: pass the opaque next_cursor from one page as cursor to get the next. Archived and trashed issues are hidden unless you ask for them with include_archived or include_trashed.

Creating an issue needs team_id and title. Everything else defaults: the state is the team's default state and the priority is none (0). Priorities are 0 none, 1 urgent, 2 high, 3 medium, 4 low.

Descriptions and comments accept markdown. Mention a user with @[Name](user:<user_id>); mentioned users are notified and subscribed.`
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
