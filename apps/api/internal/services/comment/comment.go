package comment_service

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/common/markdown"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	comment_repo "github.com/unbindapp/zigzag-api/internal/repositories/comment"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	github_sync_service "github.com/unbindapp/zigzag-api/internal/services/githubsync"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	notification_service "github.com/unbindapp/zigzag-api/internal/services/notification"
	webhook_service "github.com/unbindapp/zigzag-api/internal/services/webhook"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i CommentServiceInterface -p comment_service -s CommentService -o comment_service_iface.go
type CommentService struct {
	cfg           *config.Config
	repo          repositories.RepositoriesInterface
	issues        *issue_service.IssueService
	notifications notification_service.NotificationServiceInterface
	webhooks      webhook_service.WebhookServiceInterface
	githubSync    github_sync_service.GithubSyncServiceInterface
	now           func() time.Time
}

func NewCommentService(cfg *config.Config, repo repositories.RepositoriesInterface, issues *issue_service.IssueService, notifications notification_service.NotificationServiceInterface, webhooks webhook_service.WebhookServiceInterface, githubSync github_sync_service.GithubSyncServiceInterface) *CommentService {
	return &CommentService{cfg: cfg, repo: repo, issues: issues, notifications: notifications, webhooks: webhooks, githubSync: githubSync, now: time.Now}
}

const snippetRunes = 280

func commentNotFound() error {
	return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Comment not found")
}

func invalid(message string) error {
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, message)
}

type ListCommentsInput struct {
	IssueID uuid.UUID `query:"issue_id" required:"true" format:"uuid" doc:"The issue"`
	models.PaginationInput
}

func (self *CommentService) List(ctx context.Context, userID uuid.UUID, input *ListCommentsInput) ([]*models.CommentResponse, *string, error) {
	issue, err := self.issues.ReadableIssue(ctx, userID, input.IssueID)
	if err != nil {
		return nil, nil, err
	}
	var after *repository.TimeCursor
	if input.Cursor != "" {
		if after, err = repository.DecodeTimeCursor(input.Cursor); err != nil {
			return nil, nil, invalid("cursor is not a next_cursor from a previous page")
		}
	}
	limit := max(input.Limit, 1)
	comments, err := self.repo.Comment().ListThreads(ctx, issue.ID, after, limit+1)
	if err != nil {
		return nil, nil, err
	}
	reactions, err := self.repo.Issue().ReactionsForComments(ctx, commentIDs(comments))
	if err != nil {
		return nil, nil, err
	}
	threads := models.TransformCommentThreads(comments, reactions, userID)
	var next *string
	if len(threads) > limit {
		threads = threads[:limit]
		last := threads[limit-1]
		next = new(repository.TimeCursor{CreatedAt: last.CreatedAt, ID: last.ID}.Encode())
	}
	return threads, next, nil
}

type CreateCommentInput struct {
	IssueID  uuid.UUID  `json:"issue_id" required:"true" format:"uuid" doc:"The issue to comment on"`
	Body     string     `json:"body" required:"true" minLength:"1" maxLength:"100000" doc:"Markdown; mention people with @[Name](user:<id>)"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" required:"false" format:"uuid" doc:"Comment to reply to; a reply to a reply joins the same thread"`
}

func (self *CommentService) Create(ctx context.Context, userID uuid.UUID, input *CreateCommentInput) (*models.CommentResponse, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, invalid("body cannot be empty")
	}
	issue, err := self.issues.WritableIssue(ctx, userID, input.IssueID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	parentID, err := self.threadRoot(ctx, issue.ID, input.ParentID)
	if err != nil {
		return nil, err
	}
	return self.create(ctx, issue, &comment_repo.CreateCommentInput{IssueID: issue.ID, AuthorID: &userID, Body: body, ParentID: parentID}, nil)
}

type LinkInTx func(ctx context.Context, tx repository.TxInterface, commentID uuid.UUID) error

