package issue_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/attachment"
	"github.com/unbindapp/zigzag-api/ent/comment"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/issuehistory"
	"github.com/unbindapp/zigzag-api/ent/issuerelation"
	"github.com/unbindapp/zigzag-api/ent/projectpost"
	"github.com/unbindapp/zigzag-api/ent/reaction"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/user"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

func (self *IssueRepository) CreateHistory(ctx context.Context, tx repository.TxInterface, issueID uuid.UUID, actorID *uuid.UUID, source schema.HistorySource, changes []schema.FieldChange) error {
	if len(changes) == 0 {
		return nil
	}
	return self.base.Client(tx).IssueHistory.Create().
		SetIssueID(issueID).
		SetNillableActorID(actorID).
		SetSource(source).
		SetChanges(changes).
		Exec(ctx)
}

func (self *IssueRepository) ListHistory(ctx context.Context, issueID uuid.UUID, offset, limit int) ([]*ent.IssueHistory, error) {
	return self.base.DB.IssueHistory.Query().
		Where(issuehistory.IssueID(issueID)).
		WithActor().
		Order(issuehistory.ByCreatedAt(), issuehistory.ByID()).
		Offset(offset).
		Limit(limit).
		All(ctx)
}

func (self *IssueRepository) CreateRelation(ctx context.Context, tx repository.TxInterface, issueID, relatedIssueID uuid.UUID, relationType schema.IssueRelationType) (*ent.IssueRelation, error) {
	return self.base.Client(tx).IssueRelation.Create().
		SetIssueID(issueID).
		SetRelatedIssueID(relatedIssueID).
		SetType(relationType).
		Save(ctx)
}

func (self *IssueRepository) RelationExists(ctx context.Context, a, b uuid.UUID, relationType schema.IssueRelationType) (bool, error) {
	return self.base.DB.IssueRelation.Query().
		Where(
			issuerelation.TypeEQ(relationType),
			issuerelation.Or(
				issuerelation.And(issuerelation.IssueID(a), issuerelation.RelatedIssueID(b)),
				issuerelation.And(issuerelation.IssueID(b), issuerelation.RelatedIssueID(a)),
			),
		).
		Exist(ctx)
}

func (self *IssueRepository) GetRelation(ctx context.Context, id uuid.UUID) (*ent.IssueRelation, error) {
	return self.base.DB.IssueRelation.Query().
		Where(issuerelation.ID(id)).
		WithIssue(func(q *ent.IssueQuery) { q.WithTeam().WithState() }).
		WithRelatedIssue(func(q *ent.IssueQuery) { q.WithTeam().WithState() }).
		Only(ctx)
}

func (self *IssueRepository) DeleteRelation(ctx context.Context, tx repository.TxInterface, id uuid.UUID) error {
	return self.base.Client(tx).IssueRelation.DeleteOneID(id).Exec(ctx)
}

