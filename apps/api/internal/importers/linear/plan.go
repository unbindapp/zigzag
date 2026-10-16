package linear

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	comment_repo "github.com/unbindapp/zigzag-api/internal/repositories/comment"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
	label_repo "github.com/unbindapp/zigzag-api/internal/repositories/label"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

const Source = string(schema.ImportSourceLinear)

const (
	EntityUsers          = "users"
	EntityTeams          = "teams"
	EntityWorkflowStates = "workflow_states"
	EntityLabels         = "labels"
	EntityProjects       = "projects"
	EntityMilestones     = "milestones"
	EntityCycles         = "cycles"
	EntityIssues         = "issues"
	EntityComments       = "comments"
	EntityAttachments    = "attachments"
	EntityRelations      = "relations"
	EntityHistory        = "history"
)

var Entities = []string{
	EntityUsers, EntityTeams, EntityWorkflowStates, EntityLabels, EntityProjects, EntityMilestones,
	EntityCycles, EntityIssues, EntityComments, EntityAttachments, EntityRelations, EntityHistory,
}

type Options struct {
	ActorID            uuid.UUID
	RestoreArchived    bool
	RestoreTrashed     bool
	CreateMissingUsers bool
	Now                time.Time
}

// Issues holds only the columns issue_repo.IssueKeys loads.
type Snapshot struct {
	Users             []*ent.User
	TakenDisplayNames []string
	Members           []*ent.WorkspaceMember
	Teams             []*ent.Team
	States            []*ent.WorkflowState
	TeamMembers       []*ent.TeamMember
	Labels            []*ent.Label
	ProjectStatuses   []*ent.ProjectStatus
	Projects          []*ent.Project
	Milestones        []*ent.ProjectMilestone
	Cycles            []*ent.Cycle
	Issues            []*ent.Issue
	Comments          []*ent.Comment
	Attachments       []*ent.Attachment
	Relations         []*ent.IssueRelation
	HistoryIssueIDs   []uuid.UUID
}

type Plan struct {
	Counts      map[string]schema.ImportEntityCount
	Warnings    []schema.ImportWarning
	UserMapping []schema.ImportUserMapping

	Users       []user_repo.ImportedUser
	Members     []workspace_repo.ImportedMember
	Teams       []team_repo.ImportedTeam
	States      []team_repo.ImportedState
	TeamMembers []team_repo.ImportedTeamMember
	Counters    map[uuid.UUID]int
	Labels      []label_repo.ImportedLabel
	Projects    []project_repo.ImportedProject
	Milestones  []project_repo.ImportedMilestone
	Cycles      []cycle_repo.ImportedCycle
	Issues      []issue_repo.ImportedIssue
	Parents     []issue_repo.ImportedParent
	Attachments []issue_repo.ImportedAttachment
	History     []issue_repo.ImportedHistory
	Comments    []comment_repo.ImportedComment
	Relations   []issue_repo.ImportedRelation
}

type countKind int

const (
	created countKind = iota
	updated
	skipped
)

type planner struct {
	export   *LinearExport
	snapshot *Snapshot
	options  Options
	plan     *Plan

	users        map[string]*uuid.UUID
	unknownUsers map[string]bool
	appearances  map[uuid.UUID]map[uuid.UUID]bool
	teams        map[string]*teamTarget
	states       map[string]uuid.UUID
	labels       map[string]*labelTarget
	labelsByName map[string][]*labelTarget
	labelsByID   map[uuid.UUID]*labelTarget
	projects     map[string]uuid.UUID
	milestones   map[string]milestoneTarget
	cycles       map[string]cycleTarget
	issues       map[string]*issueTarget
}

func Build(export *LinearExport, snapshot *Snapshot, options Options) *Plan {
	p := &planner{
		export:   export,
		snapshot: snapshot,
		options:  options,
		plan: &Plan{
			Counts:      map[string]schema.ImportEntityCount{},
			Warnings:    []schema.ImportWarning{},
			UserMapping: []schema.ImportUserMapping{},
			Counters:    map[uuid.UUID]int{},
		},
		users:        map[string]*uuid.UUID{},
		unknownUsers: map[string]bool{},
		appearances:  map[uuid.UUID]map[uuid.UUID]bool{},
		teams:        map[string]*teamTarget{},
		states:       map[string]uuid.UUID{},
		labels:       map[string]*labelTarget{},
		labelsByName: map[string][]*labelTarget{},
		labelsByID:   map[uuid.UUID]*labelTarget{},
		projects:     map[string]uuid.UUID{},
		milestones:   map[string]milestoneTarget{},
		cycles:       map[string]cycleTarget{},
		issues:       map[string]*issueTarget{},
	}
	for _, entity := range Entities {
		p.plan.Counts[entity] = schema.ImportEntityCount{}
	}
	p.planUsers()
	p.planTeams()
	p.planLabels()
	p.planProjects()
	p.planCycles()
	p.planIssues()
	p.planComments()
	p.planRelations()
	p.planTeamMembers()
	return p.plan
}

func (self *planner) count(entity string, kind countKind) {
	c := self.plan.Counts[entity]
	switch kind {
	case created:
		c.Created++
	case updated:
		c.Updated++
	case skipped:
		c.Skipped++
	}
	self.plan.Counts[entity] = c
}

func (self *planner) warn(entity, externalID, format string, args ...any) {
	self.plan.Warnings = append(self.plan.Warnings, schema.ImportWarning{
		Entity:     entity,
		ExternalID: externalID,
		Message:    fmt.Sprintf(format, args...),
	})
}

func (self *planner) appear(teamID uuid.UUID, userID *uuid.UUID) {
	if userID == nil {
		return
	}
	if self.appearances[teamID] == nil {
		self.appearances[teamID] = map[uuid.UUID]bool{}
	}
	self.appearances[teamID][*userID] = true
}

func text(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func flag(value *bool) bool {
	return value != nil && *value
}

func timeOr(value *time.Time, fallback time.Time) time.Time {
	if value == nil {
		return fallback
	}
	return value.UTC()
}

func utc(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	at := value.UTC()
	return &at
}

func (self *planner) date(entity, externalID, field string, value *string) *time.Time {
	raw := text(value)
	if raw == "" {
		return nil
	}
	if len(raw) > len(time.DateOnly) {
		raw = raw[:len(time.DateOnly)]
	}
	parsed, err := time.ParseInLocation(time.DateOnly, raw, time.UTC)
	if err != nil {
		self.warn(entity, externalID, "%s %q is not a date and was left out", field, text(value))
		return nil
	}
	return &parsed
}

func clip(value string, length int) string {
	if utf8.RuneCountInString(value) <= length {
		return value
	}
	return string([]rune(value)[:length])
}

func lower(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
