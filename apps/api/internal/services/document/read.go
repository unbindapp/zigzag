package document_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	document_repo "github.com/unbindapp/zigzag-api/internal/repositories/document"
)

type ListDocumentsInput struct {
	WorkspaceID     uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"The workspace"`
	ProjectID       uuid.UUID `query:"project_id" required:"false" format:"uuid" doc:"Only documents of this project"`
	InitiativeID    uuid.UUID `query:"initiative_id" required:"false" format:"uuid" doc:"Only documents of this initiative"`
	IssueID         uuid.UUID `query:"issue_id" required:"false" format:"uuid" doc:"Only documents of this issue"`
	Query           string    `query:"query" required:"false" maxLength:"200" doc:"Case insensitive text matched against the title"`
	IncludeArchived bool      `query:"include_archived" required:"false" doc:"Include archived documents"`
	IncludeTrashed  bool      `query:"include_trashed" required:"false" doc:"Include trashed documents"`
}

func optionalID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func (self *DocumentService) List(ctx context.Context, userID uuid.UUID, input *ListDocumentsInput) ([]*models.DocumentResponse, error) {
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID, "Workspace not found"); err != nil {
		return nil, err
	}
	visible, err := self.issues.VisibleTeams(ctx, userID, input.WorkspaceID)
	if err != nil {
		return nil, err
	}
	documents, err := self.repo.Document().List(ctx, &document_repo.ListDocumentsFilter{
		WorkspaceID:     input.WorkspaceID,
		VisibleTeamIDs:  visible.IDs(),
		ProjectID:       optionalID(input.ProjectID),
		InitiativeID:    optionalID(input.InitiativeID),
		IssueID:         optionalID(input.IssueID),
		Query:           strings.TrimSpace(input.Query),
		IncludeArchived: input.IncludeArchived,
		IncludeTrashed:  input.IncludeTrashed,
	})
	if err != nil {
		return nil, err
	}
	return models.TransformDocuments(documents), nil
}

type GetDocumentInput struct {
	DocumentID  uuid.UUID `query:"document_id" required:"false" format:"uuid" doc:"Document id; give this, or workspace_id and slug"`
	WorkspaceID uuid.UUID `query:"workspace_id" required:"false" format:"uuid" doc:"Workspace of the document when looking up by slug"`
	Slug        string    `query:"slug" required:"false" maxLength:"80" doc:"Document slug from its URL, with workspace_id"`
}

func (self *DocumentService) Get(ctx context.Context, userID uuid.UUID, input *GetDocumentInput) (*models.DocumentResponse, error) {
	if input.DocumentID != uuid.Nil {
		document, err := self.documentWithRole(ctx, userID, input.DocumentID, schema.WorkspaceRoleViewer)
		if err != nil {
			return nil, err
		}
		return models.TransformDocument(document), nil
	}
	if input.WorkspaceID == uuid.Nil || input.Slug == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Give document_id, or workspace_id and slug")
	}
	if err := self.checkRole(ctx, userID, schema.WorkspaceRoleViewer, input.WorkspaceID, "Document not found"); err != nil {
		return nil, err
	}
	document, err := self.repo.Document().GetBySlug(ctx, input.WorkspaceID, input.Slug)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Document not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.ensureIssueVisible(ctx, userID, document); err != nil {
		return nil, err
	}
	return models.TransformDocument(document), nil
}
