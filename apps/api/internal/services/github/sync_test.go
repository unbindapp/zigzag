package github_service

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/crypto"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	github_repo "github.com/unbindapp/zigzag-api/internal/repositories/github"
	"github.com/unbindapp/zigzag-api/internal/services"
	comment_service "github.com/unbindapp/zigzag-api/internal/services/comment"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	label_service "github.com/unbindapp/zigzag-api/internal/services/label"
	mocks_integrations_github "github.com/unbindapp/zigzag-api/mocks/integrations/github"
	mocks_services_comment "github.com/unbindapp/zigzag-api/mocks/services/comment"
	mocks_services_issue "github.com/unbindapp/zigzag-api/mocks/services/issue"
	mocks_services_label "github.com/unbindapp/zigzag-api/mocks/services/label"
	mocks_services_project "github.com/unbindapp/zigzag-api/mocks/services/project"
)

const (
	syncToken    = "github_pat_11ABCDEFG0123456789abcdefghijklmnop"
	issueNodeID  = "I_kwDOKbcdef6zAAAB"
	issueHTMLURL = "https://github.com/unbindapp/unbind/issues/12"
)

type GithubSyncSuite struct {
	services.ServiceTestSuite
	service     *GithubService
	issues      *mocks_services_issue.IssueServiceMock
	comments    *mocks_services_comment.CommentServiceMock
	labels      *mocks_services_label.LabelServiceMock
	projects    *mocks_services_project.ProjectServiceMock
	client      *mocks_integrations_github.GithubClientMock
	encryptor   *crypto.Encryptor
	now         time.Time
	userID      uuid.UUID
	eveID       uuid.UUID
	workspaceID uuid.UUID
	integration *ent.GithubIntegration
	team        *ent.Team
	states      map[string]*ent.WorkflowState
	repoSync    *ent.GithubRepoSync
	repoSyncs   map[string]*ent.GithubRepoSync
	link        *ent.GithubIssueLink
	extraLinks  map[string]*ent.GithubIssueLink
	issue       *ent.Issue
	bug         *ent.Label
	synced      []string
	jobs        []func()
}

