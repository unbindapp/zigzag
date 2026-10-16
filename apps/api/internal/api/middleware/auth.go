package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/auth"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
)

const (
	authMethodKey    = "auth_method"
	authMethodBearer = "bearer"
	authMethodCookie = "cookie"
	authMethodAPIKey = "api_key"

	apiKeyLastUsedInterval = time.Minute
)

func (self *Middleware) Authenticate(ctx huma.Context, next func(huma.Context)) {
	if caller, ok := apictx.MCPCallerFromContext(ctx.Context()); ok {
		self.authenticateMCPCaller(ctx, next, caller)
		return
	}

	token, fromBearer, ok := extractToken(ctx, self.cfg.CookieSecure)
	if ok && fromBearer && auth.IsAPIKey(token) {
		self.authenticateAPIKey(ctx, next, token)
		return
	}
	// a leaked MCP token must never fall through to the refresh cookie and ride a browser session
	if ok && fromBearer && auth.IsOAuthToken(token) {
		_ = huma.WriteErr(self.api, ctx, http.StatusUnauthorized, "OAuth tokens are only accepted by the MCP endpoint")
		return
	}

	if ok {
		if claims, err := self.tokenManager.Verify(token); err == nil {
			method := authMethodCookie
			if fromBearer {
				method = authMethodBearer
			}
			self.proceed(ctx, next, claims, method)
			return
		}
	}

	cookie, err := huma.ReadCookie(ctx, auth.RefreshTokenCookieName(self.cfg.CookieSecure))
	if err != nil || cookie.Value == "" {
		_ = huma.WriteErr(self.api, ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	stored, err := self.repository.Oauth().GetByRefreshToken(ctx.Context(), cookie.Value)
	if err != nil || stored.Revoked || stored.ExpiresAt.Before(time.Now()) || stored.Edges.User == nil || !stored.Edges.User.Active {
		_ = huma.WriteErr(self.api, ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	user := stored.Edges.User
	accessToken, accessExpiresAt, err := self.tokenManager.MintAccessToken(user)
	if err != nil {
		log.Errorf("auth: mint access token: %v", err)
		_ = huma.WriteErr(self.api, ctx, http.StatusInternalServerError, "Failed to process user")
		return
	}

	accessCookie := auth.AccessCookie(accessToken, accessExpiresAt, self.cfg.CookieSecure)
	ctx.AppendHeader("Set-Cookie", accessCookie.String())
	csrfCookie := auth.CSRFCookie(auth.MintCSRFToken(self.tokenManager.CSRFSecret(), cookie.Value), time.Now().Add(auth.RefreshTokenTTL), self.cfg.CookieSecure)
	ctx.AppendHeader("Set-Cookie", csrfCookie.String())

	ctx = huma.WithValue(ctx, apictx.UserKey, user)
	ctx = huma.WithValue(ctx, authMethodKey, authMethodCookie)
	next(ctx)
}

func (self *Middleware) authenticateAPIKey(ctx huma.Context, next func(huma.Context), token string) {
	now := time.Now()
	key, ok := auth.VerifyAPIKey(ctx.Context(), self.repository.APIKey(), token, now)
	if !ok || !key.Edges.User.Active {
		_ = huma.WriteErr(self.api, ctx, http.StatusUnauthorized, "Invalid or expired API key")
		return
	}

	access := permissions_repo.APIKeyAccessOf(key)
	if !self.allowNarrowed(ctx, access, "an API key", "This API key is read only") {
		return
	}

	if err := self.repository.APIKey().TouchLastUsed(ctx.Context(), key.ID, now, apiKeyLastUsedInterval); err != nil {
		log.Warnf("auth: record api key use: %v", err)
	}

	self.proceedNarrowed(ctx, next, key.Edges.User, access)
}

func (self *Middleware) authenticateMCPCaller(ctx huma.Context, next func(huma.Context), caller *apictx.MCPCaller) {
	if !self.allowNarrowed(ctx, caller.Access, "the MCP server", "This connection is read only") {
		return
	}
	self.proceedNarrowed(ctx, next, caller.User, caller.Access)
}

func (self *Middleware) allowNarrowed(ctx huma.Context, access permissions_repo.APIKeyAccess, credential, readOnlyMessage string) bool {
	op := ctx.Operation()
	if oapi.IsSessionOnly(op) {
		_ = huma.WriteErr(self.api, ctx, http.StatusForbidden, "This endpoint cannot be used with "+credential)
		return false
	}
	if action, known := oapi.ActionOf(op); known && action != oapi.Read && !access.AllowsWrites() {
		_ = huma.WriteErr(self.api, ctx, http.StatusForbidden, readOnlyMessage)
		return false
	}
	if privilege, needed := oapi.PrivilegeOf(op); needed && !access.Has(privilege) {
		_ = huma.WriteErr(self.api, ctx, http.StatusForbidden, privilegeMessage(credential, privilege))
		return false
	}
	return true
}

func privilegeMessage(credential string, privilege schema.KeyPrivilege) string {
	return "This needs the " + string(privilege) + " privilege on " + credential
}

func (self *Middleware) proceedNarrowed(ctx huma.Context, next func(huma.Context), user *ent.User, access permissions_repo.APIKeyAccess) {
	ctx = huma.WithContext(ctx, permissions_repo.WithAPIKeyAccess(ctx.Context(), access))
	ctx = huma.WithValue(ctx, apictx.UserKey, user)
	ctx = huma.WithValue(ctx, authMethodKey, authMethodAPIKey)
	next(ctx)
}

func (self *Middleware) proceed(ctx huma.Context, next func(huma.Context), claims *auth.VerifiedClaims, method string) {
	user, err := self.repository.User().GetByID(ctx.Context(), claims.UserID)
	if ent.IsNotFound(err) {
		_ = huma.WriteErr(self.api, ctx, http.StatusUnauthorized, "Authentication required")
		return
	}
	if err != nil {
		log.Errorf("auth: load user: %v", err)
		_ = huma.WriteErr(self.api, ctx, http.StatusInternalServerError, "Failed to process user")
		return
	}
	if !user.Active {
		_ = huma.WriteErr(self.api, ctx, http.StatusUnauthorized, "This account is deactivated")
		return
	}

	ctx = huma.WithValue(ctx, apictx.UserKey, user)
	ctx = huma.WithValue(ctx, authMethodKey, method)
	next(ctx)
}

func extractToken(ctx huma.Context, secure bool) (token string, fromBearer bool, ok bool) {
	if authHeader := ctx.Header("Authorization"); authHeader != "" {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return "", false, false
		}
		return strings.TrimPrefix(authHeader, "Bearer "), true, true
	}

	cookie, err := huma.ReadCookie(ctx, auth.AccessTokenCookieName(secure))
	if err != nil || cookie.Value == "" {
		return "", false, false
	}
	return cookie.Value, false, true
}
