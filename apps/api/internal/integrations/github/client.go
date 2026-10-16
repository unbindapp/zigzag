package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAPIURL  = "https://api.github.com"
	PageSize       = 100
	maxRetries     = 3
	maxRetryWait   = time.Minute
	firstRetryWait = time.Second
	maxErrorBody   = 4096
)

type APIError struct {
	Status  int
	Message string
}

func (self *APIError) Error() string {
	return fmt.Sprintf("GitHub answered %d: %s", self.Status, self.Message)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && (apiErr.Status == http.StatusNotFound || apiErr.Status == http.StatusGone)
}

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Issue struct {
	ID          int64           `json:"id"`
	NodeID      string          `json:"node_id"`
	Number      int             `json:"number"`
	Title       string          `json:"title"`
	Body        string          `json:"body"`
	State       string          `json:"state"`
	StateReason string          `json:"state_reason"`
	HTMLURL     string          `json:"html_url"`
	User        Account         `json:"user"`
	Labels      []Label         `json:"labels"`
	Assignee    *Account        `json:"assignee"`
	Comments    int             `json:"comments"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	PullRequest json.RawMessage `json:"pull_request,omitempty"`
}

func (self Issue) IsPullRequest() bool {
	return len(self.PullRequest) > 0 && string(self.PullRequest) != "null"
}

type Comment struct {
	ID        int64     `json:"id"`
	NodeID    string    `json:"node_id"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	User      Account   `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IssuePatch struct {
	Title       *string `json:"title,omitempty"`
	Body        *string `json:"body,omitempty"`
	State       string  `json:"state,omitempty"`
	StateReason string  `json:"state_reason,omitempty"`
}

type RepositoryInfo struct {
	FullName  string `json:"full_name"`
	HTMLURL   string `json:"html_url"`
	HasIssues bool   `json:"has_issues"`
}

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f client.go -i ClientInterface -p github -s Client -o client_iface.go
type Client struct {
	baseURL string
	http    *http.Client
	sleep   func(context.Context, time.Duration) error
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultAPIURL
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 30 * time.Second}, sleep: sleepContext}
}

func (self *Client) GetRepository(ctx context.Context, token, fullName string) (*RepositoryInfo, error) {
	var repository RepositoryInfo
	return &repository, self.do(ctx, token, http.MethodGet, repoPath(fullName), nil, &repository, nil)
}

func (self *Client) UpdateIssue(ctx context.Context, token, fullName string, number int, patch *IssuePatch) (*Issue, error) {
	var issue Issue
	return &issue, self.do(ctx, token, http.MethodPatch, issuePath(fullName, number), patch, &issue, nil)
}

func (self *Client) EnsureLabel(ctx context.Context, token, fullName, name, color string) error {
	err := self.do(ctx, token, http.MethodGet, repoPath(fullName)+"/labels/"+url.PathEscape(name), nil, nil, nil)
	if !IsNotFound(err) {
		return err
	}
	err = self.do(ctx, token, http.MethodPost, repoPath(fullName)+"/labels", Label{Name: name, Color: color}, nil, nil)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
		return nil
	}
	return err
}

func (self *Client) AddLabels(ctx context.Context, token, fullName string, number int, names []string) error {
	return self.do(ctx, token, http.MethodPost, issuePath(fullName, number)+"/labels", map[string][]string{"labels": names}, nil, nil)
}

func (self *Client) RemoveLabel(ctx context.Context, token, fullName string, number int, name string) error {
	err := self.do(ctx, token, http.MethodDelete, issuePath(fullName, number)+"/labels/"+url.PathEscape(name), nil, nil, nil)
	if IsNotFound(err) {
		return nil
	}
	return err
}

func (self *Client) AddAssignees(ctx context.Context, token, fullName string, number int, logins []string) error {
	return self.do(ctx, token, http.MethodPost, issuePath(fullName, number)+"/assignees", map[string][]string{"assignees": logins}, nil, nil)
}

func (self *Client) RemoveAssignees(ctx context.Context, token, fullName string, number int, logins []string) error {
	return self.do(ctx, token, http.MethodDelete, issuePath(fullName, number)+"/assignees", map[string][]string{"assignees": logins}, nil, nil)
}

func (self *Client) CreateComment(ctx context.Context, token, fullName string, number int, body string) (*Comment, error) {
	var comment Comment
	return &comment, self.do(ctx, token, http.MethodPost, issuePath(fullName, number)+"/comments", map[string]string{"body": body}, &comment, nil)
}

func (self *Client) UpdateComment(ctx context.Context, token, fullName string, commentID int64, body string) (*Comment, error) {
	var comment Comment
	return &comment, self.do(ctx, token, http.MethodPatch, commentPath(fullName, commentID), map[string]string{"body": body}, &comment, nil)
}

func (self *Client) DeleteComment(ctx context.Context, token, fullName string, commentID int64) error {
	err := self.do(ctx, token, http.MethodDelete, commentPath(fullName, commentID), nil, nil, nil)
	if IsNotFound(err) {
		return nil
	}
	return err
}

func (self *Client) ListIssues(ctx context.Context, token, fullName, state string, page int) ([]Issue, int, error) {
	query := url.Values{"state": {state}, "per_page": {strconv.Itoa(PageSize)}, "page": {strconv.Itoa(max(page, 1))}, "sort": {"created"}, "direction": {"asc"}}
	var issues []Issue
	var next int
	err := self.do(ctx, token, http.MethodGet, repoPath(fullName)+"/issues?"+query.Encode(), nil, &issues, &next)
	return issues, next, err
}

func (self *Client) ListComments(ctx context.Context, token, fullName string, number, page int) ([]Comment, int, error) {
	query := url.Values{"per_page": {strconv.Itoa(PageSize)}, "page": {strconv.Itoa(max(page, 1))}}
	var comments []Comment
	var next int
	err := self.do(ctx, token, http.MethodGet, issuePath(fullName, number)+"/comments?"+query.Encode(), nil, &comments, &next)
	return comments, next, err
}

func (self *Client) do(ctx context.Context, token, method, path string, body, out any, nextPage *int) error {
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = encoded
	}
	for attempt := 0; ; attempt++ {
		resp, err := self.send(ctx, token, method, path, payload)
		if err != nil {
			return err
		}
		wait, retry := retryAfter(resp, attempt)
		if retry && attempt < maxRetries {
			_ = resp.Body.Close()
			if err := self.sleep(ctx, wait); err != nil {
				return err
			}
			continue
		}
		return decode(resp, out, nextPage)
	}
}

func (self *Client) send(ctx context.Context, token, method, path string, payload []byte) (*http.Response, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, self.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "zigzag")
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return self.http.Do(req)
}

func decode(resp *http.Response, out any, nextPage *int) error {
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return &APIError{Status: resp.StatusCode, Message: errorMessage(raw)}
	}
	if nextPage != nil {
		*nextPage = nextPageOf(resp.Header.Get("Link"))
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func errorMessage(raw []byte) string {
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &parsed) == nil && parsed.Message != "" {
		return parsed.Message
	}
	return strings.TrimSpace(string(raw))
}

func retryAfter(resp *http.Response, attempt int) (time.Duration, bool) {
	switch {
	case resp.StatusCode >= 500, resp.StatusCode == http.StatusTooManyRequests:
	case resp.StatusCode == http.StatusForbidden && (resp.Header.Get("Retry-After") != "" || resp.Header.Get("X-RateLimit-Remaining") == "0"):
	default:
		return 0, false
	}
	if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds >= 0 {
		return min(time.Duration(seconds)*time.Second, maxRetryWait), true
	}
	if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return min(max(time.Until(time.Unix(reset, 0)), 0), maxRetryWait), true
	}
	return firstRetryWait << attempt, true
}

func nextPageOf(link string) int {
	for part := range strings.SplitSeq(link, ",") {
		target, rel, ok := strings.Cut(part, ";")
		if !ok || !strings.Contains(rel, `rel="next"`) {
			continue
		}
		parsed, err := url.Parse(strings.Trim(strings.TrimSpace(target), "<>"))
		if err != nil {
			return 0
		}
		page, _ := strconv.Atoi(parsed.Query().Get("page"))
		return page
	}
	return 0
}

func sleepContext(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func repoPath(fullName string) string {
	owner, name, _ := strings.Cut(fullName, "/")
	return "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
}

func issuePath(fullName string, number int) string {
	return repoPath(fullName) + "/issues/" + strconv.Itoa(number)
}

func commentPath(fullName string, commentID int64) string {
	return repoPath(fullName) + "/issues/comments/" + strconv.FormatInt(commentID, 10)
}
