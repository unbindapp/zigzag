package initiative_service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	initiative_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/initiative"
)

type ListInitiativesInput struct {
	WorkspaceID     uuid.UUID               `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Status          schema.InitiativeStatus `query:"status" required:"false" doc:"Only initiatives with this status: planned, active or completed"`
	OwnerID         uuid.UUID               `query:"owner_id" required:"false" format:"uuid" doc:"Only initiatives owned by this user"`
	IncludeArchived bool                    `query:"include_archived" required:"false" doc:"Include archived initiatives"`
	IncludeTrashed  bool                    `query:"include_trashed" required:"false" doc:"Include trashed initiatives"`
}

func (self *InitiativeService) List(ctx context.Context, userID uuid.UUID, input *ListInitiativesInput) ([]*models.InitiativeResponse, error) {
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	filter := &initiative_repo.ListInitiativesFilter{
		WorkspaceID: input.WorkspaceID, IncludeArchived: input.IncludeArchived, IncludeTrashed: input.IncludeTrashed,
	}
	if input.Status != "" {
		if err := validateStatus(input.Status); err != nil {
			return nil, err
		}
		filter.Status = &input.Status
	}
	if input.OwnerID != uuid.Nil {
		filter.OwnerID = &input.OwnerID
	}
	initiatives, err := self.repo.Initiative().List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return self.respond(ctx, initiatives)
}

type GetInitiativeInput struct {
	InitiativeID uuid.UUID `query:"initiative_id" required:"false" format:"uuid" doc:"Initiative id; give this, or workspace_id and slug"`
	WorkspaceID  uuid.UUID `query:"workspace_id" required:"false" format:"uuid" doc:"Workspace of the initiative when looking up by slug"`
	Slug         string    `query:"slug" required:"false" maxLength:"80" doc:"Initiative slug from its URL, with workspace_id"`
}

func (self *InitiativeService) Get(ctx context.Context, userID uuid.UUID, input *GetInitiativeInput) (*models.InitiativeResponse, error) {
	if input.InitiativeID != uuid.Nil {
		initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleViewer)
		if err != nil {
			return nil, err
		}
		return self.respondOne(ctx, initiative.ID)
	}
	if input.WorkspaceID == uuid.Nil || input.Slug == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Give initiative_id, or workspace_id and slug")
	}
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID, "Initiative not found"); err != nil {
		return nil, err
	}
	initiative, err := self.repo.Initiative().GetBySlug(ctx, input.WorkspaceID, input.Slug)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Initiative not found")
	}
	if err != nil {
		return nil, err
	}
	return self.respondOne(ctx, initiative.ID)
}

func validateStatus(status schema.InitiativeStatus) error {
	if slices.Contains(status.Values(), string(status)) {
		return nil
	}
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "status must be planned, active or completed")
}

func validateHealth(health schema.Health) error {
	if slices.Contains(health.Values(), string(health)) {
		return nil
	}
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "health must be on_track, at_risk or off_track")
}
