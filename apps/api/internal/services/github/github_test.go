package github_service

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
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
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	mocks_services_githubsync "github.com/unbindapp/zigzag-api/mocks/services/githubsync"
	mocks_services_notification "github.com/unbindapp/zigzag-api/mocks/services/notification"
	mocks_services_webhook "github.com/unbindapp/zigzag-api/mocks/services/webhook"
)

const (
	testSecret = "zzgh_0123456789abcdef"
	repoName   = "unbindapp/unbind"
	repoURL    = "https://github.com/unbindapp/unbind"
	prURL      = "https://github.com/unbindapp/unbind/pull/42"
)

type historyRow struct {
	issueID uuid.UUID
	actorID *uuid.UUID
	source  schema.HistorySource
	changes []schema.FieldChange
}

type emitted struct {
	event   string
	payload map[string]any
}

type GithubServiceSuite struct {
	services.ServiceTestSuite
	service       *GithubService
	encryptor     *crypto.Encryptor
	notifications *mocks_services_notification.NotificationServiceMock
	webhooks      *mocks_services_webhook.WebhookServiceMock
	now           time.Time
	userID        uuid.UUID
	workspaceID   uuid.UUID
	integration   *ent.GithubIntegration
	team          *ent.Team
	states        map[string]*ent.WorkflowState
	issues        map[string]*ent.Issue
	attachments   map[uuid.UUID][]*ent.Attachment
	history       []historyRow
	emitted       []emitted
	deliveries    []string
}

func (suite *GithubServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	var err error
	suite.encryptor, err = crypto.NewEncryptor(bytes.Repeat([]byte{7}, 32))
	suite.Require().NoError(err)
	suite.notifications = mocks_services_notification.NewNotificationServiceMock(suite.T())
	suite.webhooks = mocks_services_webhook.NewWebhookServiceMock(suite.T())
	cfg := &config.Config{ExternalUIUrl: "https://zz.example.com"}
	githubSync := mocks_services_githubsync.NewGithubSyncServiceMock(suite.T())
	githubSync.EXPECT().AfterIssueUpdate(mock.Anything, mock.Anything, mock.Anything).Maybe()
	issues := issue_service.NewIssueService(cfg, suite.MockRepo, suite.notifications, suite.webhooks, githubSync)
	suite.service = NewGithubService(cfg, suite.MockRepo, suite.encryptor, Dependencies{Issues: issues})
	suite.now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID, suite.workspaceID = uuid.New(), uuid.New()
	suite.history, suite.emitted, suite.deliveries = nil, nil, nil

	sealed, err := suite.encryptor.Encrypt([]byte(testSecret))
	suite.Require().NoError(err)
	suite.integration = &ent.GithubIntegration{ID: uuid.New(), WorkspaceID: suite.workspaceID, SecretEncrypted: sealed, SecretPrefix: testSecret[:secretPrefixLength]}

	teamID := uuid.New()
	suite.states = map[string]*ent.WorkflowState{}
	for i, state := range []struct {
		name string
		kind schema.StateType
	}{
		{"Triage", schema.StateTypeTriage}, {"Backlog", schema.StateTypeBacklog}, {"Todo", schema.StateTypeUnstarted},
		{"In Progress", schema.StateTypeStarted}, {"In Review", schema.StateTypeStarted},
		{"Done", schema.StateTypeCompleted}, {"Shipped", schema.StateTypeCompleted}, {"Canceled", schema.StateTypeCanceled},
	} {
		suite.states[state.name] = &ent.WorkflowState{ID: uuid.New(), TeamID: teamID, Name: state.name, Type: state.kind, Position: float64(i)}
	}
	suite.team = &ent.Team{ID: teamID, WorkspaceID: suite.workspaceID, Key: "UNB", Name: "Unbind"}
	suite.issues = map[string]*ent.Issue{}
	suite.attachments = map[uuid.UUID][]*ent.Attachment{}
	for number, state := range map[int]string{455: "Backlog", 456: "Todo", 457: "In Progress", 458: "Done", 459: "Canceled", 460: "Triage"} {
		issue := &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: teamID, Number: number, Title: "Issue", StateID: suite.states[state].ID}
		issue.Edges.Team = suite.team
		issue.Edges.State = suite.states[state]
		suite.issues["UNB-"+strconv.Itoa(number)] = issue
	}
	suite.ExpectTx()
	suite.fakeWorld()
}

