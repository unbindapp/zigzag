package webhooks_handler

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
		OperationID: "list-webhooks",
		Summary:     "List Webhooks",
		Description: "List the webhooks of a workspace with their url, events, team scope and whether they are enabled. Secrets are never listed. Needs admin.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-webhook",
		Summary:     "Create Webhook",
		Description: "Create a webhook that POSTs a signed JSON envelope {id, event, workspace_id, created_at, data} to a public url for each chosen event, optionally only for one team. Returns the signing secret once; the X-Zigzag-Signature header is sha256=<hex HMAC-SHA256 of the body>. Needs admin.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP, oapi.OpenWorld)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-webhook",
		Summary:     "Update Webhook",
		Description: "Change a webhook's url, label, events, team scope or enabled flag. Omitted fields keep their value. Needs admin.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP, oapi.OpenWorld)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-webhook",
		Summary:     "Delete Webhook",
		Description: "Delete a webhook; no more events are sent to it. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "test-webhook",
		Summary:     "Test Webhook",
		Description: "Send one signed webhook.test event to the webhook now and report the status the endpoint answered. Needs admin.",
		Path:        "/test",
		Method:      http.MethodPost,
	}, handlers.Test, oapi.NoMCP("sends a request to an outside endpoint on demand; a settings page action for people"), oapi.OpenWorld)
}
