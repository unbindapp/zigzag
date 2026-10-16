package views_handler

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
		OperationID: "list-views",
		Summary:     "List Views",
		Description: "List the saved views you own plus the views shared with the workspace. Pass a view's filters to list-issues to see its issues.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-view",
		Summary:     "Get View",
		Description: "Get a saved view with its filters (list-issues parameters) and display settings.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-view",
		Summary:     "Create View",
		Description: "Save a set of list-issues filters and display settings as a named view. Private unless shared is true. Returns the view. Needs member.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-view",
		Summary:     "Update View",
		Description: "Change a view's name, sharing, filters or display. Owners edit their views; admins also edit shared views of others. Returns the view.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-view",
		Summary:     "Delete View",
		Description: "Delete a saved view. Owners delete their views; admins also delete shared views of others. Issues are not touched.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)
}
