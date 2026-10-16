package search_handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	search_service "github.com/zigzagapp-org/zigzag-api/internal/services/search"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "search",
		Summary:     "Search",
		Description: "Search a workspace for issues (title, then description), projects (name and summary) and documents (title and content) in one call. Returns up to limit results of each kind, best matches first; trashed and archived projects and documents and trashed issues are left out.",
		Path:        "",
		Method:      http.MethodGet,
	}, handlers.Search, oapi.MCP)
}

type SearchInput struct {
	server.BaseAuthInput
	search_service.SearchInput
}

type SearchResult struct {
	Body struct {
		Data *models.SearchResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Search(ctx context.Context, input *SearchInput) (*SearchResult, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := self.srv.SearchService.Search(ctx, user.ID, &input.SearchInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SearchResult{}
	resp.Body.Data = result
	return resp, nil
}
