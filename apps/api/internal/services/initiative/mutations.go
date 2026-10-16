package initiative_service

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	initiative_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/initiative"
)

const slugBaseLength = 48

type CreateInitiativeInput struct {
	WorkspaceID uuid.UUID               `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Name        string                  `json:"name" required:"true" minLength:"1" maxLength:"80" doc:"Initiative name"`
	Description string                  `json:"description,omitempty" required:"false" maxLength:"100000" doc:"Markdown description"`
	Icon        string                  `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name"`
	Color       string                  `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color such as #5e6ad2"`
	OwnerID     uuid.UUID               `json:"owner_id,omitempty" required:"false" format:"uuid" doc:"Workspace member who owns the initiative"`
	Status      schema.InitiativeStatus `json:"status,omitempty" required:"false" doc:"planned (default), active or completed"`
	TargetDate  string                  `json:"target_date,omitempty" required:"false" format:"date" doc:"Target date, YYYY-MM-DD"`
	ProjectIDs  []uuid.UUID             `json:"project_ids,omitempty" format:"uuid" required:"false" doc:"Projects of this workspace that belong to the initiative"`
}

func (self *InitiativeService) Create(ctx context.Context, userID uuid.UUID, input *CreateInitiativeInput) (*models.InitiativeResponse, error) {
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleMember, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	color, err := names.OptionalColor(input.Color)
	if err != nil {
		return nil, err
	}
	if input.Status != "" {
		if err := validateStatus(input.Status); err != nil {
			return nil, err
		}
	}
	targetDate, err := models.ParseDate("target_date", input.TargetDate)
	if err != nil {
		return nil, err
	}
	var ownerID *uuid.UUID
	if input.OwnerID != uuid.Nil {
		ownerID = &input.OwnerID
	}
	projectIDs := uniqueIDs(input.ProjectIDs)
	if err := self.validateReferences(ctx, input.WorkspaceID, ownerID, projectIDs); err != nil {
		return nil, err
	}
	slug, err := utils.SlugWithSuffix(utils.TruncateSlug(utils.Slugify(name), slugBaseLength), 4)
	if err != nil {
		return nil, err
	}
	sortOrder, err := self.repo.Initiative().MaxSortOrder(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	var created *ent.Initiative
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		created, err = self.repo.Initiative().Create(ctx, tx, &initiative_repo.CreateInitiativeInput{
			WorkspaceID: input.WorkspaceID, Name: name, Slug: slug, Description: input.Description, Icon: input.Icon,
			Color: color, OwnerID: ownerID, Status: input.Status, TargetDate: targetDate, SortOrder: sortOrder + 1, ProjectIDs: projectIDs,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return self.respondOne(ctx, created.ID)
}

type UpdateInitiativeInput struct {
	InitiativeID uuid.UUID                `json:"initiative_id" required:"true" format:"uuid" doc:"The initiative to change"`
	Name         *string                  `json:"name,omitempty" required:"false" maxLength:"80" doc:"New name"`
	Description  *string                  `json:"description,omitempty" required:"false" maxLength:"100000" doc:"Markdown description, empty to clear"`
	Icon         *string                  `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	Color        *string                  `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color, empty to clear"`
	OwnerID      *uuid.UUID               `json:"owner_id,omitempty" required:"false" format:"uuid" doc:"New owner; the nil uuid 00000000-0000-0000-0000-000000000000 removes the owner"`
	Status       *schema.InitiativeStatus `json:"status,omitempty" required:"false" doc:"planned, active or completed"`
	TargetDate   *string                  `json:"target_date,omitempty" required:"false" doc:"Target date YYYY-MM-DD, empty to clear"`
	SortOrder    *float64                 `json:"sort_order,omitempty" required:"false" doc:"Manual order in lists"`
	ProjectIDs   []uuid.UUID              `json:"project_ids,omitempty" format:"uuid" required:"false" doc:"Replaces the projects; an empty list removes all"`
}

func (self *InitiativeService) Update(ctx context.Context, userID uuid.UUID, input *UpdateInitiativeInput) (*models.InitiativeResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	update := &initiative_repo.UpdateInitiativeInput{Description: input.Description, Icon: input.Icon, SortOrder: input.SortOrder}
	if input.Name != nil {
		name, err := names.Clean(*input.Name)
		if err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Color != nil {
		color, err := names.OptionalColor(*input.Color)
		if err != nil {
			return nil, err
		}
		update.Color = &color
	}
	if input.Status != nil {
		if err := validateStatus(*input.Status); err != nil {
			return nil, err
		}
		update.Status = input.Status
	}
	if input.TargetDate != nil {
		targetDate, err := models.ParseDate("target_date", *input.TargetDate)
		if err != nil {
			return nil, err
		}
		update.TargetDate, update.ClearTargetDate = targetDate, targetDate == nil
	}
	var ownerID *uuid.UUID
	if input.OwnerID != nil {
		update.ClearOwner = *input.OwnerID == uuid.Nil
		if !update.ClearOwner {
			ownerID, update.OwnerID = input.OwnerID, input.OwnerID
		}
	}
	if input.ProjectIDs != nil {
		update.ReplaceProjects, update.ProjectIDs = true, uniqueIDs(input.ProjectIDs)
	}
	if err := self.validateReferences(ctx, initiative.WorkspaceID, ownerID, update.ProjectIDs); err != nil {
		return nil, err
	}
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		return self.repo.Initiative().Update(ctx, tx, initiative.ID, update)
	})
	if err != nil {
		return nil, err
	}
	return self.respondOne(ctx, initiative.ID)
}

type InitiativeRefInput struct {
	InitiativeID uuid.UUID `json:"initiative_id" required:"true" format:"uuid" doc:"The initiative"`
}

func (self *InitiativeService) Delete(ctx context.Context, userID uuid.UUID, input *InitiativeRefInput) error {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleAdmin)
	if err != nil {
		return err
	}
	return self.repo.Initiative().Delete(ctx, initiative.ID)
}

func (self *InitiativeService) Archive(ctx context.Context, userID uuid.UUID, input *InitiativeRefInput) (*models.InitiativeResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	now := self.now()
	if err := self.repo.Initiative().SetArchivedAt(ctx, initiative.ID, &now); err != nil {
		return nil, err
	}
	return self.respondOne(ctx, initiative.ID)
}

func (self *InitiativeService) Unarchive(ctx context.Context, userID uuid.UUID, input *InitiativeRefInput) (*models.InitiativeResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	if err := self.repo.Initiative().SetArchivedAt(ctx, initiative.ID, nil); err != nil {
		return nil, err
	}
	return self.respondOne(ctx, initiative.ID)
}

type InitiativeProjectInput struct {
	InitiativeID uuid.UUID `json:"initiative_id" required:"true" format:"uuid" doc:"The initiative"`
	ProjectID    uuid.UUID `json:"project_id" required:"true" format:"uuid" doc:"A project of the same workspace"`
}

func (self *InitiativeService) AddProject(ctx context.Context, userID uuid.UUID, input *InitiativeProjectInput) (*models.InitiativeResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	if err := self.validateReferences(ctx, initiative.WorkspaceID, nil, []uuid.UUID{input.ProjectID}); err != nil {
		return nil, err
	}
	if err := self.repo.Initiative().AddProject(ctx, initiative.ID, input.ProjectID); err != nil {
		return nil, err
	}
	return self.respondOne(ctx, initiative.ID)
}

func (self *InitiativeService) RemoveProject(ctx context.Context, userID uuid.UUID, input *InitiativeProjectInput) (*models.InitiativeResponse, error) {
	initiative, err := self.initiativeWithRole(ctx, userID, input.InitiativeID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	if err := self.repo.Initiative().RemoveProject(ctx, initiative.ID, input.ProjectID); err != nil {
		return nil, err
	}
	return self.respondOne(ctx, initiative.ID)
}

func (self *InitiativeService) validateReferences(ctx context.Context, workspaceID uuid.UUID, ownerID *uuid.UUID, projectIDs []uuid.UUID) error {
	if ownerID != nil {
		found, err := self.repo.Project().CountWorkspaceMembers(ctx, workspaceID, []uuid.UUID{*ownerID})
		if err != nil {
			return err
		}
		if found != 1 {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The owner must be a member of this workspace")
		}
	}
	if len(projectIDs) == 0 {
		return nil
	}
	found, err := self.repo.Initiative().CountProjectsInWorkspace(ctx, workspaceID, projectIDs)
	if err != nil {
		return err
	}
	if found != len(projectIDs) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Every project must belong to this workspace")
	}
	return nil
}

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	unique := []uuid.UUID{}
	for _, id := range ids {
		if !slices.Contains(unique, id) {
			unique = append(unique, id)
		}
	}
	return unique
}

func trimmedBody(raw string) (string, error) {
	body := strings.TrimSpace(raw)
	if body == "" {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "An update needs a body")
	}
	return body, nil
}
