package github

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidSignature(t *testing.T) {
	body := []byte(`{"zen":"Keep it logically awesome."}`)
	signed := Sign("s3cret", body)
	cases := map[string]struct {
		secret string
		header string
		body   []byte
		want   bool
	}{
		"matching signature":   {secret: "s3cret", header: signed, body: body, want: true},
		"other secret":         {secret: "other", header: signed, body: body},
		"tampered body":        {secret: "s3cret", header: signed, body: []byte(`{"zen":"x"}`)},
		"missing header":       {secret: "s3cret", header: "", body: body},
		"sha1 header":          {secret: "s3cret", header: "sha1=" + signed[len("sha256="):], body: body},
		"not hex":              {secret: "s3cret", header: "sha256=zz", body: body},
		"truncated signature":  {secret: "s3cret", header: signed[:len(signed)-2], body: body},
		"uppercase hex digest": {secret: "s3cret", header: "sha256=" + strings.ToUpper(signed[len("sha256="):]), body: body, want: true},
	}
	for name, tc := range cases {
		assert.Equal(t, tc.want, ValidSignature(tc.secret, tc.header, tc.body), name)
	}
}

func TestPullRequestDisplayState(t *testing.T) {
	cases := map[string]struct {
		pr   PullRequest
		want string
	}{
		"open":            {pr: PullRequest{State: "open"}, want: StateOpen},
		"draft":           {pr: PullRequest{State: "open", Draft: true}, want: StateDraft},
		"merged":          {pr: PullRequest{State: "closed", Merged: true}, want: StateMerged},
		"closed unmerged": {pr: PullRequest{State: "closed"}, want: StateClosed},
		"closed draft":    {pr: PullRequest{State: "closed", Draft: true}, want: StateClosed},
	}
	for name, tc := range cases {
		assert.Equal(t, tc.want, tc.pr.DisplayState(), name)
	}
}

func TestPushBranch(t *testing.T) {
	cases := map[string]struct {
		push   PushEvent
		branch string
		ok     bool
	}{
		"branch":         {push: PushEvent{Ref: "refs/heads/bbedward/unb-1"}, branch: "bbedward/unb-1", ok: true},
		"tag":            {push: PushEvent{Ref: "refs/tags/v1.0.0"}},
		"deleted branch": {push: PushEvent{Ref: "refs/heads/unb-1", Deleted: true}},
	}
	for name, tc := range cases {
		branch, ok := tc.push.Branch()
		assert.Equal(t, tc.branch, branch, name)
		assert.Equal(t, tc.ok, ok, name)
	}
}
