package github

import "strings"

const (
	EventPing         = "ping"
	EventPullRequest  = "pull_request"
	EventPush         = "push"
	EventIssues       = "issues"
	EventIssueComment = "issue_comment"
)

const (
	StateOpen   = "open"
	StateDraft  = "draft"
	StateMerged = "merged"
	StateClosed = "closed"
)

type Repository struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type Account struct {
	Login string `json:"login"`
}

func (self *Account) LoginOrEmpty() string {
	if self == nil {
		return ""
	}
	return self.Login
}

type Envelope struct {
	Action     string     `json:"action"`
	Repository Repository `json:"repository"`
}

func (self Envelope) Kind(event string) string {
	if self.Action == "" {
		return event
	}
	return event + "." + self.Action
}

type PullRequest struct {
	Number  int     `json:"number"`
	Title   string  `json:"title"`
	Body    string  `json:"body"`
	HTMLURL string  `json:"html_url"`
	State   string  `json:"state"`
	Draft   bool    `json:"draft"`
	Merged  bool    `json:"merged"`
	User    Account `json:"user"`
	Head    struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
}

func (self PullRequest) DisplayState() string {
	switch {
	case self.Merged:
		return StateMerged
	case self.State == "closed":
		return StateClosed
	case self.Draft:
		return StateDraft
	default:
		return StateOpen
	}
}

type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}

type PushEvent struct {
	Ref        string     `json:"ref"`
	After      string     `json:"after"`
	Deleted    bool       `json:"deleted"`
	Repository Repository `json:"repository"`
	Sender     Account    `json:"sender"`
}

func (self PushEvent) Branch() (string, bool) {
	branch, ok := strings.CutPrefix(self.Ref, "refs/heads/")
	if !ok || branch == "" || self.Deleted {
		return "", false
	}
	return branch, true
}

func BranchURL(repository Repository, branch string) string {
	return strings.TrimRight(repository.HTMLURL, "/") + "/tree/" + branch
}

type IssuesEvent struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Label      *Label     `json:"label"`
	Assignee   *Account   `json:"assignee"`
	Repository Repository `json:"repository"`
	Sender     Account    `json:"sender"`
	Changes    struct {
		NewIssue      *Issue      `json:"new_issue"`
		NewRepository *Repository `json:"new_repository"`
	} `json:"changes"`
}

type IssueCommentEvent struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Comment    Comment    `json:"comment"`
	Repository Repository `json:"repository"`
	Sender     Account    `json:"sender"`
}
