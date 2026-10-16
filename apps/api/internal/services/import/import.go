package import_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ImportServiceInterface -p import_service -s ImportService -o import_service_iface.go
type ImportService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewImportService(cfg *config.Config, repo repositories.RepositoriesInterface) *ImportService {
	return &ImportService{cfg: cfg, repo: repo}
}
