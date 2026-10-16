package googleauth_service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/crypto"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	system_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/system"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

const (
	testClientID     = "client-123.apps.googleusercontent.com"
	testClientSecret = "GOCSPX-test-client-secret"
	testEmail        = "grace@example.com"
	testSubject      = "google-sub-1"
)

type fakeVerifier struct {
	identities map[string]*GoogleIdentity
}

func (self *fakeVerifier) Verify(_ context.Context, clientID, rawIDToken string) (*GoogleIdentity, error) {
	identity, found := self.identities[rawIDToken]
	if clientID != testClientID || !found {
		return nil, errors.New("signature does not verify")
	}
	copied := *identity
	return &copied, nil
}

type GoogleAuthSuite struct {
	services.ServiceTestSuite
	service   *GoogleAuthService
	encryptor *crypto.Encryptor
	verifier  *fakeVerifier
	tokens    *httptest.Server
	settings  *ent.SystemSetting
	now       time.Time
	userID    uuid.UUID
	challenge string
}

func (suite *GoogleAuthSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.ExpectTx()
	var err error
	suite.encryptor, err = crypto.NewEncryptor(bytes.Repeat([]byte{7}, 32))
	suite.Require().NoError(err)
	sealed, err := suite.encryptor.Encrypt([]byte(testClientSecret))
	suite.Require().NoError(err)

	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.userID = uuid.New()
	suite.settings = &ent.SystemSetting{
		GoogleAuth:                  &schema.GoogleAuthSettings{Enabled: true, ClientID: testClientID},
		GoogleClientSecretEncrypted: sealed,
	}
	suite.MockSystemRepo.EXPECT().GetSystemSettings(mock.Anything, nil).RunAndReturn(func(context.Context, repository.TxInterface) (*ent.SystemSetting, error) {
		return suite.settings, nil
	}).Maybe()

	suite.tokens = httptest.NewServer(http.HandlerFunc(suite.tokenEndpoint))
	suite.T().Cleanup(suite.tokens.Close)
	suite.verifier = &fakeVerifier{identities: map[string]*GoogleIdentity{}}

	suite.service = NewGoogleAuthService(&config.Config{ExternalUIUrl: "https://zigzag.example.com"}, suite.MockRepo, suite.encryptor, []byte("state-secret-state-secret-state-"))
	suite.service.verifier = suite.verifier
	suite.service.endpoint.TokenURL = suite.tokens.URL
	suite.service.now = func() time.Time { return suite.now }
}

func (suite *GoogleAuthSuite) tokenEndpoint(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	verifierSum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
	pkceMatches := base64.RawURLEncoding.EncodeToString(verifierSum[:]) == suite.challenge
	validClient := r.PostForm.Get("client_id") == testClientID && r.PostForm.Get("client_secret") == testClientSecret
	validRequest := r.PostForm.Get("grant_type") == "authorization_code" && r.PostForm.Get("redirect_uri") == "https://zigzag.example.com/api/go/auth/google/callback"
	code := r.PostForm.Get("code")
	w.Header().Set("Content-Type", "application/json")
	if !pkceMatches || !validClient || !validRequest || code == "bad-code" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at", "token_type": "Bearer", "expires_in": 3600, "id_token": "id-token-" + code})
}

func (suite *GoogleAuthSuite) startLogin(redirect string) (cookie, state, nonce string) {
	start, err := suite.service.StartLogin(suite.Ctx, redirect)
	suite.Require().NoError(err)
	authURL, err := url.Parse(start.AuthURL)
	suite.Require().NoError(err)
	query := authURL.Query()
	suite.challenge = query.Get("code_challenge")
	return start.Cookie, query.Get("state"), query.Get("nonce")
}

func (suite *GoogleAuthSuite) identity(nonce string) *GoogleIdentity {
	return &GoogleIdentity{Subject: testSubject, Email: "Grace@Example.com", EmailVerified: true, Name: "Grace Hopper", Nonce: nonce}
}

