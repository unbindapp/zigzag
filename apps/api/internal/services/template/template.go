package template_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i TemplateServiceInterface -p template_service -s TemplateService -o template_service_iface.go
type TemplateService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewTemplateService(cfg *config.Config, repo repositories.RepositoriesInterface) *TemplateService {
	return &TemplateService{cfg: cfg, repo: repo}
}