func (suite *GithubServiceSuite) fakeWorld() {
	suite.MockGithubRepo.EXPECT().GetByWorkspace(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, workspaceID uuid.UUID) (*ent.GithubIntegration, error) {
		if suite.integration == nil || workspaceID != suite.workspaceID {
			return nil, &ent.NotFoundError{}
		}
		return suite.integration, nil
	}).Maybe()
	suite.MockGithubRepo.EXPECT().RecordDelivery(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, integrationID uuid.UUID, fullName, event string, at time.Time) error {
		suite.Equal(suite.integration.ID, integrationID)
		suite.Equal(suite.now, at)
		suite.deliveries = append(suite.deliveries, fullName+" "+event)
		return nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().FindByIdentifier(mock.Anything, []uuid.UUID{suite.workspaceID}, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []uuid.UUID, key string, number int) ([]uuid.UUID, error) {
		issue, ok := suite.issues[key+"-"+strconv.Itoa(number)]
		if !ok {
			return []uuid.UUID{}, nil
		}
		return []uuid.UUID{issue.ID}, nil
	}).Maybe()
	byID := func(id uuid.UUID) (*ent.Issue, error) {
		for _, issue := range suite.issues {
			if issue.ID == id {
				return issue, nil
			}
		}
		return nil, &ent.NotFoundError{}
	}
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, id uuid.UUID) (*ent.Issue, error) {
		return byID(id)
	}).Maybe()
	suite.MockIssueRepo.EXPECT().LockWithEdges(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, id uuid.UUID) (*ent.Issue, error) {
		return byID(id)
	}).Maybe()
	suite.MockTeamRepo.EXPECT().ListStates(mock.Anything, suite.team.ID).RunAndReturn(func(context.Context, uuid.UUID) ([]*ent.WorkflowState, error) {
		ordered := make([]*ent.WorkflowState, 0, len(suite.states))
		for _, name := range []string{"Triage", "Backlog", "Todo", "In Progress", "In Review", "Done", "Shipped", "Canceled"} {
			ordered = append(ordered, suite.states[name])
		}
		return ordered, nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().Attachments(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, issueID uuid.UUID) ([]*ent.Attachment, error) {
		return suite.attachments[issueID], nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().AttachmentsByURL(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, issueID uuid.UUID, urls []string) ([]*ent.Attachment, error) {
		var found []*ent.Attachment
		for _, attachment := range suite.attachments[issueID] {
			for _, url := range urls {
				if attachment.URL == url {
					found = append(found, attachment)
				}
			}
		}
		return found, nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().CreateAttachment(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, input *issue_repo.CreateAttachmentInput) (*ent.Attachment, error) {
		created := &ent.Attachment{ID: uuid.New(), IssueID: input.IssueID, CreatorID: input.CreatorID, Title: input.Title, URL: input.URL, SourceType: input.SourceType, Metadata: roundTrip(input.Metadata)}
		suite.attachments[input.IssueID] = append(suite.attachments[input.IssueID], created)
		return created, nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().UpdateLinkedAttachment(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, id uuid.UUID, input *issue_repo.LinkedAttachmentInput) (*ent.Attachment, error) {
		for _, list := range suite.attachments {
			for _, attachment := range list {
				if attachment.ID == id {
					attachment.Title, attachment.URL, attachment.SourceType, attachment.Metadata = input.Title, input.URL, input.SourceType, roundTrip(input.Metadata)
					return attachment, nil
				}
			}
		}
		return nil, &ent.NotFoundError{}
	}).Maybe()
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, id uuid.UUID, fields *issue_repo.IssueFields) error {
		issue, err := byID(id)
		if err != nil {
			return err
		}
		issue.StateID, issue.CompletedAt, issue.StartedAt, issue.CycleID = fields.StateID, fields.CompletedAt, fields.StartedAt, fields.CycleID
		for _, state := range suite.states {
			if state.ID == fields.StateID {
				issue.Edges.State = state
			}
		}
		return nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, issueID uuid.UUID, actorID *uuid.UUID, source schema.HistorySource, changes []schema.FieldChange) error {
		suite.history = append(suite.history, historyRow{issueID: issueID, actorID: actorID, source: source, changes: changes})
		return nil
	}).Maybe()
	suite.MockIssueRepo.EXPECT().Counts(mock.Anything, mock.Anything).Return(map[uuid.UUID]models.IssueCounts{}, nil).Maybe()
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil).Maybe()
	suite.webhooks.EXPECT().Emit(mock.Anything, suite.workspaceID, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ uuid.UUID, event string, payload any) error {
		suite.emitted = append(suite.emitted, emitted{event: event, payload: payload.(map[string]any)})
		return nil
	}).Maybe()
}

