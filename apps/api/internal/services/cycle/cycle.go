package cycle_service

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i CycleServiceInterface -p cycle_service -s CycleService -o cycle_service_iface.go
type CycleService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
	now  func() time.Time
}

func NewCycleService(cfg *config.Config, repo repositories.RepositoriesInterface) *CycleService {
	return &CycleService{cfg: cfg, repo: repo, now: time.Now}
}

func (self *CycleService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *CycleService) visibleTeam(ctx context.Context, userID, teamID uuid.UUID) (*ent.Team, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, team.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Team not found")
	}
	if !team.Private || self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID) == nil {
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

func (self *CycleService) adminTeam(ctx context.Context, userID, teamID uuid.UUID) (*ent.Team, error) {
	team, err := self.visibleTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID); err != nil {
		return nil, err
	}
	return team, nil
}

func (self *CycleService) visibleCycle(ctx context.Context, userID, cycleID uuid.UUID) (*ent.Cycle, *ent.Team, error) {
	found, err := self.repo.Cycle().GetByID(ctx, cycleID)
	if ent.IsNotFound(err) {
		return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Cycle not found")
	}
	if err != nil {
		return nil, nil, err
	}
	team, err := self.visibleTeam(ctx, userID, found.TeamID)
	if errdefs.IsNotFound(err) {
		return nil, nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Cycle not found")
	}
	if err != nil {
		return nil, nil, err
	}
	return found, team, nil
}

func (self *CycleService) adminCycle(ctx context.Context, userID, cycleID uuid.UUID) (*ent.Cycle, *ent.Team, error) {
	found, team, err := self.visibleCycle(ctx, userID, cycleID)
	if err != nil {
		return nil, nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID); err != nil {
		return nil, nil, err
	}
	return found, team, nil
}

func (self *CycleService) respond(ctx context.Context, team *ent.Team, cycles []*ent.Cycle) ([]*models.CycleResponse, error) {
	ids := make([]uuid.UUID, len(cycles))
	for i, found := range cycles {
		ids[i] = found.ID
	}
	tallies, err := self.repo.Cycle().Tally(ctx, nil, ids)
	if err != nil {
		return nil, err
	}
	now := self.now()
	responses := make([]*models.CycleResponse, len(cycles))
	for i, found := range cycles {
		responses[i] = models.TransformCycleEntity(found, progressOf(team, tallies[found.ID]), now)
	}
	return responses, nil
}

func (self *CycleService) respondOne(ctx context.Context, team *ent.Team, found *ent.Cycle) (*models.CycleResponse, error) {
	responses, err := self.respond(ctx, team, []*ent.Cycle{found})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}

func progressOf(team *ent.Team, tally map[schema.StateType]cycle_repo.StateTally) models.CycleProgress {
	measure := func(t cycle_repo.StateTally) float64 {
		if team.EstimateType == schema.EstimateTypeNone {
			return float64(t.Count)
		}
		return t.Points
	}
	var progress models.CycleProgress
	for stateType, t := range tally {
		if stateType == schema.StateTypeCanceled {
			continue
		}
		progress.Scope += measure(t)
		progress.IssueCount += t.Count
		switch stateType {
		case schema.StateTypeCompleted:
			progress.Completed += measure(t)
			progress.CompletedIssueCount += t.Count
		case schema.StateTypeStarted:
			progress.Started += measure(t)
		}
	}
	if progress.Scope > 0 {
		progress.Percent = math.Round(progress.Completed/progress.Scope*1000) / 10
	}
	return progress
}

func teamLocation(team *ent.Team) *time.Location {
	loc, err := time.LoadLocation(team.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func localDate(at time.Time, loc *time.Location) time.Time {
	local := at.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func cycleLabel(found *ent.Cycle) string {
	if found.Name != "" {
		return found.Name
	}
	return "Cycle " + strconv.Itoa(found.Number)
}

func historySource(ctx context.Context) schema.HistorySource {
	if _, ok := apictx.MCPCallerFromContext(ctx); ok {
		return schema.HistorySourceMCP
	}
	if _, ok := permissions_repo.APIKeyAccessFromContext(ctx); ok {
		return schema.HistorySourceAPI
	}
	return schema.HistorySourceUser
}
