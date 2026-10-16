package project_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/initiative"
	"github.com/zigzagapp-org/zigzag-api/ent/label"
	"github.com/zigzagapp-org/zigzag-api/ent/team"
	"github.com/zigzagapp-org/zigzag-api/ent/workspacemember"
)

func (self *ProjectRepository) CountTeamsInWorkspace(ctx context.Context, workspaceID uuid.UUID, teamIDs []uuid.UUID) (int, error) {
	return self.base.DB.Team.Query().Where(team.WorkspaceID(workspaceID), team.IDIn(teamIDs...)).Count(ctx)
}

func (self *ProjectRepository) CountWorkspaceMembers(ctx context.Context, workspaceID uuid.UUID, userIDs []uuid.UUID) (int, error) {
	return self.base.DB.WorkspaceMember.Query().Where(workspacemember.WorkspaceID(workspaceID), workspacemember.UserIDIn(userIDs...)).Count(ctx)
}

func (self *ProjectRepository) CountApplicableLabels(ctx context.Context, workspaceID uuid.UUID, labelIDs []uuid.UUID) (int, error) {
	return self.base.DB.Label.Query().
		Where(label.WorkspaceID(workspaceID), label.TeamIDIsNil(), label.IsGroup(false), label.ArchivedAtIsNil(), label.IDIn(labelIDs...)).
		Count(ctx)
}

func (self *ProjectRepository) CountInitiativesInWorkspace(ctx context.Context, workspaceID uuid.UUID, initiativeIDs []uuid.UUID) (int, error) {
	return self.base.DB.Initiative.Query().Where(initiative.WorkspaceID(workspaceID), initiative.IDIn(initiativeIDs...)).Count(ctx)
}
