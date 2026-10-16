package issue_service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

type SubscriptionInput struct {
	IssueID uuid.UUID  `json:"issue_id" required:"true" format:"uuid" doc:"The issue"`
	UserID  *uuid.UUID `json:"user_id,omitempty" required:"false" format:"uuid" doc:"Workspace member to subscribe or unsubscribe; defaults to you"`
}

func (self *IssueService) Subscribe(ctx context.Context, userID uuid.UUID, input *SubscriptionInput) (*models.IssueDetailResponse, error) {
	issue, subscriber, err := self.subscriptionTarget(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	if err := self.repo.Issue().Subscribe(ctx, nil, issue.ID, subscriber); err != nil {
		return nil, err
	}
	return self.respondDetail(ctx, userID, issue)
}

func (self *IssueService) Unsubscribe(ctx context.Context, userID uuid.UUID, input *SubscriptionInput) (*models.IssueDetailResponse, error) {
	issue, subscriber, err := self.subscriptionTarget(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	if err := self.repo.Issue().Unsubscribe(ctx, issue.ID, subscriber); err != nil {
		return nil, err
	}
	return self.respondDetail(ctx, userID, issue)
}

func (self *IssueService) subscriptionTarget(ctx context.Context, userID uuid.UUID, input *SubscriptionInput) (*ent.Issue, uuid.UUID, error) {
	issue, err := self.WritableIssue(ctx, userID, input.IssueID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, uuid.Nil, err
	}
	subscriber := userID
	if input.UserID != nil && *input.UserID != userID {
		if err := self.ensureWorkspaceMember(ctx, issue.WorkspaceID, *input.UserID, "user_id"); err != nil {
			return nil, uuid.Nil, err
		}
		subscriber = *input.UserID
	}
	return issue, subscriber, nil
}

type ReactionInput struct {
	Emoji           string     `json:"emoji" required:"true" minLength:"1" maxLength:"64" doc:"The emoji character, like 👍"`
	IssueID         *uuid.UUID `json:"issue_id,omitempty" required:"false" format:"uuid" doc:"React to this issue; give exactly one of issue_id, comment_id and project_update_id"`
	CommentID       *uuid.UUID `json:"comment_id,omitempty" required:"false" format:"uuid" doc:"React to this comment"`
	ProjectUpdateID *uuid.UUID `json:"project_update_id,omitempty" required:"false" format:"uuid" doc:"React to this project update"`
}

func reactionTargetOfIssue(issueID uuid.UUID) issue_repo.ReactionTarget {
	return issue_repo.ReactionTarget{IssueID: &issueID}
}

func (self *IssueService) AddReaction(ctx context.Context, userID uuid.UUID, input *ReactionInput) ([]*models.ReactionGroup, error) {
	target, emoji, err := self.reactionTarget(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	if err := self.repo.Issue().AddReaction(ctx, userID, emoji, target); err != nil {
		return nil, err
	}
	return self.reactions(ctx, userID, target)
}

func (self *IssueService) RemoveReaction(ctx context.Context, userID uuid.UUID, input *ReactionInput) ([]*models.ReactionGroup, error) {
	target, emoji, err := self.reactionTarget(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	if err := self.repo.Issue().RemoveReaction(ctx, userID, emoji, target); err != nil {
		return nil, err
	}
	return self.reactions(ctx, userID, target)
}

func (self *IssueService) reactions(ctx context.Context, userID uuid.UUID, target issue_repo.ReactionTarget) ([]*models.ReactionGroup, error) {
	reactions, err := self.repo.Issue().Reactions(ctx, target)
	if err != nil {
		return nil, err
	}
	return models.TransformReactionGroups(reactions, userID), nil
}

func (self *IssueService) reactionTarget(ctx context.Context, userID uuid.UUID, input *ReactionInput) (issue_repo.ReactionTarget, string, error) {
	target := issue_repo.ReactionTarget{IssueID: input.IssueID, CommentID: input.CommentID, ProjectUpdateID: input.ProjectUpdateID}
	given := 0
	for _, id := range []*uuid.UUID{input.IssueID, input.CommentID, input.ProjectUpdateID} {
		if id != nil {
			given++
		}
	}
	if given != 1 {
		return target, "", invalid("Give exactly one of issue_id, comment_id and project_update_id")
	}
	emoji := strings.TrimSpace(input.Emoji)
	if emoji == "" || utf8.RuneCountInString(emoji) > 16 || strings.ContainsAny(emoji, " \t\n") {
		return target, "", invalid("emoji must be a single emoji")
	}

	switch {
	case input.IssueID != nil:
		_, err := self.WritableIssue(ctx, userID, *input.IssueID, schema.WorkspaceRoleMember)
		return target, emoji, err
	case input.CommentID != nil:
		issueID, err := self.repo.Issue().IssueOfComment(ctx, *input.CommentID)
		if ent.IsNotFound(err) {
			return target, "", errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Comment not found")
		}
		if err != nil {
			return target, "", err
		}
		_, err = self.WritableIssue(ctx, userID, issueID, schema.WorkspaceRoleMember)
		if errdefs.IsNotFound(err) {
			return target, "", errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Comment not found")
		}
		return target, emoji, err
	default:
		workspaceID, err := self.repo.Issue().WorkspaceOfProjectUpdate(ctx, *input.ProjectUpdateID)
		notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Project update not found")
		if ent.IsNotFound(err) {
			return target, "", notFound
		}
		if err != nil {
			return target, "", err
		}
		if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID); err != nil {
			return target, "", errdefs.MaskAsNotFound(err, "Project update not found")
		}
		return target, emoji, self.check(ctx, userID, schema.WorkspaceRoleMember, workspaceID)
	}
}
