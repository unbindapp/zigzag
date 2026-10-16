package favorite_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i FavoriteServiceInterface -p favorite_service -s FavoriteService -o favorite_service_iface.go
type FavoriteService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewFavoriteService(cfg *config.Config, repo repositories.RepositoriesInterface) *FavoriteService {
	return &FavoriteService{cfg: cfg, repo: repo}
}
