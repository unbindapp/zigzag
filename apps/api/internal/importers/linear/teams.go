package linear

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
)

type stateRef struct {
	id       uuid.UUID
	name     string
	kind     schema.StateType
	position float64
}

type teamTarget struct {
	id      uuid.UUID
	key     string
	order   int
	created bool
	states  []stateRef
	numbers map[int]uuid.UUID
	counter int
}

func (self *teamTarget) stateByName(name string, kind schema.StateType) (stateRef, bool) {
	for _, state := range self.states {
		if lower(state.name) == lower(name) && (kind == "" || state.kind == kind) {
			return state, true
		}
	}
	return stateRef{}, false
}

func (self *teamTarget) firstState(kinds ...schema.StateType) (stateRef, bool) {
	for _, kind := range kinds {
		for _, state := range self.states {
			if state.kind == kind {
				return state, true
			}
		}
	}
	return stateRef{}, false
}

const maxNameLength = 64

var defaultStateColors = map[schema.StateType]string{
	schema.StateTypeTriage:    "#fc7840",
	schema.StateTypeBacklog:   "#95a2b3",
	schema.StateTypeUnstarted: "#e2e2e2",
	schema.StateTypeStarted:   "#f2c94c",
	schema.StateTypeCompleted: "#5e6ad2",
	schema.StateTypeCanceled:  "#95a2b3",
}

func stateType(raw string) (schema.StateType, bool) {
	switch lower(raw) {
	case "duplicate":
		return schema.StateTypeCanceled, true
	case "triage", "backlog", "unstarted", "started", "completed", "canceled":
		return schema.StateType(lower(raw)), true
	default:
		return schema.StateTypeBacklog, false
	}
}

func (self *planner) orderedTeams() []*teamTarget {
	targets := make([]*teamTarget, 0, len(self.teams))
	for _, target := range self.teams {
		targets = append(targets, target)
	}
	slices.SortFunc(targets, func(a, b *teamTarget) int { return cmp.Compare(a.order, b.order) })
	return targets
}

func (self *planner) planTeams() {
	byKey := map[string]*teamTarget{}
	byID := map[uuid.UUID]*teamTarget{}
	takenNames := map[string]bool{}
	for _, team := range self.snapshot.Teams {
		takenNames[lower(team.Name)] = true
		target := &teamTarget{id: team.ID, key: team.Key, numbers: map[int]uuid.UUID{}, counter: team.IssueCounter}
		byKey[team.Key] = target
		byID[team.ID] = target
	}
	for _, state := range self.snapshot.States {
		if target, ok := byID[state.TeamID]; ok {
			target.states = append(target.states, stateRef{id: state.ID, name: state.Name, kind: state.Type, position: state.Position})
		}
	}
	for _, issue := range self.snapshot.Issues {
		if target, ok := byID[issue.TeamID]; ok {
			target.numbers[issue.Number] = issue.ID
		}
	}

	for i, t := range self.export.Teams {
		key, err := names.TeamKey(t.Key)
		if err != nil {
			self.warn(EntityTeams, t.ID, "team key %q is not valid; the team and its issues were skipped", t.Key)
			self.count(EntityTeams, skipped)
			continue
		}
		if target, ok := byKey[key]; ok {
			target.order = i
			self.teams[t.ID] = target
			self.count(EntityTeams, updated)
			continue
		}

		name := clip(strings.TrimSpace(t.Name), maxNameLength)
		if name == "" {
			name = key
		}
		if takenNames[lower(name)] {
			name = clip(name, maxNameLength-len(key)-3) + " (" + key + ")"
			self.warn(EntityTeams, t.ID, "a team named %q exists already; imported as %q", t.Name, name)
		}
		takenNames[lower(name)] = true
		target := &teamTarget{id: uuid.New(), key: key, order: i, created: true, numbers: map[int]uuid.UUID{}}
		byKey[key] = target
		self.teams[t.ID] = target
		timezone := text(t.Timezone)
		if _, err := time.LoadLocation(timezone); timezone == "" || err != nil {
			timezone = "UTC"
		}
		row := team_repo.ImportedTeam{
			ID:            target.id,
			Key:           key,
			Name:          name,
			Description:   text(t.Description),
			Icon:          text(t.Icon),
			Color:         validColor(text(t.Color), ""),
			Timezone:      timezone,
			CyclesEnabled: flag(t.CyclesEnabled),
		}
		self.plan.Teams = append(self.plan.Teams, row)
		self.count(EntityTeams, created)
		if self.options.ActorID != uuid.Nil {
			self.appear(target.id, &self.options.ActorID)
		}
	}

	self.planStates()

	for i := range self.plan.Teams {
		row := &self.plan.Teams[i]
		target := byKey[row.Key]
		if len(target.states) == 0 {
			self.addDefaultStates(target)
		}
		if state, ok := target.firstState(schema.StateTypeUnstarted, schema.StateTypeBacklog, schema.StateTypeStarted); ok {
			row.DefaultStateID = &state.id
		}
		if state, ok := target.firstState(schema.StateTypeTriage); ok {
			row.TriageStateID = &state.id
			row.TriageEnabled = true
		}
	}
}

