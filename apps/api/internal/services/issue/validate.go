package issue_service

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

func (self *IssueService) teamState(ctx context.Context, team *ent.Team, stateID uuid.UUID) (*ent.WorkflowState, error) {
	state, err := self.repo.Team().GetState(ctx, stateID)
	if ent.IsNotFound(err) || (err == nil && (state.TeamID != team.ID || state.ArchivedAt != nil)) {
		return nil, invalid("state_id must be a workflow state of team " + team.Key)
	}
	return state, err
}

func (self *IssueService) defaultState(ctx context.Context, team *ent.Team, actorID uuid.UUID) (*ent.WorkflowState, error) {
	if team.TriageEnabled && team.TriageStateID != nil {
		member, err := self.repo.Team().IsMember(ctx, team.ID, actorID)
		if err != nil {
			return nil, err
		}
		if !member {
			return self.repo.Team().GetState(ctx, *team.TriageStateID)
		}
	}
	if team.DefaultStateID != nil {
		return self.repo.Team().GetState(ctx, *team.DefaultStateID)
	}
	states, err := self.repo.Team().ListStates(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	for _, preferred := range []schema.StateType{schema.StateTypeUnstarted, schema.StateTypeBacklog, schema.StateTypeStarted} {
		for _, state := range states {
			if state.Type == preferred && state.ArchivedAt == nil {
				return state, nil
			}
		}
	}
	return nil, invalid("Team " + team.Key + " has no open workflow state")
}

func (self *IssueService) ensureWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID, field string) error {
	_, err := self.repo.Workspace().GetMember(ctx, nil, workspaceID, userID)
	if ent.IsNotFound(err) {
		return invalid(field + " must be a member of the workspace")
	}
	return err
}

func (self *IssueService) workspaceMembers(ctx context.Context, workspaceID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	members := make([]uuid.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		_, err := self.repo.Workspace().GetMember(ctx, nil, workspaceID, id)
		if ent.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		members = append(members, id)
	}
	return members, nil
}

func (self *IssueService) labelSet(ctx context.Context, workspaceID, teamID uuid.UUID, current []uuid.UUID, replace, add, remove []uuid.UUID) ([]*ent.Label, error) {
	ids := current
	if replace != nil {
		ids = replace
	}
	final := []uuid.UUID{}
	for _, id := range append(slices.Clone(ids), add...) {
		if !slices.Contains(final, id) && !slices.Contains(remove, id) {
			final = append(final, id)
		}
	}
	labels, err := self.repo.Issue().LabelsByIDs(ctx, final)
	if err != nil {
		return nil, err
	}
	if len(labels) != len(final) {
		return nil, invalid("label_ids contains a label that does not exist")
	}
	fresh := map[uuid.UUID]bool{}
	for _, label := range labels {
		if slices.Contains(current, label.ID) {
			continue
		}
		fresh[label.ID] = true
		switch {
		case label.WorkspaceID != workspaceID:
			return nil, invalid("label_ids contains a label that does not exist")
		case label.TeamID != nil && *label.TeamID != teamID:
			return nil, invalid("Label " + label.Name + " belongs to another team")
		case label.IsGroup:
			return nil, invalid("Label " + label.Name + " is a group; apply one of its labels")
		case label.ArchivedAt != nil:
			return nil, invalid("Label " + label.Name + " is archived")
		}
	}
	return singleSelect(labels, fresh)
}

func singleSelect(labels []*ent.Label, fresh map[uuid.UUID]bool) ([]*ent.Label, error) {
	byGroup := map[uuid.UUID][]*ent.Label{}
	for _, label := range labels {
		if label.ParentID == nil || label.Edges.Parent == nil || !label.Edges.Parent.GroupSingleSelect {
			continue
		}
		byGroup[*label.ParentID] = append(byGroup[*label.ParentID], label)
	}
	drop := map[uuid.UUID]bool{}
	for _, members := range byGroup {
		if len(members) < 2 {
			continue
		}
		added := slices.DeleteFunc(slices.Clone(members), func(label *ent.Label) bool { return !fresh[label.ID] })
		if len(added) > 1 {
			return nil, invalid("Only one label of group " + members[0].Edges.Parent.Name + " may be applied")
		}
		if len(added) == 0 {
			continue
		}
		for _, label := range members {
			if label.ID != added[0].ID {
				drop[label.ID] = true
			}
		}
	}
	return slices.DeleteFunc(labels, func(label *ent.Label) bool { return drop[label.ID] }), nil
}

func labelIDs(labels []*ent.Label) []uuid.UUID {
	ids := make([]uuid.UUID, len(labels))
	for i, label := range labels {
		ids[i] = label.ID
	}
	return ids
}

func labelNames(labels []*ent.Label, ids []uuid.UUID) string {
	names := []string{}
	for _, label := range labels {
		if slices.Contains(ids, label.ID) {
			names = append(names, label.Name)
		}
	}
	return strings.Join(names, ", ")
}

func (self *IssueService) project(ctx context.Context, workspaceID, projectID uuid.UUID) (*ent.Project, error) {
	project, err := self.repo.Issue().GetProject(ctx, projectID)
	if ent.IsNotFound(err) || (err == nil && (project.WorkspaceID != workspaceID || project.TrashedAt != nil)) {
		return nil, invalid("project_id must be a project of this workspace")
	}
	return project, err
}

