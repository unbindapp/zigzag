package view_service

import (
	"context"
	"slices"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ViewServiceInterface -p view_service -s ViewService -o view_service_iface.go
type ViewService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewViewService(cfg *config.Config, repo repositories.RepositoriesInterface) *ViewService {
	return &ViewService{cfg: cfg, repo: repo}
}

var (
	filterKeys = []string{
		"team_id", "team_ids", "state_ids", "state_types", "assignee_ids", "creator_ids", "label_ids", "project_ids",
		"milestone_ids", "cycle_ids", "parent_id", "has_parent", "priorities", "due_before", "due_after",
		"created_after", "updated_after", "query", "subscriber_id", "include_archived",
	}
	groupByValues = []string{"none", "state", "assignee", "priority", "project", "cycle", "label", "parent"}
	orderByValues = []string{"manual", "priority", "created_at", "updated_at", "due_date", "title"}
)

func (self *ViewService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *ViewService) teamVisible(ctx context.Context, userID uuid.UUID, team *ent.Team) (bool, error) {
	if !team.Private || self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID) == nil {
		return true, nil
	}
	return self.repo.Team().IsMember(ctx, team.ID, userID)
}

func (self *ViewService) teamInWorkspace(ctx context.Context, userID, workspaceID, teamID uuid.UUID) (*ent.Team, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) || (err == nil && team.WorkspaceID != workspaceID) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
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

func (self *ViewService) visibleView(ctx context.Context, userID, viewID uuid.UUID) (*ent.View, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "View not found")
	found, err := self.repo.View().GetByID(ctx, viewID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, found.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "View not found")
	}
	if found.OwnerID != userID && !found.Shared {
		return nil, notFound
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

func (self *ViewService) editableView(ctx context.Context, userID, viewID uuid.UUID) (*ent.View, error) {
	found, err := self.visibleView(ctx, userID, viewID)
	if err != nil {
		return nil, err
	}
	role := schema.WorkspaceRoleMember
	if found.OwnerID != userID {
		role = schema.WorkspaceRoleAdmin
	}
	if err := self.check(ctx, userID, role, found.WorkspaceID); err != nil {
		return nil, err
	}
	return found, nil
}

func validKind(kind schema.ViewKind) bool {
	return slices.Contains(kind.Values(), string(kind))
}

func validateFilters(filters map[string]any) error {
	unknown := make([]string, 0)
	for key := range filters {
		if !slices.Contains(filterKeys, key) {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Unknown filters: "+strings.Join(unknown, ", ")+". Use list-issues parameter names such as "+strings.Join(filterKeys[:6], ", "))
}

func validateDisplay(display schema.ViewDisplay) error {
	if display.Layout != "list" && display.Layout != "board" {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "display.layout must be list or board")
	}
	if display.GroupBy != "" && !slices.Contains(groupByValues, display.GroupBy) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "display.group_by must be one of "+strings.Join(groupByValues, ", "))
	}
	if display.OrderBy != "" && !slices.Contains(orderByValues, display.OrderBy) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "display.order_by must be one of "+strings.Join(orderByValues, ", "))
	}
	return nil
}
