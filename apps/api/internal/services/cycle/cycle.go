package cycle_service

import (
	"context"

	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i CycleServiceInterface -p cycle_service -s CycleService -o cycle_service_iface.go
type CycleService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewCycleService(cfg *config.Config, repo repositories.RepositoriesInterface) *CycleService {
	return &CycleService{cfg: cfg, repo: repo}
}

func (self *CycleService) ProcessCycles(ctx context.Context) error {
	return nil
}