func (suite *GoogleAuthSuite) finish(code string, identity *GoogleIdentity, cookie, state string) (*LoginResult, error) {
	suite.verifier.identities["id-token-"+code] = identity
	return suite.service.FinishLogin(suite.Ctx, &FinishLoginInput{Code: code, State: state, Cookie: cookie})
}

func (suite *GoogleAuthSuite) TestStartBuildsAPKCEAuthorizationURL() {
	suite.settings.GoogleAuth.HostedDomain = "example.com"
	start, err := suite.service.StartLogin(suite.Ctx, "/acme/my-issues")
	suite.Require().NoError(err)

	authURL, err := url.Parse(start.AuthURL)
	suite.Require().NoError(err)
	query := authURL.Query()
	suite.Equal("accounts.google.com", authURL.Host)
	suite.Equal("/o/oauth2/v2/auth", authURL.Path)
	suite.Equal(testClientID, query.Get("client_id"))
	suite.Equal("code", query.Get("response_type"))
	suite.Equal("openid email profile", query.Get("scope"))
	suite.Equal("https://zigzag.example.com/api/go/auth/google/callback", query.Get("redirect_uri"))
	suite.Equal("S256", query.Get("code_challenge_method"))
	suite.Len(query.Get("code_challenge"), 43)
	suite.NotEmpty(query.Get("state"))
	suite.NotEmpty(query.Get("nonce"))
	suite.Equal("example.com", query.Get("hd"))
	suite.Equal(suite.now.Add(loginStateTTL), start.ExpiresAt)
	suite.NotContains(start.Cookie, query.Get("state"), "the cookie is signed, not the bare state")
}

func (suite *GoogleAuthSuite) TestStartRefusesWhenDisabledOrIncomplete() {
	suite.settings.GoogleAuth.Enabled = false
	_, err := suite.service.StartLogin(suite.Ctx, "/")
	suite.ErrorIs(err, ErrNotEnabled)

	suite.settings.GoogleAuth.Enabled = true
	suite.settings.GoogleClientSecretEncrypted = nil
	_, err = suite.service.StartLogin(suite.Ctx, "/")
	suite.ErrorIs(err, ErrNotEnabled)
}

