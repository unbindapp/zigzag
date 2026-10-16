package team_service

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
	team_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/team"
)

func (self *TeamService) ListStates(ctx context.Context, userID, teamID uuid.UUID) ([]*models.WorkflowStateResponse, error) {
	team, err := self.visibleTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	states, err := self.repo.Team().ListStates(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	return models.TransformWorkflowStates(states), nil
}

type CreateStateInput struct {
	TeamID      uuid.UUID        `json:"team_id" required:"true" format:"uuid" doc:"The team"`
	Name        string           `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"State name, unique in the team"`
	Type        schema.StateType `json:"type" required:"true" doc:"triage, backlog, unstarted, started, completed or canceled"`
	Color       string           `json:"color" required:"true" maxLength:"7" pattern:"^#[0-9a-fA-F]{6}$" doc:"Hex color such as #f2c94c"`
	Description string           `json:"description,omitempty" required:"false" maxLength:"500" doc:"What the state means"`
	Position    *float64         `json:"position,omitempty" required:"false" doc:"Order within its type; appended last when omitted"`
}

func (self *TeamService) CreateState(ctx context.Context, userID uuid.UUID, input *CreateStateInput) (*models.WorkflowStateResponse, error) {
	team, err := self.adminTeam(ctx, userID, input.TeamID)
	if err != nil {
		return nil, err
	}
	if !validStateType(input.Type) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "type must be triage, backlog, unstarted, started, completed or canceled")
	}
	name, err := self.freeStateName(ctx, team.ID, input.Name, nil)
	if err != nil {
		return nil, err
	}
	color, err := names.Color(input.Color)
	if err != nil {
		return nil, err
	}
	position, err := self.statePosition(ctx, team.ID, input.Type, input.Position)
	if err != nil {
		return nil, err
	}
	state, err := self.repo.Team().CreateState(ctx, nil, &team_repo.CreateStateInput{
		TeamID: team.ID, Name: name, Type: input.Type, Color: color, Description: strings.TrimSpace(input.Description), Position: position,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformWorkflowState(state), nil
}

type UpdateStateInput struct {
	StateID     uuid.UUID `json:"state_id" required:"true" format:"uuid" doc:"The state to change"`
	Name        *string   `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Color       *string   `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^#[0-9a-fA-F]{6}$" doc:"Hex color"`
	Description *string   `json:"description,omitempty" required:"false" maxLength:"500" doc:"Description, empty to clear"`
	Position    *float64  `json:"position,omitempty" required:"false" doc:"Order within its type"`
}

func (self *TeamService) UpdateState(ctx context.Context, userID uuid.UUID, input *UpdateStateInput) (*models.WorkflowStateResponse, error) {
	state, err := self.adminState(ctx, userID, input.StateID)
	if err != nil {
		return nil, err
	}
	update := &team_repo.UpdateStateInput{Description: input.Description, Position: input.Position}
	if input.Name != nil {
		name, err := self.freeStateName(ctx, state.TeamID, *input.Name, &state.ID)
		if err != nil {
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
	updated, err := self.repo.Team().UpdateState(ctx, state.ID, update)
	if err != nil {
		return nil, err
	}
	return models.TransformWorkflowState(updated), nil
}

type DeleteStateInput struct {
	StateID             uuid.UUID `json:"state_id" required:"true" format:"uuid" doc:"The state to delete"`
	MoveIssuesToStateID uuid.UUID `json:"move_issues_to_state_id" required:"true" format:"uuid" doc:"Another state of the same team that receives the deleted state's issues"`
}

func (self *TeamService) DeleteState(ctx context.Context, userID uuid.UUID, input *DeleteStateInput) error {
	state, err := self.adminState(ctx, userID, input.StateID)
	if err != nil {
		return err
	}
	if input.MoveIssuesToStateID == state.ID {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "move_issues_to_state_id must be a different state")
	}
	target, err := self.repo.Team().GetState(ctx, input.MoveIssuesToStateID)
	if ent.IsNotFound(err) || (err == nil && target.TeamID != state.TeamID) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "move_issues_to_state_id must be a state of the same team")
	}
	if err != nil {
		return err
	}
	return self.repo.Team().DeleteState(ctx, &team_repo.StateMove{From: state, To: target, ActorID: &userID, Source: apictx.HistorySource(ctx), At: self.now()})
}

func (self *TeamService) adminState(ctx context.Context, userID, stateID uuid.UUID) (*ent.WorkflowState, error) {
	state, err := self.repo.Team().GetState(ctx, stateID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "State not found")
	}
	if err != nil {
		return nil, err
	}
	if _, err := self.adminTeam(ctx, userID, state.TeamID); err != nil {
		return nil, err
	}
	return state, nil
}

func (self *TeamService) freeStateName(ctx context.Context, teamID uuid.UUID, raw string, except *uuid.UUID) (string, error) {
	name, err := names.Clean(raw)
	if err != nil {
		return "", err
	}
	taken, err := self.repo.Team().StateNameTaken(ctx, teamID, name, except)
	if err != nil {
		return "", err
	}
	if taken {
		return "", names.EnsureFree(name, []string{name}, "state", "team")
	}
	return name, nil
}

func (self *TeamService) statePosition(ctx context.Context, teamID uuid.UUID, stateType schema.StateType, requested *float64) (float64, error) {
	if requested != nil {
		return *requested, nil
	}
	return self.repo.Team().NextStatePosition(ctx, teamID, stateType)
}

func validStateType(stateType schema.StateType) bool {
	for _, value := range stateType.Values() {
		if string(stateType) == value {
			return true
		}
	}
	return false
}
