package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLogin(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"octocat", "octocat", true},
		{" @Octo-Cat ", "octo-cat", true},
		{"a", "a", true},
		{"-octocat", "-octocat", false},
		{"octocat-", "octocat-", false},
		{"octo--cat", "octo--cat", false},
		{"octo_cat", "octo_cat", false},
		{"a234567890123456789012345678901234567890", "a234567890123456789012345678901234567890", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := NormalizeLogin(tc.raw)
		assert.Equal(t, tc.want, got, tc.raw)
		assert.Equal(t, tc.ok, ok, tc.raw)
	}
}

func TestNormalizeFullName(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"unbindapp/zigzag", "unbindapp/zigzag", true},
		{"UnbindApp/Zigzag", "unbindapp/zigzag", true},
		{"https://github.com/unbindapp/zigzag.git", "unbindapp/zigzag", true},
		{" unbindapp/zig.zag_2 ", "unbindapp/zig.zag_2", true},
		{"unbindapp", "unbindapp", false},
		{"unbindapp/zigzag/issues", "unbindapp/zigzag/issues", false},
		{"unbindapp/..", "unbindapp/..", false},
		{"unbind app/zigzag", "unbind app/zigzag", false},
	}
	for _, tc := range cases {
		got, ok := NormalizeFullName(tc.raw)
		assert.Equal(t, tc.want, got, tc.raw)
		assert.Equal(t, tc.ok, ok, tc.raw)
	}
}

func TestClosureOf(t *testing.T) {
	cases := []struct {
		state, reason string
		want          Closure
	}{
		{"open", "", Open},
		{"open", "reopened", Open},
		{"closed", "completed", ClosedCompleted},
		{"closed", "", ClosedCompleted},
		{"closed", "not_planned", ClosedNotPlanned},
		{"closed", "duplicate", ClosedNotPlanned},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, ClosureOf(tc.state, tc.reason), tc.state+" "+tc.reason)
	}
}

func TestLabelColors(t *testing.T) {
	for github, zigzag := range map[string]string{"d73a4a": "#d73a4a", "FBCA04": "#fbca04", "#0e8a16": "#0e8a16", "red": "#ededed", "": "#ededed"} {
		assert.Equal(t, zigzag, ZigzagColor(github), github)
	}
	for zigzag, github := range map[string]string{"#5E6AD2": "5e6ad2", "#eb5757": "eb5757", "nope": "ededed"} {
		assert.Equal(t, github, GithubColor(zigzag), zigzag)
	}
}

func TestCommentPrefixes(t *testing.T) {
	assert.Equal(t, "**Eve Doe on Zigzag**\n\nOn it", ZigzagCommentBody("Eve Doe", "On it"))
	assert.Equal(t, "**Eve on Zigzag**\n\nx", ZigzagCommentBody("*Eve*", "x"))
	assert.Equal(t, "**@octocat on GitHub**\n\nSeen it", GithubCommentBody("octocat", "Seen it"))
	assert.True(t, FromZigzag(ZigzagCommentBody("Eve", "On it")))
	assert.False(t, FromZigzag(GithubCommentBody("octocat", "Seen it")))
	assert.False(t, FromZigzag("I read **this on Zigzag**\nyesterday"))
}

func TestInboundTravelsInTheContext(t *testing.T) {
	_, ok := InboundFrom(context.Background())
	assert.False(t, ok)
	inbound, ok := InboundFrom(WithInbound(context.Background(), "Closed on GitHub", "https://github.com/a/b/issues/1"))
	assert.True(t, ok)
	assert.Equal(t, Inbound{Label: "Closed on GitHub", URL: "https://github.com/a/b/issues/1"}, inbound)
}
