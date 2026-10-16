package template_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i TemplateServiceInterface -p template_service -s TemplateService -o template_service_iface.go
type TemplateService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewTemplateService(cfg *config.Config, repo repositories.RepositoriesInterface) *TemplateService {
	return &TemplateService{cfg: cfg, repo: repo}
}

func (self *TemplateService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *TemplateService) teamInWorkspace(ctx context.Context, userID, workspaceID, teamID uuid.UUID) (*ent.Team, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) || (err == nil && team.WorkspaceID != workspaceID) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if !team.Private || self.check(ctx, userID, schema.WorkspaceRoleAdmin, workspaceID) == nil {
		return team, nil
	}
	member, err := self.repo.Team().IsMember(ctx, team.ID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, notFound
	}
	return team, nil
}

func (self *TemplateService) visibleTemplate(ctx context.Context, userID, templateID uuid.UUID) (*ent.Template, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Template not found")
	found, err := self.repo.Template().GetByID(ctx, templateID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, found.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Template not found")
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

func (self *TemplateService) adminTemplate(ctx context.Context, userID, templateID uuid.UUID) (*ent.Template, error) {
	found, err := self.visibleTemplate(ctx, userID, templateID)
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, found.WorkspaceID); err != nil {
		return nil, err
	}
	return found, nil
}

func (self *TemplateService) validateData(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID, data schema.TemplateData) error {
	if data.Priority != nil && (*data.Priority < 0 || *data.Priority > 4) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "data.priority must be 0 (none) through 4 (low)")
	}
	if data.StateID != nil {
		if err := self.ensureTeamState(ctx, teamID, *data.StateID); err != nil {
			return err
		}
	}
	if err := self.ensureLabels(ctx, workspaceID, teamID, data.LabelIDs); err != nil {
		return err
	}
	if data.AssigneeID != nil {
		_, err := self.repo.Workspace().GetMember(ctx, nil, workspaceID, *data.AssigneeID)
		if ent.IsNotFound(err) {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "data.assignee_id must be a member of this workspace")
		}
		if err != nil {
			return err
		}
	}
	if data.ProjectID != nil {
		found, err := self.repo.Template().ProjectInWorkspace(ctx, workspaceID, *data.ProjectID)
		if err != nil {
			return err
		}
		if !found {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "data.project_id must be a project of this workspace")
		}
	}
	for _, sub := range data.SubIssues {
		if sub.Title == "" {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Every sub-issue needs a title")
		}
	}
	return nil
}

func (self *TemplateService) ensureTeamState(ctx context.Context, teamID *uuid.UUID, stateID uuid.UUID) error {
	if teamID == nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "data.state_id needs a team template, since workflow states belong to a team")
	}
	state, err := self.repo.Team().GetState(ctx, stateID)
	if ent.IsNotFound(err) || (err == nil && state.TeamID != *teamID) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "data.state_id must be a workflow state of the template's team")
	}
	return err
}

func (self *TemplateService) ensureLabels(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID, labelIDs []uuid.UUID) error {
	labels, err := self.repo.Label().GetMany(ctx, labelIDs)
	if err != nil {
		return err
	}
	invalid := errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "data.label_ids must be workspace labels or labels of the template's team, and not label groups")
	unique := map[uuid.UUID]bool{}
	for _, id := range labelIDs {
		unique[id] = true
	}
	if len(labels) != len(unique) {
		return invalid
	}
	for _, found := range labels {
		if found.WorkspaceID != workspaceID || found.IsGroup {
			return invalid
		}
		if found.TeamID != nil && (teamID == nil || *found.TeamID != *teamID) {
			return invalid
		}
	}
	return nil
}
