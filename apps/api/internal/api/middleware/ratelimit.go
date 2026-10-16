package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
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

func (self *RateLimiter) Limit(name string, limit int, period time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := self.now()
			windowStart := now.Truncate(period)
			key := fmt.Sprintf("%s:%s:%d", name, clientAddress(r), windowStart.Unix())
			if self.hit(key, now, windowStart.Add(period)) <= limit {
				next.ServeHTTP(w, r)
				return
			}
			retryAfter := windowStart.Add(period).Sub(now).Seconds()
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter)+1))
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
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
