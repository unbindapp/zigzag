package connectedapps_handler

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
		OperationID: "get-connected-app-client",
		Summary:     "Describe OAuth Client",
		Description: "Describe the client asking for access on the consent page. The name is self reported by the client.",
		Path:        "/client",
		Method:      http.MethodGet,
	}, handlers.GetClient, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "approve-connected-app",
		Summary:     "Approve Connected App",
		Description: "Grant an OAuth client access within the current user's own permissions. Returns the URL to send the browser to.",
		Path:        "/approve",
		Method:      http.MethodPost,
	}, handlers.Approve, oapi.SessionOnly, oapi.Confirm)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "deny-connected-app",
		Summary:     "Deny Connected App",
		Description: "Refuse an OAuth client. Returns the URL to send the browser to.",
		Path:        "/deny",
		Method:      http.MethodPost,
	}, handlers.Deny, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-connected-apps",
		Summary:     "List Connected Apps",
		Description: "List the OAuth clients the current user has granted access to.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "revoke-connected-app",
		Summary:     "Revoke Connected App",
		Description: "Revoke a connected app. Its access and refresh tokens stop working immediately.",
		Path:        "/revoke",
		Method:      http.MethodDelete,
	}, handlers.Revoke, oapi.SessionOnly)
}