func roundTrip(metadata map[string]any) map[string]any {
	encoded, _ := json.Marshal(metadata)
	var decoded map[string]any
	_ = json.Unmarshal(encoded, &decoded)
	return decoded
}

func (suite *GithubServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *GithubServiceSuite) deliver(event string, payload any) (bool, error) {
	body, err := json.Marshal(payload)
	suite.Require().NoError(err)
	return suite.service.HandleWebhook(context.Background(), &Delivery{WorkspaceID: suite.workspaceID, Event: event, Signature: github.Sign(testSecret, body), Body: body})
}

type prPayload struct {
	action string
	title  string
	body   string
	branch string
	state  string
	draft  bool
	merged bool
	sha    string
}

func (self prPayload) build() map[string]any {
	state := self.state
	if state == "" {
		state = "open"
	}
	sha := self.sha
	if sha == "" {
		sha = "abc123"
	}
	return map[string]any{
		"action": self.action,
		"number": 42,
		"pull_request": map[string]any{
			"number": 42, "title": self.title, "body": self.body, "html_url": prURL, "state": state,
			"draft": self.draft, "merged": self.merged, "user": map[string]any{"login": "bbedward"},
			"head": map[string]any{"ref": self.branch, "sha": sha},
		},
		"repository": map[string]any{"full_name": repoName, "html_url": repoURL},
	}
}

func pushPayload(ref, after string) map[string]any {
	return map[string]any{
		"ref": ref, "after": after, "deleted": false,
		"repository": map[string]any{"full_name": repoName, "html_url": repoURL},
		"sender":     map[string]any{"login": "bbedward"},
	}
}

func (suite *GithubServiceSuite) stateOf(identifier string) string {
	return suite.issues[identifier].Edges.State.Name
}

func (suite *GithubServiceSuite) onlyAttachment(identifier string) *ent.Attachment {
	list := suite.attachments[suite.issues[identifier].ID]
	suite.Require().Len(list, 1)
	return list[0]
}

func (suite *GithubServiceSuite) historyOf(identifier string) []historyRow {
	var rows []historyRow
	for _, row := range suite.history {
		if row.issueID == suite.issues[identifier].ID {
			rows = append(rows, row)
		}
	}
	return rows
}

func (suite *GithubServiceSuite) TestGetIsMaskedForNonAdmins() {
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetGithubIntegrationInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *GithubServiceSuite) TestGetShowsTheWebhookURLBeforeConnecting() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.integration = nil
	status, err := suite.service.Get(suite.Ctx, suite.userID, &GetGithubIntegrationInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.False(status.Connected)
	suite.Nil(status.Integration)
	suite.Equal("https://zz.example.com/api/go/integrations/github/webhook?workspace_id="+suite.workspaceID.String(), status.WebhookURL)
}

