package documents_handler

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
		OperationID: "list-documents",
		Summary:     "List Documents",
		Description: "List a workspace's documents, most recently updated first, optionally only those of a project, initiative or issue. Archived and trashed ones are hidden unless include_archived or include_trashed is set. Not paginated.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-document",
		Summary:     "Get Document",
		Description: "Get a document with its markdown content by document_id, or by workspace_id and slug.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-document",
		Summary:     "Create Document",
		Description: "Create a markdown document in a workspace from workspace_id and title, optionally attached to a project, initiative or issue of that workspace. Returns the document. Needs member.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-document",
		Summary:     "Update Document",
		Description: "Change a document's title, icon or content, or move it to another project, initiative or issue. Omitted fields keep their value. Returns the document. Needs member.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-document",
		Summary:     "Delete Document",
		Description: "Delete a document for good. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "archive-document",
		Summary:     "Archive Document",
		Description: "Archive a document so it is hidden from lists unless include_archived is set. Returns the document. Needs member.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unarchive-document",
		Summary:     "Unarchive Document",
		Description: "Bring an archived document back into lists. Returns the document. Needs member.",
		Path:        "/unarchive",
		Method:      http.MethodPost,
	}, handlers.Unarchive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "trash-document",
		Summary:     "Trash Document",
		Description: "Move a document to the trash; it is hidden from lists unless include_trashed is set. Returns the document. Needs member.",
		Path:        "/trash",
		Method:      http.MethodPost,
	}, handlers.Trash, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "restore-document",
		Summary:     "Restore Document",
		Description: "Take a document out of the trash. Returns the document. Needs member.",
		Path:        "/restore",
		Method:      http.MethodPost,
	}, handlers.Restore, oapi.MCP)
}
