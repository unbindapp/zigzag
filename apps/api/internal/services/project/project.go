package project_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ProjectServiceInterface -p project_service -s ProjectService -o project_service_iface.go
type ProjectService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewProjectService(cfg *config.Config, repo repositories.RepositoriesInterface) *ProjectService {
	return &ProjectService{cfg: cfg, repo: repo}
}
