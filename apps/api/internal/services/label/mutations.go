package label_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	label_repo "github.com/unbindapp/zigzag-api/internal/repositories/label"
)

type CreateLabelInput struct {
	WorkspaceID       uuid.UUID  `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	TeamID            *uuid.UUID `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Make it a label of this team only; omit for a workspace label every team shares"`
	Name              string     `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"Label name, unique among its siblings"`
	Color             string     `json:"color" required:"true" maxLength:"7" pattern:"^#[0-9a-fA-F]{6}$" doc:"Hex color such as #eb5757"`
	Description       string     `json:"description,omitempty" required:"false" maxLength:"500" doc:"What the label means"`
	ParentID          *uuid.UUID `json:"parent_id,omitempty" required:"false" format:"uuid" doc:"Put the label in this group; the group must have the same team"`
	IsGroup           bool       `json:"is_group,omitempty" required:"false" doc:"Create a label group that holds other labels; groups cannot be applied to issues"`
	GroupSingleSelect bool       `json:"group_single_select,omitempty" required:"false" doc:"For groups: an issue may carry only one label of the group"`
}

func (self *LabelService) Create(ctx context.Context, userID uuid.UUID, input *CreateLabelInput) (*models.LabelResponse, error) {
	if err := self.adminOfWorkspace(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	if input.TeamID != nil {
		if _, err := self.teamInWorkspace(ctx, userID, input.WorkspaceID, *input.TeamID); err != nil {
			return nil, err
		}
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	color, err := names.Color(input.Color)
	if err != nil {
		return nil, err
	}
	if input.GroupSingleSelect && !input.IsGroup {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "group_single_select applies to label groups only")
	}
	if input.ParentID != nil {
		if input.IsGroup {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A label group cannot be inside another group")
		}
		if err := self.ensureGroup(ctx, input.WorkspaceID, input.TeamID, *input.ParentID); err != nil {
			return nil, err
		}
	}

	var created *ent.Label
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.ensureNameFree(ctx, tx, input.WorkspaceID, input.TeamID, input.ParentID, name, nil); err != nil {
			return err
		}
		sortOrder, err := self.repo.Label().NextSortOrder(ctx, tx, input.WorkspaceID, input.TeamID, input.ParentID)
		if err != nil {
			return err
		}
		created, err = self.repo.Label().Create(ctx, tx, &label_repo.CreateLabelInput{
			WorkspaceID:       input.WorkspaceID,
			TeamID:            input.TeamID,
			Name:              name,
			Color:             color,
			Description:       strings.TrimSpace(input.Description),
			ParentID:          input.ParentID,
			IsGroup:           input.IsGroup,
			GroupSingleSelect: input.GroupSingleSelect,
			SortOrder:         sortOrder,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return models.TransformLabelEntity(created), nil
}

type UpdateLabelInput struct {
	LabelID           uuid.UUID  `json:"label_id" required:"true" format:"uuid" doc:"The label to change"`
	Name              *string    `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Color             *string    `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^#[0-9a-fA-F]{6}$" doc:"Hex color"`
	Description       *string    `json:"description,omitempty" required:"false" maxLength:"500" doc:"Description, empty to clear"`
	ParentID          *uuid.UUID `json:"parent_id,omitempty" required:"false" format:"uuid" doc:"Move the label into this group; the nil uuid 00000000-0000-0000-0000-000000000000 takes it out of its group"`
	GroupSingleSelect *bool      `json:"group_single_select,omitempty" required:"false" doc:"For groups: an issue may carry only one label of the group"`
	SortOrder         *float64   `json:"sort_order,omitempty" required:"false" doc:"Position among its siblings, lower first"`
}

func (self *LabelService) Update(ctx context.Context, userID uuid.UUID, input *UpdateLabelInput) (*models.LabelResponse, error) {
	current, err := self.adminLabel(ctx, userID, input.LabelID)
	if err != nil {
		return nil, err
	}
	update := &label_repo.UpdateLabelInput{GroupSingleSelect: input.GroupSingleSelect, SortOrder: input.SortOrder}
	name := current.Name
	if input.Name != nil {
		if name, err = names.Clean(*input.Name); err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Color != nil {
		color, err := names.Color(*input.Color)
		if err != nil {
			return nil, err
		}
		update.Color = &color
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		update.Description = &description
	}
	if input.GroupSingleSelect != nil && !current.IsGroup {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "group_single_select applies to label groups only")
	}
	parentID := current.ParentID
	if input.ParentID != nil {
		parentID, err = self.resolveNewParent(ctx, current, *input.ParentID)
		if err != nil {
			return nil, err
		}
		update.ParentID = parentID
		update.ClearParent = parentID == nil
	}

	var updated *ent.Label
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.ensureNameFree(ctx, tx, current.WorkspaceID, current.TeamID, parentID, name, &current.ID); err != nil {
			return err
		}
		updated, err = self.repo.Label().Update(ctx, tx, current.ID, update)
		return err
	})
	if err != nil {
		return nil, err
	}
	return models.TransformLabelEntity(updated), nil
}

type LabelRefInput struct {
	LabelID uuid.UUID `json:"label_id" required:"true" format:"uuid" doc:"The label"`
}

func (self *LabelService) Archive(ctx context.Context, userID uuid.UUID, input *LabelRefInput) (*models.LabelResponse, error) {
	current, err := self.adminLabel(ctx, userID, input.LabelID)
	if err != nil {
		return nil, err
	}
	now := self.now()
	updated, err := self.repo.Label().SetArchived(ctx, current.ID, &now)
	if err != nil {
		return nil, err
	}
	return models.TransformLabelEntity(updated), nil
}

func (self *LabelService) Unarchive(ctx context.Context, userID uuid.UUID, input *LabelRefInput) (*models.LabelResponse, error) {
	current, err := self.adminLabel(ctx, userID, input.LabelID)
	if err != nil {
		return nil, err
	}
	updated, err := self.repo.Label().SetArchived(ctx, current.ID, nil)
	if err != nil {
		return nil, err
	}
	return models.TransformLabelEntity(updated), nil
}

func (self *LabelService) Delete(ctx context.Context, userID uuid.UUID, input *LabelRefInput) error {
	current, err := self.adminLabel(ctx, userID, input.LabelID)
	if err != nil {
		return err
	}
	ids := []uuid.UUID{current.ID}
	if current.IsGroup {
		children, err := self.repo.Label().Children(ctx, current.ID)
		if err != nil {
			return err
		}
		for _, child := range children {
			ids = append(ids, child.ID)
		}
	}
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		byIssue, err := self.repo.Label().IssuesWithLabels(ctx, tx, ids)
		if err != nil {
			return err
		}
		if err := self.repo.Label().RecordHistory(ctx, tx, removalHistory(ctx, userID, byIssue)); err != nil {
			return err
		}
		return self.repo.Label().Delete(ctx, tx, ids)
	})
}

