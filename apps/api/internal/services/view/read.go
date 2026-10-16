package view_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	view_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/view"
)

type ListViewsInput struct {
	WorkspaceID uuid.UUID       `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID      uuid.UUID       `query:"team_id" required:"false" format:"uuid" doc:"Only views of this team"`
	Kind        schema.ViewKind `query:"kind" required:"false" doc:"issue or project"`
}

func (self *ViewService) List(ctx context.Context, userID uuid.UUID, input *ListViewsInput) ([]*models.ViewResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	filter := view_repo.ListViewsFilter{WorkspaceID: input.WorkspaceID, ViewerID: userID}
	if input.TeamID != uuid.Nil {
		if _, err := self.teamInWorkspace(ctx, userID, input.WorkspaceID, input.TeamID); err != nil {
			return nil, err
		}
		filter.TeamID = &input.TeamID
	}
	if input.Kind != "" {
		if !validKind(input.Kind) {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "kind must be issue or project")
		}
		filter.Kind = &input.Kind
	}
	views, err := self.repo.View().ListVisible(ctx, filter)
	if err != nil {
		return nil, err
	}
	visible := make([]*models.ViewResponse, 0, len(views))
	hidden := map[uuid.UUID]bool{}
	for _, found := range views {
		if found.TeamID != nil && filter.TeamID == nil {
			hide, err := self.hiddenTeam(ctx, userID, found, hidden)
			if err != nil {
				return nil, err
			}
			if hide {
				continue
			}
		}
		visible = append(visible, models.TransformViewEntity(found, userID))
	}
	return visible, nil
}

func (self *ViewService) hiddenTeam(ctx context.Context, userID uuid.UUID, found *ent.View, known map[uuid.UUID]bool) (bool, error) {
	if hide, seen := known[*found.TeamID]; seen {
		return hide, nil
	}
	_, err := self.teamInWorkspace(ctx, userID, found.WorkspaceID, *found.TeamID)
	if err != nil && !errdefs.IsNotFound(err) {
		return false, err
	}
	known[*found.TeamID] = err != nil
	return err != nil, nil
}

type GetViewInput struct {
	ViewID uuid.UUID `query:"view_id" required:"true" format:"uuid" doc:"The view"`
}

func (self *ViewService) Get(ctx context.Context, userID uuid.UUID, input *GetViewInput) (*models.ViewResponse, error) {
	found, err := self.visibleView(ctx, userID, input.ViewID)
	if err != nil {
		return nil, err
	}
	return models.TransformViewEntity(found, userID), nil
}
