package document_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i DocumentServiceInterface -p document_service -s DocumentService -o document_service_iface.go
type DocumentService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewDocumentService(cfg *config.Config, repo repositories.RepositoriesInterface) *DocumentService {
	return &DocumentService{cfg: cfg, repo: repo}
}
