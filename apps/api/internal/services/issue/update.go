package issue_service

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/markdown"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
)

type UpdateIssueInput struct {
	IssueID           uuid.UUID   `json:"issue_id" required:"true" format:"uuid" doc:"The issue to change; omitted fields keep their value"`
	Title             *string     `json:"title,omitempty" required:"false" minLength:"1" maxLength:"255" doc:"New title"`
	Description       *string     `json:"description,omitempty" required:"false" maxLength:"100000" doc:"New markdown description; empty clears it"`
	StateID           *uuid.UUID  `json:"state_id,omitempty" required:"false" format:"uuid" doc:"Workflow state of the issue's team"`
	Priority          *int        `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	Estimate          *float64    `json:"estimate,omitempty" required:"false" minimum:"-1" doc:"Estimate in the team's scale; -1 clears it"`
	AssigneeID        *uuid.UUID  `json:"assignee_id,omitempty" required:"false" format:"uuid" doc:"Workspace member to assign; the nil uuid 00000000-0000-0000-0000-000000000000 unassigns"`
	LabelIDs          []uuid.UUID `json:"label_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Replaces every label; an empty list removes all"`
	AddLabelIDs       []uuid.UUID `json:"add_label_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Labels to add; a label of a single-select group replaces its sibling"`
	RemoveLabelIDs    []uuid.UUID `json:"remove_label_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Labels to remove"`
	ProjectID         *uuid.UUID  `json:"project_id,omitempty" required:"false" format:"uuid" doc:"Project of the workspace; the nil uuid removes it and its milestone"`
	MilestoneID       *uuid.UUID  `json:"milestone_id,omitempty" required:"false" format:"uuid" doc:"Milestone of the issue's project; the nil uuid removes it"`
	CycleID           *uuid.UUID  `json:"cycle_id,omitempty" required:"false" format:"uuid" doc:"Cycle of the issue's team; the nil uuid removes it"`
	ParentID          *uuid.UUID  `json:"parent_id,omitempty" required:"false" format:"uuid" doc:"Parent issue; the nil uuid makes it top level"`
	DueDate           *string     `json:"due_date,omitempty" required:"false" doc:"Due date YYYY-MM-DD; empty clears it"`
	SortOrder         *float64    `json:"sort_order,omitempty" required:"false" doc:"Manual order in lists and boards, ascending"`
	SubIssueSortOrder *float64    `json:"sub_issue_sort_order,omitempty" required:"false" doc:"Order among the parent's sub-issues, ascending"`
	SnoozedUntil      *string     `json:"snoozed_until,omitempty" required:"false" doc:"RFC 3339 time to snooze the issue until; empty clears it"`
	TeamID            *uuid.UUID  `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Move the issue to another team of the workspace; it gets a new identifier and the old one keeps resolving"`
}

type issueOutcome struct {
	issueID         uuid.UUID
	workspaceID     uuid.UUID
	changes         []schema.FieldChange
	assigned        *uuid.UUID
	unassigned      *uuid.UUID
	stateChanged    *ent.WorkflowState
	priorityChanged bool
	mentioned       []uuid.UUID
}

func (self *IssueService) Update(ctx context.Context, userID uuid.UUID, input *UpdateIssueInput) (*models.IssueResponse, error) {
	current, err := self.WritableIssue(ctx, userID, input.IssueID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	var outcome *issueOutcome
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		locked, err := self.repo.Issue().LockWithEdges(ctx, tx, current.ID)
		if err != nil {
			return err
		}
		outcome, err = self.applyUpdate(ctx, tx, userID, locked, input)
		return err
	})
	if err != nil {
		return nil, err
	}
	response, err := self.reload(ctx, userID, current.ID)
	if err != nil {
		return nil, err
	}
	self.afterUpdate(ctx, userID, outcome, response)
	return response, nil
}