func (suite *GithubServiceSuite) TestGetListsRepositoriesWithoutTheSecret() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.integration.Edges.Repositories = []*ent.GithubRepository{{ID: uuid.New(), FullName: repoName, LastEvent: "push", LastEventAt: suite.now}}
	status, err := suite.service.Get(suite.Ctx, suite.userID, &GetGithubIntegrationInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.True(status.Connected)
	suite.Equal(testSecret[:secretPrefixLength], status.Integration.SecretPrefix)
	suite.Require().Len(status.Integration.Repositories, 1)
	suite.Equal(repoName, status.Integration.Repositories[0].FullName)
	suite.Equal("push", status.Integration.Repositories[0].LastEvent)
	encoded, _ := json.Marshal(status)
	suite.NotContains(string(encoded), testSecret)
}

func (suite *GithubServiceSuite) TestSettingsWritesNeedAdmin() {
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	ref := &GithubIntegrationRefInput{WorkspaceID: suite.workspaceID}
	_, err := suite.service.Create(suite.Ctx, suite.userID, ref)
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	_, err = suite.service.RegenerateSecret(suite.Ctx, suite.userID, ref)
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	_, err = suite.service.Delete(suite.Ctx, suite.userID, ref)
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *GithubServiceSuite) TestCreateReturnsTheSecretOnceAndStoresItSealed() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	existing := suite.integration
	suite.integration = nil
	var stored *github_repo.CreateIntegrationInput
	suite.MockGithubRepo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, input *github_repo.CreateIntegrationInput) (*ent.GithubIntegration, error) {
		stored = input
		return &ent.GithubIntegration{ID: existing.ID, WorkspaceID: input.WorkspaceID, SecretEncrypted: input.SecretEncrypted, SecretPrefix: input.SecretPrefix}, nil
	}).Once()

	created, err := suite.service.Create(suite.Ctx, suite.userID, &GithubIntegrationRefInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.True(strings.HasPrefix(created.Secret, "zzgh_"))
	suite.Len(created.Secret, len("zzgh_")+2*secretBytes)
	suite.Equal(created.Secret[:secretPrefixLength], created.SecretPrefix)
	suite.Equal(suite.userID, stored.CreatedByID)
	suite.NotContains(string(stored.SecretEncrypted), created.Secret)
	opened, err := suite.encryptor.Decrypt(stored.SecretEncrypted)
	suite.Require().NoError(err)
	suite.Equal(created.Secret, string(opened))
	suite.Contains(created.WebhookURL, "workspace_id="+suite.workspaceID.String())
}

func (suite *GithubServiceSuite) TestCreateRefusesASecondConnection() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &GithubIntegrationRefInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrConflict)
}

func (suite *GithubServiceSuite) TestRegenerateReplacesTheSecret() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockGithubRepo.EXPECT().UpdateSecret(mock.Anything, suite.integration.ID, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ uuid.UUID, sealed []byte, prefix string) (*ent.GithubIntegration, error) {
		updated := *suite.integration
		updated.SecretEncrypted, updated.SecretPrefix = sealed, prefix
		suite.integration = &updated
		return &updated, nil
	}).Once()

	regenerated, err := suite.service.RegenerateSecret(suite.Ctx, suite.userID, &GithubIntegrationRefInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.NotEqual(testSecret, regenerated.Secret)

	_, err = suite.deliver(github.EventPing, map[string]any{"zen": "hi"})
	suite.ErrorIs(err, ErrInvalidSignature, "deliveries signed with the old secret are refused")
	body := []byte(`{"zen":"hi"}`)
	handled, err := suite.service.HandleWebhook(suite.Ctx, &Delivery{WorkspaceID: suite.workspaceID, Event: github.EventPing, Signature: github.Sign(regenerated.Secret, body), Body: body})
	suite.Require().NoError(err)
	suite.True(handled)
}

