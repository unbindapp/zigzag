package document_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i DocumentServiceInterface -p document_service -s DocumentService -o document_service_iface.go
type DocumentService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
	now  func() time.Time
}

func NewDocumentService(cfg *config.Config, repo repositories.RepositoriesInterface) *DocumentService {
	return &DocumentService{cfg: cfg, repo: repo, now: time.Now}
}

func (self *DocumentService) check(ctx context.Context, userID uuid.UUID, role schema.WorkspaceRole, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}})
}

// checkRole hides the workspace from outsiders, then requires role.
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
	return document, nil
}

func (self *DocumentService) ensureParents(ctx context.Context, workspaceID uuid.UUID, projectID, initiativeID, issueID *uuid.UUID) error {
	parents := []struct {
		id      *uuid.UUID
		resolve func(context.Context, uuid.UUID) (uuid.UUID, error)
		message string
	}{
		{projectID, self.repo.Project().ResolveWorkspaceOfProject, "The project must belong to this workspace"},
		{initiativeID, self.repo.Initiative().ResolveWorkspaceOfInitiative, "The initiative must belong to this workspace"},
		{issueID, self.repo.Issue().ResolveWorkspaceOfIssue, "The issue must belong to this workspace"},
	}
	for _, parent := range parents {
		if parent.id == nil {
			continue
		}
		found, err := parent.resolve(ctx, *parent.id)
		if ent.IsNotFound(err) || (err == nil && found != workspaceID) {
			return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, parent.message)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
