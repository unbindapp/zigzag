package auth_handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	googleauth_service "github.com/zigzagapp-org/zigzag-api/internal/services/googleauth"
)

const (
	signInErrorNoAccount    = "no_account"
	signInErrorGoogleFailed = "google_failed"
)

type AuthProvidersResponse struct {
	Body struct {
		Data *models.AuthProvidersResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) GetAuthProviders(ctx context.Context, _ *server.EmptyInput) (*AuthProvidersResponse, error) {
	providers, err := self.srv.GoogleAuthService.Providers(ctx)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &AuthProvidersResponse{}
	resp.Body.Data = providers
	return resp, nil
}

type RedirectResponse struct {
	Status       int
	Location     string        `header:"Location"`
	CacheControl string        `header:"Cache-Control"`
	SetCookie    []http.Cookie `header:"Set-Cookie"`
}

type StartGoogleLoginInput struct {
	Redirect string `query:"redirect" required:"false" maxLength:"2048" doc:"App path to return to after signing in, for example /acme/my-issues"`
}

func (self *HandlerGroup) StartGoogleLogin(ctx context.Context, input *StartGoogleLoginInput) (*RedirectResponse, error) {
	start, err := self.srv.GoogleAuthService.StartLogin(ctx, input.Redirect)
	if err != nil {
		log.Warn("Google sign-in could not start", "err", err)
		return redirect(self.signInErrorURL(signInErrorGoogleFailed)), nil
	}
	return redirect(start.AuthURL, auth.GoogleLoginCookie(start.Cookie, start.ExpiresAt, self.srv.Cfg.CookieSecure)), nil
}

type FinishGoogleLoginInput struct {
	Code              string      `query:"code" required:"false" maxLength:"2048" doc:"Authorization code from Google"`
	State             string      `query:"state" required:"false" maxLength:"256" doc:"State echoed back by Google"`
	Error             string      `query:"error" required:"false" maxLength:"256" doc:"Error from Google, for example access_denied"`
	StateCookieSecure http.Cookie `cookie:"__Host-google_login"`
	StateCookie       http.Cookie `cookie:"google_login"`
}

func (self *HandlerGroup) FinishGoogleLogin(ctx context.Context, input *FinishGoogleLoginInput) (*RedirectResponse, error) {
	cookie := input.StateCookieSecure.Value
	if !self.srv.Cfg.CookieSecure {
		cookie = input.StateCookie.Value
	}
	result, err := self.srv.GoogleAuthService.FinishLogin(ctx, &googleauth_service.FinishLoginInput{
		Code:   input.Code,
		State:  input.State,
		Error:  input.Error,
		Cookie: cookie,
	})
	cleared := auth.ClearedGoogleLoginCookie(self.srv.Cfg.CookieSecure)
	if errors.Is(err, googleauth_service.ErrNoAccount) {
		return redirect(self.signInErrorURL(signInErrorNoAccount), cleared), nil
	}
	if err != nil {
		log.Warn("Google sign-in failed", "err", err)
		return redirect(self.signInErrorURL(signInErrorGoogleFailed), cleared), nil
	}
	cookies, err := self.srv.IssueSession(ctx, result.User)
	if err != nil {
		log.Error("Failed to issue a session after Google sign-in", "user_id", result.User.ID, "err", err)
		return redirect(self.signInErrorURL(signInErrorGoogleFailed), cleared), nil
	}
	return redirect(self.srv.Cfg.ExternalUIUrl+result.Redirect, append(cookies, cleared)...), nil
}

func redirect(location string, cookies ...http.Cookie) *RedirectResponse {
	return &RedirectResponse{Status: http.StatusFound, Location: location, CacheControl: "no-store", SetCookie: cookies}
}

func (self *HandlerGroup) signInErrorURL(code string) string {
	return self.srv.Cfg.ExternalUIUrl + "/sign-in?" + url.Values{"error": {code}}.Encode()
}
