package github_service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
)

const (
	SourceTypePullRequest = "github"
	SourceTypeBranch      = "github_branch"
)

var ErrInvalidSignature = errors.New("the X-Hub-Signature-256 header does not match the workspace's GitHub secret")

var (
	openStates  = []schema.StateType{schema.StateTypeTriage, schema.StateTypeBacklog, schema.StateTypeUnstarted}
	mergeStates = []schema.StateType{schema.StateTypeTriage, schema.StateTypeBacklog, schema.StateTypeUnstarted, schema.StateTypeStarted}
)

type Delivery struct {
	WorkspaceID uuid.UUID
	Event       string
	Signature   string
	Body        []byte
}

type move int

const (
	stay move = iota
	moveOpened
	moveMerged
)

type linkPlan struct {
	label      string
	attachment issue_repo.LinkedAttachmentInput
	replaceURL string
	move       move
	skip       func(*ent.Attachment) bool
}

func (self *GithubService) HandleWebhook(ctx context.Context, delivery *Delivery) (bool, error) {
	integration, err := self.repo.Github().GetByWorkspace(ctx, delivery.WorkspaceID)
	if ent.IsNotFound(err) {
		return false, ErrInvalidSignature
	}
	if err != nil {
		return false, err
	}
	secret, err := self.encryptor.Decrypt(integration.SecretEncrypted)
	if err != nil {
		return false, err
	}
	if !github.ValidSignature(string(secret), delivery.Signature, delivery.Body) {
		return false, ErrInvalidSignature
	}

	var envelope github.Envelope
	if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
		return false, nil
	}
	if err := self.repo.Github().RecordDelivery(ctx, integration.ID, envelope.Repository.FullName, envelope.Kind(delivery.Event), self.now()); err != nil {
		return false, err
	}

	ctx = permissions_repo.WithSystemCaller(ctx)
	switch delivery.Event {
	case github.EventPing:
		return true, nil
	case github.EventPullRequest:
		var event github.PullRequestEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return false, nil
		}
		return self.handlePullRequest(ctx, delivery.WorkspaceID, &event)
	case github.EventPush:
		var event github.PushEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return false, nil
		}
		return self.handlePush(ctx, delivery.WorkspaceID, &event)
	case github.EventIssues:
		var event github.IssuesEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return false, nil
		}
		return self.handleIssues(ctx, integration, &event)
	case github.EventIssueComment:
		var event github.IssueCommentEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return false, nil
		}
		return self.handleIssueComment(ctx, integration, &event)
	default:
		return false, nil
	}
}

func (self *GithubService) handlePullRequest(ctx context.Context, workspaceID uuid.UUID, event *github.PullRequestEvent) (bool, error) {
	pr := event.PullRequest
	state := pr.DisplayState()
	number := strconv.Itoa(pr.Number)
	plan := linkPlan{
		attachment: issue_repo.LinkedAttachmentInput{
			Title:      "#" + number + " " + pr.Title,
			URL:        pr.HTMLURL,
			SourceType: SourceTypePullRequest,
			Metadata: map[string]any{
				"repo": event.Repository.FullName, "number": pr.Number, "state": state,
				"branch": pr.Head.Ref, "author": pr.User.Login, "head_sha": pr.Head.SHA,
			},
		},
		replaceURL: github.BranchURL(event.Repository, pr.Head.Ref),
	}
	opens := state == github.StateOpen
	switch event.Action {
	case "opened":
		plan.label = "Pull request #" + number + " opened"
		if state == github.StateDraft {
			plan.label = "Draft pull request #" + number + " opened"
		}
	case "reopened":
		plan.label = "Pull request #" + number + " reopened"
	case "ready_for_review":
		plan.label = "Pull request #" + number + " ready for review"
	case "edited", "synchronize":
		plan.label = "Pull request #" + number + " updated"
		opens = false
	case "closed":
		plan.label = "Pull request #" + number + " closed"
		if state == github.StateMerged {
			plan.label = "Pull request #" + number + " merged"
		}
	default:
		return false, nil
	}
	if opens {
		plan.move = moveOpened
	}

	refs := github.References(pr.Head.Ref, pr.Title, pr.Body)
	return true, self.link(ctx, workspaceID, refs, func(ref github.Reference) linkPlan {
		if state == github.StateMerged && ref.Closes {
			merged := plan
			merged.move = moveMerged
			return merged
		}
		return plan
	})
}

