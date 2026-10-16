package team_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	cycle_service "github.com/unbindapp/zigzag-api/internal/services/cycle"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i TeamServiceInterface -p team_service -s TeamService -o team_service_iface.go
type TeamService struct {
	repo   repositories.RepositoriesInterface
	cycles cycle_service.CycleServiceInterface
	now    func() time.Time
}

func NewTeamService(repo repositories.RepositoriesInterface, cycles cycle_service.CycleServiceInterface) *TeamService {
	return &TeamService{repo: repo, cycles: cycles, now: time.Now}
}

func (self *TeamService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *TeamService) visibleTeam(ctx context.Context, userID, teamID uuid.UUID) (*ent.Team, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	team, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if err := self.ensureVisible(ctx, userID, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (self *TeamService) ensureVisible(ctx context.Context, userID uuid.UUID, team *ent.Team) error {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, team.WorkspaceID); err != nil {
		return errdefs.MaskAsNotFound(err, "Team not found")
	}
	if !team.Private {
		return nil
	}
	if self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID) == nil {
		return nil
	}
	member, err := self.repo.Team().IsMember(ctx, team.ID, userID)
	if err != nil {
		return err
	}
	if !member {
		return notFound
	}
	return nil
}

func (self *TeamService) adminTeam(ctx context.Context, userID, teamID uuid.UUID) (*ent.Team, error) {
	team, err := self.visibleTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, team.WorkspaceID); err != nil {
		return nil, err
	}
	return team, nil
}

func (self *TeamService) respond(ctx context.Context, userID uuid.UUID, teams []*ent.Team) ([]*models.TeamResponse, error) {
	ids := make([]uuid.UUID, len(teams))
	for i, team := range teams {
		ids[i] = team.ID
	}
	counts, err := self.repo.Team().MemberCounts(ctx, ids)
	if err != nil {
		return nil, err
	}
	active, err := self.repo.Cycle().ActiveCycleIDs(ctx, ids, self.now())
	if err != nil {
		return nil, err
	}
	responses := make([]*models.TeamResponse, len(teams))
	for i, team := range teams {
		member, err := self.repo.Team().IsMember(ctx, team.ID, userID)
		if err != nil {
			return nil, err
		}
		var activeCycle *uuid.UUID
		if id, ok := active[team.ID]; ok {
			activeCycle = &id
		}
		responses[i] = models.TransformTeamEntity(team, counts[team.ID], member, activeCycle)
	}
	return responses, nil
}

func (self *TeamService) respondOne(ctx context.Context, userID uuid.UUID, team *ent.Team) (*models.TeamResponse, error) {
	responses, err := self.respond(ctx, userID, []*ent.Team{team})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}