func (self *planner) planStates() {
	ordered := slices.Clone(self.export.WorkflowStates)
	slices.SortStableFunc(ordered, func(a, b LinearWorkflowState) int { return cmp.Compare(a.Position, b.Position) })
	for _, s := range ordered {
		target, ok := self.teams[s.TeamID]
		if !ok {
			self.warn(EntityWorkflowStates, s.ID, "state %q belongs to a team that is not imported", s.Name)
			self.count(EntityWorkflowStates, skipped)
			continue
		}
		kind, known := stateType(s.Type)
		if !known {
			self.warn(EntityWorkflowStates, s.ID, "state type %q is unknown; imported as backlog", s.Type)
		}
		if state, ok := target.stateByName(s.Name, kind); ok {
			self.states[s.ID] = state.id
			self.count(EntityWorkflowStates, skipped)
			continue
		}
		if state, ok := target.stateByName(s.Name, ""); ok {
			self.warn(EntityWorkflowStates, s.ID, "state %q exists with type %s; issues in it use the existing state", s.Name, state.kind)
			self.states[s.ID] = state.id
			self.count(EntityWorkflowStates, skipped)
			continue
		}
		name, err := names.Clean(s.Name)
		if err != nil {
			self.warn(EntityWorkflowStates, s.ID, "state has no name and was skipped")
			self.count(EntityWorkflowStates, skipped)
			continue
		}
		name = clip(name, maxNameLength)
		state := stateRef{id: uuid.New(), name: name, kind: kind, position: s.Position}
		target.states = append(target.states, state)
		self.states[s.ID] = state.id
		self.plan.States = append(self.plan.States, team_repo.ImportedState{
			ID: state.id, TeamID: target.id, Name: name, Type: kind, Color: validColor(text(s.Color), defaultStateColors[kind]), Position: s.Position,
		})
		self.count(EntityWorkflowStates, created)
	}
}

func (self *planner) addDefaultStates(target *teamTarget) {
	positions := map[schema.StateType]float64{}
	for _, d := range team_repo.DefaultStates {
		state := stateRef{id: uuid.New(), name: d.Name, kind: d.Type, position: positions[d.Type]}
		positions[d.Type]++
		target.states = append(target.states, state)
		self.plan.States = append(self.plan.States, team_repo.ImportedState{
			ID: state.id, TeamID: target.id, Name: state.name, Type: state.kind, Color: d.Color, Position: state.position,
		})
		self.count(EntityWorkflowStates, created)
	}
}

func validColor(raw, fallback string) string {
	if raw == "" {
		return fallback
	}
	color, err := names.Color(raw)
	if err != nil {
		return fallback
	}
	return color
}
