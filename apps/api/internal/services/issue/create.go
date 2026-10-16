package issue_service

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/common/markdown"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

type CreateIssueInput struct {
	TeamID        uuid.UUID   `json:"team_id" required:"true" format:"uuid" doc:"Team the issue belongs to; its key prefixes the identifier"`
	Title         string      `json:"title" required:"true" minLength:"1" maxLength:"255" doc:"One line summary"`
	Description   string      `json:"description,omitempty" required:"false" maxLength:"100000" doc:"Markdown; mention people with @[Name](user:<id>)"`
	StateID       *uuid.UUID  `json:"state_id,omitempty" required:"false" format:"uuid" doc:"Workflow state of the team; defaults to the team's default state, or triage for people outside a team with triage on"`
	Priority      *int        `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none (default), 1 urgent, 2 high, 3 medium, 4 low"`
	Estimate      *float64    `json:"estimate,omitempty" required:"false" minimum:"0" doc:"Estimate in the team's scale"`
	AssigneeID    *uuid.UUID  `json:"assignee_id,omitempty" required:"false" format:"uuid" doc:"Workspace member to assign"`
	LabelIDs      []uuid.UUID `json:"label_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Workspace labels or labels of this team; not label groups"`
	ProjectID     *uuid.UUID  `json:"project_id,omitempty" required:"false" format:"uuid" doc:"Project of the workspace"`
	MilestoneID   *uuid.UUID  `json:"milestone_id,omitempty" required:"false" format:"uuid" doc:"Milestone of the project; sets project_id when omitted"`
	CycleID       *uuid.UUID  `json:"cycle_id,omitempty" required:"false" format:"uuid" doc:"Cycle of the team"`
	ParentID      *uuid.UUID  `json:"parent_id,omitempty" required:"false" format:"uuid" doc:"Parent issue in the same workspace, making this a sub-issue"`
	DueDate       string      `json:"due_date,omitempty" required:"false" format:"date" doc:"Due date, YYYY-MM-DD"`
	SubscriberIDs []uuid.UUID `json:"subscriber_ids,omitempty" format:"uuid" required:"false" maxItems:"50" doc:"Workspace members to subscribe besides you, the assignee and mentioned people"`
	TemplateID    *uuid.UUID  `json:"template_id,omitempty" required:"false" format:"uuid" doc:"Issue template whose values fill every field you leave out; its sub-issues are created too"`
	SortOrder     *float64    `json:"sort_order,omitempty" required:"false" doc:"Manual order; defaults to the top of the state"`
}

type createdIssue struct {
	issue     *ent.Issue
	assignee  *uuid.UUID
	mentioned []uuid.UUID
}

func (self *IssueService) Create(ctx context.Context, userID uuid.UUID, input *CreateIssueInput) (*models.IssueResponse, error) {
	return self.create(ctx, userID, input, nil)
}

type LinkInTx func(ctx context.Context, tx repository.TxInterface, issueID uuid.UUID) error

func (self *IssueService) CreateFromGithub(ctx context.Context, actorID uuid.UUID, input *CreateIssueInput, link LinkInTx) (*models.IssueResponse, error) {
	if !permissions_repo.IsSystemCaller(ctx) {
		return nil, errdefs.ErrUnauthorized
	}
	return self.create(ctx, actorID, input, link)
}

func (self *IssueService) create(ctx context.Context, userID uuid.UUID, input *CreateIssueInput, link LinkInTx) (*models.IssueResponse, error) {
	team, err := self.repo.Team().GetByID(ctx, input.TeamID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, team.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Team not found")
	}
	visible, err := self.canSeeTeam(ctx, userID, team)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Team not found")
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleMember, team.WorkspaceID); err != nil {
		return nil, err
	}

	var subIssues []schema.TemplateSubIssue
	if input.TemplateID != nil {
		if subIssues, err = self.applyTemplate(ctx, team, input); err != nil {
			return nil, err
		}
	}

	var created *createdIssue
	var children []*createdIssue
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		created, err = self.insert(ctx, tx, userID, team, input)
		if err != nil {
			return err
		}
		if link != nil {
			if err := link(ctx, tx, created.issue.ID); err != nil {
				return err
			}
		}
		for _, sub := range subIssues {
			child, err := self.insert(ctx, tx, userID, team, &CreateIssueInput{TeamID: team.ID, Title: sub.Title, Description: sub.Description, ParentID: &created.issue.ID})
			if err != nil {
				return err
			}
			children = append(children, child)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	response, err := self.announceCreated(ctx, userID, team, created)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if _, err := self.announceCreated(ctx, userID, team, child); err != nil {
			log.Error("Failed to announce a template sub-issue", "issue", child.issue.ID, "err", err)
		}
	}
	return response, nil
}

