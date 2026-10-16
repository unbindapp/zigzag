package document_service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	"github.com/unbindapp/zigzag-api/internal/models"
	document_repo "github.com/unbindapp/zigzag-api/internal/repositories/document"
)

const (
	slugBaseLength = 48
	untitled       = "Untitled"
)

type CreateDocumentInput struct {
	WorkspaceID  uuid.UUID `json:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	Title        string    `json:"title" required:"true" maxLength:"255" doc:"Document title; empty becomes Untitled"`
	Content      string    `json:"content,omitempty" required:"false" maxLength:"500000" doc:"Markdown content"`
	Icon         string    `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name"`
	ProjectID    uuid.UUID `json:"project_id,omitempty" required:"false" format:"uuid" doc:"Project the document belongs to"`
	InitiativeID uuid.UUID `json:"initiative_id,omitempty" required:"false" format:"uuid" doc:"Initiative the document belongs to"`
	IssueID      uuid.UUID `json:"issue_id,omitempty" required:"false" format:"uuid" doc:"Issue the document belongs to"`
}

func (self *DocumentService) Create(ctx context.Context, userID uuid.UUID, input *CreateDocumentInput) (*models.DocumentResponse, error) {
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleMember, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	projectID, initiativeID, issueID := optionalID(input.ProjectID), optionalID(input.InitiativeID), optionalID(input.IssueID)
	if err := self.ensureParents(ctx, input.WorkspaceID, projectID, initiativeID, issueID); err != nil {
		return nil, err
	}
	title := cleanTitle(input.Title)
	slug, err := utils.SlugWithSuffix(utils.TruncateSlug(utils.Slugify(title), slugBaseLength), 4)
	if err != nil {
		return nil, err
	}
	created, err := self.repo.Document().Create(ctx, nil, &document_repo.CreateDocumentInput{
		WorkspaceID: input.WorkspaceID, Title: title, Slug: slug, Icon: input.Icon, Content: input.Content,
		ProjectID: projectID, InitiativeID: initiativeID, IssueID: issueID, CreatorID: &userID,
	})
	if err != nil {
		return nil, err
	}
	return self.reload(ctx, created.ID)
}

type UpdateDocumentInput struct {
	DocumentID   uuid.UUID  `json:"document_id" required:"true" format:"uuid" doc:"The document to change"`
	Title        *string    `json:"title,omitempty" required:"false" maxLength:"255" doc:"New title; empty becomes Untitled"`
	Content      *string    `json:"content,omitempty" required:"false" maxLength:"500000" doc:"New markdown content, replacing the old"`
	Icon         *string    `json:"icon,omitempty" required:"false" maxLength:"64" doc:"Emoji or icon name, empty to clear"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty" required:"false" format:"uuid" doc:"Move to this project; the nil uuid detaches it"`
	InitiativeID *uuid.UUID `json:"initiative_id,omitempty" required:"false" format:"uuid" doc:"Move to this initiative; the nil uuid detaches it"`
	IssueID      *uuid.UUID `json:"issue_id,omitempty" required:"false" format:"uuid" doc:"Move to this issue; the nil uuid detaches it"`
}

func (self *DocumentService) Update(ctx context.Context, userID uuid.UUID, input *UpdateDocumentInput) (*models.DocumentResponse, error) {
	document, err := self.documentWithRole(ctx, userID, input.DocumentID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	update := &document_repo.UpdateDocumentInput{Content: input.Content, Icon: input.Icon, UpdatedByID: userID}
	if input.Title != nil {
		title := cleanTitle(*input.Title)
		update.Title = &title
	}
	var projectID, initiativeID, issueID *uuid.UUID
	update.ProjectID, update.ClearProject, projectID = parentChange(input.ProjectID)
	update.InitiativeID, update.ClearInitiative, initiativeID = parentChange(input.InitiativeID)
	update.IssueID, update.ClearIssue, issueID = parentChange(input.IssueID)
	if err := self.ensureParents(ctx, document.WorkspaceID, projectID, initiativeID, issueID); err != nil {
		return nil, err
	}
	if err := self.repo.Document().Update(ctx, document.ID, update); err != nil {
		return nil, err
	}
	return self.reload(ctx, document.ID)
}

type DocumentRefInput struct {
	DocumentID uuid.UUID `json:"document_id" required:"true" format:"uuid" doc:"The document"`
}

func (self *DocumentService) Delete(ctx context.Context, userID uuid.UUID, input *DocumentRefInput) error {
	document, err := self.documentWithRole(ctx, userID, input.DocumentID, schema.WorkspaceRoleAdmin)
	if err != nil {
		return err
	}
	return self.repo.Document().Delete(ctx, document.ID)
}

func (self *DocumentService) Archive(ctx context.Context, userID uuid.UUID, input *DocumentRefInput) (*models.DocumentResponse, error) {
	now := self.now()
	return self.setLifecycle(ctx, userID, input.DocumentID, self.repo.Document().SetArchivedAt, &now)
}

func (self *DocumentService) Unarchive(ctx context.Context, userID uuid.UUID, input *DocumentRefInput) (*models.DocumentResponse, error) {
	return self.setLifecycle(ctx, userID, input.DocumentID, self.repo.Document().SetArchivedAt, nil)
}

func (self *DocumentService) Trash(ctx context.Context, userID uuid.UUID, input *DocumentRefInput) (*models.DocumentResponse, error) {
	now := self.now()
	return self.setLifecycle(ctx, userID, input.DocumentID, self.repo.Document().SetTrashedAt, &now)
}

func (self *DocumentService) Restore(ctx context.Context, userID uuid.UUID, input *DocumentRefInput) (*models.DocumentResponse, error) {
	return self.setLifecycle(ctx, userID, input.DocumentID, self.repo.Document().SetTrashedAt, nil)
}

func (self *DocumentService) setLifecycle(ctx context.Context, userID, documentID uuid.UUID, set func(context.Context, uuid.UUID, *time.Time) error, at *time.Time) (*models.DocumentResponse, error) {
	document, err := self.documentWithRole(ctx, userID, documentID, schema.WorkspaceRoleMember)
	if err != nil {
		return nil, err
	}
	if err := set(ctx, document.ID, at); err != nil {
		return nil, err
	}
	return self.reload(ctx, document.ID)
}

func (self *DocumentService) reload(ctx context.Context, documentID uuid.UUID) (*models.DocumentResponse, error) {
	document, err := self.repo.Document().GetByID(ctx, documentID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Document not found")
	}
	if err != nil {
		return nil, err
	}
	return models.TransformDocument(document), nil
}

func cleanTitle(raw string) string {
	title := strings.TrimSpace(raw)
	if title == "" {
		return untitled
	}
	return title
}

// parentChange reads an optional parent id where the nil uuid detaches: set, clear, and the id to validate.
func parentChange(id *uuid.UUID) (*uuid.UUID, bool, *uuid.UUID) {
	if id == nil {
		return nil, false, nil
	}
	if *id == uuid.Nil {
		return nil, true, nil
	}
	return id, false, id
}
