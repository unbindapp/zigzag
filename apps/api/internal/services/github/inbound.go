package github_service

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	github_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/github"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
	label_service "github.com/zigzagapp-org/zigzag-api/internal/services/label"
)

const (
	maxTitleLength = 255
	untitled       = "Untitled GitHub issue"
)

var errAlreadyLinked = errors.New("the GitHub issue is already linked")

var (
	openTypes   = []schema.StateType{schema.StateTypeTriage, schema.StateTypeBacklog, schema.StateTypeUnstarted, schema.StateTypeStarted}
	staleChecks = []string{"edited", "closed", "reopened"}
)

func (self *GithubService) syncedRepo(ctx context.Context, integration *ent.GithubIntegration, fullName string) (*ent.GithubRepoSync, error) {
	name, ok := github.NormalizeFullName(fullName)
	if !ok {
		return nil, nil
	}
	repoSync, err := self.repo.Github().RepoSyncByName(ctx, integration.ID, name)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return repoSync, err
}

func (self *GithubService) linkOf(ctx context.Context, integration *ent.GithubIntegration, nodeID string) (*ent.GithubIssueLink, error) {
	if nodeID == "" {
		return nil, nil
	}
	link, err := self.repo.Github().IssueLinkByNodeID(ctx, nodeID)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if link.Edges.RepoSync == nil || link.Edges.RepoSync.IntegrationID != integration.ID {
		return nil, nil
	}
	return link, nil
}

func (self *GithubService) handleIssues(ctx context.Context, integration *ent.GithubIntegration, event *github.IssuesEvent) (bool, error) {
	repoSync, err := self.syncedRepo(ctx, integration, event.Repository.FullName)
	if repoSync == nil || err != nil {
		return false, err
	}
	link, err := self.linkOf(ctx, integration, event.Issue.NodeID)
	if err != nil {
		return false, err
	}
	switch event.Action {
	case "deleted":
		return link != nil, self.recordInbound(ctx, repoSync, self.deleteLinked(ctx, link, event))
	case "transferred":
		return link != nil, self.recordInbound(ctx, repoSync, self.transfer(ctx, integration, link, event))
	case "opened", "edited", "labeled", "unlabeled", "assigned", "unassigned", "closed", "reopened":
	default:
		return false, nil
	}
	if link == nil {
		_, err := self.importIssue(ctx, repoSync, &event.Issue)
		if !errors.Is(err, errAlreadyLinked) {
			return true, self.recordInbound(ctx, repoSync, err)
		}
		if link, err = self.linkOf(ctx, integration, event.Issue.NodeID); link == nil || err != nil {
			return true, self.recordInbound(ctx, repoSync, err)
		}
	}
	if event.Action == "opened" {
		return true, nil
	}
	return true, self.recordInbound(ctx, repoSync, self.applyIssueEvent(ctx, link, event))
}

func (self *GithubService) recordInbound(ctx context.Context, repoSync *ent.GithubRepoSync, err error) error {
	message := errdefs.DetailMessage(err)
	if err != nil {
		log.Error("Failed to apply a GitHub issue event", "repo", repoSync.FullName, "err", err)
	}
	if recordErr := self.repo.Github().RecordSyncResult(ctx, repoSync.ID, message, self.now()); recordErr != nil {
		return errors.Join(err, recordErr)
	}
	return err
}

func (self *GithubService) importIssue(ctx context.Context, repoSync *ent.GithubRepoSync, issue *github.Issue) (*models.IssueResponse, error) {
	workspaceID := repoSync.Edges.Integration.WorkspaceID
	team, err := self.repo.Team().GetByID(ctx, repoSync.TeamID)
	if err != nil {
		return nil, err
	}
	state, err := self.stateFor(ctx, team, github.ClosureOf(issue.State, issue.StateReason))
	if err != nil {
		return nil, err
	}
	input := &issue_service.CreateIssueInput{
		TeamID: team.ID, Title: clampTitle(issue.Title), Description: issue.Body, StateID: &state.ID, ProjectID: repoSync.ProjectID,
	}
	for _, label := range issue.Labels {
		id, err := self.labelFor(ctx, workspaceID, team.ID, label)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(input.LabelIDs, id) {
			input.LabelIDs = append(input.LabelIDs, id)
		}
	}
	if assignee := self.memberByLogin(ctx, workspaceID, issue.Assignee); assignee != uuid.Nil {
		input.AssigneeID = &assignee
	}
	actor := self.memberByLogin(ctx, workspaceID, &issue.User)
	label := byLine("Synced from GitHub #"+strconv.Itoa(issue.Number), actor, issue.User.Login)
	created, err := self.issues.CreateFromGithub(github.WithInbound(ctx, label, issue.HTMLURL), actor, input, func(ctx context.Context, tx repository.TxInterface, issueID uuid.UUID) error {
		_, err := self.repo.Github().CreateIssueLink(ctx, tx, &github_repo.IssueLinkInput{
			IssueID: issueID, RepoSyncID: repoSync.ID, Number: issue.Number, NodeID: issue.NodeID, HTMLURL: issue.HTMLURL, GithubUpdatedAt: issue.UpdatedAt,
		})
		return err
	})
	if ent.IsConstraintError(err) {
		return nil, errAlreadyLinked
	}
	return created, err
}

