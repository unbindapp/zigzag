package initiative_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i InitiativeServiceInterface -p initiative_service -s InitiativeService -o initiative_service_iface.go
type InitiativeService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewInitiativeService(cfg *config.Config, repo repositories.RepositoriesInterface) *InitiativeService {
	return &InitiativeService{cfg: cfg, repo: repo}
}
