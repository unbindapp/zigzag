package apictx

import (
	"context"

	"github.com/unbindapp/zigzag-api/ent/schema"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

func HistorySource(ctx context.Context) schema.HistorySource {
	if _, ok := MCPCallerFromContext(ctx); ok {
		return schema.HistorySourceMCP
	}
	if _, ok := permissions_repo.APIKeyAccessFromContext(ctx); ok {
		return schema.HistorySourceAPI
	}
	if permissions_repo.IsSystemCaller(ctx) {
		return schema.HistorySourceSystem
	}
	return schema.HistorySourceUser
}