func (suite *GoogleAuthSuite) TestExistingUserLinksOnFirstLoginAndJoinsPendingInvitations() {
	cookie, state, nonce := suite.startLogin("/acme/my-issues")
	existing := &ent.User{ID: suite.userID, Email: testEmail, Active: true, PasswordHash: "h"}
	inv := &ent.Invitation{ID: uuid.New(), WorkspaceID: uuid.New(), Email: testEmail, Role: schema.WorkspaceRoleMember}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(existing, nil).Once()
	suite.MockUserRepo.EXPECT().LinkGoogleSubject(suite.Ctx, mock.Anything, suite.userID, testSubject).Return(true, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, mock.Anything, testEmail, suite.now).Return([]*ent.Invitation{inv}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ClaimInvitation(suite.Ctx, mock.Anything, inv.ID, suite.now).Return(true, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, inv.WorkspaceID, suite.userID, schema.WorkspaceRoleMember).Return(&ent.WorkspaceMember{}, nil).Once()

	result, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.Require().NoError(err)
	suite.Equal(suite.userID, result.User.ID)
	suite.Equal(utils.ToPtr(testSubject), result.User.GoogleSubject)
	suite.Equal("/acme/my-issues", result.Redirect)
}

func (suite *GoogleAuthSuite) TestLinkedUserSignsInWithTheSameSubject() {
	cookie, state, nonce := suite.startLogin("/")
	existing := &ent.User{ID: suite.userID, Email: testEmail, Active: true, GoogleSubject: utils.ToPtr(testSubject)}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(existing, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, mock.Anything, testEmail, suite.now).Return(nil, nil).Once()

	result, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.Require().NoError(err)
	suite.Equal(suite.userID, result.User.ID)
	suite.MockUserRepo.AssertNotCalled(suite.T(), "LinkGoogleSubject", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *GoogleAuthSuite) TestSubjectMismatchIsRefused() {
	cookie, state, nonce := suite.startLogin("/")
	existing := &ent.User{ID: suite.userID, Email: testEmail, Active: true, GoogleSubject: utils.ToPtr("another-subject")}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(existing, nil).Once()

	_, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.ErrorIs(err, ErrSubjectMismatch)
}

func (suite *GoogleAuthSuite) TestSubjectLinkedConcurrentlyIsRefused() {
	cookie, state, nonce := suite.startLogin("/")
	existing := &ent.User{ID: suite.userID, Email: testEmail, Active: true, PasswordHash: "h"}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(existing, nil).Once()
	suite.MockUserRepo.EXPECT().LinkGoogleSubject(suite.Ctx, mock.Anything, suite.userID, testSubject).Return(false, nil).Once()

	_, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.ErrorIs(err, ErrSubjectMismatch)
}

func (suite *GoogleAuthSuite) TestDeactivatedUserIsRefused() {
	cookie, state, nonce := suite.startLogin("/")
	deactivated := &ent.User{ID: suite.userID, Email: testEmail, Active: false, PasswordHash: "h", GoogleSubject: utils.ToPtr(testSubject)}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(deactivated, nil).Once()

	_, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.ErrorIs(err, ErrDeactivated)
}

func (suite *GoogleAuthSuite) TestInvitedUserIsCreatedAndJoinsWorkspaceAndTeams() {
	cookie, state, nonce := suite.startLogin("/")
	teamID := uuid.New()
	inv := &ent.Invitation{ID: uuid.New(), WorkspaceID: uuid.New(), Email: testEmail, Role: schema.WorkspaceRoleViewer, TeamIds: []uuid.UUID{teamID}}
	created := &ent.User{ID: suite.userID, Email: testEmail, Active: true, GoogleSubject: utils.ToPtr(testSubject)}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, nil, testEmail, suite.now).Return([]*ent.Invitation{inv}, nil).Once()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "grace").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, &user_repo.CreateUserInput{
		Email: testEmail, Name: "Grace Hopper", DisplayName: "grace", Active: true, GoogleSubject: utils.ToPtr(testSubject),
	}).Return(created, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, mock.MatchedBy(func(tx any) bool { return tx != nil }), testEmail, suite.now).Return([]*ent.Invitation{inv}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ClaimInvitation(suite.Ctx, mock.Anything, inv.ID, suite.now).Return(true, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, inv.WorkspaceID, suite.userID, schema.WorkspaceRoleViewer).Return(&ent.WorkspaceMember{}, nil).Once()
	suite.MockTeamRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, teamID, suite.userID, false).Return(nil).Once()

	result, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.Require().NoError(err)
	suite.Equal(suite.userID, result.User.ID)
}

func (suite *GoogleAuthSuite) TestInvitationClaimedConcurrentlyWithoutPublicSignupCreatesNothing() {
	cookie, state, nonce := suite.startLogin("/")
	inv := &ent.Invitation{ID: uuid.New(), WorkspaceID: uuid.New(), Email: testEmail, Role: schema.WorkspaceRoleMember}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, mock.Anything, testEmail, suite.now).Return([]*ent.Invitation{inv}, nil).Twice()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "grace").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).Return(&ent.User{ID: suite.userID}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ClaimInvitation(suite.Ctx, mock.Anything, inv.ID, suite.now).Return(false, nil).Once()

	_, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.ErrorIs(err, ErrNoAccount, "the transaction rolls the new account back")
}

