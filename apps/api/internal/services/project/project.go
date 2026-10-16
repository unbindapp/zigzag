package project_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/models"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	notification_service "github.com/unbindapp/zigzag-api/internal/services/notification"
	webhook_service "github.com/unbindapp/zigzag-api/internal/services/webhook"
)

const projectUpdatedEvent = "project.updated"

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ProjectServiceInterface -p project_service -s ProjectService -o project_service_iface.go
type ProjectService struct {
	cfg           *config.Config
	repo          repositories.RepositoriesInterface
	notifications notification_service.NotificationServiceInterface
	webhooks      webhook_service.WebhookServiceInterface
	now           func() time.Time
}

func NewProjectService(cfg *config.Config, repo repositories.RepositoriesInterface, notifications notification_service.NotificationServiceInterface, webhooks webhook_service.WebhookServiceInterface) *ProjectService {
	return &ProjectService{cfg: cfg, repo: repo, notifications: notifications, webhooks: webhooks, now: time.Now}
}

func (self *ProjectService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *ProjectService) checkVisible(ctx context.Context, userID, workspaceID uuid.UUID, notFound string) error {
	return errdefs.MaskAsNotFound(self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID), notFound)
}

func (self *ProjectService) checkRole(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID, notFound string) error {
	if err := self.checkVisible(ctx, userID, workspaceID, notFound); err != nil {
		return err
	}
	return self.check(ctx, userID, role, workspaceID)
}

func (self *ProjectService) getProject(ctx context.Context, projectID uuid.UUID) (*ent.Project, error) {
	project, err := self.repo.Project().GetByID(ctx, projectID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Project not found")
	}
	return project, err
}

func (self *ProjectService) projectWithRole(ctx context.Context, userID, projectID uuid.UUID, role schema.WorkspaceRole) (*ent.Project, error) {
	project, err := self.getProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if err := self.checkRole(ctx, userID, role, project.WorkspaceID, "Project not found"); err != nil {
		return nil, err
	}
	return project, nil
}

func (self *ProjectService) checkProjectRole(ctx context.Context, userID, projectID uuid.UUID, role schema.WorkspaceRole, notFound string) error {
	workspaceID, err := self.repo.Project().ResolveWorkspaceOfProject(ctx, projectID)
	if ent.IsNotFound(err) {
		return errdefs.NewCustomError(errdefs.ErrTypeNotFound, notFound)
	}
	if err != nil {
		return err
	}
	return self.checkRole(ctx, userID, role, workspaceID, notFound)
}

func (self *ProjectService) respond(ctx context.Context, projects []*ent.Project) ([]*models.ProjectResponse, error) {
	projectIDs := make([]uuid.UUID, len(projects))
	milestoneIDs := []uuid.UUID{}
	for i, project := range projects {
		projectIDs[i] = project.ID
		for _, milestone := range project.Edges.Milestones {
			milestoneIDs = append(milestoneIDs, milestone.ID)
		}
	}
	totals, err := self.repo.Project().IssueTotals(ctx, projectIDs)
	if err != nil {
		return nil, err
	}
	milestoneTotals, err := self.repo.Project().MilestoneIssueTotals(ctx, milestoneIDs)
	if err != nil {
		return nil, err
	}
	return models.TransformProjects(projects, totals, milestoneTotals), nil
}

func (self *ProjectService) respondOne(ctx context.Context, project *ent.Project) (*models.ProjectResponse, error) {
	responses, err := self.respond(ctx, []*ent.Project{project})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}

func (self *ProjectService) reload(ctx context.Context, projectID uuid.UUID) (*models.ProjectResponse, error) {
	project, err := self.getProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return self.respondOne(ctx, project)
}

func (self *ProjectService) notify(ctx context.Context, event models.NotificationEvent) {
	if len(event.RecipientIDs) == 0 {
		return
	}
	if err := self.notifications.Notify(ctx, event); err != nil {
		log.Error("Failed to record project notification", "type", event.Type, "err", err)
	}
}

func (self *ProjectService) emitUpdated(ctx context.Context, project *models.ProjectResponse) {
	if err := self.webhooks.Emit(ctx, project.WorkspaceID, projectUpdatedEvent, project); err != nil {
		log.Error("Failed to emit project webhook", "project_id", project.ID, "err", err)
	}
}
