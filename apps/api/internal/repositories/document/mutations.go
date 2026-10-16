package document_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
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
		SetNillableUpdatedByID(input.CreatorID).
		Save(ctx)
}

type UpdateDocumentInput struct {
	Title           *string
	Icon            *string
	Content         *string
	ProjectID       *uuid.UUID
	ClearProject    bool
	InitiativeID    *uuid.UUID
	ClearInitiative bool
	IssueID         *uuid.UUID
	ClearIssue      bool
	UpdatedByID     uuid.UUID
}

func (self *DocumentRepository) Update(ctx context.Context, id uuid.UUID, input *UpdateDocumentInput) error {
	m := self.base.DB.Document.UpdateOneID(id).SetUpdatedByID(input.UpdatedByID)
	if input.Title != nil {
		m.SetTitle(*input.Title)
	}
	if input.Icon != nil {
		m.SetIcon(*input.Icon)
	}
	if input.Content != nil {
		m.SetContent(*input.Content)
	}
	if input.ProjectID != nil {
		m.SetProjectID(*input.ProjectID)
	}
	if input.ClearProject {
		m.ClearProjectID()
	}
	if input.InitiativeID != nil {
		m.SetInitiativeID(*input.InitiativeID)
	}
	if input.ClearInitiative {
		m.ClearInitiativeID()
	}
	if input.IssueID != nil {
		m.SetIssueID(*input.IssueID)
	}
	if input.ClearIssue {
		m.ClearIssueID()
	}
	return m.Exec(ctx)
}

func (self *DocumentRepository) SetArchivedAt(ctx context.Context, id uuid.UUID, at *time.Time) error {
	m := self.base.DB.Document.UpdateOneID(id)
	if at == nil {
		return m.ClearArchivedAt().Exec(ctx)
	}
	return m.SetArchivedAt(*at).Exec(ctx)
}

func (self *DocumentRepository) SetTrashedAt(ctx context.Context, id uuid.UUID, at *time.Time) error {
	m := self.base.DB.Document.UpdateOneID(id)
	if at == nil {
		return m.ClearTrashedAt().Exec(ctx)
	}
	return m.SetTrashedAt(*at).Exec(ctx)
}

func (self *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return self.base.DB.Document.DeleteOneID(id).Exec(ctx)
}
