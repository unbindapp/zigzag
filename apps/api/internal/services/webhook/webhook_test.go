package webhook_service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/oauthserver"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	webhook_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/webhook"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type received struct {
	headers http.Header
	body    []byte
}

type WebhookServiceSuite struct {
	services.ServiceTestSuite
	service     *WebhookService
	userID      uuid.UUID
	workspaceID uuid.UUID
	endpoint    *httptest.Server
	mu          sync.Mutex
	received    []received
	failures    int
}

func (suite *WebhookServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewWebhookService(&config.Config{}, suite.MockRepo)
	suite.service.async = func(fn func()) { fn() }
	suite.service.retryDelays = []time.Duration{0, 0, 0}
	suite.service.now = func() time.Time { return time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC) }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.received = nil
	suite.failures = 0
	suite.endpoint = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		suite.mu.Lock()
		defer suite.mu.Unlock()
		suite.received = append(suite.received, received{headers: r.Header.Clone(), body: body})
		if suite.failures > 0 {
			suite.failures--
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	suite.service.client = newClient(oauthserver.AllowAllIPPolicy)
}

func (suite *WebhookServiceSuite) TearDownTest() {
	suite.endpoint.Close()
}

func (suite *WebhookServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *WebhookServiceSuite) hook(events ...string) *ent.Webhook {
	return &ent.Webhook{ID: uuid.New(), WorkspaceID: suite.workspaceID, URL: suite.endpoint.URL, Secret: "whsec_test", Enabled: true, Events: events, Label: "Bridge"}
}

func (suite *WebhookServiceSuite) validInput() *CreateWebhookInput {
	return &CreateWebhookInput{WorkspaceID: suite.workspaceID, URL: "https://hooks.example.com/zigzag", Label: " Bridge ", Events: []string{"issue.created"}, Enabled: true}
}

func (suite *WebhookServiceSuite) TestCreateNeedsAdmin() {
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, suite.validInput())
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *WebhookServiceSuite) TestCreateValidatesTheInput() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	cases := map[string]func(*CreateWebhookInput){
		"empty url":         func(in *CreateWebhookInput) { in.URL = " " },
		"not http":          func(in *CreateWebhookInput) { in.URL = "ftp://hooks.example.com" },
		"relative":          func(in *CreateWebhookInput) { in.URL = "/hooks" },
		"credentials":       func(in *CreateWebhookInput) { in.URL = "https://user:pass@hooks.example.com" },
		"localhost":         func(in *CreateWebhookInput) { in.URL = "http://localhost:8080/hook" },
		"loopback ip":       func(in *CreateWebhookInput) { in.URL = "http://127.0.0.1/hook" },
		"private ip":        func(in *CreateWebhookInput) { in.URL = "http://10.0.0.5/hook" },
		"metadata ip":       func(in *CreateWebhookInput) { in.URL = "http://169.254.169.254/latest" },
		"no events":         func(in *CreateWebhookInput) { in.Events = nil },
		"unknown event":     func(in *CreateWebhookInput) { in.Events = []string{"issue.exploded"} },
		"label too long":    func(in *CreateWebhookInput) { in.Label = strings.Repeat("x", 101) },
		"url too long":      func(in *CreateWebhookInput) { in.URL = "https://hooks.example.com/" + strings.Repeat("x", 2048) },
		"ipv6 unique local": func(in *CreateWebhookInput) { in.URL = "http://[fd00::1]/hook" },
	}
	for name, mutate := range cases {
		input := suite.validInput()
		mutate(input)
		_, err := suite.service.Create(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}
}

func (suite *WebhookServiceSuite) TestCreateRefusesTeamsOfOtherWorkspaces() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	teamID := uuid.New()
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, teamID).Return(&ent.Team{ID: teamID, WorkspaceID: uuid.New()}, nil).Once()
	input := suite.validInput()
	input.TeamID = &teamID
	_, err := suite.service.Create(suite.Ctx, suite.userID, input)
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *WebhookServiceSuite) TestCreateReturnsTheSecretOnce() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	teamID := uuid.New()
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, teamID).Return(&ent.Team{ID: teamID, WorkspaceID: suite.workspaceID}, nil).Once()
	var stored *webhook_repo.CreateWebhookInput
	suite.MockWebhookRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ repository.TxInterface, input *webhook_repo.CreateWebhookInput) (*ent.Webhook, error) {
		stored = input
		return &ent.Webhook{ID: uuid.New(), WorkspaceID: input.WorkspaceID, TeamID: input.TeamID, URL: input.URL, Secret: input.Secret, Events: input.Events, Label: input.Label, Enabled: input.Enabled}, nil
	}).Once()

	input := suite.validInput()
	input.TeamID = &teamID
	input.Events = []string{"issue.created", "comment.created", "issue.created"}
	created, err := suite.service.Create(suite.Ctx, suite.userID, input)
	suite.Require().NoError(err)
	suite.True(strings.HasPrefix(created.Secret, "whsec_"))
	suite.Len(created.Secret, len("whsec_")+48)
	suite.Equal(stored.Secret, created.Secret)
	suite.Equal("Bridge", created.Label)
	suite.Equal([]string{"issue.created", "comment.created"}, created.Events)
	suite.Equal(&teamID, created.TeamID)
}

