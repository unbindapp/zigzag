package label_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
)

type ListLabelsInput struct {
	WorkspaceID     uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID          uuid.UUID `query:"team_id" required:"false" format:"uuid" doc:"Also return the labels of this team. Without it only workspace labels come back"`
	AllTeams        bool      `query:"all_teams" required:"false" doc:"Also return the labels of every team you can see; ignored when team_id is set"`
	IncludeArchived bool      `query:"include_archived" required:"false" doc:"Include archived labels"`
}

func (self *LabelService) List(ctx context.Context, userID uuid.UUID, input *ListLabelsInput) ([]*models.LabelTreeResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	teamIDs, err := self.teamScope(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	labels, err := self.repo.Label().List(ctx, input.WorkspaceID, teamIDs, input.IncludeArchived)
	if err != nil {
		return nil, err
	}
	return models.TransformLabelTree(labels), nil
}

func (self *LabelService) teamScope(ctx context.Context, userID uuid.UUID, input *ListLabelsInput) ([]uuid.UUID, error) {
	if input.TeamID != uuid.Nil {
		team, err := self.teamInWorkspace(ctx, userID, input.WorkspaceID, input.TeamID)
		if err != nil {
			return nil, err
		}
		return []uuid.UUID{team.ID}, nil
	}
	if !input.AllTeams {
		return nil, nil
	}
	teams, err := self.repo.Team().ListByWorkspace(ctx, input.WorkspaceID, true)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(teams))
	for _, team := range teams {
		visible, err := self.teamVisible(ctx, userID, team)
		if err != nil {
			return nil, err
		}
		if visible {
			ids = append(ids, team.ID)
		}
	}
	return ids, nil
}
