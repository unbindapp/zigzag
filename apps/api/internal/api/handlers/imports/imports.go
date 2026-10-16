package imports_handler

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
)

const (
	maxExportBytes    = 64 << 20
	exportReadTimeout = 5 * time.Minute
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID:     "import-linear",
		Summary:         "Import Linear Export",
		Description:     "Import a Linear export (users, teams, states, labels, projects, milestones, cycles, issues with comments, attachments, relations and state history) into a workspace. Runs as a dry run unless dry_run is false: the report then shows exactly what would be created, updated and skipped, with warnings and how each Linear user maps to an account. Issues keep their numbers, and running it again updates instead of duplicating. Returns the report. Needs admin.",
		Path:            "/linear",
		Method:          http.MethodPost,
		MaxBodyBytes:    maxExportBytes,
		BodyReadTimeout: exportReadTimeout,
	}, handlers.ImportLinear, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-imports",
		Summary:     "List Imports",
		Description: "List the 50 most recent imports of a workspace, newest first, dry runs included, each with its status, who started it and its full report. Needs admin.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)
}