func (suite *GithubSyncSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	var err error
	suite.encryptor, err = crypto.NewEncryptor(bytes.Repeat([]byte{9}, 32))
	suite.Require().NoError(err)
	suite.issues = mocks_services_issue.NewIssueServiceMock(suite.T())
	suite.comments = mocks_services_comment.NewCommentServiceMock(suite.T())
	suite.labels = mocks_services_label.NewLabelServiceMock(suite.T())
	suite.projects = mocks_services_project.NewProjectServiceMock(suite.T())
	suite.client = mocks_integrations_github.NewGithubClientMock(suite.T())
	suite.service = NewGithubService(&config.Config{ExternalUIUrl: "https://zz.example.com"}, suite.MockRepo, suite.encryptor, Dependencies{
		Issues: suite.issues, Comments: suite.comments, Labels: suite.labels, Projects: suite.projects, Client: suite.client,
	})
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.jobs = nil
	suite.service.run = func(job func()) { suite.jobs = append(suite.jobs, job) }
	suite.userID, suite.eveID, suite.workspaceID = uuid.New(), uuid.New(), uuid.New()
	suite.synced = nil
	suite.extraLinks = map[string]*ent.GithubIssueLink{}

	sealedSecret, err := suite.encryptor.Encrypt([]byte(testSecret))
	suite.Require().NoError(err)
	sealedToken, err := suite.encryptor.Encrypt([]byte(syncToken))
	suite.Require().NoError(err)
	suite.integration = &ent.GithubIntegration{ID: uuid.New(), WorkspaceID: suite.workspaceID, SecretEncrypted: sealedSecret, TokenEncrypted: sealedToken, TokenPrefix: "github_pat_11AB"}

	teamID := uuid.New()
	suite.states = map[string]*ent.WorkflowState{}
	ordered := []*ent.WorkflowState{}
	for i, state := range []struct {
		name string
		kind schema.StateType
	}{
		{"Backlog", schema.StateTypeBacklog}, {"Todo", schema.StateTypeUnstarted}, {"In Progress", schema.StateTypeStarted},
		{"Done", schema.StateTypeCompleted}, {"Shipped", schema.StateTypeCompleted}, {"Canceled", schema.StateTypeCanceled},
	} {
		suite.states[state.name] = &ent.WorkflowState{ID: uuid.New(), TeamID: teamID, Name: state.name, Type: state.kind, Position: float64(i)}
		ordered = append(ordered, suite.states[state.name])
	}
	suite.team = &ent.Team{ID: teamID, WorkspaceID: suite.workspaceID, Key: "UNB", Name: "Unbind", DefaultStateID: &suite.states["Todo"].ID}
	projectID := uuid.New()
	suite.repoSync = &ent.GithubRepoSync{ID: uuid.New(), IntegrationID: suite.integration.ID, FullName: repoName, TeamID: teamID, ProjectID: &projectID}
	suite.repoSync.Edges.Integration = suite.integration
	suite.repoSync.Edges.Team = suite.team
	suite.bug = &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#d73a4a"}
	suite.issue = &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: teamID, Number: 7, Title: "Logs stop streaming", Description: "Body", StateID: suite.states["Todo"].ID}
	suite.issue.Edges.State = suite.states["Todo"]
	suite.issue.Edges.Team = suite.team
	suite.link = &ent.GithubIssueLink{ID: uuid.New(), IssueID: suite.issue.ID, RepoSyncID: suite.repoSync.ID, Number: 12, NodeID: issueNodeID, HTMLURL: issueHTMLURL, GithubUpdatedAt: time.Date(2026, 10, 16, 10, 0, 0, 0, time.UTC)}
	suite.link.Edges.RepoSync = suite.repoSync

	suite.MockGithubRepo.EXPECT().GetByWorkspace(mock.Anything, suite.workspaceID).Return(suite.integration, nil).Maybe()
	suite.MockGithubRepo.EXPECT().RecordDelivery(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.repoSyncs = map[string]*ent.GithubRepoSync{repoName: suite.repoSync}
	suite.MockGithubRepo.EXPECT().RepoSyncByName(mock.Anything, suite.integration.ID, mock.Anything).RunAndReturn(func(_ context.Context, _ uuid.UUID, fullName string) (*ent.GithubRepoSync, error) {
		found, ok := suite.repoSyncs[fullName]
		if !ok {
			return nil, &ent.NotFoundError{}
		}
		return found, nil
	}).Maybe()
	suite.MockGithubRepo.EXPECT().IssueLinkByNodeID(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, nodeID string) (*ent.GithubIssueLink, error) {
		if extra, ok := suite.extraLinks[nodeID]; ok {
			return extra, nil
		}
		if suite.link == nil || nodeID != suite.link.NodeID {
			return nil, &ent.NotFoundError{}
		}
		return suite.link, nil
	}).Maybe()
	suite.MockGithubRepo.EXPECT().RecordSyncResult(mock.Anything, suite.repoSync.ID, mock.Anything, suite.now).RunAndReturn(func(_ context.Context, _ uuid.UUID, message string, _ time.Time) error {
		suite.synced = append(suite.synced, message)
		return nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, suite.issue.ID).RunAndReturn(func(context.Context, repository.TxInterface, uuid.UUID) (*ent.Issue, error) {
		return suite.issue, nil
	}).Maybe()
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, teamID).Return(suite.team, nil).Maybe()
	suite.MockTeamRepo.EXPECT().ListStates(mock.Anything, teamID).Return(ordered, nil).Maybe()
	suite.MockLabelRepo.EXPECT().List(mock.Anything, suite.workspaceID, []uuid.UUID{teamID}, false).Return([]*ent.Label{suite.bug}, nil).Maybe()
	eve := &ent.User{ID: suite.eveID, Name: "Eve", GithubLogin: new("eve-gh")}
	suite.MockUserRepo.EXPECT().GetByGithubLogin(mock.Anything, "eve-gh").Return(eve, nil).Maybe()
	suite.MockUserRepo.EXPECT().GetByGithubLogin(mock.Anything, mock.Anything).Return(nil, &ent.NotFoundError{}).Maybe()
	suite.MockWorkspaceRepo.EXPECT().GetMember(mock.Anything, nil, suite.workspaceID, suite.eveID).Return(&ent.WorkspaceMember{}, nil).Maybe()
}