func (self *GithubService) applyIssueEvent(ctx context.Context, link *ent.GithubIssueLink, event *github.IssuesEvent) error {
	issue := &event.Issue
	if slices.Contains(staleChecks, event.Action) && issue.UpdatedAt.Before(link.GithubUpdatedAt) {
		return nil
	}
	current, err := self.repo.Issue().GetWithEdges(ctx, nil, link.IssueID)
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	input, label, err := self.eventUpdate(ctx, current, event)
	if err != nil {
		return err
	}
	if input != nil {
		actor := self.memberByLogin(ctx, current.WorkspaceID, &event.Sender)
		inbound := github.WithInbound(ctx, byLine(label, actor, event.Sender.Login), issue.HTMLURL)
		if _, err := self.issues.Update(inbound, actor, input); err != nil {
			return err
		}
	}
	return self.repo.Github().RecordInbound(ctx, link.ID, issue.UpdatedAt)
}

func (self *GithubService) eventUpdate(ctx context.Context, current *ent.Issue, event *github.IssuesEvent) (*issue_service.UpdateIssueInput, string, error) {
	issue := &event.Issue
	input := &issue_service.UpdateIssueInput{IssueID: current.ID}
	switch event.Action {
	case "edited":
		title := clampTitle(issue.Title)
		if title == current.Title && issue.Body == current.Description {
			return nil, "", nil
		}
		if title != current.Title {
			input.Title = &title
		}
		if issue.Body != current.Description {
			input.Description = &issue.Body
		}
		return input, "Edited on GitHub", nil
	case "labeled":
		if event.Label == nil {
			return nil, "", nil
		}
		id, err := self.labelFor(ctx, current.WorkspaceID, current.TeamID, *event.Label)
		if err != nil || slices.ContainsFunc(current.Edges.Labels, func(label *ent.Label) bool { return label.ID == id }) {
			return nil, "", err
		}
		input.AddLabelIDs = []uuid.UUID{id}
		return input, "Labeled on GitHub", nil
	case "unlabeled":
		if event.Label == nil {
			return nil, "", nil
		}
		for _, label := range current.Edges.Labels {
			if strings.EqualFold(label.Name, event.Label.Name) {
				input.RemoveLabelIDs = append(input.RemoveLabelIDs, label.ID)
			}
		}
		if len(input.RemoveLabelIDs) == 0 {
			return nil, "", nil
		}
		return input, "Unlabeled on GitHub", nil
	case "assigned":
		assignee := self.memberByLogin(ctx, current.WorkspaceID, event.Assignee)
		if assignee == uuid.Nil || (current.AssigneeID != nil && *current.AssigneeID == assignee) {
			return nil, "", nil
		}
		input.AssigneeID = &assignee
		return input, "Assigned on GitHub", nil
	case "unassigned":
		assignee := current.Edges.Assignee
		if assignee == nil || assignee.GithubLogin == nil || event.Assignee == nil || !strings.EqualFold(*assignee.GithubLogin, event.Assignee.Login) {
			return nil, "", nil
		}
		nobody := uuid.Nil
		input.AssigneeID = &nobody
		return input, "Unassigned on GitHub", nil
	case "closed", "reopened":
		closure := github.ClosureOf(issue.State, issue.StateReason)
		if current.Edges.State != nil && matchesClosure(current.Edges.State.Type, closure) {
			return nil, "", nil
		}
		team, err := self.repo.Team().GetByID(ctx, current.TeamID)
		if err != nil {
			return nil, "", err
		}
		state, err := self.stateFor(ctx, team, closure)
		if err != nil {
			return nil, "", err
		}
		input.StateID = &state.ID
		if closure == github.Open {
			return input, "Reopened on GitHub", nil
		}
		return input, "Closed on GitHub", nil
	}
	return nil, "", nil
}

func (self *GithubService) deleteLinked(ctx context.Context, link *ent.GithubIssueLink, event *github.IssuesEvent) error {
	if link == nil {
		return nil
	}
	label := byLine("Deleted on GitHub", uuid.Nil, event.Sender.Login)
	_, err := self.issues.Trash(github.WithInbound(ctx, label, event.Issue.HTMLURL), uuid.Nil, &issue_service.IssueRefInput{IssueID: link.IssueID})
	if err != nil && !errdefs.IsNotFound(err) {
		return err
	}
	return self.repo.Github().DeleteIssueLink(ctx, link.ID)
}

