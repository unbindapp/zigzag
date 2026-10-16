package comment_repo

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/comment"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type ImportedComment struct {
	ID         uuid.UUID
	Existing   bool
	ExternalID string
	IssueID    uuid.UUID
	AuthorID   *uuid.UUID
	Body       string
	ParentID   *uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (self *CommentRepository) ImportedComments(ctx context.Context, workspaceID uuid.UUID, source string) ([]*ent.Comment, error) {
	return self.base.DB.Comment.Query().
		Where(comment.ImportSource(source), comment.HasIssueWith(issue.WorkspaceID(workspaceID))).
		All(ctx)
}

// ImportComments expects every thread root before its replies.
func (self *CommentRepository) ImportComments(ctx context.Context, tx repository.TxInterface, source string, comments []ImportedComment) error {
	db := self.base.Client(tx)
	var builders []*ent.CommentCreate
	for _, c := range comments {
		if c.Existing {
			update := db.Comment.UpdateOneID(c.ID).SetBody(c.Body).SetUpdatedAt(c.UpdatedAt)
			if c.AuthorID == nil {
				update.ClearAuthorID()
			} else {
				update.SetAuthorID(*c.AuthorID)
			}
			if err := update.Exec(ctx); err != nil {
				return err
			}
			continue
		}
		builders = append(builders, db.Comment.Create().
			SetID(c.ID).
			SetIssueID(c.IssueID).
			SetNillableAuthorID(c.AuthorID).
			SetBody(c.Body).
			SetNillableParentID(c.ParentID).
			SetImportSource(source).
			SetImportExternalID(c.ExternalID).
			SetCreatedAt(c.CreatedAt).
			SetUpdatedAt(c.UpdatedAt))
	}
	for batch := range slices.Chunk(builders, 500) {
		if err := db.Comment.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