func (suite *GithubSyncSuite) deliver(event string, payload map[string]any) (bool, error) {
	body, err := json.Marshal(payload)
	suite.Require().NoError(err)
	return suite.service.HandleWebhook(context.Background(), &Delivery{WorkspaceID: suite.workspaceID, Event: event, Signature: github.Sign(testSecret, body), Body: body})
}

type issueEvent struct {
	action      string
	sender      string
	state       string
	stateReason string
	title       string
	body        string
	updatedAt   string
	labels      []map[string]any
	assignee    map[string]any
	label       map[string]any
	pullRequest bool
	changes     map[string]any
}

func (self issueEvent) issue() map[string]any {
	state, title, updated := self.state, self.title, self.updatedAt
	if state == "" {
		state = "open"
	}
	if title == "" {
		title = "Logs stop streaming"
	}
	if updated == "" {
		updated = "2026-10-16T11:00:00Z"
	}
	labels := self.labels
	if labels == nil {
		labels = []map[string]any{}
	}
	issue := map[string]any{
		"id": 3500000012, "node_id": issueNodeID, "number": 12, "title": title, "body": self.body, "state": state,
		"state_reason": self.stateReason, "html_url": issueHTMLURL, "user": map[string]any{"login": "octocat"},
		"labels": labels, "assignee": self.assignee, "comments": 0,
		"created_at": "2026-10-16T09:00:00Z", "updated_at": updated,
	}
	if self.pullRequest {
		issue["pull_request"] = map[string]any{"url": "https://api.github.com/repos/unbindapp/unbind/pulls/12"}
	}
	return issue
}

func (self issueEvent) build() map[string]any {
	sender := self.sender
	if sender == "" {
		sender = "octocat"
	}
	payload := map[string]any{
		"action": self.action, "issue": self.issue(), "repository": map[string]any{"full_name": "UnbindApp/Unbind", "html_url": repoURL},
		"sender": map[string]any{"login": sender},
	}
	if self.label != nil {
		payload["label"] = self.label
	}
	if self.assignee != nil {
		payload["assignee"] = self.assignee
	}
	if self.changes != nil {
		payload["changes"] = self.changes
	}
	return payload
}

func inboundLabel(ctx context.Context) string {
	inbound, ok := github.InboundFrom(ctx)
	if !ok {
		return ""
	}
	return inbound.Label
}

func (suite *GithubSyncSuite) expectUpdate(actor uuid.UUID, label string, check func(*issue_service.UpdateIssueInput)) {
	suite.issues.EXPECT().Update(mock.Anything, actor, mock.Anything).RunAndReturn(func(ctx context.Context, _ uuid.UUID, input *issue_service.UpdateIssueInput) (*models.IssueResponse, error) {
		suite.Equal(label, inboundLabel(ctx))
		suite.Equal(suite.issue.ID, input.IssueID)
		check(input)
		return &models.IssueResponse{}, nil
	}).Once()
}

func (suite *GithubSyncSuite) expectRecordInbound(updatedAt string) {
	at, err := time.Parse(time.RFC3339, updatedAt)
	suite.Require().NoError(err)
	suite.MockGithubRepo.EXPECT().RecordInbound(mock.Anything, suite.link.ID, at).Return(nil).Once()
}

func (suite *GithubSyncSuite) TestEventsOfRepositoriesThatAreNotSyncedAreIgnored() {
	payload := issueEvent{action: "opened"}.build()
	payload["repository"] = map[string]any{"full_name": "unbindapp/website"}
	handled, err := suite.deliver(github.EventIssues, payload)
	suite.Require().NoError(err)
	suite.False(handled)
}