func (suite *GoogleAuthSuite) TestPublicSignupCreatesAnAccountWithoutInvitations() {
	suite.settings.AllowPublicSignup = true
	cookie, state, nonce := suite.startLogin("/")
	identity := suite.identity(nonce)
	identity.Name = ""
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, mock.Anything, testEmail, suite.now).Return(nil, nil).Twice()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "grace").Return(&ent.User{}, nil).Once()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "grace2").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, &user_repo.CreateUserInput{
		Email: testEmail, Name: "grace2", DisplayName: "grace2", Active: true, GoogleSubject: utils.ToPtr(testSubject),
	}).Return(&ent.User{ID: suite.userID}, nil).Once()

	result, err := suite.finish("good-code", identity, cookie, state)
	suite.Require().NoError(err)
	suite.Equal(suite.userID, result.User.ID)
}

func (suite *GoogleAuthSuite) TestNoAccountWithoutInvitationOrPublicSignup() {
	cookie, state, nonce := suite.startLogin("/")
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, nil, testEmail, suite.now).Return(nil, nil).Once()

	_, err := suite.finish("good-code", suite.identity(nonce), cookie, state)
	suite.ErrorIs(err, ErrNoAccount)
	suite.MockUserRepo.AssertNotCalled(suite.T(), "Create", mock.Anything, mock.Anything, mock.Anything)
}

func (suite *GoogleAuthSuite) TestUnverifiedEmailIsRefused() {
	cookie, state, nonce := suite.startLogin("/")
	identity := suite.identity(nonce)
	identity.EmailVerified = false

	_, err := suite.finish("good-code", identity, cookie, state)
	suite.ErrorIs(err, ErrEmailNotVerified)
	suite.MockUserRepo.AssertNotCalled(suite.T(), "GetByEmail", mock.Anything, mock.Anything)
}

func (suite *GoogleAuthSuite) TestHostedDomainIsEnforced() {
	suite.settings.GoogleAuth.HostedDomain = "example.com"
	for name, hd := range map[string]string{"consumer account": "", "other workspace": "other.com"} {
		suite.Run(name, func() {
			cookie, state, nonce := suite.startLogin("/")
			identity := suite.identity(nonce)
			identity.HostedDomain = hd
			_, err := suite.finish("good-code", identity, cookie, state)
			suite.ErrorIs(err, ErrWrongDomain)
		})
	}

	cookie, state, nonce := suite.startLogin("/")
	identity := suite.identity(nonce)
	identity.HostedDomain = "Example.com"
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, testEmail).Return(&ent.User{ID: suite.userID, Email: testEmail, Active: true, GoogleSubject: utils.ToPtr(testSubject)}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ListPendingInvitationsForEmail(suite.Ctx, mock.Anything, testEmail, suite.now).Return(nil, nil).Once()
	_, err := suite.finish("good-code", identity, cookie, state)
	suite.NoError(err)
}

func (suite *GoogleAuthSuite) TestBadStateIsRefusedBeforeTalkingToGoogle() {
	cookie, state, nonce := suite.startLogin("/")
	otherCookie, _, _ := suite.startLogin("/")
	suite.verifier.identities["id-token-good-code"] = suite.identity(nonce)

	payload, err := auth.VerifySignedValue(suite.service.stateSecret, loginStatePurpose, cookie)
	suite.Require().NoError(err)
	forged := base64.RawURLEncoding.EncodeToString(bytes.Replace(payload, []byte(`"r":"/"`), []byte(`"r":"/x"`), 1)) + cookie[len(base64.RawURLEncoding.EncodeToString(payload)):]

	cases := map[string]FinishLoginInput{
		"missing cookie":         {Code: "good-code", State: state},
		"missing state":          {Code: "good-code", Cookie: cookie},
		"state of another start": {Code: "good-code", State: state, Cookie: otherCookie},
		"tampered cookie":        {Code: "good-code", State: state, Cookie: forged},
		"garbage cookie":         {Code: "good-code", State: state, Cookie: "not-a-cookie"},
	}
	for name, input := range cases {
		_, err := suite.service.FinishLogin(suite.Ctx, &input)
		suite.ErrorIs(err, ErrBadState, name)
	}

	suite.now = suite.now.Add(loginStateTTL)
	_, err = suite.service.FinishLogin(suite.Ctx, &FinishLoginInput{Code: "good-code", State: state, Cookie: cookie})
	suite.ErrorIs(err, ErrBadState, "expired")
	suite.MockUserRepo.AssertNotCalled(suite.T(), "GetByEmail", mock.Anything, mock.Anything)
}