type BulkUpdateIssuesInput struct {
	IssueIDs       []uuid.UUID `json:"issue_ids" format:"uuid" required:"true" nullable:"false" minItems:"1" maxItems:"250" doc:"Issues to change together in one transaction"`
	StateID        *uuid.UUID  `json:"state_id,omitempty" required:"false" format:"uuid" doc:"Workflow state; must belong to every issue's team"`
	Priority       *int        `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	AssigneeID     *uuid.UUID  `json:"assignee_id,omitempty" required:"false" format:"uuid" doc:"Assignee; the nil uuid unassigns"`
	AddLabelIDs    []uuid.UUID `json:"add_label_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Labels to add"`
	RemoveLabelIDs []uuid.UUID `json:"remove_label_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Labels to remove"`
	ProjectID      *uuid.UUID  `json:"project_id,omitempty" required:"false" format:"uuid" doc:"Project; the nil uuid removes it"`
	CycleID        *uuid.UUID  `json:"cycle_id,omitempty" required:"false" format:"uuid" doc:"Cycle of every issue's team; the nil uuid removes it"`
	ParentID       *uuid.UUID  `json:"parent_id,omitempty" required:"false" format:"uuid" doc:"Parent issue; the nil uuid makes them top level"`
	DueDate        *string     `json:"due_date,omitempty" required:"false" doc:"Due date YYYY-MM-DD; empty clears it"`
	TeamID         *uuid.UUID  `json:"team_id,omitempty" required:"false" format:"uuid" doc:"Move every issue to this team"`
}

func (self *IssueService) BulkUpdate(ctx context.Context, userID uuid.UUID, input *BulkUpdateIssuesInput) ([]*models.IssueResponse, error) {
	ids := without(input.IssueIDs)
	issues := make([]*ent.Issue, len(ids))
	for i, id := range ids {
		issue, err := self.WritableIssue(ctx, userID, id, schema.WorkspaceRoleMember)
		if err != nil {
			return nil, err
		}
		issues[i] = issue
	}
	outcomes := make([]*issueOutcome, len(issues))
	err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		locked, err := self.lockAll(ctx, tx, ids)
		if err != nil {
			return err
		}
		for i, issue := range locked {
			outcome, err := self.applyUpdate(ctx, tx, userID, issue, &UpdateIssueInput{
				IssueID: issue.ID, StateID: input.StateID, Priority: input.Priority, AssigneeID: input.AssigneeID,
				AddLabelIDs: input.AddLabelIDs, RemoveLabelIDs: input.RemoveLabelIDs, ProjectID: input.ProjectID,
				CycleID: input.CycleID, ParentID: input.ParentID, DueDate: input.DueDate, TeamID: input.TeamID,
			})
			if err != nil {
				return prefixError(models.Identifier(issue.Edges.Team.Key, issue.Number), err)
			}
			outcomes[i] = outcome
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	reloaded, err := self.repo.Issue().GetManyWithEdges(ctx, nil, ids)
	if err != nil {
		return nil, err
	}
	responses, err := self.respond(ctx, userID, reloaded)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*issueOutcome, len(outcomes))
	for _, outcome := range outcomes {
		byID[outcome.issueID] = outcome
	}
	for _, response := range responses {
		if outcome, ok := byID[response.ID]; ok {
			self.afterUpdate(ctx, userID, outcome, response)
		}
	}
	return responses, nil
}

func (self *IssueService) lockAll(ctx context.Context, tx repository.TxInterface, ids []uuid.UUID) ([]*ent.Issue, error) {
	byID := make(map[uuid.UUID]*ent.Issue, len(ids))
	for _, id := range slices.SortedFunc(slices.Values(ids), func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) }) {
		issue, err := self.repo.Issue().LockWithEdges(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		byID[id] = issue
	}
	locked := make([]*ent.Issue, len(ids))
	for i, id := range ids {
		locked[i] = byID[id]
	}
	return locked, nil
}

func prefixError(identifier string, err error) error {
	var custom *errdefs.CustomError
	if !errors.As(err, &custom) {
		return err
	}
	return errdefs.NewCustomError(custom.Type, identifier+": "+custom.Message)
}

func isNil(id *uuid.UUID) bool {
	return id != nil && *id == uuid.Nil
}

func sameID(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func idValue(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

func (self *IssueService) applyUpdate(ctx context.Context, tx repository.TxInterface, userID uuid.UUID, current *ent.Issue, input *UpdateIssueInput) (*issueOutcome, error) {
	now := self.now()
	fields := issue_repo.FieldsOf(current)
	outcome := &issueOutcome{issueID: current.ID, workspaceID: current.WorkspaceID}
	team := current.Edges.Team
	currentLabels := labelIDs(current.Edges.Labels)
	replaceLabels := input.LabelIDs
	removeLabels := slices.Clone(input.RemoveLabelIDs)
	record := func(change schema.FieldChange) { outcome.changes = append(outcome.changes, change) }

	if input.TeamID != nil && *input.TeamID != current.TeamID {
		target, err := self.moveTarget(ctx, userID, current, *input.TeamID)
		if err != nil {
			return nil, err
		}
		if err := self.repo.Issue().RecordMove(ctx, tx, current.ID, current.TeamID, current.Number); err != nil {
			return nil, err
		}
		if fields.Number, err = self.repo.Team().NextIssueNumber(ctx, tx, target.ID); err != nil {
			return nil, err
		}
		fields.TeamID = target.ID
		record(schema.FieldChange{Field: "team", From: current.TeamID.String(), To: target.ID.String(), FromLabel: models.Identifier(team.Key, current.Number), ToLabel: models.Identifier(target.Key, fields.Number)})
		if input.StateID == nil {
			state, err := self.matchingState(ctx, target, current.Edges.State, userID)
			if err != nil {
				return nil, err
			}
			fields.StateID = state.ID
		}
		if input.CycleID == nil {
			fields.CycleID = nil
		}
		for _, label := range current.Edges.Labels {
			if label.TeamID != nil && *label.TeamID != target.ID {
				removeLabels = append(removeLabels, label.ID)
			}
		}
		team = target
	}

	if input.StateID != nil {
		fields.StateID = *input.StateID
	}
	if fields.StateID != current.StateID {
		state, err := self.teamState(ctx, team, fields.StateID)
		if err != nil {
			return nil, err
		}
		applyStateTimestamps(&fields, current.Edges.State, state, now)
		record(schema.FieldChange{Field: "state", From: current.StateID.String(), To: state.ID.String(), FromLabel: current.Edges.State.Name, ToLabel: state.Name})
		outcome.stateChanged = state
		if state.Type == schema.StateTypeStarted && fields.CycleID == nil && input.CycleID == nil {
			if fields.CycleID, err = self.activeCycle(ctx, team); err != nil {
				return nil, err
			}
		}
	}

	if input.Title != nil {
		title, err := validateTitle(*input.Title)
		if err != nil {
			return nil, err
		}
		if title != current.Title {
			fields.Title = title
			record(schema.FieldChange{Field: "title", From: current.Title, To: title})
		}
	}
	if input.Description != nil && *input.Description != current.Description {
		fields.Description = *input.Description
		fields.DescriptionText = markdown.PlainText(*input.Description)
		record(schema.FieldChange{Field: "description"})
		before := markdown.Mentions(current.Description)
		added := slices.DeleteFunc(markdown.Mentions(*input.Description), func(id uuid.UUID) bool { return slices.Contains(before, id) })
		mentioned, err := self.workspaceMembers(ctx, current.WorkspaceID, added)
		if err != nil {
			return nil, err
		}
		outcome.mentioned = mentioned
	}
	if input.Priority != nil && *input.Priority != current.Priority {
		fields.Priority = *input.Priority
		record(schema.FieldChange{Field: "priority", From: current.Priority, To: *input.Priority, FromLabel: models.PriorityLabel(current.Priority), ToLabel: models.PriorityLabel(*input.Priority)})
		outcome.priorityChanged = true
	}
	if input.Estimate != nil {
		var estimate *float64
		if *input.Estimate != -1 {
			if err := validateEstimate(team, *input.Estimate); err != nil {
				return nil, err
			}
			estimate = input.Estimate
		}
		if !sameFloat(estimate, current.Estimate) {
			fields.Estimate = estimate
			record(schema.FieldChange{Field: "estimate", From: floatValue(current.Estimate), To: floatValue(estimate)})
		}
	}
	if input.AssigneeID != nil {
		if err := self.updateAssignee(ctx, current, input.AssigneeID, &fields, outcome); err != nil {
			return nil, err
		}
	}
	if err := self.updateProject(ctx, current, input, &fields, record); err != nil {
		return nil, err
	}
	if input.CycleID != nil {
		fields.CycleID = nil
		if !isNil(input.CycleID) {
			if _, err := self.cycle(ctx, team, *input.CycleID); err != nil {
				return nil, err
			}
			fields.CycleID = input.CycleID
		}
	}
	if !sameID(fields.CycleID, current.CycleID) {
		toLabel, err := self.cycleLabelOf(ctx, fields.CycleID)
		if err != nil {
			return nil, err
		}
		record(schema.FieldChange{Field: "cycle", From: idValue(current.CycleID), To: idValue(fields.CycleID), FromLabel: cycleLabel(current.Edges.Cycle), ToLabel: toLabel})
	}
	if input.ParentID != nil {
		fields.ParentID = nil
		parentLabel := ""
		if !isNil(input.ParentID) {
			parent, err := self.parent(ctx, tx, userID, current.WorkspaceID, &current.ID, *input.ParentID)
			if err != nil {
				return nil, err
			}
			fields.ParentID = &parent.ID
			parentLabel = models.Identifier(parent.Edges.Team.Key, parent.Number)
		}
		if !sameID(fields.ParentID, current.ParentID) {
			change := schema.FieldChange{Field: "parent", From: idValue(current.ParentID), To: idValue(fields.ParentID), ToLabel: parentLabel}
			if previous := current.Edges.Parent; previous != nil && previous.Edges.Team != nil {
				change.FromLabel = models.Identifier(previous.Edges.Team.Key, previous.Number)
			}
			record(change)
		}
	}
	if input.DueDate != nil {
		due, err := models.ParseDate("due_date", *input.DueDate)
		if err != nil {
			return nil, err
		}
		if !sameDate(due, current.DueDate) {
			fields.DueDate = due
			record(schema.FieldChange{Field: "due_date", From: dateValue(current.DueDate), To: dateValue(due)})
		}
	}
	if input.SortOrder != nil {
		fields.SortOrder = *input.SortOrder
	}
	if input.SubIssueSortOrder != nil {
		fields.SubIssueSortOrder = *input.SubIssueSortOrder
	}
	if input.SnoozedUntil != nil {
		snoozed, err := parseTime("snoozed_until", *input.SnoozedUntil)
		if err != nil {
			return nil, err
		}
		fields.SnoozedUntil = snoozed
	}

	labels, err := self.labelSet(ctx, current.WorkspaceID, team.ID, currentLabels, replaceLabels, input.AddLabelIDs, removeLabels)
	if err != nil {
		return nil, err
	}
	finalLabels := labelIDs(labels)
	added := slices.DeleteFunc(slices.Clone(finalLabels), func(id uuid.UUID) bool { return slices.Contains(currentLabels, id) })
	removed := slices.DeleteFunc(slices.Clone(currentLabels), func(id uuid.UUID) bool { return slices.Contains(finalLabels, id) })
	if len(added) > 0 {
		record(schema.FieldChange{Field: "labels_added", To: uuidStrings(added), ToLabel: labelNames(labels, added)})
	}
	if len(removed) > 0 {
		record(schema.FieldChange{Field: "labels_removed", From: uuidStrings(removed), FromLabel: labelNames(current.Edges.Labels, removed)})
	}

	if err := self.repo.Issue().Update(ctx, tx, current.ID, &fields); err != nil {
		return nil, err
	}
	if err := self.respaceIfCrowded(ctx, tx, current, &fields, input); err != nil {
		return nil, err
	}
	if err := self.repo.Issue().SetLabels(ctx, tx, current.ID, added, removed); err != nil {
		return nil, err
	}
	subscribe := slices.Clone(outcome.mentioned)
	if outcome.assigned != nil {
		subscribe = append(subscribe, *outcome.assigned)
	}
	if err := self.repo.Issue().Subscribe(ctx, tx, current.ID, subscribe...); err != nil {
		return nil, err
	}
	if err := self.recordHistory(ctx, tx, current.ID, actorOf(userID), outcome.changes); err != nil {
		return nil, err
	}
	return outcome, nil
}

func (self *IssueService) moveTarget(ctx context.Context, userID uuid.UUID, current *ent.Issue, teamID uuid.UUID) (*ent.Team, error) {
	target, err := self.repo.Team().GetByID(ctx, teamID)
	if ent.IsNotFound(err) || (err == nil && target.WorkspaceID != current.WorkspaceID) {
		return nil, invalid("team_id must be a team of the issue's workspace")
	}
	if err != nil {
		return nil, err
	}
	visible, err := self.canSeeTeam(ctx, userID, target)
	if err != nil {
		return nil, err
	}
	if !visible || target.ArchivedAt != nil {
		return nil, invalid("team_id must be a team of the issue's workspace")
	}
	return target, nil
}

func (self *IssueService) matchingState(ctx context.Context, team *ent.Team, from *ent.WorkflowState, userID uuid.UUID) (*ent.WorkflowState, error) {
	states, err := self.repo.Team().ListStates(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	states = slices.DeleteFunc(states, func(state *ent.WorkflowState) bool { return state.ArchivedAt != nil })
	for _, state := range states {
		if strings.EqualFold(state.Name, from.Name) {
			return state, nil
		}
	}
	for _, state := range states {
		if state.Type == from.Type {
			return state, nil
		}
	}
	return self.defaultState(ctx, team, userID)
}

func (self *IssueService) updateAssignee(ctx context.Context, current *ent.Issue, assigneeID *uuid.UUID, fields *issue_repo.IssueFields, outcome *issueOutcome) error {
	var next *uuid.UUID
	if !isNil(assigneeID) {
		if err := self.ensureWorkspaceMember(ctx, current.WorkspaceID, *assigneeID, "assignee_id"); err != nil {
			return err
		}
		next = assigneeID
	}
	if sameID(next, current.AssigneeID) {
		return nil
	}
	fields.AssigneeID = next
	change := schema.FieldChange{Field: "assignee", From: idValue(current.AssigneeID), To: idValue(next)}
	if current.Edges.Assignee != nil {
		change.FromLabel = current.Edges.Assignee.Name
	}
	if next != nil {
		if assignee, err := self.repo.User().GetByID(ctx, *next); err == nil {
			change.ToLabel = assignee.Name
		}
	}
	outcome.changes = append(outcome.changes, change)
	outcome.assigned, outcome.unassigned = next, current.AssigneeID
	return nil
}

func (self *IssueService) updateProject(ctx context.Context, current *ent.Issue, input *UpdateIssueInput, fields *issue_repo.IssueFields, record func(schema.FieldChange)) error {
	projectName, milestoneName := "", ""
	if input.ProjectID != nil {
		fields.ProjectID = nil
		if !isNil(input.ProjectID) {
			project, err := self.project(ctx, current.WorkspaceID, *input.ProjectID)
			if err != nil {
				return err
			}
			fields.ProjectID = input.ProjectID
			projectName = project.Name
		}
	}
	if input.MilestoneID != nil {
		fields.MilestoneID = nil
		if !isNil(input.MilestoneID) {
			milestone, err := self.milestone(ctx, *input.MilestoneID)
			if err != nil {
				return err
			}
			milestoneName = milestone.Name
			if fields.ProjectID == nil && input.ProjectID == nil {
				project, err := self.project(ctx, current.WorkspaceID, milestone.ProjectID)
				if err != nil {
					return err
				}
				fields.ProjectID = &milestone.ProjectID
				projectName = project.Name
			}
			if fields.ProjectID == nil || milestone.ProjectID != *fields.ProjectID {
				return invalid("milestone_id must be a milestone of the issue's project")
			}
			fields.MilestoneID = &milestone.ID
		}
	}
	if fields.MilestoneID != nil && !sameID(fields.ProjectID, current.ProjectID) && input.MilestoneID == nil {
		fields.MilestoneID = nil
	}
	if !sameID(fields.ProjectID, current.ProjectID) {
		change := schema.FieldChange{Field: "project", From: idValue(current.ProjectID), To: idValue(fields.ProjectID), ToLabel: projectName}
		if current.Edges.Project != nil {
			change.FromLabel = current.Edges.Project.Name
		}
		record(change)
	}
	if !sameID(fields.MilestoneID, current.MilestoneID) {
		change := schema.FieldChange{Field: "milestone", From: idValue(current.MilestoneID), To: idValue(fields.MilestoneID), ToLabel: milestoneName}
		if current.Edges.Milestone != nil {
			change.FromLabel = current.Edges.Milestone.Name
		}
		record(change)
	}
	return nil
}

func (self *IssueService) afterUpdate(ctx context.Context, userID uuid.UUID, outcome *issueOutcome, response *models.IssueResponse) {
	self.announce(ctx, actorOf(userID), outcome, response)
}

func (self *IssueService) announce(ctx context.Context, actor *uuid.UUID, outcome *issueOutcome, response *models.IssueResponse) {
	if outcome == nil || len(outcome.changes) == 0 {
		return
	}
	event := func(kind schema.NotificationType, recipients []uuid.UUID, body string) {
		self.notify(ctx, models.NotificationEvent{
			Type: kind, WorkspaceID: response.WorkspaceID, ActorID: actor, RecipientIDs: without(recipients, actor),
			IssueID: &response.ID, Title: issueTitle(response), Body: body,
		})
	}
	if outcome.assigned != nil {
		event(schema.NotificationIssueAssigned, []uuid.UUID{*outcome.assigned}, "Assigned to you")
	}
	if outcome.unassigned != nil {
		event(schema.NotificationIssueUnassigned, []uuid.UUID{*outcome.unassigned}, "Unassigned from you")
	}
	if len(outcome.mentioned) > 0 {
		event(schema.NotificationIssueMentioned, outcome.mentioned, markdown.PlainText(response.Description))
	}
	if outcome.stateChanged != nil || outcome.priorityChanged {
		subscribers, err := self.repo.Issue().SubscriberIDs(ctx, response.ID)
		if err != nil {
			subscribers = nil
		}
		if outcome.stateChanged != nil {
			event(schema.NotificationIssueStateChanged, subscribers, "Moved to "+outcome.stateChanged.Name)
		}
		if outcome.priorityChanged {
			event(schema.NotificationIssuePriorityChanged, subscribers, "Priority set to "+response.PriorityLabel)
		}
	}
	self.emit(ctx, response.WorkspaceID, "issue.updated", map[string]any{"issue": response, "changes": outcome.changes})
	self.githubSync.AfterIssueUpdate(ctx, response.ID, outcome.changes)
}

func (self *IssueService) respaceIfCrowded(ctx context.Context, tx repository.TxInterface, current *ent.Issue, fields *issue_repo.IssueFields, input *UpdateIssueInput) error {
	if input.SortOrder != nil {
		if err := self.repo.Issue().RespaceIfCrowded(ctx, tx, current.WorkspaceID, current.ID, fields.SortOrder); err != nil {
			return err
		}
	}
	if input.SubIssueSortOrder == nil || fields.ParentID == nil {
		return nil
	}
	return self.repo.Issue().RespaceSubIssuesIfCrowded(ctx, tx, *fields.ParentID, current.ID, fields.SubIssueSortOrder)
}

func (self *IssueService) cycleLabelOf(ctx context.Context, cycleID *uuid.UUID) (string, error) {
	if cycleID == nil {
		return "", nil
	}
	found, err := self.repo.Cycle().GetByID(ctx, *cycleID)
	if err != nil {
		return "", err
	}
	return cycleLabel(found), nil
}

func cycleLabel(cycle *ent.Cycle) string {
	if cycle == nil {
		return ""
	}
	if cycle.Name != "" {
		return cycle.Name
	}
	return "Cycle " + strconv.Itoa(cycle.Number)
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func sameFloat(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func floatValue(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func sameDate(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Format(models.DateFormat) == b.Format(models.DateFormat)
}

func dateValue(value *time.Time) any {
	formatted := models.FormatDate(value)
	if formatted == nil {
		return nil
	}
	return *formatted
}

func parseTime(field, raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, invalid(field + " must be an RFC 3339 time like 2026-10-16T09:00:00Z")
	}
	return &parsed, nil
}