func (suite *GithubServiceSuite) TestRegenerateAndDeleteNeedAConnection() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.integration = nil
	ref := &GithubIntegrationRefInput{WorkspaceID: suite.workspaceID}
	_, err := suite.service.RegenerateSecret(suite.Ctx, suite.userID, ref)
	suite.ErrorIs(err, errdefs.ErrNotFound)
	_, err = suite.service.Delete(suite.Ctx, suite.userID, ref)
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *GithubServiceSuite) TestDeleteRemovesTheConnection() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockGithubRepo.EXPECT().Delete(mock.Anything, suite.integration.ID).Return(nil).Once()
	id, err := suite.service.Delete(suite.Ctx, suite.userID, &GithubIntegrationRefInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Equal(suite.integration.ID, id)
}

func (suite *GithubServiceSuite) TestBadSignaturesAreRefusedBeforeAnythingIsRecorded() {
	body := []byte(`{"zen":"hi","repository":{"full_name":"unbindapp/unbind"}}`)
	cases := map[string]*Delivery{
		"other secret":      {WorkspaceID: suite.workspaceID, Event: "ping", Signature: github.Sign("guess", body), Body: body},
		"missing signature": {WorkspaceID: suite.workspaceID, Event: "ping", Body: body},
		"tampered body":     {WorkspaceID: suite.workspaceID, Event: "ping", Signature: github.Sign(testSecret, body), Body: append([]byte(" "), body...)},
		"unknown workspace": {WorkspaceID: uuid.New(), Event: "ping", Signature: github.Sign(testSecret, body), Body: body},
	}
	for name, delivery := range cases {
		_, err := suite.service.HandleWebhook(suite.Ctx, delivery)
		suite.ErrorIs(err, ErrInvalidSignature, name)
	}
	suite.Empty(suite.deliveries)
}

func (suite *GithubServiceSuite) TestPingRecordsTheRepository() {
	handled, err := suite.deliver(github.EventPing, map[string]any{"zen": "hi", "repository": map[string]any{"full_name": repoName}})
	suite.Require().NoError(err)
	suite.True(handled)
	suite.Equal([]string{repoName + " ping"}, suite.deliveries)
}

func (suite *GithubServiceSuite) TestOtherEventsAreIgnored() {
	suite.MockGithubRepo.EXPECT().RepoSyncByName(mock.Anything, suite.integration.ID, repoName).Return(nil, &ent.NotFoundError{}).Once()
	handled, err := suite.deliver("issues", map[string]any{"action": "opened", "repository": map[string]any{"full_name": repoName}})
	suite.Require().NoError(err)
	suite.False(handled)
	handled, err = suite.deliver(github.EventPullRequest, prPayload{action: "labeled", title: "UNB-455 x"}.build())
	suite.Require().NoError(err)
	suite.False(handled)
	suite.Equal([]string{repoName + " issues.opened", repoName + " pull_request.labeled"}, suite.deliveries)
	suite.Empty(suite.history)
}

func (suite *GithubServiceSuite) TestOpenedPullRequestLinksAndStartsTheIssue() {
	handled, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "Limit websocket reads", branch: "bbedward/unb-455-limit"}.build())
	suite.Require().NoError(err)
	suite.True(handled)

	attachment := suite.onlyAttachment("UNB-455")
	suite.Equal("#42 Limit websocket reads", attachment.Title)
	suite.Equal(prURL, attachment.URL)
	suite.Equal(SourceTypePullRequest, attachment.SourceType)
	suite.Nil(attachment.CreatorID)
	suite.Equal(map[string]any{"repo": repoName, "number": float64(42), "state": "open", "branch": "bbedward/unb-455-limit", "author": "bbedward", "head_sha": "abc123"}, attachment.Metadata)
	suite.Equal("In Progress", suite.stateOf("UNB-455"))

	rows := suite.historyOf("UNB-455")
	suite.Require().Len(rows, 1)
	suite.Nil(rows[0].actorID)
	suite.Equal(schema.HistorySourceSystem, rows[0].source)
	suite.Equal(schema.FieldChange{Field: "github", To: prURL, ToLabel: "Pull request #42 opened"}, rows[0].changes[0])
	suite.Equal("state", rows[0].changes[1].Field)
	suite.Equal("Backlog", rows[0].changes[1].FromLabel)
	suite.Equal("In Progress", rows[0].changes[1].ToLabel)

	suite.Require().Len(suite.emitted, 1)
	suite.Equal("issue.updated", suite.emitted[0].event)
	suite.Equal(rows[0].changes, suite.emitted[0].payload["changes"])
	suite.Equal(suite.issues["UNB-455"].ID, suite.emitted[0].payload["issue"].(*models.IssueResponse).ID)
}

