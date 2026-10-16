package templates_handler

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
		OperationID: "list-templates",
		Summary:     "List Templates",
		Description: "List the workspace templates, plus a team's templates with team_id. Pass a template_id to create-issue to start an issue from it.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-template",
		Summary:     "Get Template",
		Description: "Get a template with the values a new issue starts with.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-template",
		Summary:     "Create Template",
		Description: "Create an issue, project or document template for the workspace or one team. Referenced labels, state, assignee and project must exist. Needs admin.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-template",
		Summary:     "Update Template",
		Description: "Rename, describe, reorder or replace the values of a template. Omitted fields keep their value. Needs admin.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-template",
		Summary:     "Delete Template",
		Description: "Delete a template. Issues created from it stay as they are. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)
}
