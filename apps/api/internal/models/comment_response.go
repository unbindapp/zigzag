package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
)

type AttachmentResponse struct {
	ID         uuid.UUID      `json:"id" format:"uuid"`
	IssueID    uuid.UUID      `json:"issue_id" format:"uuid"`
	Title      string         `json:"title"`
	Subtitle   string         `json:"subtitle"`
	URL        string         `json:"url"`
	SourceType string         `json:"source_type" doc:"upload, link, github (a pull request), github_branch or import"`
	UploadID   *uuid.UUID     `json:"upload_id" nullable:"true" format:"uuid"`
	Metadata   map[string]any `json:"metadata" nullable:"false"`
	SortOrder  float64        `json:"sort_order"`
	Creator    *UserSummary   `json:"creator,omitempty" required:"false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func TransformAttachmentEntity(entity *ent.Attachment) *AttachmentResponse {
	metadata := entity.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	return &AttachmentResponse{
		ID:         entity.ID,
		IssueID:    entity.IssueID,
		Title:      entity.Title,
		Subtitle:   entity.Subtitle,
		URL:        entity.URL,
		SourceType: entity.SourceType,
		UploadID:   entity.UploadID,
		Metadata:   metadata,
		SortOrder:  entity.SortOrder,
		Creator:    TransformUserSummary(entity.Edges.Creator),
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
	}
}

func TransformAttachmentEntities(entities []*ent.Attachment) []*AttachmentResponse {
	responses := make([]*AttachmentResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformAttachmentEntity(entity)
	}
	return responses
}

type CommentResponse struct {
	ID         uuid.UUID               `json:"id" format:"uuid"`
	IssueID    uuid.UUID               `json:"issue_id" format:"uuid"`
	Author     *UserSummary            `json:"author,omitempty" required:"false"`
	Body       string                  `json:"body" doc:"Markdown"`
	ParentID   *uuid.UUID              `json:"parent_id" nullable:"true" format:"uuid" doc:"The comment this one replies to"`
	EditedAt   *time.Time              `json:"edited_at" nullable:"true"`
	ResolvedAt *time.Time              `json:"resolved_at" nullable:"true"`
	ResolvedBy *UserSummary            `json:"resolved_by,omitempty" required:"false"`
	Reactions  []*ReactionGroup        `json:"reactions" nullable:"false"`
	Replies    []*CommentReplyResponse `json:"replies" nullable:"false" doc:"Replies, oldest first"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}

type CommentReplyResponse struct {
	ID         uuid.UUID        `json:"id" format:"uuid"`
	IssueID    uuid.UUID        `json:"issue_id" format:"uuid"`
	Author     *UserSummary     `json:"author,omitempty" required:"false"`
	Body       string           `json:"body" doc:"Markdown"`
	ParentID   *uuid.UUID       `json:"parent_id" nullable:"true" format:"uuid" doc:"The comment this one replies to"`
	EditedAt   *time.Time       `json:"edited_at" nullable:"true"`
	ResolvedAt *time.Time       `json:"resolved_at" nullable:"true"`
	ResolvedBy *UserSummary     `json:"resolved_by,omitempty" required:"false"`
	Reactions  []*ReactionGroup `json:"reactions" nullable:"false"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

func TransformCommentReply(entity *ent.Comment, reactions []*ent.Reaction, callerID uuid.UUID) *CommentReplyResponse {
	return &CommentReplyResponse{
		ID:         entity.ID,
		IssueID:    entity.IssueID,
		Author:     TransformUserSummary(entity.Edges.Author),
		Body:       entity.Body,
		ParentID:   entity.ParentID,
		EditedAt:   entity.EditedAt,
		ResolvedAt: entity.ResolvedAt,
		ResolvedBy: TransformUserSummary(entity.Edges.ResolvedBy),
		Reactions:  TransformReactionGroups(reactions, callerID),
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
	}
}

func TransformCommentEntity(entity *ent.Comment, reactions []*ent.Reaction, callerID uuid.UUID) *CommentResponse {
	reply := TransformCommentReply(entity, reactions, callerID)
	return &CommentResponse{
		ID:         reply.ID,
		IssueID:    reply.IssueID,
		Author:     reply.Author,
		Body:       reply.Body,
		ParentID:   reply.ParentID,
		EditedAt:   reply.EditedAt,
		ResolvedAt: reply.ResolvedAt,
		ResolvedBy: reply.ResolvedBy,
		Reactions:  reply.Reactions,
		Replies:    []*CommentReplyResponse{},
		CreatedAt:  reply.CreatedAt,
		UpdatedAt:  reply.UpdatedAt,
	}
}

func TransformCommentThreads(entities []*ent.Comment, reactions map[uuid.UUID][]*ent.Reaction, callerID uuid.UUID) []*CommentResponse {
	threads := []*CommentResponse{}
	byID := map[uuid.UUID]*CommentResponse{}
	for _, entity := range entities {
		if entity.ParentID != nil {
			continue
		}
		response := TransformCommentEntity(entity, reactions[entity.ID], callerID)
		byID[entity.ID] = response
		threads = append(threads, response)
	}
	for _, entity := range entities {
		if entity.ParentID == nil {
			continue
		}
		parent, ok := byID[*entity.ParentID]
		if !ok {
			continue
		}
		parent.Replies = append(parent.Replies, TransformCommentReply(entity, reactions[entity.ID], callerID))
	}
	return threads
}

type FieldChange = schema.FieldChange

type HistoryResponse struct {
	ID        uuid.UUID            `json:"id" format:"uuid"`
	IssueID   uuid.UUID            `json:"issue_id" format:"uuid"`
	Actor     *UserSummary         `json:"actor,omitempty" required:"false" doc:"Missing for system changes and imports"`
	Source    schema.HistorySource `json:"source"`
	Changes   []FieldChange        `json:"changes" nullable:"false"`
	CreatedAt time.Time            `json:"created_at"`
}

func TransformHistoryEntity(entity *ent.IssueHistory) *HistoryResponse {
	changes := entity.Changes
	if changes == nil {
		changes = []FieldChange{}
	}
	return &HistoryResponse{
		ID:        entity.ID,
		IssueID:   entity.IssueID,
		Actor:     TransformUserSummary(entity.Edges.Actor),
		Source:    entity.Source,
		Changes:   changes,
		CreatedAt: entity.CreatedAt,
	}
}

func TransformHistoryEntities(entities []*ent.IssueHistory) []*HistoryResponse {
	responses := make([]*HistoryResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformHistoryEntity(entity)
	}
	return responses
}