func (self *IssueService) announceCreated(ctx context.Context, userID uuid.UUID, team *ent.Team, created *createdIssue) (*models.IssueResponse, error) {
	response, err := self.reload(ctx, userID, created.issue.ID)
	if err != nil {
		return nil, err
	}
	self.notifyCreated(ctx, userID, created, response)
	self.emit(ctx, team.WorkspaceID, "issue.created", response)
	return response, nil
}

func (self *IssueService) applyTemplate(ctx context.Context, team *ent.Team, input *CreateIssueInput) ([]schema.TemplateSubIssue, error) {
	template, err := self.repo.Template().GetByID(ctx, *input.TemplateID)
	if ent.IsNotFound(err) || (err == nil && (template.WorkspaceID != team.WorkspaceID || template.Kind != schema.TemplateKindIssue || (template.TeamID != nil && *template.TeamID != team.ID))) {
		return nil, invalid("template_id must be an issue template of this workspace or team")
	}
	if err != nil {
		return nil, err
	}
	data := template.Data
	if input.Description == "" {
		input.Description = data.Description
	}
	if input.Priority == nil {
		input.Priority = data.Priority
	}
	if input.Estimate == nil {
		input.Estimate = data.Estimate
	}
	if input.AssigneeID == nil {
		input.AssigneeID = data.AssigneeID
	}
	if input.ProjectID == nil {
		input.ProjectID = data.ProjectID
	}
	if input.StateID == nil && data.StateID != nil {
		if state, err := self.repo.Team().GetState(ctx, *data.StateID); err == nil && state.TeamID == team.ID {
			input.StateID = data.StateID
		}
	}
	for _, id := range data.LabelIDs {
		if !slices.Contains(input.LabelIDs, id) {
			input.LabelIDs = append(input.LabelIDs, id)
		}
	}
	return data.SubIssues, nil
}

