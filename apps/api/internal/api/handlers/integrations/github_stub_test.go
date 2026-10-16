package integrations_handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

type stubCall struct {
	Method string
	Path   string
	Auth   string
	Body   map[string]any
}

type githubStub struct {
	mu     sync.Mutex
	calls  []stubCall
	server *httptest.Server
}

func (suite *GithubWebhookSuite) githubAPI() string {
	suite.api = &githubStub{}
	suite.api.server = httptest.NewServer(http.HandlerFunc(suite.api.serve))
	suite.T().Cleanup(suite.api.server.Close)
	return suite.api.server.URL
}

func (self *githubStub) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	call := stubCall{Method: r.Method, Path: r.URL.Path, Auth: r.Header.Get("Authorization")}
	_ = json.Unmarshal(raw, &call.Body)
	self.mu.Lock()
	self.calls = append(self.calls, call)
	self.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodPatch:
		_, _ = w.Write([]byte(`{"number":7,"updated_at":"2026-10-16T12:00:00Z"}`))
	case r.Method == http.MethodPost && len(r.URL.Path) > 9 && r.URL.Path[len(r.URL.Path)-9:] == "/comments":
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":9001,"body":"x","created_at":"2026-10-16T12:00:00Z"}`))
	case r.Method == http.MethodGet && len(r.URL.Path) > 8 && r.URL.Path[:8] == "/repos/u":
		_, _ = w.Write([]byte(`{"full_name":"unbindapp/unbind","has_issues":true}`))
	default:
		_, _ = w.Write([]byte(`{}`))
	}
}

func (self *githubStub) recorded() []stubCall {
	self.mu.Lock()
	defer self.mu.Unlock()
	return append([]stubCall(nil), self.calls...)
}