func (suite *GithubServiceSuite) TestEveryOpeningActionStartsUnstartedIssues() {
	for _, action := range []string{"opened", "reopened", "ready_for_review"} {
		suite.Run(action, func() {
			issue := suite.issues["UNB-456"]
			issue.StateID, issue.Edges.State = suite.states["Todo"].ID, suite.states["Todo"]
			_, err := suite.deliver(github.EventPullRequest, prPayload{action: action, title: "[UNB-456] Fix"}.build())
			suite.Require().NoError(err)
			suite.Equal("In Progress", suite.stateOf("UNB-456"))
		})
	}
	suite.Len(suite.attachments[suite.issues["UNB-456"].ID], 1, "the same pull request updates one attachment")
}

func (suite *GithubServiceSuite) TestOpeningMovesTriageAndBacklogButNotStartedOrDoneIssues() {
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "UNB-460 UNB-455 UNB-457 UNB-458 UNB-459"}.build())
	suite.Require().NoError(err)
	suite.Equal("In Progress", suite.stateOf("UNB-460"))
	suite.Equal("In Progress", suite.stateOf("UNB-455"))
	suite.Equal("In Progress", suite.stateOf("UNB-457"))
	suite.Equal("Done", suite.stateOf("UNB-458"))
	suite.Equal("Canceled", suite.stateOf("UNB-459"))
	for _, identifier := range []string{"UNB-457", "UNB-458", "UNB-459"} {
		suite.Len(suite.attachments[suite.issues[identifier].ID], 1, identifier)
		rows := suite.historyOf(identifier)
		suite.Require().Len(rows, 1, identifier)
		suite.Len(rows[0].changes, 1, "%s only records the link", identifier)
	}
}

func (suite *GithubServiceSuite) TestTeamPullRequestOpenedStateWins() {
	suite.team.GithubPrOpenedStateID = &suite.states["In Review"].ID
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "UNB-455"}.build())
	suite.Require().NoError(err)
	suite.Equal("In Review", suite.stateOf("UNB-455"))
}

func (suite *GithubServiceSuite) TestDraftPullRequestsWaitForReview() {
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "UNB-455", draft: true}.build())
	suite.Require().NoError(err)
	suite.Equal("Backlog", suite.stateOf("UNB-455"))
	suite.Equal("draft", suite.onlyAttachment("UNB-455").Metadata["state"])
	suite.Equal("Draft pull request #42 opened", suite.historyOf("UNB-455")[0].changes[0].ToLabel)

	_, err = suite.deliver(github.EventPullRequest, prPayload{action: "ready_for_review", title: "UNB-455"}.build())
	suite.Require().NoError(err)
	suite.Equal("In Progress", suite.stateOf("UNB-455"))
	suite.Equal("open", suite.onlyAttachment("UNB-455").Metadata["state"])
}

func (suite *GithubServiceSuite) TestEditsRefreshTheAttachmentAndLinkNewIssuesWithoutMoving() {
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "UNB-455 Limit reads"}.build())
	suite.Require().NoError(err)
	suite.history, suite.emitted = nil, nil

	_, err = suite.deliver(github.EventPullRequest, prPayload{action: "edited", title: "UNB-455 Limit websocket reads", body: "Also UNB-456"}.build())
	suite.Require().NoError(err)
	suite.Equal("#42 UNB-455 Limit websocket reads", suite.onlyAttachment("UNB-455").Title)
	suite.Equal("#42 UNB-455 Limit websocket reads", suite.onlyAttachment("UNB-456").Title)
	suite.Equal("Todo", suite.stateOf("UNB-456"), "an edit links new issues without moving them")
	suite.Len(suite.history, 2)
	suite.Equal("Pull request #42 updated", suite.history[0].changes[0].ToLabel)
	suite.Len(suite.emitted, 2)
}

