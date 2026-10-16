package labels_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-labels",
		Summary:     "List Labels",
		Description: "List the workspace labels, plus a team's labels with team_id or every visible team's with all_teams. Groups come with their labels in children; archived labels show with include_archived. Not paginated.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-label",
		Summary:     "Create Label",
		Description: "Create a workspace label, or a team label with team_id, from workspace_id, name and color. Set is_group for a group, or parent_id to put the label in a group. Returns the label. Needs admin.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-label",
		Summary:     "Update Label",
		Description: "Rename, recolor, describe, regroup or reorder a label. Omitted fields keep their value. Returns the label. Needs admin.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "archive-label",
		Summary:     "Archive Label",
		Description: "Archive a label, and the labels of a group. Archived labels stay on issues but are hidden from pickers. Needs admin.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "unarchive-label",
		Summary:     "Unarchive Label",
		Description: "Restore an archived label, and the labels of a group. Needs admin.",
		Path:        "/unarchive",
		Method:      http.MethodPost,
	}, handlers.Unarchive, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-label",
		Summary:     "Delete Label",
		Description: "Delete a label permanently and remove it from every issue. Deleting a group deletes its labels too. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)
}
