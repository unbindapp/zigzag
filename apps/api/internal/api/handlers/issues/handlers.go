package issues_handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	comment_service "github.com/zigzagapp-org/zigzag-api/internal/services/comment"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
)

type ListInput struct {
	server.BaseAuthInput
	issue_service.ListIssuesInput
}

type IssueListResponse struct {
	Body struct {
		Data       []*models.IssueResponse `json:"data" nullable:"false"`
		NextCursor *string                 `json:"next_cursor" nullable:"true" doc:"Pass as cursor to get the next page; null on the last page"`
		Total      int                     `json:"total" doc:"Issues matching the filters across all pages"`
	}
}

func (self *HandlerGroup) List(ctx context.Context, input *ListInput) (*IssueListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := self.srv.IssueService.List(ctx, user.ID, &input.ListIssuesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &IssueListResponse{}
	resp.Body.Data = result.Issues
	resp.Body.NextCursor = result.NextCursor
	resp.Body.Total = result.Total
	return resp, nil
}

type GetInput struct {
	server.BaseAuthInput
	issue_service.GetIssueInput
}

type IssueResponse struct {
	Body struct {
		Data *models.IssueResponse `json:"data" nullable:"false"`
	}
}

func issueResponse(issue *models.IssueResponse) *IssueResponse {
	resp := &IssueResponse{}
	resp.Body.Data = issue
	return resp
}

type IssueDetailResponse struct {
	Body struct {
		Data *models.IssueDetailResponse `json:"data" nullable:"false"`
	}
}

func issueDetailResponse(issue *models.IssueDetailResponse) *IssueDetailResponse {
	resp := &IssueDetailResponse{}
	resp.Body.Data = issue
	return resp
}

func (self *HandlerGroup) Get(ctx context.Context, input *GetInput) (*IssueDetailResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issue, err := self.srv.IssueService.Get(ctx, user.ID, &input.GetIssueInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return issueDetailResponse(issue), nil
}

type CreateInput struct {
	server.BaseAuthInput
	Body *issue_service.CreateIssueInput `required:"true"`
}

func (self *HandlerGroup) Create(ctx context.Context, input *CreateInput) (*IssueResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issue, err := self.srv.IssueService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return issueResponse(issue), nil
}

type UpdateInput struct {
	server.BaseAuthInput
	Body *issue_service.UpdateIssueInput `required:"true"`
}

func (self *HandlerGroup) Update(ctx context.Context, input *UpdateInput) (*IssueResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issue, err := self.srv.IssueService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return issueResponse(issue), nil
}

type IssueRefInput struct {
	server.BaseAuthInput
	Body *issue_service.IssueRefInput `required:"true"`
}

type DeletedResponse struct {
	Body struct {
		Data models.DeletedResponse `json:"data" nullable:"false"`
	}
}

func deletedResponse(id uuid.UUID) *DeletedResponse {
	resp := &DeletedResponse{}
	resp.Body.Data = models.DeletedResponse{ID: id, Deleted: true}
	return resp
}

func (self *HandlerGroup) Delete(ctx context.Context, input *IssueRefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.IssueService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	return deletedResponse(input.Body.IssueID), nil
}

func (self *HandlerGroup) lifecycle(ctx context.Context, input *IssueRefInput, apply func(context.Context, uuid.UUID, *issue_service.IssueRefInput) (*models.IssueResponse, error)) (*IssueResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issue, err := apply(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return issueResponse(issue), nil
}

func (self *HandlerGroup) Trash(ctx context.Context, input *IssueRefInput) (*IssueResponse, error) {
	return self.lifecycle(ctx, input, self.srv.IssueService.Trash)
}

func (self *HandlerGroup) Restore(ctx context.Context, input *IssueRefInput) (*IssueResponse, error) {
	return self.lifecycle(ctx, input, self.srv.IssueService.Restore)
}

func (self *HandlerGroup) Archive(ctx context.Context, input *IssueRefInput) (*IssueResponse, error) {
	return self.lifecycle(ctx, input, self.srv.IssueService.Archive)
}

func (self *HandlerGroup) Unarchive(ctx context.Context, input *IssueRefInput) (*IssueResponse, error) {
	return self.lifecycle(ctx, input, self.srv.IssueService.Unarchive)
}

type BulkUpdateInput struct {
	server.BaseAuthInput
	Body *issue_service.BulkUpdateIssuesInput `required:"true"`
}

type IssuesResponse struct {
	Body struct {
		Data []*models.IssueResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) BulkUpdate(ctx context.Context, input *BulkUpdateInput) (*IssuesResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issues, err := self.srv.IssueService.BulkUpdate(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &IssuesResponse{}
	resp.Body.Data = issues
	return resp, nil
}

type HistoryInput struct {
	server.BaseAuthInput
	issue_service.ListIssueHistoryInput
}

type HistoryListResponse struct {
	Body struct {
		Data       []*models.HistoryResponse `json:"data" nullable:"false"`
		NextCursor *string                   `json:"next_cursor" nullable:"true" doc:"Pass as cursor to get the next page; null on the last page"`
	}
}

func (self *HandlerGroup) ListHistory(ctx context.Context, input *HistoryInput) (*HistoryListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	entries, next, err := self.srv.IssueService.ListHistory(ctx, user.ID, &input.ListIssueHistoryInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &HistoryListResponse{}
	resp.Body.Data = entries
	resp.Body.NextCursor = next
	return resp, nil
}

type SearchInput struct {
	server.BaseAuthInput
	issue_service.SearchIssuesInput
}

func (self *HandlerGroup) Search(ctx context.Context, input *SearchInput) (*IssuesResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issues, err := self.srv.IssueService.Search(ctx, user.ID, &input.SearchIssuesInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &IssuesResponse{}
	resp.Body.Data = issues
	return resp, nil
}

type ListCommentsInput struct {
	server.BaseAuthInput
	comment_service.ListCommentsInput
}

type CommentListResponse struct {
	Body struct {
		Data       []*models.CommentResponse `json:"data" nullable:"false"`
		NextCursor *string                   `json:"next_cursor" nullable:"true" doc:"Pass as cursor to get the next page of threads; null on the last page"`
	}
}

func (self *HandlerGroup) ListComments(ctx context.Context, input *ListCommentsInput) (*CommentListResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	comments, next, err := self.srv.CommentService.List(ctx, user.ID, &input.ListCommentsInput)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &CommentListResponse{}
	resp.Body.Data = comments
	resp.Body.NextCursor = next
	return resp, nil
}

type CommentResponse struct {
	Body struct {
		Data *models.CommentResponse `json:"data" nullable:"false"`
	}
}

func commentResponse(comment *models.CommentResponse) *CommentResponse {
	resp := &CommentResponse{}
	resp.Body.Data = comment
	return resp
}

type CreateCommentInput struct {
	server.BaseAuthInput
	Body *comment_service.CreateCommentInput `required:"true"`
}

func (self *HandlerGroup) CreateComment(ctx context.Context, input *CreateCommentInput) (*CommentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	comment, err := self.srv.CommentService.Create(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return commentResponse(comment), nil
}

type UpdateCommentInput struct {
	server.BaseAuthInput
	Body *comment_service.UpdateCommentInput `required:"true"`
}

func (self *HandlerGroup) UpdateComment(ctx context.Context, input *UpdateCommentInput) (*CommentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	comment, err := self.srv.CommentService.Update(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return commentResponse(comment), nil
}

type CommentRefInput struct {
	server.BaseAuthInput
	Body *comment_service.CommentRefInput `required:"true"`
}

func (self *HandlerGroup) DeleteComment(ctx context.Context, input *CommentRefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.CommentService.Delete(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	return deletedResponse(input.Body.CommentID), nil
}

func (self *HandlerGroup) ResolveComment(ctx context.Context, input *CommentRefInput) (*CommentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	comment, err := self.srv.CommentService.Resolve(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return commentResponse(comment), nil
}

func (self *HandlerGroup) UnresolveComment(ctx context.Context, input *CommentRefInput) (*CommentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	comment, err := self.srv.CommentService.Unresolve(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return commentResponse(comment), nil
}

type CreateRelationInput struct {
	server.BaseAuthInput
	Body *issue_service.CreateRelationInput `required:"true"`
}

type RelationResponse struct {
	Body struct {
		Data *models.RelationResponse `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) CreateRelation(ctx context.Context, input *CreateRelationInput) (*RelationResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	relation, err := self.srv.IssueService.CreateRelation(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &RelationResponse{}
	resp.Body.Data = relation
	return resp, nil
}

type RelationRefInput struct {
	server.BaseAuthInput
	Body *issue_service.RelationRefInput `required:"true"`
}

func (self *HandlerGroup) DeleteRelation(ctx context.Context, input *RelationRefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.IssueService.DeleteRelation(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	return deletedResponse(input.Body.RelationID), nil
}

type AttachmentResponse struct {
	Body struct {
		Data *models.AttachmentResponse `json:"data" nullable:"false"`
	}
}

func attachmentResponse(attachment *models.AttachmentResponse) *AttachmentResponse {
	resp := &AttachmentResponse{}
	resp.Body.Data = attachment
	return resp
}

type CreateAttachmentInput struct {
	server.BaseAuthInput
	Body *issue_service.CreateAttachmentInput `required:"true"`
}

func (self *HandlerGroup) CreateAttachment(ctx context.Context, input *CreateAttachmentInput) (*AttachmentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	attachment, err := self.srv.IssueService.CreateAttachment(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return attachmentResponse(attachment), nil
}

type UpdateAttachmentInput struct {
	server.BaseAuthInput
	Body *issue_service.UpdateAttachmentInput `required:"true"`
}

func (self *HandlerGroup) UpdateAttachment(ctx context.Context, input *UpdateAttachmentInput) (*AttachmentResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	attachment, err := self.srv.IssueService.UpdateAttachment(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return attachmentResponse(attachment), nil
}

type AttachmentRefInput struct {
	server.BaseAuthInput
	Body *issue_service.AttachmentRefInput `required:"true"`
}

func (self *HandlerGroup) DeleteAttachment(ctx context.Context, input *AttachmentRefInput) (*DeletedResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := self.srv.IssueService.DeleteAttachment(ctx, user.ID, input.Body); err != nil {
		return nil, oapi.MapError(err)
	}
	return deletedResponse(input.Body.AttachmentID), nil
}

type SubscriptionInput struct {
	server.BaseAuthInput
	Body *issue_service.SubscriptionInput `required:"true"`
}

func (self *HandlerGroup) Subscribe(ctx context.Context, input *SubscriptionInput) (*IssueDetailResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issue, err := self.srv.IssueService.Subscribe(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return issueDetailResponse(issue), nil
}

func (self *HandlerGroup) Unsubscribe(ctx context.Context, input *SubscriptionInput) (*IssueDetailResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	issue, err := self.srv.IssueService.Unsubscribe(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	return issueDetailResponse(issue), nil
}

type ReactionInput struct {
	server.BaseAuthInput
	Body *issue_service.ReactionInput `required:"true"`
}

type ReactionsResponse struct {
	Body struct {
		Data []*models.ReactionGroup `json:"data" nullable:"false"`
	}
}

func (self *HandlerGroup) AddReaction(ctx context.Context, input *ReactionInput) (*ReactionsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	reactions, err := self.srv.IssueService.AddReaction(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ReactionsResponse{}
	resp.Body.Data = reactions
	return resp, nil
}

func (self *HandlerGroup) RemoveReaction(ctx context.Context, input *ReactionInput) (*ReactionsResponse, error) {
	user, err := self.srv.AuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	reactions, err := self.srv.IssueService.RemoveReaction(ctx, user.ID, input.Body)
	if err != nil {
		return nil, oapi.MapError(err)
	}
	resp := &ReactionsResponse{}
	resp.Body.Data = reactions
	return resp, nil
}
