package auth_handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	user_service "github.com/zigzagapp-org/zigzag-api/internal/services/user"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "login",
		Summary:     "Login",
		Description: "Sign in with email and password. On success, sets the session cookies.",
		Path:        "/login",
		Method:      http.MethodPost,
		Errors:      []int{http.StatusUnauthorized},
	}, handlers.Login, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "logout",
		Summary:     "Logout",
		Description: "Revoke the current refresh token and clear the session cookies.",
		Path:        "/logout",
		Method:      http.MethodPost,
	}, handlers.Logout, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "request-password-reset",
		Summary:     "Request Password Reset",
		Description: "Email a one-time reset link when the account exists and mail is configured. Always answers success.",
		Path:        "/request-password-reset",
		Method:      http.MethodPost,
	}, handlers.RequestPasswordReset, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "reset-password",
		Summary:     "Reset Password",
		Description: "Set a new password with the token from a reset link. Signs out every existing session.",
		Path:        "/reset-password",
		Method:      http.MethodPost,
	}, handlers.ResetPassword, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-auth-providers",
		Summary:     "Get Auth Providers",
		Description: "List the sign-in methods offered besides email and password.",
		Path:        "/providers",
		Method:      http.MethodGet,
	}, handlers.GetAuthProviders, oapi.Public, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID:   "start-google-login",
		Summary:       "Start Google Login",
		Description:   "Redirect the browser to Google to sign in. Answers 302 to Google, or to the sign-in page when Google sign-in is off.",
		Path:          "/google/start",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusFound,
	}, handlers.StartGoogleLogin, oapi.Public, oapi.SessionOnly, oapi.NoMCP("browser redirect flow"))

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID:   "finish-google-login",
		Summary:       "Finish Google Login",
		Description:   "Google redirects here after sign-in. Sets the session cookies and redirects into the app, or back to the sign-in page with an error.",
		Path:          "/google/callback",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusFound,
	}, handlers.FinishGoogleLogin, oapi.Public, oapi.SessionOnly, oapi.NoMCP("browser redirect flow"))
}

type LoginInput struct {
	Body struct {
		Email    string `json:"email" format:"email" required:"true" maxLength:"254" doc:"Account email"`
		Password string `json:"password" required:"true" minLength:"1" maxLength:"72" doc:"Account password"`
	}
}

type SessionUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type SessionResponse struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Data SessionUser `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Login(ctx context.Context, input *LoginInput) (*SessionResponse, error) {
	user, err := self.srv.UserService.Login(ctx, input.Body.Email, input.Body.Password)
	if errors.Is(err, user_service.ErrInvalidCredentials) {
		return nil, huma.Error401Unauthorized("Invalid email or password")
	}
	if err != nil {
		return nil, oapi.MapError(err)
	}
	cookies, err := self.srv.IssueSession(ctx, user)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SessionResponse{SetCookie: cookies}
	resp.Body.Data = SessionUser{ID: user.ID.String(), Email: user.Email}
	return resp, nil
}

type LogoutInput struct {
	RefreshTokenSecure http.Cookie `cookie:"__Host-refresh_token"`
	RefreshToken       http.Cookie `cookie:"refresh_token"`
}

type SuccessResponse struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Data models.SuccessResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) Logout(ctx context.Context, input *LogoutInput) (*SuccessResponse, error) {
	refreshToken := input.RefreshTokenSecure.Value
	if refreshToken == "" {
		refreshToken = input.RefreshToken.Value
	}
	if err := self.srv.UserService.Logout(ctx, refreshToken); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &SuccessResponse{SetCookie: auth.ClearedSessionCookies(self.srv.Cfg.CookieSecure)}
	resp.Body.Data.Success = true
	return resp, nil
}

type RequestPasswordResetInput struct {
	Body struct {
		Email string `json:"email" required:"true" maxLength:"254" doc:"Email of the account"`
	}
}

type PlainSuccessResponse struct {
	Body struct {
		Data models.SuccessResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) RequestPasswordReset(ctx context.Context, input *RequestPasswordResetInput) (*PlainSuccessResponse, error) {
	if err := self.srv.UserService.RequestPasswordReset(ctx, input.Body.Email); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &PlainSuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}

type ResetPasswordInput struct {
	Body *user_service.ResetPasswordInput `required:"true"`
}

func (self *HandlerGroup) ResetPassword(ctx context.Context, input *ResetPasswordInput) (*PlainSuccessResponse, error) {
	if err := self.srv.UserService.ResetPassword(ctx, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &PlainSuccessResponse{}
	resp.Body.Data.Success = true
	return resp, nil
}