func removalHistory(ctx context.Context, userID uuid.UUID, byIssue map[uuid.UUID][]*ent.Label) []label_repo.HistoryEntry {
	source := apictx.HistorySource(ctx)
	entries := make([]label_repo.HistoryEntry, 0, len(byIssue))
	for issueID, labels := range byIssue {
		ids := make([]string, len(labels))
		labelNames := make([]string, len(labels))
		for i, removed := range labels {
			ids[i] = removed.ID.String()
			labelNames[i] = removed.Name
		}
		entries = append(entries, label_repo.HistoryEntry{
			IssueID: issueID,
			ActorID: &userID,
			Source:  source,
			Changes: []schema.FieldChange{{Field: "labels_removed", From: ids, FromLabel: strings.Join(labelNames, ", ")}},
		})
	}
	return entries
}

func (self *LabelService) resolveNewParent(ctx context.Context, current *ent.Label, requested uuid.UUID) (*uuid.UUID, error) {
	if requested == uuid.Nil {
		return nil, nil
	}
	if current.IsGroup {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A label group cannot be inside another group")
	}
	if requested == current.ID {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A label cannot be its own group")
	}
	if err := self.ensureGroup(ctx, current.WorkspaceID, current.TeamID, requested); err != nil {
		return nil, err
	}
	return &requested, nil
}

func (self *LabelService) ensureGroup(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID, parentID uuid.UUID) error {
	parent, err := self.repo.Label().GetByID(ctx, parentID)
	if ent.IsNotFound(err) || (err == nil && parent.WorkspaceID != workspaceID) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "parent_id must be a label group of this workspace")
	}
	if err != nil {
		return err
	}
	if !parent.IsGroup {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "parent_id must be a label group, not a plain label")
	}
	if !sameTeam(parent.TeamID, teamID) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "A label and its group must both be workspace labels or belong to the same team")
	}
	return nil
}

func (self *LabelService) ensureNameFree(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, teamID, parentID *uuid.UUID, name string, except *uuid.UUID) error {
	taken, err := self.repo.Label().NameTaken(ctx, tx, workspaceID, teamID, parentID, name, except)
	if err != nil {
		return err
	}
	if !taken {
		return nil
	}
	scope := "workspace"
	switch {
	case parentID != nil:
		scope = "group"
	case teamID != nil:
		scope = "team"
	}
	return names.EnsureFree(name, []string{name}, "label", scope)
}

func sameTeam(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