func (self *IssueRepository) Relations(ctx context.Context, issueID uuid.UUID) (outgoing, incoming []*ent.IssueRelation, err error) {
	outgoing, err = self.base.DB.IssueRelation.Query().
		Where(issuerelation.IssueID(issueID), issuerelation.HasRelatedIssueWith(issue.TrashedAtIsNil())).
		WithRelatedIssue(func(q *ent.IssueQuery) { q.WithTeam().WithState() }).
		Order(issuerelation.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	incoming, err = self.base.DB.IssueRelation.Query().
		Where(issuerelation.RelatedIssueID(issueID), issuerelation.HasIssueWith(issue.TrashedAtIsNil())).
		WithIssue(func(q *ent.IssueQuery) { q.WithTeam().WithState() }).
		Order(issuerelation.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	return outgoing, incoming, nil
}

type CreateAttachmentInput struct {
	IssueID    uuid.UUID
	CreatorID  *uuid.UUID
	Title      string
	Subtitle   string
	URL        string
	SourceType string
	UploadID   *uuid.UUID
}

func (self *IssueRepository) CreateAttachment(ctx context.Context, tx repository.TxInterface, input *CreateAttachmentInput) (*ent.Attachment, error) {
	count, err := self.base.Client(tx).Attachment.Query().Where(attachment.IssueID(input.IssueID)).Count(ctx)
	if err != nil {
		return nil, err
	}
	created, err := self.base.Client(tx).Attachment.Create().
		SetIssueID(input.IssueID).
		SetNillableCreatorID(input.CreatorID).
		SetTitle(input.Title).
		SetSubtitle(input.Subtitle).
		SetURL(input.URL).
		SetSourceType(input.SourceType).
		SetNillableUploadID(input.UploadID).
		SetSortOrder(float64(count)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return self.getAttachment(ctx, tx, created.ID)
}

func (self *IssueRepository) GetAttachment(ctx context.Context, id uuid.UUID) (*ent.Attachment, error) {
	return self.getAttachment(ctx, nil, id)
}

func (self *IssueRepository) getAttachment(ctx context.Context, tx repository.TxInterface, id uuid.UUID) (*ent.Attachment, error) {
	return self.base.Client(tx).Attachment.Query().Where(attachment.ID(id)).WithCreator().Only(ctx)
}

func (self *IssueRepository) UpdateAttachment(ctx context.Context, id uuid.UUID, title, subtitle *string) (*ent.Attachment, error) {
	err := self.base.DB.Attachment.UpdateOneID(id).
		SetNillableTitle(title).
		SetNillableSubtitle(subtitle).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return self.GetAttachment(ctx, id)
}

func (self *IssueRepository) DeleteAttachment(ctx context.Context, tx repository.TxInterface, id uuid.UUID) error {
	return self.base.Client(tx).Attachment.DeleteOneID(id).Exec(ctx)
}

func (self *IssueRepository) Attachments(ctx context.Context, issueID uuid.UUID) ([]*ent.Attachment, error) {
	return self.base.DB.Attachment.Query().
		Where(attachment.IssueID(issueID)).
		WithCreator().
		Order(attachment.BySortOrder(), attachment.ByCreatedAt()).
		All(ctx)
}

func (self *IssueRepository) Subscribe(ctx context.Context, tx repository.TxInterface, issueID uuid.UUID, userIDs ...uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}
	existing, err := self.base.Client(tx).Issue.Query().Where(issue.ID(issueID)).QuerySubscribers().Where(user.IDIn(userIDs...)).IDs(ctx)
	if err != nil {
		return err
	}
	known := make(map[uuid.UUID]bool, len(existing))
	for _, id := range existing {
		known[id] = true
	}
	fresh := []uuid.UUID{}
	for _, id := range userIDs {
		if known[id] {
			continue
		}
		known[id] = true
		fresh = append(fresh, id)
	}
	if len(fresh) == 0 {
		return nil
	}
	return self.base.Client(tx).Issue.UpdateOneID(issueID).AddSubscriberIDs(fresh...).Exec(ctx)
}

func (self *IssueRepository) Unsubscribe(ctx context.Context, issueID, userID uuid.UUID) error {
	return self.base.DB.Issue.UpdateOneID(issueID).RemoveSubscriberIDs(userID).Exec(ctx)
}

func (self *IssueRepository) Subscribers(ctx context.Context, issueID uuid.UUID) ([]*ent.User, error) {
	return self.base.DB.Issue.Query().Where(issue.ID(issueID)).QuerySubscribers().Order(user.ByName()).All(ctx)
}

func (self *IssueRepository) SubscriberIDs(ctx context.Context, issueID uuid.UUID) ([]uuid.UUID, error) {
	return self.base.DB.Issue.Query().Where(issue.ID(issueID)).QuerySubscribers().Where(user.Active(true)).IDs(ctx)
}

type ReactionTarget struct {
	IssueID         *uuid.UUID
	CommentID       *uuid.UUID
	ProjectUpdateID *uuid.UUID
}

func (self ReactionTarget) predicate() func(*ent.ReactionQuery) *ent.ReactionQuery {
	return func(q *ent.ReactionQuery) *ent.ReactionQuery {
		switch {
		case self.IssueID != nil:
			return q.Where(reaction.IssueID(*self.IssueID))
		case self.CommentID != nil:
			return q.Where(reaction.CommentID(*self.CommentID))
		default:
			return q.Where(reaction.ProjectUpdateID(*self.ProjectUpdateID))
		}
	}
}

func (self *IssueRepository) AddReaction(ctx context.Context, userID uuid.UUID, emoji string, target ReactionTarget) error {
	exists, err := target.predicate()(self.base.DB.Reaction.Query().Where(reaction.UserID(userID), reaction.Emoji(emoji))).Exist(ctx)
	if err != nil || exists {
		return err
	}
	err = self.base.DB.Reaction.Create().
		SetUserID(userID).
		SetEmoji(emoji).
		SetNillableIssueID(target.IssueID).
		SetNillableCommentID(target.CommentID).
		SetNillableProjectUpdateID(target.ProjectUpdateID).
		Exec(ctx)
	if ent.IsConstraintError(err) {
		return nil
	}
	return err
}

func (self *IssueRepository) RemoveReaction(ctx context.Context, userID uuid.UUID, emoji string, target ReactionTarget) error {
	ids, err := target.predicate()(self.base.DB.Reaction.Query().Where(reaction.UserID(userID), reaction.Emoji(emoji))).IDs(ctx)
	if err != nil || len(ids) == 0 {
		return err
	}
	_, err = self.base.DB.Reaction.Delete().Where(reaction.IDIn(ids...)).Exec(ctx)
	return err
}

func (self *IssueRepository) Reactions(ctx context.Context, target ReactionTarget) ([]*ent.Reaction, error) {
	return target.predicate()(self.base.DB.Reaction.Query()).
		WithUser().
		Order(reaction.ByCreatedAt()).
		All(ctx)
}

func (self *IssueRepository) ReactionsForComments(ctx context.Context, commentIDs []uuid.UUID) (map[uuid.UUID][]*ent.Reaction, error) {
	byComment := map[uuid.UUID][]*ent.Reaction{}
	if len(commentIDs) == 0 {
		return byComment, nil
	}
	reactions, err := self.base.DB.Reaction.Query().
		Where(reaction.CommentIDIn(commentIDs...)).
		WithUser().
		Order(reaction.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, entity := range reactions {
		byComment[*entity.CommentID] = append(byComment[*entity.CommentID], entity)
	}
	return byComment, nil
}

func (self *IssueRepository) IssueOfComment(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Comment.Query().Where(comment.ID(commentID)).Select(comment.FieldIssueID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.IssueID, nil
}

func (self *IssueRepository) WorkspaceOfProjectUpdate(ctx context.Context, updateID uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.ProjectPost.Query().Where(projectpost.ID(updateID)).QueryProject().Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