func (self *CommentService) create(ctx context.Context, issue *ent.Issue, input *comment_repo.CreateCommentInput, link LinkInTx) (*models.CommentResponse, error) {
	mentioned, err := self.members(ctx, issue.WorkspaceID, markdown.Mentions(input.Body))
	if err != nil {
		return nil, err
	}
	subscribe := mentioned
	if input.AuthorID != nil {
		subscribe = append([]uuid.UUID{*input.AuthorID}, mentioned...)
	}

	var created *ent.Comment
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		created, err = self.repo.Comment().Create(ctx, tx, input)
		if err != nil {
			return err
		}
		if link != nil {
			if err := link(ctx, tx, created.ID); err != nil {
				return err
			}
		}
		return self.repo.Issue().Subscribe(ctx, tx, issue.ID, subscribe...)
	})
	if err != nil {
		return nil, err
	}

	viewer := uuid.Nil
	if input.AuthorID != nil {
		viewer = *input.AuthorID
	}
	response, err := self.respondOne(ctx, viewer, created.ID)
	if err != nil {
		return nil, err
	}
	title := models.Identifier(issue.Edges.Team.Key, issue.Number) + " " + issue.Title
	self.notify(ctx, input.AuthorID, issue, created.ID, schema.NotificationCommentMentioned, mentioned, title, input.Body)
	subscribers, err := self.repo.Issue().SubscriberIDs(ctx, issue.ID)
	if err != nil {
		log.Error("Failed to load issue subscribers", "issue", issue.ID, "err", err)
	}
	others := []uuid.UUID{}
	for _, id := range subscribers {
		if id != viewer && !slices.Contains(mentioned, id) {
			others = append(others, id)
		}
	}
	self.notify(ctx, input.AuthorID, issue, created.ID, schema.NotificationIssueCommented, others, title, input.Body)
	if err := self.webhooks.Emit(ctx, issue.WorkspaceID, "comment.created", map[string]any{"comment": response, "issue_id": issue.ID}); err != nil {
		log.Error("Failed to emit webhook", "event", "comment.created", "err", err)
	}
	self.githubSync.AfterCommentCreate(ctx, created.ID)
	return response, nil
}

func (self *CommentService) threadRoot(ctx context.Context, issueID uuid.UUID, parentID *uuid.UUID) (*uuid.UUID, error) {
	if parentID == nil {
		return nil, nil
	}
	parent, err := self.repo.Comment().GetByID(ctx, *parentID)
	if ent.IsNotFound(err) || (err == nil && parent.IssueID != issueID) {
		return nil, invalid("parent_id must be a comment on the same issue")
	}
	if err != nil {
		return nil, err
	}
	if parent.ParentID != nil {
		return parent.ParentID, nil
	}
	return &parent.ID, nil
}

type UpdateCommentInput struct {
	CommentID uuid.UUID `json:"comment_id" required:"true" format:"uuid" doc:"The comment to edit; only its author or an admin may"`
	Body      string    `json:"body" required:"true" minLength:"1" maxLength:"100000" doc:"New markdown body"`
}

func (self *CommentService) Update(ctx context.Context, userID uuid.UUID, input *UpdateCommentInput) (*models.CommentResponse, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, invalid("body cannot be empty")
	}
	comment, issue, err := self.ownedComment(ctx, userID, input.CommentID)
	if err != nil {
		return nil, err
	}
	before := markdown.Mentions(comment.Body)
	added := []uuid.UUID{}
	for _, id := range markdown.Mentions(body) {
		if !slices.Contains(before, id) {
			added = append(added, id)
		}
	}
	mentioned, err := self.members(ctx, issue.WorkspaceID, added)
	if err != nil {
		return nil, err
	}
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Comment().UpdateBody(ctx, tx, comment.ID, body, self.now()); err != nil {
			return err
		}
		return self.repo.Issue().Subscribe(ctx, tx, issue.ID, mentioned...)
	})
	if err != nil {
		return nil, err
	}
	title := models.Identifier(issue.Edges.Team.Key, issue.Number) + " " + issue.Title
	self.notify(ctx, actorOf(userID), issue, comment.ID, schema.NotificationCommentMentioned, mentioned, title, body)
	self.githubSync.AfterCommentUpdate(ctx, comment.ID)
	return self.respondOne(ctx, userID, comment.ID)
}

type CommentRefInput struct {
	CommentID uuid.UUID `json:"comment_id" required:"true" format:"uuid" doc:"The comment"`
}

func (self *CommentService) Delete(ctx context.Context, userID uuid.UUID, input *CommentRefInput) error {
	comment, _, err := self.ownedComment(ctx, userID, input.CommentID)
	if err != nil {
		return err
	}
	if err := self.repo.Comment().Delete(ctx, comment.ID); err != nil {
		return err
	}
	self.githubSync.AfterCommentDelete(ctx, comment.ID)
	return nil
}

func (self *CommentService) Resolve(ctx context.Context, userID uuid.UUID, input *CommentRefInput) (*models.CommentResponse, error) {
	comment, err := self.writableComment(ctx, userID, input.CommentID)
	if err != nil {
		return nil, err
	}
	if comment.ResolvedAt == nil {
		now := self.now()
		if err := self.repo.Comment().SetResolved(ctx, comment.ID, &userID, &now); err != nil {
			return nil, err
		}
	}
	return self.respondOne(ctx, userID, comment.ID)
}

