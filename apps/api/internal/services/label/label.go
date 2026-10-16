package label_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i LabelServiceInterface -p label_service -s LabelService -o label_service_iface.go
type LabelService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
	now  func() time.Time
}

func NewLabelService(cfg *config.Config, repo repositories.RepositoriesInterface) *LabelService {
	return &LabelService{cfg: cfg, repo: repo, now: time.Now}
}

func (self *LabelService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *LabelService) adminOfWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID); err != nil {
		return errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	return self.check(ctx, userID, schema.WorkspaceRoleAdmin, workspaceID)
}

func (self *LabelService) teamVisible(ctx context.Context, userID uuid.UUID, team *ent.Team) (bool, error) {
	if !team.Private || self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID) == nil {
		return true, nil
	}
	return self.repo.Team().IsMember(ctx, team.ID, userID)
}

func (self *LabelService) teamInWorkspace(ctx context.Context, userID, workspaceID, teamID uuid.UUID) (*ent.Team, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if team.WorkspaceID != workspaceID {
		return nil, notFound
	}
	visible, err := self.teamVisible(ctx, userID, team)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, notFound
	}
	return team, nil
}

func (self *LabelService) visibleLabel(ctx context.Context, userID, labelID uuid.UUID) (*ent.Label, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Label not found")
	found, err := self.repo.Label().GetByID(ctx, labelID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, found.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Label not found")
	}
	if found.TeamID == nil {
		return found, nil
	}
	if _, err := self.teamInWorkspace(ctx, userID, found.WorkspaceID, *found.TeamID); err != nil {
		if errdefs.IsNotFound(err) {
			return nil, notFound
		}
		return nil, err
	}
	return found, nil
}

func (self *LabelService) adminLabel(ctx context.Context, userID, labelID uuid.UUID) (*ent.Label, error) {
	found, err := self.visibleLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, found.WorkspaceID); err != nil {
		return nil, err
	}
	return found, nil
}
