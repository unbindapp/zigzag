package document_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/document"
)

type ListDocumentsFilter struct {
	WorkspaceID     uuid.UUID
	ProjectID       *uuid.UUID
	InitiativeID    *uuid.UUID
	IssueID         *uuid.UUID
	Query           string
	IncludeArchived bool
	IncludeTrashed  bool
}

func withDocumentEdges(q *ent.DocumentQuery) *ent.DocumentQuery {
	return q.WithProject().WithInitiative().WithCreator().WithUpdatedBy()
}

func (self *DocumentRepository) List(ctx context.Context, filter *ListDocumentsFilter) ([]*ent.Document, error) {
	q := withDocumentEdges(self.base.DB.Document.Query().Where(document.WorkspaceID(filter.WorkspaceID)))
	if !filter.IncludeArchived {
		q.Where(document.ArchivedAtIsNil())
	}
	if !filter.IncludeTrashed {
		q.Where(document.TrashedAtIsNil())
	}
	if filter.ProjectID != nil {
		q.Where(document.ProjectID(*filter.ProjectID))
	}
	if filter.InitiativeID != nil {
		q.Where(document.InitiativeID(*filter.InitiativeID))
	}
	if filter.IssueID != nil {
		q.Where(document.IssueID(*filter.IssueID))
	}
	if filter.Query != "" {
		q.Where(document.TitleContainsFold(filter.Query))
	}
	return q.Order(ent.Desc(document.FieldUpdatedAt)).All(ctx)
}

func (self *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Document, error) {
	return withDocumentEdges(self.base.DB.Document.Query().Where(document.ID(id))).Only(ctx)
}

func (self *DocumentRepository) GetBySlug(ctx context.Context, workspaceID uuid.UUID, slug string) (*ent.Document, error) {
	return withDocumentEdges(self.base.DB.Document.Query().Where(document.WorkspaceID(workspaceID), document.Slug(slug))).Only(ctx)
}

func (self *DocumentRepository) ResolveWorkspaceOfDocument(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Document.Query().Where(document.ID(id)).Select(document.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
