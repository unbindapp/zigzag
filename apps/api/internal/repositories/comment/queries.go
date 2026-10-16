package comment_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/comment"
	"github.com/unbindapp/zigzag-api/ent/issue"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateCommentInput struct {
	IssueID  uuid.UUID
	AuthorID *uuid.UUID
	Body     string
	ParentID *uuid.UUID
}

func (self *CommentRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateCommentInput) (*ent.Comment, error) {
	return self.base.Client(tx).Comment.Create().
		SetIssueID(input.IssueID).
		SetNillableAuthorID(input.AuthorID).
		SetBody(input.Body).
		SetNillableParentID(input.ParentID).
		Save(ctx)
}

func (self *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Comment, error) {
	return self.base.DB.Comment.Get(ctx, id)
}

func (self *CommentRepository) ResolveWorkspaceOfComment(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Comment.Query().Where(comment.ID(id)).QueryIssue().Select(issue.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
