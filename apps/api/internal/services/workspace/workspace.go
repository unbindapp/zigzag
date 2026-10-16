package workspace_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	email_service "github.com/unbindapp/zigzag-api/internal/services/email"
)

const invitationTTL = 7 * 24 * time.Hour

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i WorkspaceServiceInterface -p workspace_service -s WorkspaceService -o workspace_service_iface.go
type WorkspaceService struct {
	cfg   *config.Config
	repo  repositories.RepositoriesInterface
	email email_service.EmailServiceInterface
	now   func() time.Time
}

func NewWorkspaceService(cfg *config.Config, repo repositories.RepositoriesInterface, email email_service.EmailServiceInterface) *WorkspaceService {
	return &WorkspaceService{cfg: cfg, repo: repo, email: email, now: time.Now}
}

func (self *WorkspaceService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *WorkspaceService) checkVisible(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID); err != nil {
		return errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	return self.check(ctx, userID, role, workspaceID)
}
