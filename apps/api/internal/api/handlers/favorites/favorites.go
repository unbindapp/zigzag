package favorites_handler

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
		OperationID: "list-favorites",
		Summary:     "List Favorites",
		Description: "List your favorites in a workspace in your order, each with the entity's current name and url_params. Entities that are gone or hidden from you are left out.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "add-favorite",
		Summary:     "Add Favorite",
		Description: "Star an issue, project, cycle, view, document, label, team or initiative for yourself. Adding one that is already a favorite returns it unchanged.",
		Path:        "/add",
		Method:      http.MethodPost,
	}, handlers.Add, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "remove-favorite",
		Summary:     "Remove Favorite",
		Description: "Unstar an entity. Only your own favorites change; the entity is not touched.",
		Path:        "/remove",
		Method:      http.MethodDelete,
	}, handlers.Remove, oapi.MCP, oapi.Risk("low"))

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "reorder-favorites",
		Summary:     "Reorder Favorites",
		Description: "Put your favorites in a new order by listing their favorite ids, first on top. Returns the reordered list.",
		Path:        "/reorder",
		Method:      http.MethodPut,
	}, handlers.Reorder, oapi.MCP)
}
