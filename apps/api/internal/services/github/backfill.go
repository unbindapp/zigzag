package github_service

import (
	"context"
	"errors"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/models"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
)

type BackfillGithubRepoSyncInput struct {
	RepoSyncID uuid.UUID `json:"repo_sync_id" required:"true" format:"uuid" doc:"The synced repository to import existing issues from"`
	SyncClosed *bool     `json:"sync_closed,omitempty" required:"false" doc:"Include closed issues; defaults to the repository's setting and becomes it"`
}

func (self *GithubService) Backfill(ctx context.Context, userID uuid.UUID, input *BackfillGithubRepoSyncInput) (*models.GithubRepoSyncResponse, error) {
	repoSync, err := self.adminRepoSync(ctx, userID, input.RepoSyncID)
	if err != nil {
		return nil, err
	}
	token, err := self.token(repoSync.Edges.Integration)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Add a GitHub token before importing existing issues")
	}
	syncClosed := repoSync.SyncClosed
	if input.SyncClosed != nil {
		syncClosed = *input.SyncClosed
	}
	started, err := self.repo.Github().StartBackfill(ctx, repoSync.ID, syncClosed)
	if err != nil {
		return nil, err
	}
	if !started {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, "An import of "+repoSync.FullName+" is already running")
	}
	running, err := self.repo.Github().GetRepoSync(ctx, repoSync.ID)
	if err != nil {
		return nil, err
	}
	systemCtx := permissions_repo.WithSystemCaller(context.Background())
	self.run(func() { self.backfill(systemCtx, running, token) })
	return self.respondRepoSync(ctx, running)
}

func (self *GithubService) RecoverBackfills(ctx context.Context) error {
	return self.repo.Github().FailInterruptedBackfills(ctx)
}

func (self *GithubService) backfill(ctx context.Context, repoSync *ent.GithubRepoSync, token string) {
	report := &schema.GithubBackfillReport{}
	state := github.StateOpen
	if repoSync.SyncClosed {
		state = "all"
	}
	var failed, listing error
	for page := 1; page > 0; {
		issues, next, err := self.client.ListIssues(ctx, token, repoSync.FullName, state, page)
		if err != nil {
			listing = err
			break
		}
		for i := range issues {
			if err := self.backfillIssue(ctx, repoSync, token, &issues[i], report); err != nil {
				log.Error("Failed to import a GitHub issue", "repo", repoSync.FullName, "number", issues[i].Number, "err", err)
				report.Skipped++
				failed = err
			}
		}
		if err := self.repo.Github().RecordBackfillProgress(ctx, repoSync.ID, report); err != nil {
			log.Error("Failed to record GitHub import progress", "repo", repoSync.FullName, "err", err)
		}
		page = next
	}
	status, message := schema.GithubBackfillCompleted, ""
	switch {
	case listing != nil:
		status, message = schema.GithubBackfillFailed, errdefs.DetailMessage(listing)
	case failed != nil:
		message = "Some issues were skipped: " + errdefs.DetailMessage(failed)
	}
	if err := self.repo.Github().FinishBackfill(ctx, repoSync.ID, status, report, message, self.now()); err != nil {
		log.Error("Failed to finish the GitHub import", "repo", repoSync.FullName, "err", err)
	}
}

func (self *GithubService) backfillIssue(ctx context.Context, repoSync *ent.GithubRepoSync, token string, issue *github.Issue, report *schema.GithubBackfillReport) error {
	if issue.IsPullRequest() {
		report.Skipped++
		return nil
	}
	integration := repoSync.Edges.Integration
	link, err := self.linkOf(ctx, integration, issue.NodeID)
	if err != nil {
		return err
	}
	if link == nil {
		_, err := self.importIssue(ctx, repoSync, issue)
		if err != nil && !errors.Is(err, errAlreadyLinked) {
			return err
		}
		if err == nil {
			report.IssuesCreated++
		}
		if link, err = self.linkOf(ctx, integration, issue.NodeID); link == nil || err != nil {
			return err
		}
	} else {
		changed, err := self.reconcile(ctx, link, issue)
		if err != nil {
			return err
		}
		if !changed {
			report.Skipped++
			return nil
		}
		report.IssuesUpdated++
	}
	if issue.Comments == 0 {
		return nil
	}
	return self.backfillComments(ctx, link, integration.WorkspaceID, token, repoSync.FullName, issue.Number, report)
}

func (self *GithubService) reconcile(ctx context.Context, link *ent.GithubIssueLink, issue *github.Issue) (bool, error) {
	if !issue.UpdatedAt.After(link.GithubUpdatedAt) {
		return false, nil
	}
	current, err := self.repo.Issue().GetWithEdges(ctx, nil, link.IssueID)
	if err != nil {
		return false, err
	}
	input := &issue_service.UpdateIssueInput{IssueID: current.ID}
	changed := false
	if title := clampTitle(issue.Title); title != current.Title {
		input.Title, changed = &title, true
	}
	if issue.Body != current.Description {
		input.Description, changed = &issue.Body, true
	}
	closure := github.ClosureOf(issue.State, issue.StateReason)
	if current.Edges.State == nil || !matchesClosure(current.Edges.State.Type, closure) {
		team, err := self.repo.Team().GetByID(ctx, current.TeamID)
		if err != nil {
			return false, err
		}
		state, err := self.stateFor(ctx, team, closure)
		if err != nil {
			return false, err
		}
		input.StateID, changed = &state.ID, true
	}
	for _, label := range issue.Labels {
		id, err := self.labelFor(ctx, current.WorkspaceID, current.TeamID, label)
		if err != nil {
			return false, err
		}
		if !slices.ContainsFunc(current.Edges.Labels, func(existing *ent.Label) bool { return existing.ID == id }) && !slices.Contains(input.AddLabelIDs, id) {
			input.AddLabelIDs, changed = append(input.AddLabelIDs, id), true
		}
	}
	if assignee := self.memberByLogin(ctx, current.WorkspaceID, issue.Assignee); assignee != uuid.Nil && (current.AssigneeID == nil || *current.AssigneeID != assignee) {
		input.AssigneeID, changed = &assignee, true
	}
	if changed {
		label := "Synced from GitHub #" + strconv.Itoa(issue.Number)
		if _, err := self.issues.Update(github.WithInbound(ctx, label, issue.HTMLURL), uuid.Nil, input); err != nil {
			return false, err
		}
	}
	return changed, self.repo.Github().RecordInbound(ctx, link.ID, issue.UpdatedAt)
}

func (self *GithubService) backfillComments(ctx context.Context, link *ent.GithubIssueLink, workspaceID uuid.UUID, token, fullName string, number int, report *schema.GithubBackfillReport) error {
	for page := 1; page > 0; {
		comments, next, err := self.client.ListComments(ctx, token, fullName, number, page)
		if err != nil {
			return err
		}
		for i := range comments {
			created, err := self.importComment(ctx, link, workspaceID, &comments[i])
			if err != nil {
				return err
			}
			if created {
				report.CommentsCreated++
			}
		}
		page = next
	}
	return nil
}
