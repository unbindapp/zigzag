package label_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i LabelServiceInterface -p label_service -s LabelService -o label_service_iface.go
type LabelService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewLabelService(cfg *config.Config, repo repositories.RepositoriesInterface) *LabelService {
	return &LabelService{cfg: cfg, repo: repo}
}
