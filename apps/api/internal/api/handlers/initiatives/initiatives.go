package initiatives_handler

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
		OperationID: "list-initiatives",
		Summary:     "List Initiatives",
		Description: "List a workspace's initiatives with owner, projects, health and progress over their projects' issues. Archived and trashed ones are hidden unless include_archived or include_trashed is set. Not paginated.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-initiative",
		Summary:     "Get Initiative",
		Description: "Get an initiative by initiative_id, or by workspace_id and slug, with its projects, health and progress.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-initiative",
		Summary:     "Create Initiative",
		Description: "Create an initiative grouping projects of a workspace toward a goal from workspace_id and name. Status defaults to planned. Returns the initiative. Needs member.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-initiative",
		Summary:     "Update Initiative",
		Description: "Change an initiative's name, description, owner, status, target date or projects. Omitted fields keep their value; project_ids replaces the list. Returns the initiative. Needs member.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-initiative",
		Summary:     "Delete Initiative",
		Description: "Delete an initiative with its updates and documents. Its projects stay. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "archive-initiative",
		Summary:     "Archive Initiative",
		Description: "Archive an initiative so it is hidden from lists unless include_archived is set. Returns the initiative. Needs member.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unarchive-initiative",
		Summary:     "Unarchive Initiative",
		Description: "Bring an archived initiative back into lists. Returns the initiative. Needs member.",
		Path:        "/unarchive",
		Method:      http.MethodPost,
	}, handlers.Unarchive, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "add-initiative-project",
		Summary:     "Add Initiative Project",
		Description: "Add a project of the same workspace to an initiative. Returns the initiative. Needs member.",
		Path:        "/projects/add",
		Method:      http.MethodPost,
	}, handlers.AddProject, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "remove-initiative-project",
		Summary:     "Remove Initiative Project",
		Description: "Take a project out of an initiative. The project itself is kept. Returns the initiative. Needs member.",
		Path:        "/projects/remove",
		Method:      http.MethodDelete,
	}, handlers.RemoveProject, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-initiative-updates",
		Summary:     "List Initiative Updates",
		Description: "List an initiative's status updates, newest first, with author and health.",
		Path:        "/updates/list",
		Method:      http.MethodGet,
	}, handlers.ListUpdates, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-initiative-update",
		Summary:     "Create Initiative Update",
		Description: "Post a status update on an initiative with a markdown body and a health. The initiative's health follows the latest update. Returns the update. Needs member.",
		Path:        "/updates/create",
		Method:      http.MethodPost,
	}, handlers.CreateUpdate, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-initiative-update",
		Summary:     "Delete Initiative Update",
		Description: "Delete an initiative update. Only its author or an admin may.",
		Path:        "/updates/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteUpdate, oapi.MCP)
}
