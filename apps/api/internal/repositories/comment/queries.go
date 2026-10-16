package comment_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/comment"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type CreateCommentInput struct {
	IssueID   uuid.UUID
	AuthorID  *uuid.UUID
	Body      string
	ParentID  *uuid.UUID
	CreatedAt *time.Time
}

func (self *CommentRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateCommentInput) (*ent.Comment, error) {
	return self.base.Client(tx).Comment.Create().
		SetIssueID(input.IssueID).
		SetNillableAuthorID(input.AuthorID).
		SetBody(input.Body).
		SetNillableParentID(input.ParentID).
		SetNillableCreatedAt(input.CreatedAt).
		Save(ctx)
}

func (self *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Comment, error) {
	return self.base.DB.Comment.Get(ctx, id)
}

func (self *CommentRepository) GetWithEdges(ctx context.Context, id uuid.UUID) (*ent.Comment, error) {
	return self.base.DB.Comment.Query().
		Where(comment.ID(id)).
		WithAuthor().
		WithResolvedBy().
		WithReplies(func(q *ent.CommentQuery) {
			q.WithAuthor().WithResolvedBy().Order(comment.ByCreatedAt(), comment.ByID())
		}).
		Only(ctx)
}

func (self *CommentRepository) ResolveWorkspaceOfComment(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Comment.Query().Where(comment.ID(id)).QueryIssue().Select(issue.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *CommentRepository) ListThreads(ctx context.Context, issueID uuid.UUID, after *repository.TimeCursor, limit int) ([]*ent.Comment, error) {
	query := self.base.DB.Comment.Query().Where(comment.IssueID(issueID), comment.ParentIDIsNil())
	if after != nil {
		query = query.Where(comment.Or(
			comment.CreatedAtGT(after.CreatedAt),
			comment.And(comment.CreatedAt(after.CreatedAt), comment.IDGT(after.ID)),
		))
	}
	roots, err := query.
		WithAuthor().
		WithResolvedBy().
		Order(comment.ByCreatedAt(), comment.ByID()).
		Limit(limit).
		All(ctx)
	if err != nil || len(roots) == 0 {
		return roots, err
	}
	ids := make([]uuid.UUID, len(roots))
	for i, root := range roots {
		ids[i] = root.ID
	}
	replies, err := self.base.DB.Comment.Query().
		Where(comment.ParentIDIn(ids...)).
		WithAuthor().
		WithResolvedBy().
		Order(comment.ByCreatedAt(), comment.ByID()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return append(roots, replies...), nil
}

func (self *CommentRepository) UpdateBody(ctx context.Context, tx repository.TxInterface, id uuid.UUID, body string, editedAt time.Time) error {
	return self.base.Client(tx).Comment.UpdateOneID(id).SetBody(body).SetEditedAt(editedAt).Exec(ctx)
}

func (self *CommentRepository) SetResolved(ctx context.Context, id uuid.UUID, resolvedBy *uuid.UUID, resolvedAt *time.Time) error {
	update := self.base.DB.Comment.UpdateOneID(id)
	if resolvedAt == nil {
		return update.ClearResolvedAt().ClearResolvedByID().Exec(ctx)
	}
	return update.SetResolvedAt(*resolvedAt).SetNillableResolvedByID(resolvedBy).Exec(ctx)
}

func (self *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Comment.DeleteOneID(id).Exec(ctx)
}
