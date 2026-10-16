package initiative_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	initiative_repo "github.com/unbindapp/zigzag-api/internal/repositories/initiative"
)

type InitiativeQueryInput struct {
	InitiativeID uuid.UUID `query:"initiative_id" required:"true" format:"uuid" doc:"The initiative"`
}

func (self *InitiativeService) ListUpdates(ctx context.Context, userID uuid.UUID, input *InitiativeQueryInput) ([]*models.InitiativeUpdateResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleViewer)
	if err != nil {
		return nil, err
	}
	updates, err := self.repo.Initiative().ListUpdates(ctx, initiative.ID)
	if err != nil {
		return nil, err
	}
	return models.TransformInitiativeUpdates(updates), nil
}

type CreateInitiativeUpdateInput struct {
	InitiativeID uuid.UUID     `json:"initiative_id" required:"true" format:"uuid" doc:"The initiative"`
	Body         string        `json:"body" required:"true" minLength:"1" maxLength:"50000" doc:"Markdown text of the update"`
	Health       schema.Health `json:"health" required:"true" doc:"on_track, at_risk or off_track; becomes the initiative's health"`
}

func (self *InitiativeService) CreateUpdate(ctx context.Context, userID uuid.UUID, input *CreateInitiativeUpdateInput) (*models.InitiativeUpdateResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	body, err := trimmedBody(input.Body)
	if err != nil {
		return nil, err
	}
	if err := validateHealth(input.Health); err != nil {
		return nil, err
	}
	created, err := self.repo.Initiative().CreateUpdate(ctx, &initiative_repo.CreateUpdateInput{
		InitiativeID: initiative.ID, AuthorID: userID, Body: body, Health: input.Health,
	})
	if err != nil {
		return nil, err
	}
	update, err := self.repo.Initiative().GetUpdate(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	return models.TransformInitiativeUpdate(update), nil
}

type InitiativeUpdateRefInput struct {
	UpdateID uuid.UUID `json:"update_id" required:"true" format:"uuid" doc:"The initiative update to delete"`
}

func (self *InitiativeService) DeleteUpdate(ctx context.Context, userID uuid.UUID, input *InitiativeUpdateRefInput) error {
	update, err := self.repo.Initiative().GetUpdate(ctx, input.UpdateID)
	if ent.IsNotFound(err) {
		return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Initiative update not found")
	}
	if err != nil {
		return err
	}
	workspaceID, err := self.repo.Initiative().ResolveWorkspaceOfInitiative(ctx, update.InitiativeID)
	if err != nil {
		return err
	}
	role := schema.WorkspaceRoleAdmin
	if update.AuthorID != nil && *update.AuthorID == userID {
		role = schema.WorkspaceRoleMember
	}
	if err := self.checkRole(ctx, userID, role, workspaceID, "Initiative update not found"); err != nil {
		return err
	}
	return self.repo.Initiative().DeleteUpdate(ctx, update.ID)
}
