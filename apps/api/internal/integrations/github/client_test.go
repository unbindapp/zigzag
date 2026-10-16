package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, *[]time.Duration) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := NewClient(server.URL)
	waits := &[]time.Duration{}
	client.sleep = func(_ context.Context, wait time.Duration) error {
		*waits = append(*waits, wait)
		return nil
	}
	return client, waits
}

func TestClientSendsTheTokenAndPayload(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/repos/unbindapp/zigzag/issues/12", r.URL.Path)
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		var body map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, map[string]any{"state": "closed", "state_reason": "not_planned"}, body)
		_, _ = w.Write([]byte(`{"number":12,"updated_at":"2026-10-16T12:00:00Z"}`))
	})
	issue, err := client.UpdateIssue(context.Background(), "secret-token", "unbindapp/zigzag", 12, &IssuePatch{State: "closed", StateReason: "not_planned"})
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC), issue.UpdatedAt)
}

func TestClientRetriesServerErrorsWithBackoff(t *testing.T) {
	var attempts atomic.Int32
	client, waits := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"id":7}`))
	})
	comment, err := client.CreateComment(context.Background(), "t", "unbindapp/zigzag", 12, "hi")
	require.NoError(t, err)
	assert.Equal(t, int64(7), comment.ID)
	assert.Equal(t, []time.Duration{time.Second, 2 * time.Second}, *waits)
}

func TestClientGivesUpAfterThreeRetries(t *testing.T) {
	var attempts atomic.Int32
	client, waits := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"message":"Service Unavailable"}`))
	})
	err := client.AddLabels(context.Background(), "t", "unbindapp/zigzag", 12, []string{"bug"})
	assert.EqualError(t, err, "GitHub answered 503: Service Unavailable")
	assert.Equal(t, int32(4), attempts.Load())
	assert.Len(t, *waits, 3)
}

func TestClientWaitsOutSecondaryRateLimits(t *testing.T) {
	var attempts atomic.Int32
	client, waits := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.Header().Set("Retry-After", "7")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"You have exceeded a secondary rate limit"}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, client.AddAssignees(context.Background(), "t", "unbindapp/zigzag", 12, []string{"eve"}))
	assert.Equal(t, []time.Duration{7 * time.Second}, *waits)
}

func TestClientDoesNotRetryPlainForbiddenOrNotFound(t *testing.T) {
	var attempts atomic.Int32
	client, waits := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Resource not accessible by personal access token"}`))
	})
	_, err := client.UpdateIssue(context.Background(), "t", "unbindapp/zigzag", 12, &IssuePatch{})
	assert.EqualError(t, err, "GitHub answered 403: Resource not accessible by personal access token")
	assert.NoError(t, client.RemoveLabel(context.Background(), "t", "unbindapp/zigzag", 12, "bug"), "a label that is already gone is fine")
	assert.NoError(t, client.DeleteComment(context.Background(), "t", "unbindapp/zigzag", 9))
	assert.Equal(t, int32(3), attempts.Load())
	assert.Empty(t, *waits)
}

func TestClientFollowsTheLinkHeader(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/unbindapp/zigzag/issues", r.URL.Path)
		assert.Equal(t, "all", r.URL.Query().Get("state"))
		assert.Equal(t, "100", r.URL.Query().Get("per_page"))
		if r.URL.Query().Get("page") == "1" {
			w.Header().Set("Link", `<https://api.github.com/repositories/1/issues?state=all&per_page=100&page=2>; rel="next", <https://api.github.com/repositories/1/issues?state=all&per_page=100&page=3>; rel="last"`)
			_, _ = w.Write([]byte(`[{"number":1,"node_id":"I_1"},{"number":2,"node_id":"PR_2","pull_request":{"url":"x"}}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})
	issues, next, err := client.ListIssues(context.Background(), "t", "unbindapp/zigzag", "all", 1)
	require.NoError(t, err)
	assert.Equal(t, 2, next)
	require.Len(t, issues, 2)
	assert.False(t, issues[0].IsPullRequest())
	assert.True(t, issues[1].IsPullRequest())
	_, next, err = client.ListIssues(context.Background(), "t", "unbindapp/zigzag", "all", 2)
	require.NoError(t, err)
	assert.Zero(t, next)
}

func TestEnsureLabelCreatesOnlyMissingLabels(t *testing.T) {
	var created []map[string]any
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/unbindapp/zigzag/labels/needs triage":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"name":"bug"}`))
		case r.Method == http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			created = append(created, body)
			w.WriteHeader(http.StatusCreated)
		}
	})
	require.NoError(t, client.EnsureLabel(context.Background(), "t", "unbindapp/zigzag", "bug", "d73a4a"))
	require.NoError(t, client.EnsureLabel(context.Background(), "t", "unbindapp/zigzag", "needs triage", "fbca04"))
	assert.Equal(t, []map[string]any{{"name": "needs triage", "color": "fbca04"}}, created)
}
