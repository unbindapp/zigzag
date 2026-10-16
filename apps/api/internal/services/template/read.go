package template_service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
)

type ListTemplatesInput struct {
	WorkspaceID uuid.UUID           `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID      uuid.UUID           `query:"team_id" required:"false" format:"uuid" doc:"Also return this team's templates. Without it only workspace templates come back"`
	Kind        schema.TemplateKind `query:"kind" required:"false" doc:"issue, project or document"`
}

func (self *TemplateService) List(ctx context.Context, userID uuid.UUID, input *ListTemplatesInput) ([]*models.TemplateResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	var teamIDs []uuid.UUID
	if input.TeamID != uuid.Nil {
		team, err := self.teamInWorkspace(ctx, userID, input.WorkspaceID, input.TeamID)
		if err != nil {
			return nil, err
		}
		teamIDs = []uuid.UUID{team.ID}
	}
	var kind *schema.TemplateKind
	if input.Kind != "" {
		if !slices.Contains(input.Kind.Values(), string(input.Kind)) {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "kind must be issue, project or document")
		}
		kind = &input.Kind
	}
	templates, err := self.repo.Template().List(ctx, input.WorkspaceID, teamIDs, kind)
	if err != nil {
		return nil, err
	}
	return models.TransformTemplateEntities(templates), nil
}

type GetTemplateInput struct {
	TemplateID uuid.UUID `query:"template_id" required:"true" format:"uuid" doc:"The template"`
}

func (self *TemplateService) Get(ctx context.Context, userID uuid.UUID, input *GetTemplateInput) (*models.TemplateResponse, error) {
	found, err := self.visibleTemplate(ctx, userID, input.TemplateID)
	if err != nil {
		return nil, err
	}
	return models.TransformTemplateEntity(found), nil
}