func (suite *WebhookServiceSuite) TestListMasksNonAdmins() {
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListWebhooksInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *WebhookServiceSuite) TestListDoesNotExposeSecrets() {
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockWebhookRepo.EXPECT().List(suite.Ctx, suite.workspaceID).Return([]*ent.Webhook{suite.hook("issue.created")}, nil).Once()
	hooks, err := suite.service.List(suite.Ctx, suite.userID, &ListWebhooksInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	raw, _ := json.Marshal(hooks)
	suite.NotContains(string(raw), "whsec_test")
}

func (suite *WebhookServiceSuite) TestUpdateHidesWebhooksOutsideTheWorkspace() {
	hook := suite.hook("issue.created")
	suite.MockWebhookRepo.EXPECT().GetByID(suite.Ctx, hook.ID).Return(hook, nil)
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateWebhookInput{WebhookID: hook.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *WebhookServiceSuite) TestUpdateNeedsAdmin() {
	hook := suite.hook("issue.created")
	suite.MockWebhookRepo.EXPECT().GetByID(suite.Ctx, hook.ID).Return(hook, nil)
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateWebhookInput{WebhookID: hook.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.service.Delete(suite.Ctx, suite.userID, &WebhookRefInput{WebhookID: hook.ID}), errdefs.ErrUnauthorized)
	_, err = suite.service.Test(suite.Ctx, suite.userID, &WebhookRefInput{WebhookID: hook.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *WebhookServiceSuite) TestUpdateMissingWebhook() {
	id := uuid.New()
	suite.MockWebhookRepo.EXPECT().GetByID(suite.Ctx, id).Return(nil, &ent.NotFoundError{}).Once()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateWebhookInput{WebhookID: id})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *WebhookServiceSuite) TestUpdateClearsTheTeamWithTheNilUUID() {
	teamID := uuid.New()
	hook := suite.hook("issue.created")
	hook.TeamID = &teamID
	suite.MockWebhookRepo.EXPECT().GetByID(suite.Ctx, hook.ID).Return(hook, nil)
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	enabled := false
	suite.MockWebhookRepo.EXPECT().Update(suite.Ctx, hook.ID, mock.MatchedBy(func(input *webhook_repo.UpdateWebhookInput) bool {
		return input.ClearTeam && input.TeamID == nil && input.Enabled != nil && !*input.Enabled && input.URL == nil
	})).Return(hook, nil).Once()

	nilTeam := uuid.Nil
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateWebhookInput{WebhookID: hook.ID, TeamID: &nilTeam, Enabled: &enabled})
	suite.NoError(err)

	bad := "http://192.168.1.1/x"
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateWebhookInput{WebhookID: hook.ID, URL: &bad})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateWebhookInput{WebhookID: hook.ID, Events: []string{"nope"}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *WebhookServiceSuite) TestEmitSignsAndFiltersByEventAndTeam() {
	teamID, otherTeam := uuid.New(), uuid.New()
	wanted := suite.hook("issue.created", "issue.updated")
	unsubscribed := suite.hook("comment.created")
	teamScoped := suite.hook("issue.created")
	teamScoped.TeamID = &teamID
	otherScoped := suite.hook("issue.created")
	otherScoped.TeamID = &otherTeam
	suite.MockWebhookRepo.EXPECT().ListEnabled(suite.Ctx, suite.workspaceID).Return([]*ent.Webhook{wanted, unsubscribed, teamScoped, otherScoped}, nil).Once()

	err := suite.service.Emit(suite.Ctx, suite.workspaceID, "issue.created", map[string]any{"id": "abc", "team_id": teamID})
	suite.Require().NoError(err)
	suite.Require().Len(suite.received, 2)
	for _, got := range suite.received {
		suite.Equal("issue.created", got.headers.Get(EventHeader))
		mac := hmac.New(sha256.New, []byte("whsec_test"))
		mac.Write(got.body)
		suite.Equal("sha256="+hex.EncodeToString(mac.Sum(nil)), got.headers.Get(SignatureHeader))
		var envelope struct {
			Event       string         `json:"event"`
			WorkspaceID uuid.UUID      `json:"workspace_id"`
			Data        map[string]any `json:"data"`
		}
		suite.Require().NoError(json.Unmarshal(got.body, &envelope))
		suite.Equal("issue.created", envelope.Event)
		suite.Equal(suite.workspaceID, envelope.WorkspaceID)
		suite.Equal("abc", envelope.Data["id"])
	}
}

func (suite *WebhookServiceSuite) TestEmitResolvesTheTeamOfAComment() {
	teamID, issueID := uuid.New(), uuid.New()
	scoped := suite.hook("comment.created")
	scoped.TeamID = &teamID
	suite.MockWebhookRepo.EXPECT().ListEnabled(suite.Ctx, suite.workspaceID).Return([]*ent.Webhook{scoped}, nil).Once()
	suite.MockWebhookRepo.EXPECT().TeamOfIssue(suite.Ctx, issueID).Return(teamID, nil).Once()

	suite.Require().NoError(suite.service.Emit(suite.Ctx, suite.workspaceID, "comment.created", map[string]any{"issue_id": issueID}))
	suite.Len(suite.received, 1)
}

func (suite *WebhookServiceSuite) TestEmitMatchesTheTeamOfAnUpdatedIssue() {
	teamID, otherTeam := uuid.New(), uuid.New()
	scoped := suite.hook("issue.updated")
	scoped.TeamID = &teamID
	other := suite.hook("issue.updated")
	other.TeamID = &otherTeam
	suite.MockWebhookRepo.EXPECT().ListEnabled(suite.Ctx, suite.workspaceID).Return([]*ent.Webhook{scoped, other}, nil).Once()

	payload := map[string]any{"issue": map[string]any{"id": uuid.New(), "team_id": teamID}, "changes": []any{}}
	suite.Require().NoError(suite.service.Emit(suite.Ctx, suite.workspaceID, "issue.updated", payload))
	suite.Len(suite.received, 1, "only the webhook of the issue's team receives it")
}

func (suite *WebhookServiceSuite) TestEmitRetriesThreeTimes() {
	suite.MockWebhookRepo.EXPECT().ListEnabled(suite.Ctx, suite.workspaceID).Return([]*ent.Webhook{suite.hook("issue.deleted")}, nil).Twice()

	suite.failures = 2
	suite.Require().NoError(suite.service.Emit(suite.Ctx, suite.workspaceID, "issue.deleted", map[string]any{}))
	suite.Len(suite.received, 3)
	suite.Equal(suite.received[0].headers.Get(DeliveryHeader), suite.received[2].headers.Get(DeliveryHeader))

	suite.received = nil
	suite.failures = 10
	suite.Require().NoError(suite.service.Emit(suite.Ctx, suite.workspaceID, "issue.deleted", map[string]any{}))
	suite.Len(suite.received, 4)
}

func (suite *WebhookServiceSuite) TestEmitWithoutSubscribersSendsNothing() {
	suite.MockWebhookRepo.EXPECT().ListEnabled(suite.Ctx, suite.workspaceID).Return([]*ent.Webhook{}, nil).Once()
	suite.Require().NoError(suite.service.Emit(suite.Ctx, suite.workspaceID, "cycle.updated", map[string]any{}))
	suite.Empty(suite.received)
}

func (suite *WebhookServiceSuite) TestTestReportsTheEndpointAnswer() {
	hook := suite.hook("issue.created")
	suite.MockWebhookRepo.EXPECT().GetByID(suite.Ctx, hook.ID).Return(hook, nil)
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)

	result, err := suite.service.Test(suite.Ctx, suite.userID, &WebhookRefInput{WebhookID: hook.ID})
	suite.Require().NoError(err)
	suite.True(result.Delivered)
	suite.Equal(http.StatusNoContent, result.StatusCode)
	suite.Equal("webhook.test", suite.received[0].headers.Get(EventHeader))

	suite.failures = 1
	result, err = suite.service.Test(suite.Ctx, suite.userID, &WebhookRefInput{WebhookID: hook.ID})
	suite.Require().NoError(err)
	suite.False(result.Delivered)
	suite.Equal(http.StatusBadGateway, result.StatusCode)
	suite.Contains(result.Error, "502")
}

func (suite *WebhookServiceSuite) TestDeliveryRefusesPrivateAddresses() {
	suite.service.client = newClient(oauthserver.DefaultIPPolicy)
	hook := suite.hook("issue.created")
	suite.MockWebhookRepo.EXPECT().GetByID(suite.Ctx, hook.ID).Return(hook, nil)
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)

	result, err := suite.service.Test(suite.Ctx, suite.userID, &WebhookRefInput{WebhookID: hook.ID})
	suite.Require().NoError(err)
	suite.False(result.Delivered)
	suite.Contains(result.Error, "not allowed")
	suite.Empty(suite.received)
}

func TestWebhookServiceSuite(t *testing.T) {
	suite.Run(t, new(WebhookServiceSuite))
}
