package github

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushOnDefaultBranch(t *testing.T) {
	cases := []struct {
		name  string
		event PushEvent
		want  bool
	}{
		{name: "default branch", event: PushEvent{Ref: "refs/heads/master", Repository: Repository{DefaultBranch: "master"}}, want: true},
		{name: "feature branch", event: PushEvent{Ref: "refs/heads/unb-12", Repository: Repository{DefaultBranch: "master"}}},
		{name: "tag named like the default branch", event: PushEvent{Ref: "refs/tags/master", Repository: Repository{DefaultBranch: "master"}}},
		{name: "deleted default branch", event: PushEvent{Ref: "refs/heads/master", Deleted: true, Repository: Repository{DefaultBranch: "master"}}},
		{name: "unknown default branch", event: PushEvent{Ref: "refs/heads/master"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.event.OnDefaultBranch())
		})
	}
}

func TestPushCommitDisplay(t *testing.T) {
	long := strings.Repeat("a", 80)
	cases := []struct {
		name    string
		commit  PushCommit
		sha     string
		subject string
	}{
		{name: "multi-line message", commit: PushCommit{ID: "7f3d2a9c1e4b5f60718293a4b5c6d7e8f9012345", Message: "Cap websocket frames\n\nFixes UNB-5"}, sha: "7f3d2a9", subject: "Cap websocket frames"},
		{name: "crlf message", commit: PushCommit{ID: "abc", Message: "Cap frames\r\n\r\nbody"}, sha: "abc", subject: "Cap frames"},
		{name: "long subject", commit: PushCommit{ID: "7f3d2a9c", Message: long + "\nbody"}, sha: "7f3d2a9", subject: long[:72]},
		{name: "long multibyte subject", commit: PushCommit{ID: "7f3d2a9c", Message: strings.Repeat("é", 80)}, sha: "7f3d2a9", subject: strings.Repeat("é", 72)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.sha, tc.commit.ShortSHA())
			assert.Equal(t, tc.subject, tc.commit.Subject())
		})
	}
}

func TestPushCommitAuthorAndURL(t *testing.T) {
	repository := Repository{FullName: "unbindapp/unbind", HTMLURL: "https://github.com/unbindapp/unbind/"}
	commit := PushCommit{ID: "7f3d2a9c"}
	commit.Author.Name = "Mona Lisa Octocat"
	assert.Equal(t, "Mona Lisa Octocat", commit.AuthorName())
	assert.Equal(t, "https://github.com/unbindapp/unbind/commit/7f3d2a9c", commit.HTMLURL(repository))

	commit.Author.Username = "bbedward"
	commit.URL = "https://github.com/unbindapp/unbind/commit/7f3d2a9cfull"
	assert.Equal(t, "bbedward", commit.AuthorName())
	assert.Equal(t, commit.URL, commit.HTMLURL(repository))
}
