package team_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
)

type ListTeamsInput struct {
	WorkspaceID     uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	IncludeArchived bool      `query:"include_archived" required:"false" doc:"Include archived teams"`
}

func (self *TeamService) List(ctx context.Context, userID uuid.UUID, input *ListTeamsInput) ([]*models.TeamResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	teams, err := self.repo.Team().ListByWorkspace(ctx, input.WorkspaceID, input.IncludeArchived)
	if err != nil {
		return nil, err
	}
	visible := make([]*ent.Team, 0, len(teams))
	for _, team := range teams {
		err := self.ensureVisible(ctx, userID, team)
		if errdefs.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		visible = append(visible, team)
	}
	return self.respond(ctx, userID, visible)
}

type GetTeamInput struct {
	TeamID      uuid.UUID `query:"team_id" required:"false" format:"uuid" doc:"Team id; give this, or workspace_id and key"`
	WorkspaceID uuid.UUID `query:"workspace_id" required:"false" format:"uuid" doc:"Workspace of the team when looking up by key"`
	Key         string    `query:"key" required:"false" maxLength:"7" doc:"Team key such as ENG, with workspace_id"`
}

func (self *TeamService) Get(ctx context.Context, userID uuid.UUID, input *GetTeamInput) (*models.TeamResponse, error) {
	teamID := input.TeamID
	if teamID == uuid.Nil {
		if input.WorkspaceID == uuid.Nil || input.Key == "" {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Give team_id, or workspace_id and key")
		}
		if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
			return nil, errdefs.MaskAsNotFound(err, "Team not found")
		}
		team, err := self.repo.Team().GetByKey(ctx, input.WorkspaceID, input.Key)
		if err != nil {
			return nil, err
		}
		teamID = team.ID
	}
	team, err := self.visibleTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	return self.respondOne(ctx, userID, team)
}