func (self *GithubService) handlePush(ctx context.Context, workspaceID uuid.UUID, event *github.PushEvent) (bool, error) {
	branch, ok := event.Branch()
	if !ok {
		return false, nil
	}
	plan := linkPlan{
		label: "Branch " + branch + " pushed",
		attachment: issue_repo.LinkedAttachmentInput{
			Title:      branch,
			URL:        github.BranchURL(event.Repository, branch),
			SourceType: SourceTypeBranch,
			Metadata: map[string]any{
				"repo": event.Repository.FullName, "branch": branch, "author": event.Sender.Login, "head_sha": event.After,
			},
		},
		move: moveOpened,
		skip: func(attachment *ent.Attachment) bool {
			return attachment.SourceType == SourceTypePullRequest &&
				attachment.Metadata["repo"] == event.Repository.FullName && attachment.Metadata["branch"] == branch
		},
	}
	refs := github.References(branch)
	return true, self.link(ctx, workspaceID, refs, func(github.Reference) linkPlan { return plan })
}

func (self *GithubService) link(ctx context.Context, workspaceID uuid.UUID, refs []github.Reference, planFor func(github.Reference) linkPlan) error {
	issues, err := self.resolve(ctx, workspaceID, refs)
	if err != nil {
		return err
	}
	var failures []error
	for _, target := range issues {
		if err := self.linkIssue(ctx, target.issueID, planFor(target.ref)); err != nil {
			log.Error("Failed to apply a GitHub event to an issue", "issue", target.issueID, "err", err)
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

type resolvedRef struct {
	issueID uuid.UUID
	ref     github.Reference
}

func (self *GithubService) resolve(ctx context.Context, workspaceID uuid.UUID, refs []github.Reference) ([]resolvedRef, error) {
	var resolved []resolvedRef
	index := map[uuid.UUID]int{}
	for _, ref := range refs {
		ids, err := self.repo.Issue().FindByIdentifier(ctx, []uuid.UUID{workspaceID}, ref.Key, ref.Number)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if i, seen := index[id]; seen {
				resolved[i].ref.Closes = resolved[i].ref.Closes || ref.Closes
				continue
			}
			index[id] = len(resolved)
			resolved = append(resolved, resolvedRef{issueID: id, ref: ref})
		}
	}
	return resolved, nil
}

func (self *GithubService) linkIssue(ctx context.Context, issueID uuid.UUID, plan linkPlan) error {
	issue, err := self.repo.Issue().GetWithEdges(ctx, nil, issueID)
	if err != nil {
		return err
	}
	if plan.skip != nil {
		attachments, err := self.repo.Issue().Attachments(ctx, issueID)
		if err != nil {
			return err
		}
		for _, attachment := range attachments {
			if plan.skip(attachment) {
				return nil
			}
		}
	}
	link := &issue_service.GithubLink{IssueID: issueID, Label: plan.label, Attachment: plan.attachment, ReplaceURL: plan.replaceURL}
	if plan.move != stay && issue.Edges.Team != nil {
		state, err := self.targetState(ctx, issue.Edges.Team, plan.move)
		if err != nil {
			return err
		}
		link.State = state
		link.MoveFrom = openStates
		if plan.move == moveMerged {
			link.MoveFrom = mergeStates
		}
	}
	return self.issues.ApplyGithubLink(ctx, link)
}

func (self *GithubService) targetState(ctx context.Context, team *ent.Team, kind move) (*ent.WorkflowState, error) {
	states, err := self.repo.Team().ListStates(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	configured, fallback := team.GithubPrOpenedStateID, schema.StateTypeStarted
	if kind == moveMerged {
		configured, fallback = team.GithubPrMergedStateID, schema.StateTypeCompleted
	}
	var first *ent.WorkflowState
	for _, state := range states {
		if configured != nil && state.ID == *configured {
			return state, nil
		}
		if first == nil && state.Type == fallback {
			first = state
		}
	}
	return first, nil
}