func (suite *GithubServiceSuite) TestSynchronizeOnlyRefreshesMetadataQuietly() {
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "UNB-455"}.build())
	suite.Require().NoError(err)
	suite.history, suite.emitted = nil, nil

	_, err = suite.deliver(github.EventPullRequest, prPayload{action: "synchronize", title: "UNB-455", sha: "def456"}.build())
	suite.Require().NoError(err)
	suite.Equal("def456", suite.onlyAttachment("UNB-455").Metadata["head_sha"])
	suite.Empty(suite.history)
	suite.Empty(suite.emitted)
}

func (suite *GithubServiceSuite) TestMergeCompletesIssuesItFixes() {
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "Limit reads", branch: "bbedward/unb-456-limit", body: "fixes UNB-455, UNB-458 and UNB-459\n\nSee UNB-457"}.build())
	suite.Require().NoError(err)
	suite.history = nil

	_, err = suite.deliver(github.EventPullRequest, prPayload{action: "closed", title: "Limit reads", branch: "bbedward/unb-456-limit", body: "fixes UNB-455, UNB-458 and UNB-459\n\nSee UNB-457", state: "closed", merged: true}.build())
	suite.Require().NoError(err)
	suite.Equal("Done", suite.stateOf("UNB-455"))
	suite.Equal("In Progress", suite.stateOf("UNB-456"), "a branch reference keeps its state")
	suite.Equal("In Progress", suite.stateOf("UNB-457"), "a bare mention keeps its state")
	suite.Equal("Done", suite.stateOf("UNB-458"))
	suite.Equal("Canceled", suite.stateOf("UNB-459"), "canceled issues stay canceled")
	for _, identifier := range []string{"UNB-455", "UNB-456", "UNB-457", "UNB-458", "UNB-459"} {
		suite.Equal("merged", suite.onlyAttachment(identifier).Metadata["state"], identifier)
	}
	rows := suite.historyOf("UNB-455")
	suite.Require().Len(rows, 1)
	suite.Equal("Pull request #42 merged", rows[0].changes[0].ToLabel)
	suite.Equal("Done", rows[0].changes[1].ToLabel)
}

func (suite *GithubServiceSuite) TestTeamPullRequestMergedStateWins() {
	suite.team.GithubPrMergedStateID = &suite.states["Shipped"].ID
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "closed", title: "Closes UNB-457", state: "closed", merged: true}.build())
	suite.Require().NoError(err)
	suite.Equal("Shipped", suite.stateOf("UNB-457"))
}

func (suite *GithubServiceSuite) TestClosingWithoutMergingChangesNoState() {
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "closed", title: "fixes UNB-457", state: "closed"}.build())
	suite.Require().NoError(err)
	suite.Equal("In Progress", suite.stateOf("UNB-457"))
	suite.Equal("closed", suite.onlyAttachment("UNB-457").Metadata["state"])
	suite.Equal("Pull request #42 closed", suite.historyOf("UNB-457")[0].changes[0].ToLabel)
}

func (suite *GithubServiceSuite) TestUnknownIdentifiersAreIgnored() {
	handled, err := suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "fixes UNB-9999 and ENG-1", branch: "feature/zz-12"}.build())
	suite.Require().NoError(err)
	suite.True(handled)
	suite.Empty(suite.attachments)
	suite.Empty(suite.history)
	suite.Empty(suite.emitted)
}

func (suite *GithubServiceSuite) TestTwoIdentifiersOfOneIssueLinkItOnce() {
	suite.issues["OLD-3"] = suite.issues["UNB-455"]
	_, err := suite.deliver(github.EventPullRequest, prPayload{action: "closed", title: "OLD-3", body: "resolves UNB-455", state: "closed", merged: true}.build())
	suite.Require().NoError(err)
	suite.Len(suite.attachments[suite.issues["UNB-455"].ID], 1)
	suite.Len(suite.historyOf("UNB-455"), 1)
	suite.Equal("Done", suite.stateOf("UNB-455"))
}

