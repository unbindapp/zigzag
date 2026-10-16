package github_service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	comment_service "github.com/unbindapp/zigzag-api/internal/services/comment"
)

const commentedLabel = "Commented on GitHub"

func (self *GithubService) handleIssueComment(ctx context.Context, integration *ent.GithubIntegration, event *github.IssueCommentEvent) (bool, error) {
	if event.Issue.IsPullRequest() {
		return false, nil
	}
	switch event.Action {
	case "created", "edited", "deleted":
	default:
		return false, nil
	}
	repoSync, err := self.syncedRepo(ctx, integration, event.Repository.FullName)
	if repoSync == nil || err != nil {
		return false, err
	}
	link, err := self.linkOf(ctx, integration, event.Issue.NodeID)
	if err != nil {
		return false, err
	}
	if link == nil && event.Action == "deleted" {
		return true, nil
	}
	if link == nil {
		if _, err := self.importIssue(ctx, repoSync, &event.Issue); err != nil && !errors.Is(err, errAlreadyLinked) {
			return true, self.recordInbound(ctx, repoSync, err)
		}
		if link, err = self.linkOf(ctx, integration, event.Issue.NodeID); link == nil || err != nil {
			return true, self.recordInbound(ctx, repoSync, err)
		}
	}
	workspaceID := repoSync.Edges.Integration.WorkspaceID
	switch event.Action {
	case "deleted":
		err = self.deleteComment(ctx, event.Comment.ID)
	case "edited":
		err = self.editComment(ctx, link, workspaceID, &event.Comment)
	default:
		_, err = self.importComment(ctx, link, workspaceID, &event.Comment)
	}
	return true, self.recordInbound(ctx, repoSync, err)
}

func (self *GithubService) importComment(ctx context.Context, link *ent.GithubIssueLink, workspaceID uuid.UUID, comment *github.Comment) (bool, error) {
	if github.FromZigzag(comment.Body) || strings.TrimSpace(comment.Body) == "" {
		return false, nil
	}
	_, err := self.repo.Github().CommentLinkByGithubID(ctx, comment.ID)
	if err == nil || !ent.IsNotFound(err) {
		return false, err
	}
	actor := self.memberByLogin(ctx, workspaceID, &comment.User)
	input := &comment_service.GithubCommentInput{
		IssueID:   link.IssueID,
		Body:      commentBody(actor, comment),
		CreatedAt: comment.CreatedAt,
		Link: func(ctx context.Context, tx repository.TxInterface, commentID uuid.UUID) error {
			return self.repo.Github().CreateCommentLink(ctx, tx, commentID, link.ID, comment.ID)
		},
	}
	if actor != uuid.Nil {
		input.AuthorID = &actor
	}
	_, err = self.comments.CreateFromGithub(github.WithInbound(ctx, commentedLabel, comment.HTMLURL), input)
	if ent.IsConstraintError(err) {
		return false, nil
	}
	return err == nil, err
}

func (self *GithubService) editComment(ctx context.Context, link *ent.GithubIssueLink, workspaceID uuid.UUID, comment *github.Comment) error {
	if github.FromZigzag(comment.Body) || strings.TrimSpace(comment.Body) == "" {
		return nil
	}
	existing, err := self.repo.Github().CommentLinkByGithubID(ctx, comment.ID)
	if ent.IsNotFound(err) {
		_, err = self.importComment(ctx, link, workspaceID, comment)
		return err
	}
	if err != nil {
		return err
	}
	actor := self.memberByLogin(ctx, workspaceID, &comment.User)
	_, err = self.comments.Update(github.WithInbound(ctx, "Edited a comment on GitHub", comment.HTMLURL), actor, &comment_service.UpdateCommentInput{
		CommentID: existing.CommentID, Body: commentBody(actor, comment),
	})
	if errdefs.IsNotFound(err) {
		return self.repo.Github().DeleteCommentLink(ctx, existing.ID)
	}
	return err
}

func (self *GithubService) deleteComment(ctx context.Context, githubCommentID int64) error {
	existing, err := self.repo.Github().CommentLinkByGithubID(ctx, githubCommentID)
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = self.comments.Delete(github.WithInbound(ctx, "Deleted a comment on GitHub", ""), uuid.Nil, &comment_service.CommentRefInput{CommentID: existing.CommentID})
	if err != nil && !errdefs.IsNotFound(err) {
		return err
	}
	return self.repo.Github().DeleteCommentLink(ctx, existing.ID)
}

func commentBody(actor uuid.UUID, comment *github.Comment) string {
	if actor != uuid.Nil {
		return comment.Body
	}
	return github.GithubCommentBody(strings.ToLower(comment.User.Login), comment.Body)
}
