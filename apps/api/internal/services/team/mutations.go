package team_service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
)

type CreateTeamInput struct {
	WorkspaceID uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Name        string    `json:"name" required:"true" minLength:"1" maxLength:"64" doc:"Team name, unique in the workspace"`
	Key         string    `json:"key" required:"true" minLength:"1" maxLength:"7" pattern:"^[A-Za-z][A-Za-z0-9]{0,6}$" doc:"Issue prefix, 1 to 7 uppercase letters or digits starting with a letter, unique in the workspace"`
	Description string    `json:"description,omitempty" required:"false" maxLength:"500" doc:"What the team works on"`
	Icon        string    `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name"`
	Color       string    `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color such as #5e6ad2"`
	Private     bool      `json:"private,omitempty" required:"false" doc:"Private teams are visible to their members and workspace admins only"`
	Timezone    string    `json:"timezone,omitempty" required:"false" maxLength:"64" doc:"IANA time zone for cycles, defaults to UTC"`
}

func (self *TeamService) Create(ctx context.Context, userID uuid.UUID, input *CreateTeamInput) (*models.TeamResponse, error) {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleAdmin, input.WorkspaceID); err != nil {
		return nil, err
	}
	name, err := names.Clean(input.Name)
	if err != nil {
		return nil, err
	}
	key, err := names.TeamKey(input.Key)
	if err != nil {
		return nil, err
	}
	if err := validateTimezone(input.Timezone); err != nil {
		return nil, err
	}
	color, err := names.OptionalColor(input.Color)
	if err != nil {
		return nil, err
	}

	var created *ent.Team
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.ensureFree(ctx, tx, input.WorkspaceID, key, name, nil); err != nil {
			return err
		}
		created, err = self.repo.Team().Create(ctx, tx, &team_repo.CreateTeamInput{
			WorkspaceID: input.WorkspaceID,
			Name:        name,
			Key:         key,
			Description: strings.TrimSpace(input.Description),
			Icon:        input.Icon,
			Color:       color,
			Private:     input.Private,
			Timezone:    input.Timezone,
			CreatorID:   userID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return self.respondOne(ctx, userID, created)
}

type UpdateTeamInput struct {
	TeamID              uuid.UUID            `json:"team_id" required:"true" format:"uuid" doc:"The team to change"`
	Name                *string              `json:"name,omitempty" required:"false" maxLength:"64" doc:"New name"`
	Key                 *string              `json:"key,omitempty" required:"false" maxLength:"7" pattern:"^[A-Za-z][A-Za-z0-9]{0,6}$" doc:"New issue prefix; existing identifiers change with it"`
	Description         *string              `json:"description,omitempty" required:"false" maxLength:"500" doc:"Description, empty to clear"`
	Icon                *string              `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	Color               *string              `json:"color,omitempty" required:"false" maxLength:"7" pattern:"^(#[0-9a-fA-F]{6})?$" doc:"Hex color, empty to clear"`
	Private             *bool                `json:"private,omitempty" required:"false" doc:"Visible to members and workspace admins only"`
	Timezone            *string              `json:"timezone,omitempty" required:"false" maxLength:"64" doc:"IANA time zone"`
	CyclesEnabled       *bool                `json:"cycles_enabled,omitempty" required:"false" doc:"Turn cycles on or off"`
	CycleDurationWeeks  *int                 `json:"cycle_duration_weeks,omitempty" required:"false" minimum:"1" maximum:"8" doc:"Weeks per cycle"`
	CycleCooldownWeeks  *int                 `json:"cycle_cooldown_weeks,omitempty" required:"false" minimum:"0" maximum:"4" doc:"Weeks between cycles"`
	CycleStartDay       *int                 `json:"cycle_start_day,omitempty" required:"false" minimum:"1" maximum:"7" doc:"1 Monday through 7 Sunday"`
	CycleAutoMove       *bool                `json:"cycle_auto_move,omitempty" required:"false" doc:"Move unfinished issues to the next cycle"`
	CycleAutoAddStarted *bool                `json:"cycle_auto_add_started,omitempty" required:"false" doc:"An issue that starts joins the active cycle"`
	EstimateType        *schema.EstimateType `json:"estimate_type,omitempty" required:"false" doc:"none, exponential, fibonacci, linear or tshirt"`
	EstimateAllowZero   *bool                `json:"estimate_allow_zero,omitempty" required:"false" doc:"Allow a zero estimate"`
	EstimateExtended    *bool                `json:"estimate_extended,omitempty" required:"false" doc:"Offer the extended estimate scale"`
	TriageEnabled       *bool                `json:"triage_enabled,omitempty" required:"false" doc:"New issues from outside the team land in triage"`
	DefaultStateID      *uuid.UUID           `json:"default_state_id,omitempty" required:"false" format:"uuid" doc:"State new issues get, one of this team's states"`
	AutoCloseMonths     *int                 `json:"auto_close_months,omitempty" required:"false" minimum:"0" maximum:"24" doc:"Close stale issues after this many months, 0 turns it off"`
	AutoArchiveMonths   *int                 `json:"auto_archive_months,omitempty" required:"false" minimum:"1" maximum:"24" doc:"Archive completed and canceled issues after this many months"`
}

func (self *TeamService) Update(ctx context.Context, userID uuid.UUID, input *UpdateTeamInput) (*models.TeamResponse, error) {
	team, err := self.adminTeam(ctx, userID, input.TeamID)
	if err != nil {
		return nil, err
	}
	update := &team_repo.UpdateTeamInput{
		Description: input.Description, Icon: input.Icon, Color: input.Color, Private: input.Private,
		CyclesEnabled: input.CyclesEnabled, CycleDurationWeeks: input.CycleDurationWeeks, CycleCooldownWeeks: input.CycleCooldownWeeks,
		CycleStartDay: input.CycleStartDay, CycleAutoMove: input.CycleAutoMove, CycleAutoAddStarted: input.CycleAutoAddStarted,
		EstimateType: input.EstimateType, EstimateAllowZero: input.EstimateAllowZero, EstimateExtended: input.EstimateExtended,
		TriageEnabled: input.TriageEnabled, AutoArchiveMonths: input.AutoArchiveMonths,
	}
	if input.Color != nil {
		color, err := names.OptionalColor(*input.Color)
		if err != nil {
			return nil, err
		}
		update.Color = &color
	}
	name, key := "", ""
	if input.Name != nil {
		if name, err = names.Clean(*input.Name); err != nil {
			return nil, err
		}
		update.Name = &name
	}
	if input.Key != nil {
		if key, err = names.TeamKey(*input.Key); err != nil {
			return nil, err
		}
		update.Key = &key
	}
	if input.Timezone != nil {
		if err := validateTimezone(*input.Timezone); err != nil {
			return nil, err
		}
		update.Timezone = input.Timezone
	}
	if input.EstimateType != nil && !validEstimateType(*input.EstimateType) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "estimate_type must be none, exponential, fibonacci, linear or tshirt")
	}
	if input.AutoCloseMonths != nil {
		update.AutoCloseMonths = input.AutoCloseMonths
		update.ClearAutoClose = *input.AutoCloseMonths == 0
		if update.ClearAutoClose {
			update.AutoCloseMonths = nil
		}
	}
	if input.DefaultStateID != nil {
		if err := self.ensureTeamState(ctx, team.ID, *input.DefaultStateID); err != nil {
			return nil, err
		}
		update.DefaultStateID = input.DefaultStateID
	}

	var updated *ent.Team
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.ensureFree(ctx, tx, team.WorkspaceID, key, name, &team.ID); err != nil {
			return err
		}
		if input.TriageEnabled != nil && *input.TriageEnabled && team.TriageStateID == nil {
			triage, err := self.ensureTriageState(ctx, tx, team.ID)
			if err != nil {
				return err
			}
			update.TriageStateID = &triage
		}
		updated, err = self.repo.Team().Update(ctx, tx, team.ID, update)
		return err
	})
	if err != nil {
		return nil, err
	}
	if updated.CyclesEnabled && !team.CyclesEnabled {
		if err := self.cycles.PlanTeamCycles(ctx, updated.ID); err != nil {
			log.Error("Planning the first cycles failed; the scheduler retries", "team", updated.ID, "err", err)
		}
	}
	return self.respondOne(ctx, userID, updated)
}

type DeleteTeamInput struct {
	TeamID uuid.UUID `json:"team_id" required:"true" format:"uuid" doc:"The team to delete together with all its issues, states and cycles"`
}

func (self *TeamService) Delete(ctx context.Context, userID uuid.UUID, input *DeleteTeamInput) error {
	team, err := self.adminTeam(ctx, userID, input.TeamID)
	if err != nil {
		return err
	}
	return self.repo.Team().Delete(ctx, team.ID)
}

func (self *TeamService) ensureFree(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, key, name string, except *uuid.UUID) error {
	keyTaken, nameTaken, err := self.repo.Team().KeyOrNameTaken(ctx, tx, workspaceID, key, name, except)
	if err != nil {
		return err
	}
	if keyTaken {
		return errdefs.NewCustomError(errdefs.ErrTypeConflict, "A team with the key "+key+" already exists in this workspace")
	}
	if nameTaken {
		return names.EnsureFree(name, []string{name}, "team", "workspace")
	}
	return nil
}

func (self *TeamService) ensureTeamState(ctx context.Context, teamID, stateID uuid.UUID) error {
	state, err := self.repo.Team().GetState(ctx, stateID)
	if ent.IsNotFound(err) || (err == nil && state.TeamID != teamID) {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "The state must belong to this team")
	}
	return err
}

func (self *TeamService) ensureTriageState(ctx context.Context, tx repository.TxInterface, teamID uuid.UUID) (uuid.UUID, error) {
	states, err := self.repo.Team().ListStates(ctx, teamID)
	if err != nil {
		return uuid.Nil, err
	}
	for _, state := range states {
		if state.Type == schema.StateTypeTriage {
			return state.ID, nil
		}
	}
	created, err := self.repo.Team().CreateState(ctx, tx, &team_repo.CreateStateInput{TeamID: teamID, Name: "Triage", Type: schema.StateTypeTriage, Color: "#fc7840"})
	if err != nil {
		return uuid.Nil, err
	}
	return created.ID, nil
}

func validateTimezone(timezone string) error {
	if timezone == "" {
		return nil
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Timezone must be an IANA time zone such as Europe/Berlin")
	}
	return nil
}

func validEstimateType(estimate schema.EstimateType) bool {
	for _, value := range estimate.Values() {
		if string(estimate) == value {
			return true
		}
	}
	return false
}
