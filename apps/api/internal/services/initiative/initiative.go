package initiative_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i InitiativeServiceInterface -p initiative_service -s InitiativeService -o initiative_service_iface.go
type InitiativeService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
	now  func() time.Time
}

func NewInitiativeService(cfg *config.Config, repo repositories.RepositoriesInterface) *InitiativeService {
	return &InitiativeService{cfg: cfg, repo: repo, now: time.Now}
}

func (self *InitiativeService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *InitiativeService) checkRole(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID, notFound string) error {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID); err != nil {
		return errdefs.MaskAsNotFound(err, notFound)
	}
	if role == schema.WorkspaceRoleViewer {
		return nil
	}
	return self.check(ctx, userID, role, workspaceID)
}

func (self *InitiativeService) initiativeWithRole(ctx context.Context, userID, initiativeID uuid.UUID, role schema.WorkspaceRole) (*ent.Initiative, error) {
	initiative, err := self.repo.Initiative().GetByID(ctx, initiativeID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Initiative not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.checkRole(ctx, userID, role, initiative.WorkspaceID, "Initiative not found"); err != nil {
		return nil, err
	}
	return initiative, nil
}

func (self *InitiativeService) respond(ctx context.Context, initiatives []*ent.Initiative) ([]*models.InitiativeResponse, error) {
	ids := make([]uuid.UUID, len(initiatives))
	projectIDs := []uuid.UUID{}
	for i, initiative := range initiatives {
		ids[i] = initiative.ID
		for _, project := range initiative.Edges.Projects {
			projectIDs = append(projectIDs, project.ID)
		}
	}
	healths, err := self.repo.Initiative().LatestHealths(ctx, ids)
	if err != nil {
		return nil, err
	}
	totals, err := self.repo.Project().IssueTotals(ctx, projectIDs)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.InitiativeResponse, len(initiatives))
	for i, initiative := range initiatives {
		parts := make([]*models.IssueTotals, len(initiative.Edges.Projects))
		for j, project := range initiative.Edges.Projects {
			parts[j] = totals[project.ID]
		}
		var health *schema.Health
		if found, ok := healths[initiative.ID]; ok {
			health = &found
		}
		responses[i] = models.TransformInitiative(initiative, health, models.SumIssueTotals(parts...))
	}
	return responses, nil
}

func (self *InitiativeService) respondOne(ctx context.Context, initiativeID uuid.UUID) (*models.InitiativeResponse, error) {
	initiative, err := self.repo.Initiative().GetByID(ctx, initiativeID)
	if err != nil {
		return nil, err
	}
	responses, err := self.respond(ctx, []*ent.Initiative{initiative})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}
