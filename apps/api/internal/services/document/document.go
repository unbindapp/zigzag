package document_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	issue_service "github.com/zigzagapp-org/zigzag-api/internal/services/issue"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i DocumentServiceInterface -p document_service -s DocumentService -o document_service_iface.go
type DocumentService struct {
	cfg    *config.Config
	repo   repositories.RepositoriesInterface
	issues IssueAccess
	now    func() time.Time
}

type IssueAccess interface {
	VisibleTeams(ctx context.Context, userID, workspaceID uuid.UUID) (*issue_service.VisibleTeams, error)
	ReadableIssue(ctx context.Context, userID, issueID uuid.UUID) (*ent.Issue, error)
}

func NewDocumentService(cfg *config.Config, repo repositories.RepositoriesInterface, issues IssueAccess) *DocumentService {
	return &DocumentService{cfg: cfg, repo: repo, issues: issues, now: time.Now}
}

func (self *DocumentService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

func (self *DocumentService) checkRole(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID, notFound string) error {
	if err := self.check(ctx, userID, schema.WorkspaceRoleViewer, workspaceID); err != nil {
		return errdefs.MaskAsNotFound(err, notFound)
	}
	if role == schema.WorkspaceRoleViewer {
		return nil
	}
	return self.check(ctx, userID, role, workspaceID)
}

func (self *DocumentService) documentWithRole(ctx context.Context, userID, documentID uuid.UUID, role schema.WorkspaceRole) (*ent.Document, error) {
	document, err := self.repo.Document().GetByID(ctx, documentID)
	if ent.IsNotFound(err) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Document not found")
	}
	if err != nil {
		return nil, err
	}
	if err := self.checkRole(ctx, userID, role, document.WorkspaceID, "Document not found"); err != nil {
		return nil, err
	}
	if err := self.ensureIssueVisible(ctx, userID, document); err != nil {
		return nil, err
	}
	return document, nil
}

func (self *DocumentService) ensureIssueVisible(ctx context.Context, userID uuid.UUID, document *ent.Document) error {
	if document.IssueID == nil {
		return nil
	}
	_, err := self.issues.ReadableIssue(ctx, userID, *document.IssueID)
	if errdefs.IsNotFound(err) {
		return errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Document not found")
	}
	return err
}

func (self *DocumentService) ensureParents(ctx context.Context, userID, workspaceID uuid.UUID, projectID, initiativeID, issueID *uuid.UUID) error {
	parents := []struct {
		id      *uuid.UUID
		resolve func(context.Context, uuid.UUID) (uuid.UUID, error)
		message string
	}{
		{projectID, self.repo.Project().ResolveWorkspaceOfProject, "The project must belong to this workspace"},
		{initiativeID, self.repo.Initiative().ResolveWorkspaceOfInitiative, "The initiative must belong to this workspace"},
		{issueID, func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
			return self.readableIssueWorkspace(ctx, userID, id)
		}, "The issue must belong to this workspace"},
	}
	for _, parent := range parents {
		if parent.id == nil {
			continue
		}
		found, err := parent.resolve(ctx, *parent.id)
		if ent.IsNotFound(err) || errdefs.IsNotFound(err) || (err == nil && found != workspaceID) {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, parent.message)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *DocumentService) readableIssueWorkspace(ctx context.Context, userID, issueID uuid.UUID) (uuid.UUID, error) {
	issue, err := self.issues.ReadableIssue(ctx, userID, issueID)
	if err != nil {
		return uuid.Nil, err
	}
	return issue.WorkspaceID, nil
}
