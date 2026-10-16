package github_sync_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
)

func (self *GithubSyncService) pushNewComment(ctx context.Context, commentID uuid.UUID) {
	comment, err := self.repo.Comment().GetWithEdges(ctx, commentID)
	if err != nil {
		self.record(ctx, nil, err)
		return
	}
	target, err := self.issueTarget(ctx, comment.IssueID)
	if err == nil && target != nil {
		err = self.createComment(ctx, target, comment)
	}
	self.record(ctx, target, err)
}

func (self *GithubSyncService) createComment(ctx context.Context, target *target, comment *ent.Comment) error {
	created, err := self.client.CreateComment(ctx, target.token, target.repoSync.FullName, target.link.Number, zigzagBody(comment))
	if err != nil {
		return err
	}
	return self.repo.Github().CreateCommentLink(ctx, nil, comment.ID, target.link.ID, created.ID)
}

func (self *GithubSyncService) pushEditedComment(ctx context.Context, commentID uuid.UUID) {
	link, target, err := self.commentTarget(ctx, commentID)
	if err == nil && target != nil {
		var comment *ent.Comment
		if comment, err = self.repo.Comment().GetWithEdges(ctx, commentID); err == nil {
			_, err = self.client.UpdateComment(ctx, target.token, target.repoSync.FullName, link.GithubCommentID, zigzagBody(comment))
		}
	}
	self.record(ctx, target, err)
}

func (self *GithubSyncService) pushDeletedComment(ctx context.Context, commentID uuid.UUID) {
	link, target, err := self.commentTarget(ctx, commentID)
	if err == nil && target != nil {
		if err = self.client.DeleteComment(ctx, target.token, target.repoSync.FullName, link.GithubCommentID); err == nil {
			err = self.repo.Github().DeleteCommentLink(ctx, link.ID)
		}
	}
	self.record(ctx, target, err)
}

func (self *GithubSyncService) commentTarget(ctx context.Context, commentID uuid.UUID) (*ent.GithubCommentLink, *target, error) {
	link, err := self.repo.Github().CommentLinkByComment(ctx, commentID)
	if ent.IsNotFound(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if link.Edges.IssueLink == nil {
		return nil, nil, nil
	}
	target, err := self.targetOf(ctx, link.Edges.IssueLink)
	return link, target, err
}

func zigzagBody(comment *ent.Comment) string {
	name := "Someone"
	if author := comment.Edges.Author; author != nil {
		name = author.Name
		if name == "" {
			name = author.DisplayName
		}
	}
	return github.ZigzagCommentBody(name, comment.Body)
}
