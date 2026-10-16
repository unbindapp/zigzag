package notification_service

import (
	"context"

	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	email_service "github.com/unbindapp/zigzag-api/internal/services/email"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i NotificationServiceInterface -p notification_service -s NotificationService -o notification_service_iface.go
type NotificationService struct {
	cfg   *config.Config
	repo  repositories.RepositoriesInterface
	email email_service.EmailServiceInterface
}

func NewNotificationService(cfg *config.Config, repo repositories.RepositoriesInterface, email email_service.EmailServiceInterface) *NotificationService {
	return &NotificationService{cfg: cfg, repo: repo, email: email}
}

func (self *NotificationService) SendEmailBatch(ctx context.Context) error {
	return nil
}
