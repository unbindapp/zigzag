package template_service

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	template_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/template"
)

type CreateTemplateInput struct {
	WorkspaceID uuid.UUID           `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID      *uuid.UUID          `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Make it a template of this team; needed when data.state_id is set"`
	Kind        schema.TemplateKind `json:"kind" required:"true" doc:"issue, project or document"`
	Name        string              `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"Template name shown in the picker"`
	Description string              `json:"description,omitempty" required:"false" maxLength:"500" doc:"When to use the template"`
	Data        schema.TemplateData `json:"data" required:"true" doc:"Values a new issue starts with: title, description, priority, estimate, label_ids, assignee_id, state_id, project_id, sub_issues"`
}

func (self *TemplateService) Create(ctx context.Context, userID uuid.UUID, input *CreateTemplateInput) (*models.TemplateResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return nil, err
	}
	if input.TeamID != nil {
		if _, err := self.teamInWorkspace(ctx, userID, input.WorkspaceID, *input.TeamID); err != nil {
			return nil, err
		}
	}
	if !slices.Contains(input.Kind.Values(), string(input.Kind)) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "kind must be issue, project or document")
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	if err := self.validateData(ctx, input.WorkspaceID, input.TeamID, input.Data); err != nil {
		return nil, err
	}

	var created *ent.Template
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		sortOrder, err := self.repo.Template().NextSortOrder(ctx, tx, input.WorkspaceID)
		if err != nil {
			return err
		}
		created, err = self.repo.Template().Create(ctx, tx, &template_repo.CreateTemplateInput{
			WorkspaceID: input.WorkspaceID,
			TeamID:      input.TeamID,
			Kind:        input.Kind,
			Name:        name,
			Description: strings.TrimSpace(input.Description),
			Data:        input.Data,
			SortOrder:   sortOrder,
			CreatorID:   &userID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return models.TransformTemplateEntity(created), nil
}

type UpdateTemplateInput struct {
	TemplateID  uuid.UUID            `json:"template_id" required:"true" format:"uuid" doc:"The template to change"`
	Name        *string              `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Description *string              `json:"description,omitempty" required:"false" maxLength:"500" doc:"Description, empty to clear"`
	Data        *schema.TemplateData `json:"data,omitempty" required:"false" doc:"Replaces all template values"`
	SortOrder   *float64             `json:"sort_order,omitempty" required:"false" doc:"Position in the picker, lower first"`
}

func (self *TemplateService) Update(ctx context.Context, userID uuid.UUID, input *UpdateTemplateInput) (*models.TemplateResponse, error) {
	current, err := self.adminTemplate(ctx, userID, input.TemplateID)
	if err != nil {
		return nil, err
	}
	update := &template_repo.UpdateTemplateInput{Data: input.Data, SortOrder: input.SortOrder}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		update.Description = &description
	}
	if input.Data != nil {
		if err := self.validateData(ctx, current.WorkspaceID, current.TeamID, *input.Data); err != nil {
			return nil, err
		}
	}
	updated, err := self.repo.Template().Update(ctx, current.ID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformTemplateEntity(updated), nil
}

type DeleteTemplateInput struct {
	TemplateID uuid.UUID `json:"template_id" required:"true" format:"uuid" doc:"The template to delete; issues created from it are not touched"`
}

func (self *TemplateService) Delete(ctx context.Context, userID uuid.UUID, input *DeleteTemplateInput) error {
	current, err := self.adminTemplate(ctx, userID, input.TemplateID)
	if err != nil {
		return err
	}
	return self.repo.Template().Delete(ctx, current.ID)
}
