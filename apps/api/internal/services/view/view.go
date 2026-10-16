package view_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ViewServiceInterface -p view_service -s ViewService -o view_service_iface.go
type ViewService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewViewService(cfg *config.Config, repo repositories.RepositoriesInterface) *ViewService {
	return &ViewService{cfg: cfg, repo: repo}
}
