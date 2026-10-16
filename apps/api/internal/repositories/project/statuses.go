package project_repo

import (
	"context"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/project"
	"github.com/unbindapp/zigzag-api/ent/projectstatus"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

func orderStatuses(s *sql.Selector) {
	column := s.C(projectstatus.FieldType)
	expr := "CASE " + column
	for i, statusType := range schema.ProjectStatusType("").Values() {
		expr += " WHEN '" + statusType + "' THEN " + strconv.Itoa(i)
	}
	s.OrderExpr(sql.Expr(expr + " END"))
	s.OrderBy(s.C(projectstatus.FieldPosition), s.C(projectstatus.FieldCreatedAt))
}

func (self *ProjectRepository) ListStatuses(ctx context.Context, workspaceID uuid.UUID) ([]*ent.ProjectStatus, error) {
	return self.base.DB.ProjectStatus.Query().
		Where(projectstatus.WorkspaceID(workspaceID)).
		Order(orderStatuses).
		All(ctx)
}

func (self *ProjectRepository) GetStatus(ctx context.Context, statusID uuid.UUID) (*ent.ProjectStatus, error) {
	return self.base.DB.ProjectStatus.Get(ctx, statusID)
}

// DefaultStatus is the first backlog status, or the first status of any type when there is none.
func (self *ProjectRepository) DefaultStatus(ctx context.Context, workspaceID uuid.UUID) (*ent.ProjectStatus, error) {
	statuses, err := self.ListStatuses(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if len(statuses) == 0 {
		return nil, &ent.NotFoundError{}
	}
	for _, status := range statuses {
		if status.Type == schema.ProjectStatusBacklog {
			return status, nil
		}
	}
	return statuses[0], nil
}

func (self *ProjectRepository) StatusNameTaken(ctx context.Context, workspaceID uuid.UUID, name string, except *uuid.UUID) (bool, error) {
	q := self.base.DB.ProjectStatus.Query().Where(projectstatus.WorkspaceID(workspaceID), projectstatus.NameEqualFold(name))
	if except != nil {
		q.Where(projectstatus.IDNEQ(*except))
	}
	return q.Exist(ctx)
}

func (self *ProjectRepository) NextStatusPosition(ctx context.Context, workspaceID uuid.UUID, statusType schema.ProjectStatusType) (float64, error) {
	last, err := self.base.DB.ProjectStatus.Query().
		Where(projectstatus.WorkspaceID(workspaceID), projectstatus.TypeEQ(statusType)).
		Order(projectstatus.ByPosition(sql.OrderDesc())).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.Position + 1, nil
}

type CreateStatusInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Type        schema.ProjectStatusType
	Color       string
	Description string
	Position    float64
}

func (self *ProjectRepository) CreateStatus(ctx context.Context, input *CreateStatusInput) (*ent.ProjectStatus, error) {
	return self.base.DB.ProjectStatus.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetName(input.Name).
		SetType(input.Type).
		SetColor(input.Color).
		SetDescription(input.Description).
		SetPosition(input.Position).
		Save(ctx)
}

type UpdateStatusInput struct {
	Name        *string
	Color       *string
	Description *string
	Position    *float64
}

func (self *ProjectRepository) UpdateStatus(ctx context.Context, statusID uuid.UUID, input *UpdateStatusInput) (*ent.ProjectStatus, error) {
	m := self.base.DB.ProjectStatus.UpdateOneID(statusID)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.Color != nil {
		m.SetColor(*input.Color)
	}
	if input.Description != nil {
		m.SetDescription(*input.Description)
	}
	if input.Position != nil {
		m.SetPosition(*input.Position)
	}
	return m.Save(ctx)
}

func (self *ProjectRepository) DeleteStatus(ctx context.Context, tx repository.TxInterface, statusID, moveToStatusID uuid.UUID) error {
	db := self.base.Client(tx)
	if err := db.Project.Update().Where(project.StatusID(statusID)).SetStatusID(moveToStatusID).Exec(ctx); err != nil {
		return err
	}
	return db.ProjectStatus.DeleteOneID(statusID).Exec(ctx)
}