func (suite *GithubSyncSuite) TestOpenedCreatesTheIssueWithItsLabelsAssigneeAndProject() {
	suite.link = nil
	created := &ent.Label{ID: uuid.New()}
	suite.labels.EXPECT().Create(mock.Anything, uuid.Nil, &label_service.CreateLabelInput{WorkspaceID: suite.workspaceID, Name: "needs triage", Color: "#fbca04"}).Return(&models.LabelResponse{ID: created.ID}, nil).Once()
	suite.issues.EXPECT().CreateFromGithub(mock.Anything, uuid.Nil, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, _ uuid.UUID, input *issue_service.CreateIssueInput, link issue_service.LinkInTx) (*models.IssueResponse, error) {
		inbound, ok := github.InboundFrom(ctx)
		suite.True(ok)
		suite.Equal(github.Inbound{Label: "Synced from GitHub #12 by @octocat", URL: issueHTMLURL}, inbound)
		suite.Equal(suite.team.ID, input.TeamID)
		suite.Equal("Logs stop streaming", input.Title)
		suite.Equal("Seen on staging", input.Description)
		suite.Equal(suite.states["Todo"].ID, *input.StateID)
		suite.Equal(suite.repoSync.ProjectID, input.ProjectID)
		suite.Equal([]uuid.UUID{suite.bug.ID, created.ID}, input.LabelIDs)
		suite.Equal(suite.eveID, *input.AssigneeID)
		issueID := uuid.New()
		suite.MockGithubRepo.EXPECT().CreateIssueLink(mock.Anything, mock.Anything, &github_repo.IssueLinkInput{
			IssueID: issueID, RepoSyncID: suite.repoSync.ID, Number: 12, NodeID: issueNodeID, HTMLURL: issueHTMLURL, GithubUpdatedAt: time.Date(2026, 10, 16, 11, 0, 0, 0, time.UTC),
		}).Return(&ent.GithubIssueLink{}, nil).Once()
		return &models.IssueResponse{}, link(ctx, suite.NewTxMock(), issueID)
	}).Once()

	handled, err := suite.deliver(github.EventIssues, issueEvent{
		action: "opened", body: "Seen on staging",
		labels:   []map[string]any{{"name": "bug", "color": "d73a4a"}, {"name": "needs triage", "color": "FBCA04"}},
		assignee: map[string]any{"login": "Eve-GH"},
	}.build())
	suite.Require().NoError(err)
	suite.True(handled)
	suite.Equal([]string{""}, suite.synced)
}

func (suite *GithubSyncSuite) TestClosedIssuesStartInTheirClosedState() {
	suite.link = nil
	for reason, state := range map[string]string{"completed": "Done", "not_planned": "Canceled", "duplicate": "Canceled"} {
		suite.issues.EXPECT().CreateFromGithub(mock.Anything, uuid.Nil, mock.MatchedBy(func(input *issue_service.CreateIssueInput) bool {
			return *input.StateID == suite.states[state].ID
		}), mock.Anything).Return(&models.IssueResponse{}, nil).Once()
		_, err := suite.deliver(github.EventIssues, issueEvent{action: "opened", state: "closed", stateReason: reason}.build())
		suite.Require().NoError(err, reason)
	}
}

func (suite *GithubSyncSuite) TestOpenedTwiceIsIdempotent() {
	handled, err := suite.deliver(github.EventIssues, issueEvent{action: "opened"}.build())
	suite.Require().NoError(err)
	suite.True(handled)
}

