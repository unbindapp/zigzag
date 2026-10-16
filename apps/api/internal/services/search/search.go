package search_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i SearchServiceInterface -p search_service -s SearchService -o search_service_iface.go
type SearchService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewSearchService(cfg *config.Config, repo repositories.RepositoriesInterface) *SearchService {
	return &SearchService{cfg: cfg, repo: repo}
}
