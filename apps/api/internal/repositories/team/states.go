package team_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

func (self *TeamRepository) ListStates(ctx context.Context, teamID uuid.UUID) ([]*ent.WorkflowState, error) {
	return self.base.DB.WorkflowState.Query().
		Where(workflowstate.TeamID(teamID), workflowstate.ArchivedAtIsNil()).
		Order(orderStates).
		All(ctx)
}

func (self *TeamRepository) GetState(ctx context.Context, stateID uuid.UUID) (*ent.WorkflowState, error) {
	return self.base.DB.WorkflowState.Query().Where(workflowstate.ID(stateID)).WithTeam().Only(ctx)
}

func (self *TeamRepository) StateNameTaken(ctx context.Context, teamID uuid.UUID, name string, except *uuid.UUID) (bool, error) {
	q := self.base.DB.WorkflowState.Query().Where(workflowstate.TeamID(teamID), workflowstate.NameEqualFold(name))
	if except != nil {
		q = q.Where(workflowstate.IDNEQ(*except))
	}
	return q.Exist(ctx)
}

func (self *TeamRepository) NextStatePosition(ctx context.Context, teamID uuid.UUID, stateType schema.StateType) (float64, error) {
	last, err := self.base.DB.WorkflowState.Query().
		Where(workflowstate.TeamID(teamID), workflowstate.TypeEQ(stateType)).
		Order(ent.Desc(workflowstate.FieldPosition)).
		First(ctx)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.Position + 1, nil
}

type CreateStateInput struct {
	TeamID      uuid.UUID
	Name        string
	Type        schema.StateType
	Color       string
	Description string
	Position    float64
}

func (self *TeamRepository) CreateState(ctx context.Context, input *CreateStateInput) (*ent.WorkflowState, error) {
	return self.base.DB.WorkflowState.Create().
		SetTeamID(input.TeamID).
		SetName(input.Name).
		SetType(input.Type).
		SetColor(input.Color).
		SetDescription(input.Description).
		SetPosition(input.Position).
		Save(ctx)
}

type UpdateStateInput struct {
	Name        *string
	Color       *string
	Description *string
	Position    *float64
}

func (self *TeamRepository) UpdateState(ctx context.Context, stateID uuid.UUID, input *UpdateStateInput) (*ent.WorkflowState, error) {
	m := self.base.DB.WorkflowState.UpdateOneID(stateID)
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

func (self *TeamRepository) CountStates(ctx context.Context, teamID uuid.UUID) (int, error) {
	return self.base.DB.WorkflowState.Query().Where(workflowstate.TeamID(teamID), workflowstate.ArchivedAtIsNil()).Count(ctx)
}

func (self *TeamRepository) DeleteState(ctx context.Context, stateID, moveTo uuid.UUID) error {
	return self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		db := tx.Client()
		if _, err := db.Issue.Update().Where(issue.StateID(stateID)).SetStateID(moveTo).Save(ctx); err != nil {
			return err
		}
		if _, err := db.Team.Update().Where(team.DefaultStateID(stateID)).SetDefaultStateID(moveTo).Save(ctx); err != nil {
			return err
		}
		if _, err := db.Team.Update().Where(team.TriageStateID(stateID)).ClearTriageStateID().Save(ctx); err != nil {
			return err
		}
		return db.WorkflowState.DeleteOneID(stateID).Exec(ctx)
	})
}
