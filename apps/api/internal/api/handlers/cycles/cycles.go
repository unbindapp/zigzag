package cycles_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-cycles",
		Summary:     "List Cycles",
		Description: "List a team's cycles, newest number first, each with is_active, is_past, is_future and progress (scope, completed, started, percent, issue counts). Archived cycles show with include_archived. Not paginated.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-cycle",
		Summary:     "Get Cycle",
		Description: "Get a cycle by cycle_id, or by team_id and number, with progress and its daily scope history for a burn-up chart. List its issues with list-issues and cycle_ids.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-cycle",
		Summary:     "Create Cycle",
		Description: "Create a cycle for a team with cycles turned on from team_id, starts_at and ends_at. It gets the next number; its dates may not overlap another cycle. Teams with cycles on also get cycles planned automatically. Returns the cycle. Needs admin.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-cycle",
		Summary:     "Update Cycle",
		Description: "Rename, describe or move the dates of a cycle. Omitted fields keep their value. Returns the cycle. Needs admin.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-cycle",
		Summary:     "Delete Cycle",
		Description: "Delete a cycle. Its issues stay and leave the cycle. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "archive-cycle",
		Summary:     "Archive Cycle",
		Description: "Archive a cycle so it is hidden from list-cycles unless include_archived is set. Returns the cycle. Needs admin.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)
}
