package project_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/api/apictx"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
	project_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/project"
)

type ProjectQueryInput struct {
	ProjectID uuid.UUID `query:"project_id" required:"true" format:"uuid" doc:"The project"`
}

func (self *ProjectService) ListMilestones(ctx context.Context, userID uuid.UUID, input *ProjectQueryInput) ([]*models.MilestoneResponse, error) {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleViewer)
	if err != nil {
		return nil, err
	}
	milestones, err := self.repo.Project().ListMilestones(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	return self.respondMilestones(ctx, milestones)
}

type CreateMilestoneInput struct {
	ProjectID   uuid.UUID `json:"project_id" required:"true" format:"uuid" doc:"The project"`
	Name        string    `json:"name" required:"true" minLength:"1" maxLength:"80" doc:"Milestone name"`
	Description string    `json:"description,omitempty" required:"false" maxLength:"5000" doc:"What the milestone delivers"`
	TargetDate  string    `json:"target_date,omitempty" required:"false" format:"date" doc:"Target date, YYYY-MM-DD"`
	SortOrder   *float64  `json:"sort_order,omitempty" required:"false" doc:"Order within the project; appended last when omitted"`
}

func (self *ProjectService) CreateMilestone(ctx context.Context, userID uuid.UUID, input *CreateMilestoneInput) (*models.MilestoneResponse, error) {
	project, err := self.projectWithRole(ctx, userID, input.ProjectID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	targetDate, err := models.ParseDate("target_date", input.TargetDate)
	if err != nil {
		return nil, err
	}
	sortOrder := input.SortOrder
	if sortOrder == nil {
		next, err := self.repo.Project().NextMilestoneSortOrder(ctx, project.ID)
		if err != nil {
			return nil, err
		}
		sortOrder = &next
	}
	milestone, err := self.repo.Project().CreateMilestone(ctx, &project_repo.CreateMilestoneInput{
		ProjectID: project.ID, Name: name, Description: strings.TrimSpace(input.Description), TargetDate: targetDate, SortOrder: *sortOrder,
	})
	if err != nil {
		return nil, err
	}
	return self.respondMilestone(ctx, milestone)
}

type UpdateMilestoneInput struct {
	MilestoneID uuid.UUID `json:"milestone_id" required:"true" format:"uuid" doc:"The milestone to change"`
	Name        *string   `json:"name,omitempty" required:"false" maxLength:"80" doc:"New name"`
	Description *string   `json:"description,omitempty" required:"false" maxLength:"5000" doc:"Description, empty to clear"`
	TargetDate  *string   `json:"target_date,omitempty" required:"false" doc:"Target date YYYY-MM-DD, empty to clear"`
	SortOrder   *float64  `json:"sort_order,omitempty" required:"false" doc:"Order within the project"`
}

func (self *ProjectService) UpdateMilestone(ctx context.Context, userID uuid.UUID, input *UpdateMilestoneInput) (*models.MilestoneResponse, error) {
	milestone, err := self.memberMilestone(ctx, userID, input.MilestoneID)
	if err != nil {
		return nil, err
	}
	update := &project_repo.UpdateMilestoneInput{SortOrder: input.SortOrder}
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
	if input.TargetDate != nil {
		targetDate, err := models.ParseDate("target_date", *input.TargetDate)
		if err != nil {
			return nil, err
		}
		update.TargetDate, update.ClearTargetDate = targetDate, targetDate == nil
	}
	updated, err := self.repo.Project().UpdateMilestone(ctx, milestone.ID, update)
	if err != nil {
		return nil, err
	}
	return self.respondMilestone(ctx, updated)
}

type MilestoneRefInput struct {
	MilestoneID uuid.UUID `json:"milestone_id" required:"true" format:"uuid" doc:"The milestone to delete; its issues stay in the project without a milestone"`
}

func (self *ProjectService) DeleteMilestone(ctx context.Context, userID uuid.UUID, input *MilestoneRefInput) error {
	milestone, err := self.memberMilestone(ctx, userID, input.MilestoneID)
	if err != nil {
		return err
	}
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		issueIDs, err := self.repo.Project().IssueIDsInMilestone(ctx, tx, milestone.ID)
		if err != nil {
			return err
		}
		change := schema.FieldChange{Field: "milestone", From: milestone.ID.String(), FromLabel: milestone.Name}
		entries := make([]issue_repo.HistoryEntry, len(issueIDs))
		for i, issueID := range issueIDs {
			entries[i] = issue_repo.HistoryEntry{IssueID: issueID, ActorID: &userID, Source: apictx.HistorySource(ctx), Changes: []schema.FieldChange{change}}
		}
		if err := self.repo.Issue().CreateHistoryMany(ctx, tx, entries); err != nil {
			return err
		}
		return self.repo.Project().DeleteMilestone(ctx, tx, milestone.ID)
	})
}

func (self *ProjectService) memberMilestone(ctx context.Context, userID, milestoneID uuid.UUID) (*ent.ProjectMilestone, error) {
	milestone, err := self.repo.Project().GetMilestone(ctx, milestoneID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Milestone not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.checkProjectRole(ctx, userID, milestone.ProjectID, schema.WorkspaceRoleMember, "Milestone not found"); err != nil {
		return nil, err
	}
	return milestone, nil
}

func (self *ProjectService) respondMilestones(ctx context.Context, milestones []*ent.ProjectMilestone) ([]*models.MilestoneResponse, error) {
	ids := make([]uuid.UUID, len(milestones))
	for i, milestone := range milestones {
		ids[i] = milestone.ID
	}
	totals, err := self.repo.Project().MilestoneIssueTotals(ctx, ids)
	if err != nil {
		return nil, err
	}
	return models.TransformMilestones(milestones, totals), nil
}

func (self *ProjectService) respondMilestone(ctx context.Context, milestone *ent.ProjectMilestone) (*models.MilestoneResponse, error) {
	responses, err := self.respondMilestones(ctx, []*ent.ProjectMilestone{milestone})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}