func (suite *GoogleAuthSuite) TestGoogleErrorsAndInvalidTokensFail() {
	cookie, state, nonce := suite.startLogin("/")

	_, err := suite.service.FinishLogin(suite.Ctx, &FinishLoginInput{Error: "access_denied", State: state, Cookie: cookie})
	suite.ErrorIs(err, ErrDenied)

	_, err = suite.finish("bad-code", suite.identity(nonce), cookie, state)
	suite.ErrorIs(err, ErrExchange)

	_, err = suite.service.FinishLogin(suite.Ctx, &FinishLoginInput{Code: "unsigned-code", State: state, Cookie: cookie})
	suite.ErrorIs(err, ErrInvalidIDToken, "the verifier rejects the id token")

	_, err = suite.finish("good-code", suite.identity("another-nonce"), cookie, state)
	suite.ErrorIs(err, ErrInvalidIDToken, "the nonce must match the one sent to Google")
}

func (suite *GoogleAuthSuite) TestSettingsAreSystemAdminOnly() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(errdefs.ErrUnauthorized).Times(3)
	_, err := suite.service.GetSettings(suite.Ctx, suite.userID)
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	_, err = suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateGoogleAuthSettingsInput{ClientID: testClientID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.service.ClearSettings(suite.Ctx, suite.userID), errdefs.ErrUnauthorized)
	suite.MockSystemRepo.AssertNotCalled(suite.T(), "UpdateSystemSettings", mock.Anything, mock.Anything)
}

func (suite *GoogleAuthSuite) TestGetNeverReturnsTheSecret() {
	suite.settings.GoogleAuth.HostedDomain = "example.com"
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil).Once()
	resp, err := suite.service.GetSettings(suite.Ctx, suite.userID)
	suite.Require().NoError(err)
	suite.True(resp.Enabled)
	suite.True(resp.Configured)
	suite.Equal(testClientID, resp.ClientID)
	suite.Equal("example.com", resp.HostedDomain)
	suite.Equal("GOCSPX-tes", resp.SecretPrefix)
	suite.Equal("https://zigzag.example.com/api/go/auth/google/callback", resp.RedirectURI)
	raw, err := json.Marshal(resp)
	suite.Require().NoError(err)
	suite.NotContains(string(raw), testClientSecret)
}

func (suite *GoogleAuthSuite) TestUpdateEncryptsTheSecretAndNormalizes() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil).Once()
	suite.MockSystemRepo.EXPECT().UpdateSystemSettings(suite.Ctx, mock.Anything).RunAndReturn(func(_ context.Context, input *system_repo.SystemSettingUpdateInput) (*ent.SystemSetting, error) {
		suite.Equal(&schema.GoogleAuthSettings{Enabled: true, ClientID: testClientID, HostedDomain: "example.com"}, input.GoogleAuth)
		suite.NotContains(string(input.GoogleSecretEncrypted), "new-secret-value")
		decrypted, err := suite.encryptor.Decrypt(input.GoogleSecretEncrypted)
		suite.Require().NoError(err)
		suite.Equal("new-secret-value", string(decrypted))
		return &ent.SystemSetting{GoogleAuth: input.GoogleAuth, GoogleClientSecretEncrypted: input.GoogleSecretEncrypted}, nil
	}).Once()

	resp, err := suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateGoogleAuthSettingsInput{
		Enabled: true, ClientID: "  " + testClientID + " ", ClientSecret: utils.ToPtr(" new-secret-value "), HostedDomain: utils.ToPtr(" Example.COM "),
	})
	suite.Require().NoError(err)
	suite.Equal("new-secr", resp.SecretPrefix, "short secrets show at most half")
}

