package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type RateLimiter struct {
	mu      sync.Mutex
	windows map[string]*window
	now     func() time.Time
}

type window struct {
	count     int
	expiresAt time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{windows: map[string]*window{}, now: time.Now}
}

func (self *RateLimiter) allow(name, address string, limit int, period time.Duration) (bool, string) {
	now := self.now()
	windowStart := now.Truncate(period)
	key := fmt.Sprintf("%s:%s:%d", name, address, windowStart.Unix())
	if self.hit(key, now, windowStart.Add(period)) <= limit {
		return true, ""
	}
	retryAfter := windowStart.Add(period).Sub(now).Seconds()
	return false, strconv.Itoa(int(retryAfter) + 1)
}

func (self *RateLimiter) Limit(name string, limit int, period time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter := self.allow(name, clientAddress(r), limit, period)
			if allowed {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Retry-After", retryAfter)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":             "rate_limited",
				"error_description": "too many requests, try again later",
			})
		})
	}
}

func (self *RateLimiter) hit(key string, now, expiresAt time.Time) int {
	self.mu.Lock()
	defer self.mu.Unlock()
	if len(self.windows) > 10000 {
		for k, w := range self.windows {
			if !now.Before(w.expiresAt) {
				delete(self.windows, k)
			}
		}
	}
	current, ok := self.windows[key]
	if !ok || !now.Before(current.expiresAt) {
		current = &window{expiresAt: expiresAt}
		self.windows[key] = current
	}
	current.count++
	return current.count
}

func clientAddress(r *http.Request) string {
	return hostOf(r.RemoteAddr)
}

func hostOf(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

const credentialAttemptsPerMinute = 10

var limitedOperations = map[string]int{
	"login":                  credentialAttemptsPerMinute,
	"request-password-reset": credentialAttemptsPerMinute,
	"accept-invitation":      credentialAttemptsPerMinute,
}

func (self *Middleware) RateLimitOperations(ctx huma.Context, next func(huma.Context)) {
	limit, limited := limitedOperations[ctx.Operation().OperationID]
	if !limited {
		next(ctx)
		return
	}
	allowed, retryAfter := self.limiter.allow(ctx.Operation().OperationID, hostOf(ctx.RemoteAddr()), limit, time.Minute)
	if allowed {
		next(ctx)
		return
	}
	ctx.SetHeader("Retry-After", retryAfter)
	_ = huma.WriteErr(self.api, ctx, http.StatusTooManyRequests, "Too many attempts from this address; try again in a minute")
}
