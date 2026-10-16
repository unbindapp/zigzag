package document_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/document"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateDocumentInput struct {
	WorkspaceID  uuid.UUID
	Title        string
	Slug         string
	Icon         string
	Content      string
	ProjectID    *uuid.UUID
	InitiativeID *uuid.UUID
	IssueID      *uuid.UUID
	CreatorID    *uuid.UUID
}

func (self *DocumentRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateDocumentInput) (*ent.Document, error) {
	return self.base.Client(tx).Document.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetTitle(input.Title).
		SetSlug(input.Slug).
		SetIcon(input.Icon).
		SetContent(input.Content).
		SetNillableProjectID(input.ProjectID).
		SetNillableInitiativeID(input.InitiativeID).
		SetNillableIssueID(input.IssueID).
		SetNillableCreatorID(input.CreatorID).
		Save(ctx)
}

func (self *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Document, error) {
	return self.base.DB.Document.Get(ctx, id)
}

func (self *DocumentRepository) ResolveWorkspaceOfDocument(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.Document.Query().Where(document.ID(id)).Select(document.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}
