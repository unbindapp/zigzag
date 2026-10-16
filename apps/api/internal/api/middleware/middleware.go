package middleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

type Middleware struct {
	tokenManager   *auth.TokenManager
	repository     repositories.RepositoriesInterface
	api            huma.API
	cfg            *config.Config
	allowedOrigins []string
	limiter        *RateLimiter
}

func NewMiddleware(cfg *config.Config, repository repositories.RepositoriesInterface, api huma.API, tokenManager *auth.TokenManager, allowedOrigins []string) *Middleware {
	return &Middleware{
		tokenManager:   tokenManager,
		repository:     repository,
		api:            api,
		cfg:            cfg,
		allowedOrigins: allowedOrigins,
		limiter:        NewRateLimiter(),
	}
}
