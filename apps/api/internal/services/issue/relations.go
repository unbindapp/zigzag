package issue_service

import (
	"context"
	"net/url"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/api/apictx"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

type CreateRelationInput struct {
	IssueID        uuid.UUID                `json:"issue_id" required:"true" format:"uuid" doc:"The issue the relation starts from"`
	RelatedIssueID uuid.UUID                `json:"related_issue_id" required:"true" format:"uuid" doc:"The other issue, in the same workspace"`
	Type           schema.IssueRelationType `json:"type" required:"true" doc:"blocks: issue_id blocks related_issue_id; related: a plain link; duplicate: issue_id duplicates related_issue_id and moves to the Duplicate state"`
}

func (self *IssueService) CreateRelation(ctx context.Context, userID uuid.UUID, input *CreateRelationInput) (*models.RelationResponse, error) {
	if !slices.Contains(schema.IssueRelationType("").Values(), string(input.Type)) {
		return nil, invalid("type takes blocks, related or duplicate")
	}
	if input.IssueID == input.RelatedIssueID {
		return nil, invalid("An issue cannot relate to itself")
	}
	issue, err := self.WritableIssue(ctx, userID, input.IssueID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	related, err := self.ReadableIssue(ctx, userID, input.RelatedIssueID)
	if errdefs.IsNotFound(err) || (err == nil && related.WorkspaceID != issue.WorkspaceID) {
		return nil, invalid("related_issue_id must be an issue of the same workspace")
	}
	if err != nil {
		return nil, err
	}
	exists, err := self.repo.Issue().RelationExists(ctx, issue.ID, related.ID, input.Type)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeConflict, "These issues already have this relation")
	}

	issueIdentifier := models.Identifier(issue.Edges.Team.Key, issue.Number)
	relatedIdentifier := models.Identifier(related.Edges.Team.Key, related.Number)
	var relation *ent.IssueRelation
	var duplicateOutcome *issueOutcome
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		relation, err = self.repo.Issue().CreateRelation(ctx, tx, issue.ID, related.ID, input.Type)
		if err != nil {
			return err
		}
		source := apictx.HistorySource(ctx)
		label := string(input.Type) + " " + relatedIdentifier
		if err := self.repo.Issue().CreateHistory(ctx, tx, issue.ID, &userID, source, []schema.FieldChange{{Field: "relation_added", To: related.ID.String(), ToLabel: label}}); err != nil {
			return err
		}
		inverse := inverseLabel(input.Type) + " " + issueIdentifier
		if err := self.repo.Issue().CreateHistory(ctx, tx, related.ID, &userID, source, []schema.FieldChange{{Field: "relation_added", To: issue.ID.String(), ToLabel: inverse}}); err != nil {
			return err
		}
		if input.Type != schema.IssueRelationDuplicate {
			return nil
		}
		state, err := self.duplicateState(ctx, issue.TeamID)
		if err != nil || state == nil || state.ID == issue.StateID {
			return err
		}
		duplicateOutcome, err = self.applyUpdate(ctx, tx, userID, issue, &UpdateIssueInput{IssueID: issue.ID, StateID: &state.ID})
		return err
	})
	if err != nil {
		return nil, err
	}

	if duplicateOutcome != nil {
		if response, err := self.reload(ctx, userID, issue.ID); err == nil {
			self.afterUpdate(ctx, userID, duplicateOutcome, response)
		}
	}
	if input.Type == schema.IssueRelationBlocks && related.AssigneeID != nil {
		actor := &userID
		self.notify(ctx, models.NotificationEvent{
			Type: schema.NotificationIssueBlocking, WorkspaceID: issue.WorkspaceID, ActorID: actor,
			RecipientIDs: without([]uuid.UUID{*related.AssigneeID}, actor), IssueID: &related.ID,
			Title: relatedIdentifier + " " + related.Title, Body: "Blocked by " + issueIdentifier + " " + issue.Title,
		})
	}
	return models.TransformRelation(relation, related), nil
}

func inverseLabel(relationType schema.IssueRelationType) string {
	switch relationType {
	case schema.IssueRelationBlocks:
		return "blocked by"
	case schema.IssueRelationDuplicate:
		return "duplicated by"
	default:
		return string(relationType)
	}
}

func (self *IssueService) duplicateState(ctx context.Context, teamID uuid.UUID) (*ent.WorkflowState, error) {
	states, err := self.repo.Team().ListStates(ctx, teamID)
	if err != nil {
		return nil, err
	}
	var firstCanceled *ent.WorkflowState
	for _, state := range states {
		if state.Type != schema.StateTypeCanceled || state.ArchivedAt != nil {
			continue
		}
		if strings.EqualFold(state.Name, "Duplicate") {
			return state, nil
		}
		if firstCanceled == nil {
			firstCanceled = state
		}
	}
	return firstCanceled, nil
}

type RelationRefInput struct {
	RelationID uuid.UUID `json:"relation_id" required:"true" format:"uuid" doc:"The relation, from relation_id in get-issue's relations"`
}

