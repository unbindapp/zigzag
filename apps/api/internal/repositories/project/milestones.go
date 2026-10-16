package project_repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/projectmilestone"
)

func (self *ProjectRepository) ListMilestones(ctx context.Context, projectID uuid.UUID) ([]*ent.ProjectMilestone, error) {
	return self.base.DB.ProjectMilestone.Query().
		Where(projectmilestone.ProjectID(projectID), projectmilestone.ArchivedAtIsNil()).
		Order(milestoneOrder...).
		All(ctx)
}

func (self *ProjectRepository) GetMilestone(ctx context.Context, milestoneID uuid.UUID) (*ent.ProjectMilestone, error) {
	return self.base.DB.ProjectMilestone.Get(ctx, milestoneID)
}

func (self *ProjectRepository) NextMilestoneSortOrder(ctx context.Context, projectID uuid.UUID) (float64, error) {
	last, err := self.base.DB.ProjectMilestone.Query().
		Where(projectmilestone.ProjectID(projectID)).
		Order(projectmilestone.BySortOrder(sql.OrderDesc())).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.SortOrder + 1, nil
}

type CreateMilestoneInput struct {
	ProjectID   uuid.UUID
	Name        string
	Description string
	TargetDate  *time.Time
	SortOrder   float64
}

func (self *ProjectRepository) CreateMilestone(ctx context.Context, input *CreateMilestoneInput) (*ent.ProjectMilestone, error) {
	return self.base.DB.ProjectMilestone.Create().
		SetProjectID(input.ProjectID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetNillableTargetDate(input.TargetDate).
		SetSortOrder(input.SortOrder).
		Save(ctx)
}

type UpdateMilestoneInput struct {
	Name            *string
	Description     *string
	TargetDate      *time.Time
	ClearTargetDate bool
	SortOrder       *float64
}

func (self *ProjectRepository) UpdateMilestone(ctx context.Context, milestoneID uuid.UUID, input *UpdateMilestoneInput) (*ent.ProjectMilestone, error) {
	m := self.base.DB.ProjectMilestone.UpdateOneID(milestoneID)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.Description != nil {
		m.SetDescription(*input.Description)
	}
	if input.TargetDate != nil {
		m.SetTargetDate(*input.TargetDate)
	}
	if input.ClearTargetDate {
		m.ClearTargetDate()
	}
	if input.SortOrder != nil {
		m.SetSortOrder(*input.SortOrder)
	}
	return m.Save(ctx)
}

func (self *ProjectRepository) DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	return self.base.DB.ProjectMilestone.DeleteOneID(milestoneID).Exec(ctx)
}
