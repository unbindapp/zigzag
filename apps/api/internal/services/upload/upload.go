package upload_service

import (
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i UploadServiceInterface -p upload_service -s UploadService -o upload_service_iface.go
type UploadService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewUploadService(cfg *config.Config, repo repositories.RepositoriesInterface) *UploadService {
	return &UploadService{cfg: cfg, repo: repo}
}
