package github_service

import (
	"context"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/crypto"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i GithubServiceInterface -p github_service -s GithubService -o github_service_iface.go
type GithubService struct {
	cfg       *config.Config
	repo      repositories.RepositoriesInterface
	encryptor *crypto.Encryptor
	issues    issue_service.IssueServiceInterface
	now       func() time.Time
}

func NewGithubService(cfg *config.Config, repo repositories.RepositoriesInterface, encryptor *crypto.Encryptor, issues issue_service.IssueServiceInterface) *GithubService {
	return &GithubService{cfg: cfg, repo: repo, encryptor: encryptor, issues: issues, now: time.Now}
}

func (self *GithubService) checkAdmin(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleAdmin, WorkspaceID: workspaceID}})
}

func (self *GithubService) webhookURL(workspaceID uuid.UUID) string {
	return self.cfg.ExternalUIUrl + self.cfg.APIPathPrefix() + "/integrations/github/webhook?" + url.Values{"workspace_id": {workspaceID.String()}}.Encode()
}
