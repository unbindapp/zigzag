package import_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ImportServiceInterface -p import_service -s ImportService -o import_service_iface.go
type ImportService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
	now  func() time.Time
}

func NewImportService(cfg *config.Config, repo repositories.RepositoriesInterface) *ImportService {
	return &ImportService{cfg: cfg, repo: repo, now: time.Now}
}

func (self *ImportService) checkAdmin(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleAdmin, WorkspaceID: workspaceID}})
}

func invalid(message string) error {
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, message)
}
