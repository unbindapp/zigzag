package workspace_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
)

func (self *WorkspaceService) List(ctx context.Context, userID uuid.UUID) ([]*models.WorkspaceResponse, error) {
	set, err := self.repo.Permissions().GetUserPermissionSet(ctx, userID)
	if err != nil {
		return nil, err
	}
	workspaces, err := self.repo.Workspace().GetByIDs(ctx, set.VisibleWorkspaceIDs())
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(workspaces))
	for i, workspace := range workspaces {
		ids[i] = workspace.ID
	}
	counts, err := self.repo.Workspace().Counts(ctx, ids)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.WorkspaceResponse, len(workspaces))
	for i, workspace := range workspaces {
		role, _ := set.WorkspaceRole(workspace.ID)
		c := counts[workspace.ID]
		responses[i] = models.TransformWorkspaceEntity(workspace, role, c.Members, c.Teams)
	}
	return responses, nil
}

type GetWorkspaceInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"false" format:"uuid" doc:"Workspace id; give this or slug"`
	Slug        string    `query:"slug" required:"false" maxLength:"32" doc:"Workspace URL key; give this or workspace_id"`
}

func (self *WorkspaceService) Get(ctx context.Context, userID uuid.UUID, input *GetWorkspaceInput) (*models.WorkspaceResponse, error) {
	workspace, err := self.lookup(ctx, input)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Workspace not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, workspace.ID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	return self.respond(ctx, userID, workspace)
}

func (self *WorkspaceService) lookup(ctx context.Context, input *GetWorkspaceInput) (*ent.Workspace, error) {
	switch {
	case input.WorkspaceID != uuid.Nil:
		return self.repo.Workspace().GetByID(ctx, input.WorkspaceID)
	case input.Slug != "":
		return self.repo.Workspace().GetBySlug(ctx, input.Slug)
	default:
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Give workspace_id or slug")
	}
}

func (self *WorkspaceService) respond(ctx context.Context, userID uuid.UUID, workspace *ent.Workspace) (*models.WorkspaceResponse, error) {
	set, err := self.repo.Permissions().GetUserPermissionSet(ctx, userID)
	if err != nil {
		return nil, err
	}
	counts, err := self.repo.Workspace().Counts(ctx, []uuid.UUID{workspace.ID})
	if err != nil {
		return nil, err
	}
	role, _ := set.WorkspaceRole(workspace.ID)
	c := counts[workspace.ID]
	return models.TransformWorkspaceEntity(workspace, role, c.Members, c.Teams), nil
}