func (self *GithubService) transfer(ctx context.Context, integration *ent.GithubIntegration, link *ent.GithubIssueLink, event *github.IssuesEvent) error {
	if link == nil {
		return nil
	}
	moved, repository := event.Changes.NewIssue, event.Changes.NewRepository
	if moved == nil || repository == nil {
		return self.repo.Github().DeleteIssueLink(ctx, link.ID)
	}
	target, err := self.syncedRepo(ctx, integration, repository.FullName)
	if err != nil {
		return err
	}
	if target == nil {
		return self.repo.Github().DeleteIssueLink(ctx, link.ID)
	}
	err = self.repo.Github().Relink(ctx, link.ID, &github_repo.IssueLinkInput{
		RepoSyncID: target.ID, Number: moved.Number, NodeID: moved.NodeID, HTMLURL: moved.HTMLURL, GithubUpdatedAt: moved.UpdatedAt,
	})
	if ent.IsConstraintError(err) {
		return self.repo.Github().DeleteIssueLink(ctx, link.ID)
	}
	if err != nil {
		return err
	}
	change := schema.FieldChange{Field: "github", To: moved.HTMLURL, ToLabel: "Transferred on GitHub to " + target.FullName + "#" + strconv.Itoa(moved.Number)}
	return self.repo.Issue().CreateHistory(ctx, nil, link.IssueID, nil, schema.HistorySourceSystem, []schema.FieldChange{change})
}

func (self *GithubService) memberByLogin(ctx context.Context, workspaceID uuid.UUID, account *github.Account) uuid.UUID {
	if account == nil {
		return uuid.Nil
	}
	login, ok := github.NormalizeLogin(account.Login)
	if !ok {
		return uuid.Nil
	}
	user, err := self.repo.User().GetByGithubLogin(ctx, login)
	if err != nil {
		return uuid.Nil
	}
	if _, err := self.repo.Workspace().GetMember(ctx, nil, workspaceID, user.ID); err != nil {
		return uuid.Nil
	}
	return user.ID
}

func (self *GithubService) stateFor(ctx context.Context, team *ent.Team, closure github.Closure) (*ent.WorkflowState, error) {
	states, err := self.repo.Team().ListStates(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	wanted := []schema.StateType{schema.StateTypeUnstarted, schema.StateTypeBacklog}
	switch closure {
	case github.ClosedCompleted:
		wanted = []schema.StateType{schema.StateTypeCompleted}
	case github.ClosedNotPlanned:
		wanted = []schema.StateType{schema.StateTypeCanceled, schema.StateTypeCompleted}
	default:
		if team.DefaultStateID != nil {
			for _, state := range states {
				if state.ID == *team.DefaultStateID {
					return state, nil
				}
			}
		}
	}
	for _, kind := range wanted {
		for _, state := range states {
			if state.Type == kind {
				return state, nil
			}
		}
	}
	return nil, errors.New("team " + team.Key + " has no workflow state for this GitHub issue")
}

func matchesClosure(kind schema.StateType, closure github.Closure) bool {
	switch closure {
	case github.ClosedCompleted:
		return kind == schema.StateTypeCompleted
	case github.ClosedNotPlanned:
		return kind == schema.StateTypeCanceled
	default:
		return slices.Contains(openTypes, kind)
	}
}

func (self *GithubService) labelFor(ctx context.Context, workspaceID, teamID uuid.UUID, wanted github.Label) (uuid.UUID, error) {
	name := strings.TrimSpace(wanted.Name)
	if found, err := self.findLabel(ctx, workspaceID, teamID, name); found != uuid.Nil || err != nil {
		return found, err
	}
	created, err := self.labels.Create(ctx, uuid.Nil, &label_service.CreateLabelInput{WorkspaceID: workspaceID, Name: name, Color: github.ZigzagColor(wanted.Color)})
	if err == nil {
		return created.ID, nil
	}
	if found, findErr := self.findLabel(ctx, workspaceID, teamID, name); found != uuid.Nil || findErr != nil {
		return found, findErr
	}
	return uuid.Nil, err
}

func (self *GithubService) findLabel(ctx context.Context, workspaceID, teamID uuid.UUID, name string) (uuid.UUID, error) {
	labels, err := self.repo.Label().List(ctx, workspaceID, []uuid.UUID{teamID}, false)
	if err != nil {
		return uuid.Nil, err
	}
	found := uuid.Nil
	for _, label := range labels {
		if label.IsGroup || !strings.EqualFold(label.Name, name) {
			continue
		}
		if label.TeamID != nil && *label.TeamID == teamID {
			return label.ID, nil
		}
		if label.TeamID == nil && found == uuid.Nil {
			found = label.ID
		}
	}
	return found, nil
}

func byLine(label string, actor uuid.UUID, login string) string {
	if actor != uuid.Nil || login == "" {
		return label
	}
	return label + " by @" + login
}

func clampTitle(raw string) string {
	title := strings.TrimSpace(raw)
	if title == "" {
		return untitled
	}
	if utf8.RuneCountInString(title) > maxTitleLength {
		return string([]rune(title)[:maxTitleLength])
	}
	return title
}