func (self *IssueService) DeleteRelation(ctx context.Context, userID uuid.UUID, input *RelationRefInput) error {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Relation not found")
	relation, err := self.repo.Issue().GetRelation(ctx, input.RelationID)
	if ent.IsNotFound(err) {
		return notFound
	}
	if err != nil {
		return err
	}
	if _, err := self.WritableIssue(ctx, userID, relation.IssueID, schema.WorkspaceRoleMember); err != nil {
		if errdefs.IsNotFound(err) {
			return notFound
		}
		return err
	}
	issue, related := relation.Edges.Issue, relation.Edges.RelatedIssue
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Issue().DeleteRelation(ctx, tx, relation.ID); err != nil {
			return err
		}
		source := apictx.HistorySource(ctx)
		label := string(relation.Type) + " " + models.Identifier(related.Edges.Team.Key, related.Number)
		if err := self.repo.Issue().CreateHistory(ctx, tx, issue.ID, &userID, source, []schema.FieldChange{{Field: "relation_removed", From: related.ID.String(), FromLabel: label}}); err != nil {
			return err
		}
		inverse := inverseLabel(relation.Type) + " " + models.Identifier(issue.Edges.Team.Key, issue.Number)
		return self.repo.Issue().CreateHistory(ctx, tx, related.ID, &userID, source, []schema.FieldChange{{Field: "relation_removed", From: issue.ID.String(), FromLabel: inverse}})
	})
}

type CreateAttachmentInput struct {
	IssueID  uuid.UUID  `json:"issue_id" required:"true" format:"uuid" doc:"The issue to attach to"`
	Title    string     `json:"title" required:"true" minLength:"1" maxLength:"255" doc:"Link text"`
	URL      string     `json:"url,omitempty" required:"false" maxLength:"2048" doc:"An http or https link; give this or upload_id"`
	UploadID *uuid.UUID `json:"upload_id,omitempty" required:"false" format:"uuid" doc:"A file uploaded to the same workspace with create-upload; give this or url"`
	Subtitle string     `json:"subtitle,omitempty" required:"false" maxLength:"255" doc:"Secondary line"`
}

func (self *IssueService) CreateAttachment(ctx context.Context, userID uuid.UUID, input *CreateAttachmentInput) (*models.AttachmentResponse, error) {
	if (input.URL == "") == (input.UploadID == nil) {
		return nil, invalid("Give exactly one of url and upload_id")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, invalid("title cannot be empty")
	}
	issue, err := self.WritableIssue(ctx, userID, input.IssueID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	create := &issue_repo.CreateAttachmentInput{IssueID: issue.ID, CreatorID: &userID, Title: title, Subtitle: strings.TrimSpace(input.Subtitle)}
	if input.UploadID != nil {
		upload, err := self.repo.Upload().GetByID(ctx, *input.UploadID)
		if ent.IsNotFound(err) || (err == nil && upload.WorkspaceID != issue.WorkspaceID) {
			return nil, invalid("upload_id must be a file uploaded to this workspace")
		}
		if err != nil {
			return nil, err
		}
		create.URL, create.UploadID, create.SourceType = models.UploadURL(upload.ID), &upload.ID, "upload"
	} else {
		parsed, err := url.Parse(input.URL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, invalid("url must be an http or https link")
		}
		create.URL, create.SourceType = parsed.String(), "link"
	}

	var created *ent.Attachment
	err = self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		created, err = self.repo.Issue().CreateAttachment(ctx, tx, create)
		if err != nil {
			return err
		}
		change := schema.FieldChange{Field: "attachment_added", To: created.ID.String(), ToLabel: title}
		return self.repo.Issue().CreateHistory(ctx, tx, issue.ID, &userID, apictx.HistorySource(ctx), []schema.FieldChange{change})
	})
	if err != nil {
		return nil, err
	}
	return models.TransformAttachmentEntity(created), nil
}

type UpdateAttachmentInput struct {
	AttachmentID uuid.UUID `json:"attachment_id" required:"true" format:"uuid" doc:"The attachment"`
	Title        *string   `json:"title,omitempty" required:"false" minLength:"1" maxLength:"255" doc:"New link text"`
	Subtitle     *string   `json:"subtitle,omitempty" required:"false" maxLength:"255" doc:"New secondary line; empty clears it"`
}

func (self *IssueService) UpdateAttachment(ctx context.Context, userID uuid.UUID, input *UpdateAttachmentInput) (*models.AttachmentResponse, error) {
	attachment, err := self.writableAttachment(ctx, userID, input.AttachmentID)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, invalid("title cannot be empty")
		}
		input.Title = &title
	}
	updated, err := self.repo.Issue().UpdateAttachment(ctx, attachment.ID, input.Title, input.Subtitle)
	if err != nil {
		return nil, err
	}
	return models.TransformAttachmentEntity(updated), nil
}

type AttachmentRefInput struct {
	AttachmentID uuid.UUID `json:"attachment_id" required:"true" format:"uuid" doc:"The attachment"`
}

func (self *IssueService) DeleteAttachment(ctx context.Context, userID uuid.UUID, input *AttachmentRefInput) error {
	attachment, err := self.writableAttachment(ctx, userID, input.AttachmentID)
	if err != nil {
		return err
	}
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Issue().DeleteAttachment(ctx, tx, attachment.ID); err != nil {
			return err
		}
		change := schema.FieldChange{Field: "attachment_removed", From: attachment.ID.String(), FromLabel: attachment.Title}
		return self.repo.Issue().CreateHistory(ctx, tx, attachment.IssueID, &userID, apictx.HistorySource(ctx), []schema.FieldChange{change})
	})
}

func (self *IssueService) writableAttachment(ctx context.Context, userID, attachmentID uuid.UUID) (*ent.Attachment, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Attachment not found")
	attachment, err := self.repo.Issue().GetAttachment(ctx, attachmentID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	if _, err := self.WritableIssue(ctx, userID, attachment.IssueID, schema.WorkspaceRoleMember); err != nil {
		if errdefs.IsNotFound(err) {
			return nil, notFound
		}
		return nil, err
	}
	return attachment, nil
}