func (self *IssueService) milestone(ctx context.Context, milestoneID uuid.UUID) (*ent.ProjectMilestone, error) {
	milestone, err := self.repo.Issue().GetMilestone(ctx, milestoneID)
	if ent.IsNotFound(err) {
		return nil, invalid("milestone_id must be a milestone of the issue's project")
	}
	return milestone, err
}

func (self *IssueService) cycle(ctx context.Context, team *ent.Team, cycleID uuid.UUID) (*ent.Cycle, error) {
	cycle, err := self.repo.Cycle().GetByID(ctx, cycleID)
	if ent.IsNotFound(err) || (err == nil && cycle.TeamID != team.ID) {
		return nil, invalid("cycle_id must be a cycle of team " + team.Key)
	}
	return cycle, err
}

func (self *IssueService) parent(ctx context.Context, tx repository.TxInterface, userID, workspaceID uuid.UUID, issueID *uuid.UUID, parentID uuid.UUID) (*ent.Issue, error) {
	if issueID != nil && *issueID == parentID {
		return nil, invalid("An issue cannot be its own parent")
	}
	notInWorkspace := invalid("parent_id must be an issue of this workspace")
	parent, err := self.repo.Issue().GetWithEdges(ctx, tx, parentID)
	if ent.IsNotFound(err) || (err == nil && (parent.WorkspaceID != workspaceID || parent.TrashedAt != nil)) {
		return nil, notInWorkspace
	}
	if err != nil {
		return nil, err
	}
	visible, err := self.canSeeTeam(ctx, userID, parent.Edges.Team)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, notInWorkspace
	}
	if issueID == nil {
		return parent, nil
	}
	ancestors, err := self.repo.Issue().AncestorIDs(ctx, tx, parentID)
	if err != nil {
		return nil, err
	}
	if slices.Contains(ancestors, *issueID) {
		return nil, invalid("parent_id is a sub-issue of this issue, which would make a loop")
	}
	return parent, nil
}

func (self *IssueService) activeCycle(ctx context.Context, team *ent.Team) (*uuid.UUID, error) {
	if !team.CyclesEnabled || !team.CycleAutoAddStarted {
		return nil, nil
	}
	active, err := self.repo.Cycle().ActiveCycleIDs(ctx, []uuid.UUID{team.ID}, self.now())
	if err != nil {
		return nil, err
	}
	id, ok := active[team.ID]
	if !ok {
		return nil, nil
	}
	return &id, nil
}

func applyStateTimestamps(fields *issue_repo.IssueFields, from *ent.WorkflowState, to *ent.WorkflowState, now time.Time) {
	if from != nil && from.Type == schema.StateTypeTriage && to.Type != schema.StateTypeTriage {
		fields.TriagedAt = &now
	}
	switch to.Type {
	case schema.StateTypeStarted:
		if fields.StartedAt == nil {
			fields.StartedAt = &now
		}
		fields.CompletedAt, fields.CanceledAt = nil, nil
	case schema.StateTypeCompleted:
		fields.CompletedAt, fields.CanceledAt = &now, nil
	case schema.StateTypeCanceled:
		fields.CanceledAt, fields.CompletedAt = &now, nil
	default:
		fields.StartedAt, fields.CompletedAt, fields.CanceledAt = nil, nil, nil
	}
}

var estimateScales = map[schema.EstimateType]struct{ base, extended []float64 }{
	schema.EstimateTypeExponential: {base: []float64{1, 2, 4, 8, 16}, extended: []float64{32, 64}},
	schema.EstimateTypeFibonacci:   {base: []float64{1, 2, 3, 5, 8}, extended: []float64{13, 21}},
	schema.EstimateTypeLinear:      {base: []float64{1, 2, 3, 4, 5}, extended: []float64{6, 7}},
	schema.EstimateTypeTShirt:      {base: []float64{1, 2, 3, 5, 8}, extended: []float64{13, 21}},
}

func estimateValues(team *ent.Team) []float64 {
	scale, ok := estimateScales[team.EstimateType]
	if !ok {
		return nil
	}
	values := slices.Clone(scale.base)
	if team.EstimateExtended {
		values = append(values, scale.extended...)
	}
	if team.EstimateAllowZero {
		values = append([]float64{0}, values...)
	}
	return values
}

func validateEstimate(team *ent.Team, estimate float64) error {
	values := estimateValues(team)
	if len(values) == 0 {
		return invalid("Team " + team.Key + " does not use estimates; turn them on in its settings first")
	}
	if slices.Contains(values, estimate) {
		return nil
	}
	allowed := make([]string, len(values))
	for i, value := range values {
		allowed[i] = strconv.FormatFloat(value, 'f', -1, 64)
	}
	return invalid("estimate for team " + team.Key + " takes one of " + strings.Join(allowed, ", "))
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", invalid("title cannot be empty")
	}
	if len([]rune(title)) > 255 {
		return "", invalid("title is at most 255 characters")
	}
	return title, nil
}