func (self *IssueService) insert(ctx context.Context, tx repository.TxInterface, userID uuid.UUID, team *ent.Team, input *CreateIssueInput) (*createdIssue, error) {
	title, err := validateTitle(input.Title)
	if err != nil {
		return nil, err
	}
	create := &issue_repo.CreateIssueInput{
		WorkspaceID:     team.WorkspaceID,
		TeamID:          team.ID,
		Title:           title,
		Description:     input.Description,
		DescriptionText: markdown.PlainText(input.Description),
		CreatorID:       actorOf(userID),
	}

	state, err := self.initialState(ctx, team, userID, input.StateID)
	if err != nil {
		return nil, err
	}
	create.StateID = state.ID
	if input.Priority != nil {
		create.Priority = *input.Priority
	}
	if input.Estimate != nil {
		if err := validateEstimate(team, *input.Estimate); err != nil {
			return nil, err
		}
		create.Estimate = input.Estimate
	}
	if input.AssigneeID != nil {
		if err := self.ensureWorkspaceMember(ctx, team.WorkspaceID, *input.AssigneeID, "assignee_id"); err != nil {
			return nil, err
		}
		create.AssigneeID = input.AssigneeID
	}
	if len(input.LabelIDs) > 0 {
		labels, err := self.labelSet(ctx, team.WorkspaceID, team.ID, nil, nil, input.LabelIDs, nil)
		if err != nil {
			return nil, err
		}
		create.LabelIDs = labelIDs(labels)
	}
	if err := self.placeInProject(ctx, team.WorkspaceID, input.ProjectID, input.MilestoneID, create); err != nil {
		return nil, err
	}
	if input.CycleID != nil {
		if _, err := self.cycle(ctx, team, *input.CycleID); err != nil {
			return nil, err
		}
		create.CycleID = input.CycleID
	}
	if input.ParentID != nil {
		if _, err := self.parent(ctx, tx, userID, team.WorkspaceID, nil, *input.ParentID); err != nil {
			return nil, err
		}
		create.ParentID = input.ParentID
	}
	if create.DueDate, err = models.ParseDate("due_date", input.DueDate); err != nil {
		return nil, err
	}

	fields := issue_repo.IssueFields{}
	applyStateTimestamps(&fields, nil, state, self.now())
	create.StartedAt, create.CompletedAt, create.CanceledAt = fields.StartedAt, fields.CompletedAt, fields.CanceledAt
	if state.Type == schema.StateTypeStarted && create.CycleID == nil {
		if create.CycleID, err = self.activeCycle(ctx, team); err != nil {
			return nil, err
		}
	}

	mentioned, err := self.workspaceMembers(ctx, team.WorkspaceID, markdown.Mentions(input.Description))
	if err != nil {
		return nil, err
	}
	extra, err := self.workspaceMembers(ctx, team.WorkspaceID, input.SubscriberIDs)
	if err != nil {
		return nil, err
	}
	subscribers := append([]uuid.UUID{userID}, mentioned...)
	subscribers = append(subscribers, extra...)
	if create.AssigneeID != nil {
		subscribers = append(subscribers, *create.AssigneeID)
	}
	nobody := uuid.Nil
	create.SubscriberIDs = without(subscribers, &nobody)

	if input.SortOrder != nil {
		create.SortOrder = *input.SortOrder
	} else if create.SortOrder, err = self.repo.Issue().TopSortOrder(ctx, tx, state.ID); err != nil {
		return nil, err
	}
	if create.Number, err = self.repo.Team().NextIssueNumber(ctx, tx, team.ID); err != nil {
		return nil, err
	}
	create.SubIssueSortOrder = float64(create.Number)

	issue, err := self.repo.Issue().Create(ctx, tx, create)
	if err != nil {
		return nil, err
	}
	change := schema.FieldChange{Field: "created", ToLabel: models.Identifier(team.Key, create.Number)}
	if err := self.recordHistory(ctx, tx, issue.ID, actorOf(userID), []schema.FieldChange{change}); err != nil {
		return nil, err
	}
	return &createdIssue{issue: issue, assignee: create.AssigneeID, mentioned: mentioned}, nil
}

func (self *IssueService) initialState(ctx context.Context, team *ent.Team, userID uuid.UUID, stateID *uuid.UUID) (*ent.WorkflowState, error) {
	if stateID != nil {
		return self.teamState(ctx, team, *stateID)
	}
	return self.defaultState(ctx, team, userID)
}

func (self *IssueService) placeInProject(ctx context.Context, workspaceID uuid.UUID, projectID, milestoneID *uuid.UUID, create *issue_repo.CreateIssueInput) error {
	if projectID != nil {
		if _, err := self.project(ctx, workspaceID, *projectID); err != nil {
			return err
		}
		create.ProjectID = projectID
	}
	if milestoneID == nil {
		return nil
	}
	milestone, err := self.milestone(ctx, *milestoneID)
	if err != nil {
		return err
	}
	if create.ProjectID == nil {
		if _, err := self.project(ctx, workspaceID, milestone.ProjectID); err != nil {
			return err
		}
		create.ProjectID = &milestone.ProjectID
	}
	if milestone.ProjectID != *create.ProjectID {
		return invalid("milestone_id must be a milestone of the issue's project")
	}
	create.MilestoneID = milestoneID
	return nil
}

func issueTitle(response *models.IssueResponse) string {
	return strings.TrimSpace(response.Identifier + " " + response.Title)
}

func (self *IssueService) notifyCreated(ctx context.Context, userID uuid.UUID, created *createdIssue, response *models.IssueResponse) {
	actor := actorOf(userID)
	if created.assignee != nil {
		self.notify(ctx, models.NotificationEvent{
			Type: schema.NotificationIssueAssigned, WorkspaceID: response.WorkspaceID, ActorID: actor,
			RecipientIDs: without([]uuid.UUID{*created.assignee}, actor), IssueID: &response.ID,
			Title: issueTitle(response), Body: "Assigned to you",
		})
	}
	self.notify(ctx, models.NotificationEvent{
		Type: schema.NotificationIssueMentioned, WorkspaceID: response.WorkspaceID, ActorID: actor,
		RecipientIDs: without(created.mentioned, actor), IssueID: &response.ID,
		Title: issueTitle(response), Body: markdown.PlainText(response.Description),
	})
}
