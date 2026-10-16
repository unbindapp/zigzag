package middleware

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/auth"
)

func (self *Middleware) CSRF(ctx huma.Context, next func(huma.Context)) {
	switch ctx.Method() {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		next(ctx)
		return
	}

	switch method, _ := ctx.Context().Value(authMethodKey).(string); method {
	case authMethodBearer, authMethodAPIKey:
		next(ctx)
		return
	}

	if !self.originAllowed(ctx) {
		_ = huma.WriteErr(self.api, ctx, http.StatusForbidden, "CSRF: request origin not allowed")
		return
	}

	refresh, err := huma.ReadCookie(ctx, auth.RefreshTokenCookieName(self.cfg.CookieSecure))
	if err != nil || refresh.Value == "" {
		_ = huma.WriteErr(self.api, ctx, http.StatusForbidden, "CSRF: missing session")
		return
	}

	presented := ctx.Header("X-CSRF-Token")
	if presented == "" || !auth.VerifyCSRFToken(self.tokenManager.CSRFSecret(), refresh.Value, presented) {
		_ = huma.WriteErr(self.api, ctx, http.StatusForbidden, "CSRF: token validation failed")
		return
	}

	next(ctx)
}

func refererOrigin(referer string) string {
	scheme, rest, found := strings.Cut(referer, "://")
	if !found {
		return ""
	}
	host, _, _ := strings.Cut(rest, "/")
	return scheme + "://" + host
}

func (self *Middleware) originAllowed(ctx huma.Context) bool {
	origin := ctx.Header("Origin")
	if origin == "" {
		origin = refererOrigin(ctx.Header("Referer"))
	}
	return originMatches(origin, self.allowedOrigins)
}

func originMatches(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}

	origin = strings.ToLower(origin)
	host := origin
	if _, after, found := strings.Cut(origin, "://"); found {
		host = after
	}

	for _, pattern := range allowed {
		pattern = strings.ToLower(pattern)
		switch {
		case pattern == "*":
			return true
		case strings.HasPrefix(pattern, "*."):
			if strings.HasSuffix(host, pattern[1:]) {
				return true
			}
		case pattern == origin:
			return true
		}
	}
	return false
}
