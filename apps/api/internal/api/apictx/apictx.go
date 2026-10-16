package apictx

import (
	"context"

	"github.com/zigzagapp-org/zigzag-api/ent"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

type contextKey string

const (
	UserKey contextKey = "user"
)

type mcpCallerKey struct{}

type MCPCaller struct {
	User   *ent.User
	Access permissions_repo.APIKeyAccess
	Via    string
}

func WithMCPCaller(ctx context.Context, caller *MCPCaller) context.Context {
	return context.WithValue(ctx, mcpCallerKey{}, caller)
}

func MCPCallerFromContext(ctx context.Context) (*MCPCaller, bool) {
	caller, ok := ctx.Value(mcpCallerKey{}).(*MCPCaller)
	return caller, ok && caller != nil && caller.User != nil
}
