package webhook_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i WebhookServiceInterface -p webhook_service -s WebhookService -o webhook_service_iface.go
type WebhookService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewWebhookService(cfg *config.Config, repo repositories.RepositoriesInterface) *WebhookService {
	return &WebhookService{cfg: cfg, repo: repo}
}
