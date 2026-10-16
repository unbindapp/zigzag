package comment_service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	comment_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/comment"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

type GithubCommentInput struct {
	IssueID   uuid.UUID
	AuthorID  *uuid.UUID
	Body      string
	CreatedAt time.Time
	Link      LinkInTx
}

func (self *CommentService) CreateFromGithub(ctx context.Context, input *GithubCommentInput) (*models.CommentResponse, error) {
	if !permissions_repo.IsSystemCaller(ctx) {
		return nil, errdefs.ErrUnauthorized
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, invalid("body cannot be empty")
	}
	issue, err := self.repo.Issue().GetWithEdges(ctx, nil, input.IssueID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Issue not found")
	}
	if err != nil {
		return nil, err
	}
	createdAt := input.CreatedAt
	return self.create(ctx, issue, &comment_repo.CreateCommentInput{IssueID: issue.ID, AuthorID: input.AuthorID, Body: body, CreatedAt: &createdAt}, input.Link)
}