func (suite *GithubServiceSuite) TestPushLinksTheBranchAndStartsTheIssue() {
	handled, err := suite.deliver(github.EventPush, pushPayload("refs/heads/bbedward/unb-455-limit", "c0ffee"))
	suite.Require().NoError(err)
	suite.True(handled)
	attachment := suite.onlyAttachment("UNB-455")
	suite.Equal("bbedward/unb-455-limit", attachment.Title)
	suite.Equal(repoURL+"/tree/bbedward/unb-455-limit", attachment.URL)
	suite.Equal(SourceTypeBranch, attachment.SourceType)
	suite.Equal(map[string]any{"repo": repoName, "branch": "bbedward/unb-455-limit", "author": "bbedward", "head_sha": "c0ffee"}, attachment.Metadata)
	suite.Equal("In Progress", suite.stateOf("UNB-455"))
	suite.Equal("Branch bbedward/unb-455-limit pushed", suite.historyOf("UNB-455")[0].changes[0].ToLabel)
	suite.Equal([]string{repoName + " push"}, suite.deliveries)

	suite.history, suite.emitted = nil, nil
	_, err = suite.deliver(github.EventPush, pushPayload("refs/heads/bbedward/unb-455-limit", "beef"))
	suite.Require().NoError(err)
	suite.Equal("beef", suite.onlyAttachment("UNB-455").Metadata["head_sha"])
	suite.Empty(suite.history, "another push only refreshes the head")
	suite.Empty(suite.emitted)
}

func (suite *GithubServiceSuite) TestPullRequestTakesOverItsBranchAttachment() {
	_, err := suite.deliver(github.EventPush, pushPayload("refs/heads/unb-455", "c0ffee"))
	suite.Require().NoError(err)
	branch := suite.onlyAttachment("UNB-455")

	_, err = suite.deliver(github.EventPullRequest, prPayload{action: "opened", title: "Limit reads", branch: "unb-455"}.build())
	suite.Require().NoError(err)
	pr := suite.onlyAttachment("UNB-455")
	suite.Equal(branch.ID, pr.ID)
	suite.Equal(prURL, pr.URL)
	suite.Equal(SourceTypePullRequest, pr.SourceType)

	suite.history = nil
	_, err = suite.deliver(github.EventPush, pushPayload("refs/heads/unb-455", "beef"))
	suite.Require().NoError(err)
	suite.Equal(SourceTypePullRequest, suite.onlyAttachment("UNB-455").SourceType, "a push to a pull request's branch adds no branch link")
	suite.Empty(suite.history)
}

func (suite *GithubServiceSuite) TestPushesWithoutABranchAreIgnored() {
	deleted := pushPayload("refs/heads/unb-455", "0000000000000000000000000000000000000000")
	deleted["deleted"] = true
	for name, payload := range map[string]map[string]any{
		"tag":            pushPayload("refs/tags/unb-455", "c0ffee"),
		"deleted branch": deleted,
	} {
		handled, err := suite.deliver(github.EventPush, payload)
		suite.Require().NoError(err, name)
		suite.False(handled, name)
	}
	suite.Empty(suite.attachments)
}

func (suite *GithubServiceSuite) TestMalformedPayloadsAreIgnored() {
	body := []byte(`{"action": 12}`)
	handled, err := suite.service.HandleWebhook(suite.Ctx, &Delivery{WorkspaceID: suite.workspaceID, Event: github.EventPullRequest, Signature: github.Sign(testSecret, body), Body: body})
	suite.Require().NoError(err)
	suite.False(handled)
	suite.Empty(suite.history)
}

func TestGithubServiceSuite(t *testing.T) {
	suite.Run(t, new(GithubServiceSuite))
}
