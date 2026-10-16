package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"strings"
	"testing"
	"time"

	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/api/apictx"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	mocks_repositories "github.com/zigzagapp-org/zigzag-api/mocks/repositories"
	mocks_repository_apikey "github.com/zigzagapp-org/zigzag-api/mocks/repository/apikey"
	mocks_repository_oauth "github.com/zigzagapp-org/zigzag-api/mocks/repository/oauth"
	mocks_repository_user "github.com/zigzagapp-org/zigzag-api/mocks/repository/user"
)

type sessionHarness struct {
	api     humatest.TestAPI
	tokens  *auth.TokenManager
	users   *mocks_repository_user.UserRepositoryMock
	oauth   *mocks_repository_oauth.OauthRepositoryMock
	keys    *mocks_repository_apikey.APIKeyRepositoryMock
	reached *ent.User
}

func newSessionHarness(t *testing.T) *sessionHarness {
	t.Helper()
	_, api := humatest.New(t)
	repo := mocks_repositories.NewRepositoriesMock(t)
	h := &sessionHarness{
		api:   api,
		users: mocks_repository_user.NewUserRepositoryMock(t),
		oauth: mocks_repository_oauth.NewOauthRepositoryMock(t),
		keys:  mocks_repository_apikey.NewAPIKeyRepositoryMock(t),
	}
	repo.EXPECT().User().Return(h.users).Maybe()
	repo.EXPECT().Oauth().Return(h.oauth).Maybe()
	repo.EXPECT().APIKey().Return(h.keys).Maybe()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	h.tokens = auth.NewTokenManager(key, "https://zigzag.example.com", "zigzag-api")
	mw := NewMiddleware(&config.Config{CookieSecure: false}, repo, api, h.tokens, []string{"http://localhost:5174"})

	type out struct {
		Body struct {
			OK bool `json:"ok"`
		}
	}
	grp := huma.NewGroup(api, "/v1")
	grp.UseMiddleware(mw.Authenticate)
	grp.UseMiddleware(mw.CSRF)
	oapi.Register(grp, oapi.Read, huma.Operation{OperationID: "read", Method: http.MethodGet, Path: "/read"}, func(ctx context.Context, _ *struct{}) (*out, error) {
		h.reached, _ = ctx.Value(apictx.UserKey).(*ent.User)
		return &out{}, nil
	})
	return h
}

func TestAccessCookieAuthenticates(t *testing.T) {
	h := newSessionHarness(t)
	user := &ent.User{ID: uuid.New(), Email: "ada@example.com", Active: true}
	token, _, _ := h.tokens.MintAccessToken(user)
	h.users.EXPECT().GetByID(mock.Anything, user.ID).Return(user, nil).Once()

	resp := h.api.Get("/v1/read", "Cookie: access_token="+token)
	if resp.Code != http.StatusOK || h.reached == nil || h.reached.ID != user.ID {
		t.Fatalf("status = %d, reached %v", resp.Code, h.reached)
	}
}

func TestDeactivatedUsersAreRefused(t *testing.T) {
	h := newSessionHarness(t)
	user := &ent.User{ID: uuid.New(), Email: "ada@example.com", Active: false}
	token, _, _ := h.tokens.MintAccessToken(user)
	h.users.EXPECT().GetByID(mock.Anything, user.ID).Return(user, nil).Once()

	if resp := h.api.Get("/v1/read", "Authorization: Bearer "+token); resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.Code)
	}

	h.oauth.EXPECT().GetByRefreshToken(mock.Anything, "refresh").Return(&ent.Oauth2Token{ExpiresAt: time.Now().Add(time.Hour), Edges: ent.Oauth2TokenEdges{User: user}}, nil).Once()
	if resp := h.api.Get("/v1/read", "Cookie: refresh_token=refresh"); resp.Code != http.StatusUnauthorized {
		t.Fatalf("refresh for a deactivated user: status = %d, want 401", resp.Code)
	}

	key, _ := auth.NewAPIKey()
	h.keys.EXPECT().GetByTokenHash(mock.Anything, key.Hash).Return(&ent.APIKey{TokenHash: key.Hash, Role: "admin", FullAccess: true, Edges: ent.APIKeyEdges{User: user}}, nil).Once()
	if resp := h.api.Get("/v1/read", "Authorization: Bearer "+key.Token); resp.Code != http.StatusUnauthorized {
		t.Fatalf("api key of a deactivated user: status = %d, want 401", resp.Code)
	}
	if h.reached != nil {
		t.Fatal("handler ran for a deactivated user")
	}
}

func TestRefreshCookieMintsANewAccessToken(t *testing.T) {
	h := newSessionHarness(t)
	user := &ent.User{ID: uuid.New(), Email: "ada@example.com", Active: true}
	h.oauth.EXPECT().GetByRefreshToken(mock.Anything, "refresh").Return(&ent.Oauth2Token{ExpiresAt: time.Now().Add(time.Hour), Edges: ent.Oauth2TokenEdges{User: user}}, nil).Once()

	resp := h.api.Get("/v1/read", "Cookie: refresh_token=refresh")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", resp.Code, resp.Body.String())
	}
	cookies := strings.Join(resp.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(cookies, "access_token=") || !strings.Contains(cookies, "csrf_token="+auth.MintCSRFToken(h.tokens.CSRFSecret(), "refresh")) {
		t.Fatalf("refresh must set a new access cookie and the csrf cookie, got %s", cookies)
	}
}

func TestRevokedOrExpiredRefreshTokensAreRefused(t *testing.T) {
	h := newSessionHarness(t)
	user := &ent.User{ID: uuid.New(), Active: true}
	h.oauth.EXPECT().GetByRefreshToken(mock.Anything, "revoked").Return(&ent.Oauth2Token{Revoked: true, ExpiresAt: time.Now().Add(time.Hour), Edges: ent.Oauth2TokenEdges{User: user}}, nil).Once()
	h.oauth.EXPECT().GetByRefreshToken(mock.Anything, "expired").Return(&ent.Oauth2Token{ExpiresAt: time.Now().Add(-time.Minute), Edges: ent.Oauth2TokenEdges{User: user}}, nil).Once()
	for _, token := range []string{"revoked", "expired"} {
		if resp := h.api.Get("/v1/read", "Cookie: refresh_token="+token); resp.Code != http.StatusUnauthorized {
			t.Fatalf("%s: status = %d, want 401", token, resp.Code)
		}
	}
}
