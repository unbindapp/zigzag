package notification_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	email_service "github.com/zigzagapp-org/zigzag-api/internal/services/email"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i NotificationServiceInterface -p notification_service -s NotificationService -o notification_service_iface.go
type NotificationService struct {
	cfg   *config.Config
	repo  repositories.RepositoriesInterface
	email email_service.EmailServiceInterface
	now   func() time.Time
}

func NewNotificationService(cfg *config.Config, repo repositories.RepositoriesInterface, email email_service.EmailServiceInterface) *NotificationService {
	return &NotificationService{cfg: cfg, repo: repo, email: email, now: time.Now}
}

func (self *NotificationService) checkWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: workspaceID}})
}