func (self *CommentService) Unresolve(ctx context.Context, userID uuid.UUID, input *CommentRefInput) (*models.CommentResponse, error) {
	comment, err := self.writableComment(ctx, userID, input.CommentID)
	if err != nil {
		return nil, err
	}
	if comment.ResolvedAt != nil {
		if err := self.repo.Comment().SetResolved(ctx, comment.ID, nil, nil); err != nil {
			return nil, err
		}
	}
	return self.respondOne(ctx, userID, comment.ID)
}

func (self *CommentService) readableComment(ctx context.Context, userID, commentID uuid.UUID) (*ent.Comment, *ent.Issue, error) {
	comment, err := self.repo.Comment().GetByID(ctx, commentID)
	if ent.IsNotFound(err) {
		return nil, nil, commentNotFound()
	}
	if err != nil {
		return nil, nil, err
	}
	issue, err := self.issues.ReadableIssue(ctx, userID, comment.IssueID)
	if errdefs.IsNotFound(err) {
		return nil, nil, commentNotFound()
	}
	if err != nil {
		return nil, nil, err
	}
	return comment, issue, nil
}

func (self *CommentService) writableComment(ctx context.Context, userID, commentID uuid.UUID) (*ent.Comment, error) {
	comment, issue, err := self.readableComment(ctx, userID, commentID)
	if err != nil {
		return nil, err
	}
	if err := self.check(ctx, userID, schema.WorkspaceRoleMember, issue.WorkspaceID); err != nil {
		return nil, err
	}
	return comment, nil
}

func (self *CommentService) ownedComment(ctx context.Context, userID, commentID uuid.UUID) (*ent.Comment, *ent.Issue, error) {
	comment, issue, err := self.readableComment(ctx, userID, commentID)
	if err != nil {
		return nil, nil, err
	}
	role := schema.WorkspaceRoleAdmin
	if comment.AuthorID != nil && *comment.AuthorID == userID {
		role = schema.WorkspaceRoleMember
	}
	if err := self.check(ctx, userID, role, issue.WorkspaceID); err != nil {
		return nil, nil, err
	}
	return comment, issue, nil
}

func (self *CommentService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *CommentService) respondOne(ctx context.Context, userID, commentID uuid.UUID) (*models.CommentResponse, error) {
	comment, err := self.repo.Comment().GetWithEdges(ctx, commentID)
	if err != nil {
		return nil, err
	}
	all := append([]*ent.Comment{comment}, comment.Edges.Replies...)
	reactions, err := self.repo.Issue().ReactionsForComments(ctx, commentIDs(all))
	if err != nil {
		return nil, err
	}
	response := models.TransformCommentEntity(comment, reactions[comment.ID], userID)
	for _, reply := range comment.Edges.Replies {
		response.Replies = append(response.Replies, models.TransformCommentReply(reply, reactions[reply.ID], userID))
	}
	return response, nil
}

func (self *CommentService) members(ctx context.Context, workspaceID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	members := make([]uuid.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		_, err := self.repo.Workspace().GetMember(ctx, nil, workspaceID, id)
		if ent.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		members = append(members, id)
	}
	return members, nil
}

func (self *CommentService) notify(ctx context.Context, actor *uuid.UUID, issue *ent.Issue, commentID uuid.UUID, kind schema.NotificationType, recipients []uuid.UUID, title, body string) {
	filtered := make([]uuid.UUID, 0, len(recipients))
	for _, id := range recipients {
		if (actor == nil || id != *actor) && !slices.Contains(filtered, id) {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return
	}
	err := self.notifications.Notify(ctx, models.NotificationEvent{
		Type: kind, WorkspaceID: issue.WorkspaceID, ActorID: actor, RecipientIDs: filtered,
		IssueID: &issue.ID, CommentID: &commentID, Title: title, Body: snippet(markdown.PlainText(body)),
	})
	if err != nil {
		log.Error("Failed to record notifications", "type", kind, "err", err)
	}
}

func actorOf(userID uuid.UUID) *uuid.UUID {
	if userID == uuid.Nil {
		return nil
	}
	return &userID
}

func snippet(text string) string {
	runes := []rune(text)
	if len(runes) <= snippetRunes {
		return text
	}
	return string(runes[:snippetRunes-1]) + "…"
}

func commentIDs(comments []*ent.Comment) []uuid.UUID {
	ids := make([]uuid.UUID, len(comments))
	for i, comment := range comments {
		ids[i] = comment.ID
	}
	return ids
}