func (suite *GoogleAuthSuite) TestUpdateWithoutASecretKeepsTheStoredOne() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil).Once()
	suite.MockSystemRepo.EXPECT().UpdateSystemSettings(suite.Ctx, &system_repo.SystemSettingUpdateInput{
		GoogleAuth: &schema.GoogleAuthSettings{Enabled: true, ClientID: testClientID},
	}).Return(suite.settings, nil).Once()

	_, err := suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateGoogleAuthSettingsInput{Enabled: true, ClientID: testClientID})
	suite.NoError(err)
}

func (suite *GoogleAuthSuite) TestUpdateValidation() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil)
	cases := map[string]*UpdateGoogleAuthSettingsInput{
		"blank client id":     {ClientID: "  "},
		"empty secret":        {ClientID: testClientID, ClientSecret: utils.ToPtr(" ")},
		"domain with a path":  {ClientID: testClientID, HostedDomain: utils.ToPtr("example.com/x")},
		"email as the domain": {ClientID: testClientID, HostedDomain: utils.ToPtr("ada@example.com")},
		"bare label":          {ClientID: testClientID, HostedDomain: utils.ToPtr("localhost")},
	}
	for name, input := range cases {
		_, err := suite.service.UpdateSettings(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}

	suite.settings.GoogleClientSecretEncrypted = nil
	_, err := suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateGoogleAuthSettingsInput{Enabled: true, ClientID: testClientID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "enabling needs a stored secret")
	suite.MockSystemRepo.AssertNotCalled(suite.T(), "UpdateSystemSettings", mock.Anything, mock.Anything)
}

func (suite *GoogleAuthSuite) TestClearRemovesSettingsAndSecret() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil).Once()
	suite.MockSystemRepo.EXPECT().UpdateSystemSettings(suite.Ctx, &system_repo.SystemSettingUpdateInput{ClearGoogleAuth: true}).Return(&ent.SystemSetting{}, nil).Once()
	suite.NoError(suite.service.ClearSettings(suite.Ctx, suite.userID))
}

func (suite *GoogleAuthSuite) TestProvidersNeedEnabledClientAndSecret() {
	resp, err := suite.service.Providers(suite.Ctx)
	suite.Require().NoError(err)
	suite.True(resp.Google.Enabled)

	suite.settings.GoogleAuth.ClientID = ""
	resp, err = suite.service.Providers(suite.Ctx)
	suite.Require().NoError(err)
	suite.False(resp.Google.Enabled)

	suite.settings.GoogleAuth = nil
	resp, err = suite.service.Providers(suite.Ctx)
	suite.Require().NoError(err)
	suite.False(resp.Google.Enabled)
}

func TestGoogleAuthSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthSuite))
}

func TestSafeRedirectPath(t *testing.T) {
	cases := map[string]string{
		"":                           "/",
		"/":                          "/",
		"/acme/my-issues":            "/acme/my-issues",
		"/acme/issue/ACME-1?tab=all": "/acme/issue/ACME-1?tab=all",
		"https://evil.example":       "/",
		"//evil.example":             "/",
		"/\\evil.example":            "/",
		"\\\\evil.example":           "/",
		"javascript:alert(1)":        "/",
		"acme/my-issues":             "/",
		"/acme\r\nSet-Cookie: x=1":   "/",
		"/acme\tx":                   "/",
		"/ /evil.example":            "/",
	}
	for raw, want := range cases {
		if got := SafeRedirectPath(raw); got != want {
			t.Errorf("SafeRedirectPath(%q) = %q, want %q", raw, got, want)
		}
	}
}
