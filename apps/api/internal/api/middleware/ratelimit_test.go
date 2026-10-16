package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/zigzagapp-org/zigzag-api/config"
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

func TestCredentialOperationsAreLimitedPerAddress(t *testing.T) {
	_, api := humatest.New(t)
	mw := NewMiddleware(&config.Config{}, nil, api, nil, nil)
	mw.limiter.now = func() time.Time { return time.Date(2026, 9, 18, 12, 0, 10, 0, time.UTC) }
	api.UseMiddleware(mw.RateLimitOperations)
	type empty struct{}
	ok := func(context.Context, *empty) (*empty, error) { return &empty{}, nil }
	huma.Register(api, huma.Operation{OperationID: "login", Method: http.MethodPost, Path: "/auth/login"}, ok)
	huma.Register(api, huma.Operation{OperationID: "get-setup-status", Method: http.MethodGet, Path: "/setup/status"}, ok)

	call := func(method, path, addr string) int {
		req := httptest.NewRequest(method, path, nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		api.Adapter().ServeHTTP(rec, req)
		return rec.Code
	}
	for i := range 10 {
		if code := call(http.MethodPost, "/auth/login", "10.0.0.1:1000"); code == http.StatusTooManyRequests {
			t.Fatalf("attempt %d was limited", i+1)
		}
	}
	if code := call(http.MethodPost, "/auth/login", "10.0.0.1:2000"); code != http.StatusTooManyRequests {
		t.Fatalf("the eleventh login in a minute from one address answered %d", code)
	}
	if code := call(http.MethodPost, "/auth/login", "10.0.0.2:1000"); code == http.StatusTooManyRequests {
		t.Fatal("another address has its own budget")
	}
	for range 20 {
		if code := call(http.MethodGet, "/setup/status", "10.0.0.1:1000"); code == http.StatusTooManyRequests {
			t.Fatal("operations without a limit are never limited")
		}
	}
}
