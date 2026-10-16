package team_repo

import (
	"context"
	"slices"
	"time"

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

func (self *TeamRepository) CreateState(ctx context.Context, tx repository.TxInterface, input *CreateStateInput) (*ent.WorkflowState, error) {
	return self.base.Client(tx).WorkflowState.Create().
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
	updated, err := m.Save(ctx)
	if err != nil || input.Position == nil {
		return updated, err
	}
	respaced, err := self.base.RespaceIfCrowded(ctx, nil, repository.StatePosition, updated.TeamID, updated.ID, updated.Position)
	if err != nil || !respaced {
		return updated, err
	}
	return self.base.DB.WorkflowState.Get(ctx, stateID)
}

func (self *TeamRepository) CountStates(ctx context.Context, teamID uuid.UUID) (int, error) {
	return self.base.DB.WorkflowState.Query().Where(workflowstate.TeamID(teamID), workflowstate.ArchivedAtIsNil()).Count(ctx)
}

type StateMove struct {
	From    *ent.WorkflowState
	To      *ent.WorkflowState
	ActorID *uuid.UUID
	Source  schema.HistorySource
	At      time.Time
}

func (self *TeamRepository) DeleteState(ctx context.Context, move *StateMove) error {
	return self.base.WithTx(ctx, func(tx repository.TxInterface) error {
		db := tx.Client()
		moved, err := db.Issue.Query().Where(issue.StateID(move.From.ID)).IDs(ctx)
		if err != nil {
			return err
		}
		if err := moveIssueTimestamps(ctx, db, move); err != nil {
			return err
		}
		if _, err := db.Issue.Update().Where(issue.StateID(move.From.ID)).SetStateID(move.To.ID).Save(ctx); err != nil {
			return err
		}
		if err := recordStateMoves(ctx, db, moved, move); err != nil {
			return err
		}
		if _, err := db.Team.Update().Where(team.DefaultStateID(move.From.ID)).SetDefaultStateID(move.To.ID).Save(ctx); err != nil {
			return err
		}
		if _, err := db.Team.Update().Where(team.TriageStateID(move.From.ID)).ClearTriageStateID().Save(ctx); err != nil {
			return err
		}
		return db.WorkflowState.DeleteOneID(move.From.ID).Exec(ctx)
	})
}

func moveIssueTimestamps(ctx context.Context, db *ent.Client, move *StateMove) error {
	inState := issue.StateID(move.From.ID)
	if move.From.Type == schema.StateTypeTriage && move.To.Type != schema.StateTypeTriage {
		if _, err := db.Issue.Update().Where(inState).SetTriagedAt(move.At).Save(ctx); err != nil {
			return err
		}
	}
	update := db.Issue.Update().Where(inState)
	switch move.To.Type {
	case schema.StateTypeStarted:
		if _, err := db.Issue.Update().Where(inState, issue.StartedAtIsNil()).SetStartedAt(move.At).Save(ctx); err != nil {
			return err
		}
		update.ClearCompletedAt().ClearCanceledAt()
	case schema.StateTypeCompleted:
		update.SetCompletedAt(move.At).ClearCanceledAt()
	case schema.StateTypeCanceled:
		update.SetCanceledAt(move.At).ClearCompletedAt()
	default:
		update.ClearStartedAt().ClearCompletedAt().ClearCanceledAt()
	}
	_, err := update.Save(ctx)
	return err
}

func recordStateMoves(ctx context.Context, db *ent.Client, issueIDs []uuid.UUID, move *StateMove) error {
	change := schema.FieldChange{Field: "state", From: move.From.ID.String(), To: move.To.ID.String(), FromLabel: move.From.Name, ToLabel: move.To.Name}
	for batch := range slices.Chunk(issueIDs, 1000) {
		builders := make([]*ent.IssueHistoryCreate, len(batch))
		for i, issueID := range batch {
			builders[i] = db.IssueHistory.Create().
				SetIssueID(issueID).
				SetNillableActorID(move.ActorID).
				SetSource(move.Source).
				SetChanges([]schema.FieldChange{change})
		}
		if err := db.IssueHistory.CreateBulk(builders...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
