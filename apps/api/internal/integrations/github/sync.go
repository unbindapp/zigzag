package github

import (
	"context"
	"regexp"
	"strings"
)

const (
	StateReasonCompleted  = "completed"
	StateReasonNotPlanned = "not_planned"
	StateReasonDuplicate  = "duplicate"
	defaultLabelColor     = "ededed"
)

type Closure int

const (
	Open Closure = iota
	ClosedCompleted
	ClosedNotPlanned
)

func ClosureOf(state, stateReason string) Closure {
	if state != StateClosed {
		return Open
	}
	switch stateReason {
	case StateReasonNotPlanned, StateReasonDuplicate:
		return ClosedNotPlanned
	default:
		return ClosedCompleted
	}
}

var (
	loginPattern    = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9]|-[a-z0-9]){0,38}$`)
	fullNamePattern = regexp.MustCompile(`^[a-z0-9_.-]{1,39}/[a-z0-9_.-]{1,100}$`)
	zigzagPrefix    = regexp.MustCompile(`^\*\*[^*\n]+ on Zigzag\*\*\n`)
	hexColor        = regexp.MustCompile(`^[0-9a-f]{6}$`)
)

func NormalizeLogin(raw string) (string, bool) {
	login := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(raw), "@"))
	return login, len(login) <= 39 && loginPattern.MatchString(login)
}

func NormalizeFullName(raw string) (string, bool) {
	name := strings.ToLower(strings.Trim(strings.TrimSpace(raw), "/"))
	name = strings.TrimPrefix(name, "https://github.com/")
	name = strings.TrimSuffix(name, ".git")
	return name, fullNamePattern.MatchString(name) && !strings.HasSuffix(name, "/.") && !strings.HasSuffix(name, "/..")
}

func ZigzagCommentBody(authorName, body string) string {
	return "**" + strings.ReplaceAll(authorName, "*", "") + " on Zigzag**\n\n" + body
}

func GithubCommentBody(login, body string) string {
	return "**@" + login + " on GitHub**\n\n" + body
}

func FromZigzag(body string) bool {
	return zigzagPrefix.MatchString(body)
}

func ZigzagColor(githubColor string) string {
	color := strings.ToLower(strings.TrimPrefix(githubColor, "#"))
	if !hexColor.MatchString(color) {
		color = defaultLabelColor
	}
	return "#" + color
}

func GithubColor(zigzagColor string) string {
	return strings.TrimPrefix(ZigzagColor(zigzagColor), "#")
}

type Inbound struct {
	Label string
	URL   string
}

type inboundKey struct{}

func WithInbound(ctx context.Context, label, url string) context.Context {
	return context.WithValue(ctx, inboundKey{}, Inbound{Label: label, URL: url})
}

func InboundFrom(ctx context.Context) (Inbound, bool) {
	inbound, ok := ctx.Value(inboundKey{}).(Inbound)
	return inbound, ok
}