func (suite *GithubSyncSuite) TestARaceOnTheNodeIDAppliesTheEventToTheWinner() {
	winner := suite.link
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	suite.link = nil
	suite.issues.EXPECT().CreateFromGithub(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(context.Context, uuid.UUID, *issue_service.CreateIssueInput, issue_service.LinkInTx) (*models.IssueResponse, error) {
		suite.link = winner
		return nil, &ent.ConstraintError{}
	}).Once()
	suite.expectUpdate(uuid.Nil, "Labeled on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal([]uuid.UUID{suite.bug.ID}, input.AddLabelIDs)
	})
	handled, err := suite.deliver(github.EventIssues, issueEvent{action: "labeled", label: map[string]any{"name": "BUG", "color": "d73a4a"}}.build())
	suite.Require().NoError(err)
	suite.True(handled)
}

func (suite *GithubSyncSuite) TestEditedChangesTitleAndBody() {
	suite.expectUpdate(suite.eveID, "Edited on GitHub", func(input *issue_service.UpdateIssueInput) {
		suite.Equal("Logs stop streaming on Firefox", *input.Title)
		suite.Equal("New body", *input.Description)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err := suite.deliver(github.EventIssues, issueEvent{action: "edited", sender: "eve-gh", title: "Logs stop streaming on Firefox", body: "New body"}.build())
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestEditsOlderThanTheLastAppliedStateAreDropped() {
	for _, action := range []string{"edited", "closed", "reopened"} {
		_, err := suite.deliver(github.EventIssues, issueEvent{action: action, title: "Old title", state: "closed", updatedAt: "2026-10-16T09:59:59Z"}.build())
		suite.Require().NoError(err, action)
	}
}

func (suite *GithubSyncSuite) TestAnEchoOfOurOwnChangeFindsNothingToDo() {
	suite.expectRecordInbound("2026-10-16T10:00:00Z")
	_, err := suite.deliver(github.EventIssues, issueEvent{action: "edited", title: suite.issue.Title, body: suite.issue.Description, updatedAt: "2026-10-16T10:00:00Z"}.build())
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestUnlabeledRemovesTheLabelByName() {
	suite.issue.Edges.Labels = []*ent.Label{suite.bug}
	suite.expectUpdate(uuid.Nil, "Unlabeled on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal([]uuid.UUID{suite.bug.ID}, input.RemoveLabelIDs)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err := suite.deliver(github.EventIssues, issueEvent{action: "unlabeled", label: map[string]any{"name": "bug"}}.build())
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestAssignedAndUnassignedFollowGithubUsernames() {
	suite.expectUpdate(uuid.Nil, "Assigned on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal(suite.eveID, *input.AssigneeID)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err := suite.deliver(github.EventIssues, issueEvent{action: "assigned", assignee: map[string]any{"login": "eve-gh"}}.build())
	suite.Require().NoError(err)

	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err = suite.deliver(github.EventIssues, issueEvent{action: "assigned", assignee: map[string]any{"login": "stranger"}}.build())
	suite.Require().NoError(err, "a GitHub user without a Zigzag account is not assigned")

	suite.issue.AssigneeID = &suite.eveID
	suite.issue.Edges.Assignee = &ent.User{ID: suite.eveID, GithubLogin: new("eve-gh")}
	suite.expectUpdate(uuid.Nil, "Unassigned on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal(uuid.Nil, *input.AssigneeID)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err = suite.deliver(github.EventIssues, issueEvent{action: "unassigned", assignee: map[string]any{"login": "Eve-GH"}}.build())
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestClosedAndReopenedMoveBetweenStateTypes() {
	suite.expectUpdate(uuid.Nil, "Closed on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal(suite.states["Done"].ID, *input.StateID)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err := suite.deliver(github.EventIssues, issueEvent{action: "closed", state: "closed", stateReason: "completed"}.build())
	suite.Require().NoError(err)

	suite.issue.Edges.State = suite.states["Shipped"]
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err = suite.deliver(github.EventIssues, issueEvent{action: "closed", state: "closed", stateReason: "completed"}.build())
	suite.Require().NoError(err, "an issue already in a completed state stays where it is")

	suite.expectUpdate(uuid.Nil, "Closed on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal(suite.states["Canceled"].ID, *input.StateID)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err = suite.deliver(github.EventIssues, issueEvent{action: "closed", state: "closed", stateReason: "not_planned"}.build())
	suite.Require().NoError(err)

	suite.expectUpdate(uuid.Nil, "Reopened on GitHub by @octocat", func(input *issue_service.UpdateIssueInput) {
		suite.Equal(suite.states["Todo"].ID, *input.StateID)
	})
	suite.expectRecordInbound("2026-10-16T11:00:00Z")
	_, err = suite.deliver(github.EventIssues, issueEvent{action: "reopened", stateReason: "reopened"}.build())
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestDeletedTrashesAndUnlinks() {
	suite.issues.EXPECT().Trash(mock.Anything, uuid.Nil, &issue_service.IssueRefInput{IssueID: suite.issue.ID}).RunAndReturn(func(ctx context.Context, _ uuid.UUID, _ *issue_service.IssueRefInput) (*models.IssueResponse, error) {
		suite.Equal("Deleted on GitHub by @octocat", inboundLabel(ctx))
		return &models.IssueResponse{}, nil
	}).Once()
	suite.MockGithubRepo.EXPECT().DeleteIssueLink(mock.Anything, suite.link.ID).Return(nil).Once()
	handled, err := suite.deliver(github.EventIssues, issueEvent{action: "deleted"}.build())
	suite.Require().NoError(err)
	suite.True(handled)
}

func (suite *GithubSyncSuite) TestTransferredRelinksToASyncedRepositoryOrUnlinks() {
	target := &ent.GithubRepoSync{ID: uuid.New(), IntegrationID: suite.integration.ID, FullName: "unbindapp/website"}
	suite.repoSyncs[target.FullName] = target
	suite.MockGithubRepo.EXPECT().Relink(mock.Anything, suite.link.ID, &github_repo.IssueLinkInput{
		RepoSyncID: target.ID, Number: 3, NodeID: "I_new", HTMLURL: "https://github.com/unbindapp/website/issues/3", GithubUpdatedAt: time.Date(2026, 10, 16, 11, 5, 0, 0, time.UTC),
	}).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, nil, suite.issue.ID, (*uuid.UUID)(nil), schema.HistorySourceSystem, []schema.FieldChange{{
		Field: "github", To: "https://github.com/unbindapp/website/issues/3", ToLabel: "Transferred on GitHub to unbindapp/website#3",
	}}).Return(nil).Once()
	moved := issueEvent{action: "transferred", changes: map[string]any{
		"new_issue":      map[string]any{"number": 3, "node_id": "I_new", "html_url": "https://github.com/unbindapp/website/issues/3", "updated_at": "2026-10-16T11:05:00Z"},
		"new_repository": map[string]any{"full_name": "unbindapp/website"},
	}}
	_, err := suite.deliver(github.EventIssues, moved.build())
	suite.Require().NoError(err)

	delete(suite.repoSyncs, target.FullName)
	suite.MockGithubRepo.EXPECT().DeleteIssueLink(mock.Anything, suite.link.ID).Return(nil).Once()
	_, err = suite.deliver(github.EventIssues, moved.build())
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) commentEvent(action, login, body string, pullRequest bool) map[string]any {
	return map[string]any{
		"action": action, "issue": issueEvent{pullRequest: pullRequest}.issue(),
		"comment": map[string]any{
			"id": 2400000001, "body": body, "html_url": issueHTMLURL + "#issuecomment-2400000001", "user": map[string]any{"login": login},
			"created_at": "2026-10-16T10:30:00Z", "updated_at": "2026-10-16T10:30:00Z",
		},
		"repository": map[string]any{"full_name": repoName}, "sender": map[string]any{"login": login},
	}
}

func (suite *GithubSyncSuite) TestCommentsFromGithubKeepTheirAuthorOrNameThem() {
	suite.MockGithubRepo.EXPECT().CommentLinkByGithubID(mock.Anything, int64(2400000001)).Return(nil, &ent.NotFoundError{}).Twice()
	commentID := uuid.New()
	suite.comments.EXPECT().CreateFromGithub(mock.Anything, mock.MatchedBy(func(input *comment_service.GithubCommentInput) bool {
		return input.AuthorID == nil && input.Body == "**@octocat on GitHub**\n\nSeen it too" && input.CreatedAt.Equal(time.Date(2026, 10, 16, 10, 30, 0, 0, time.UTC))
	})).RunAndReturn(func(ctx context.Context, input *comment_service.GithubCommentInput) (*models.CommentResponse, error) {
		suite.Equal("Commented on GitHub", inboundLabel(ctx))
		suite.MockGithubRepo.EXPECT().CreateCommentLink(mock.Anything, mock.Anything, commentID, suite.link.ID, int64(2400000001)).Return(nil).Once()
		return &models.CommentResponse{}, input.Link(ctx, suite.NewTxMock(), commentID)
	}).Once()
	_, err := suite.deliver(github.EventIssueComment, suite.commentEvent("created", "OctoCat", "Seen it too", false))
	suite.Require().NoError(err)

	suite.comments.EXPECT().CreateFromGithub(mock.Anything, mock.MatchedBy(func(input *comment_service.GithubCommentInput) bool {
		return input.AuthorID != nil && *input.AuthorID == suite.eveID && input.Body == "On it"
	})).Return(&models.CommentResponse{}, nil).Once()
	_, err = suite.deliver(github.EventIssueComment, suite.commentEvent("created", "eve-gh", "On it", false))
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestCommentsAlreadyLinkedOrWrittenInZigzagAreSkipped() {
	suite.MockGithubRepo.EXPECT().CommentLinkByGithubID(mock.Anything, int64(2400000001)).Return(&ent.GithubCommentLink{}, nil).Once()
	_, err := suite.deliver(github.EventIssueComment, suite.commentEvent("created", "octocat", "Seen it too", false))
	suite.Require().NoError(err)

	_, err = suite.deliver(github.EventIssueComment, suite.commentEvent("created", "zigzag-bot", "**Eve on Zigzag**\n\nOn it", false))
	suite.Require().NoError(err)

	handled, err := suite.deliver(github.EventIssueComment, suite.commentEvent("created", "octocat", "LGTM", true))
	suite.Require().NoError(err)
	suite.False(handled, "pull request comments are not issue comments")
}

func (suite *GithubSyncSuite) TestCommentEditsAndDeletesFollowTheLink() {
	link := &ent.GithubCommentLink{ID: uuid.New(), CommentID: uuid.New(), IssueLinkID: suite.link.ID, GithubCommentID: 2400000001}
	suite.MockGithubRepo.EXPECT().CommentLinkByGithubID(mock.Anything, int64(2400000001)).Return(link, nil).Twice()
	suite.comments.EXPECT().Update(mock.Anything, uuid.Nil, &comment_service.UpdateCommentInput{CommentID: link.CommentID, Body: "**@octocat on GitHub**\n\nSeen it twice"}).Return(&models.CommentResponse{}, nil).Once()
	_, err := suite.deliver(github.EventIssueComment, suite.commentEvent("edited", "octocat", "Seen it twice", false))
	suite.Require().NoError(err)

	suite.comments.EXPECT().Delete(mock.Anything, uuid.Nil, &comment_service.CommentRefInput{CommentID: link.CommentID}).Return(errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Comment not found")).Once()
	suite.MockGithubRepo.EXPECT().DeleteCommentLink(mock.Anything, link.ID).Return(nil).Once()
	_, err = suite.deliver(github.EventIssueComment, suite.commentEvent("deleted", "octocat", "Seen it twice", false))
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestFailuresAreWrittenToTheRepository() {
	suite.issues.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Label Bug is archived")).Once()
	_, err := suite.deliver(github.EventIssues, issueEvent{action: "labeled", label: map[string]any{"name": "bug"}}.build())
	suite.Error(err)
	suite.Equal([]string{"Label Bug is archived"}, suite.synced)
}

func TestGithubSyncSuite(t *testing.T) {
	suite.Run(t, new(GithubSyncSuite))
}
