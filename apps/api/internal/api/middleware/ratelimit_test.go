package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newLimitedHandler(limit int, period time.Duration) (http.Handler, *RateLimiter) {
	limiter := NewRateLimiter()
	handler := limiter.Limit("test", limit, period)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	return handler, limiter
}

func hit(handler http.Handler, addr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", nil)
	req.RemoteAddr = addr
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	handler, limiter := newLimitedHandler(3, time.Minute)
	start := time.Date(2026, 9, 18, 12, 0, 10, 0, time.UTC)
	limiter.now = func() time.Time { return start }

	for i := 0; i < 3; i++ {
		if rec := hit(handler, "10.0.0.1:1234"); rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: status = %d", i+1, rec.Code)
		}
	}
	rec := hit(handler, "10.0.0.1:9999")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("fourth request: status = %d", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "51" {
		t.Fatalf("Retry-After = %q, want seconds left in the window", got)
	}
	if rec := hit(handler, "10.0.0.2:1234"); rec.Code != http.StatusNoContent {
		t.Fatalf("another address must have its own budget, status = %d", rec.Code)
	}

	limiter.now = func() time.Time { return start.Add(time.Minute) }
	if rec := hit(handler, "10.0.0.1:1234"); rec.Code != http.StatusNoContent {
		t.Fatalf("next window: status = %d", rec.Code)
	}
}

func TestRateLimiterSeparatesNames(t *testing.T) {
	limiter := NewRateLimiter()
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	a := limiter.Limit("a", 1, time.Minute)(ok)
	b := limiter.Limit("b", 1, time.Minute)(ok)
	hit(a, "10.0.0.1:1")
	if rec := hit(b, "10.0.0.1:1"); rec.Code != http.StatusNoContent {
		t.Fatalf("limits with different names must not share a budget, status = %d", rec.Code)
	}
	if rec := hit(a, "10.0.0.1:1"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d", rec.Code)
	}
}
