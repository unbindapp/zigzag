package github_service

import (
	"context"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/crypto"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	comment_service "github.com/unbindapp/zigzag-api/internal/services/comment"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	label_service "github.com/unbindapp/zigzag-api/internal/services/label"
	project_service "github.com/unbindapp/zigzag-api/internal/services/project"
)

type Dependencies struct {
	Issues   issue_service.IssueServiceInterface
	Comments comment_service.CommentServiceInterface
	Labels   label_service.LabelServiceInterface
	Projects project_service.ProjectServiceInterface
	Client   github.ClientInterface
}

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i GithubServiceInterface -p github_service -s GithubService -o github_service_iface.go
type GithubService struct {
	cfg       *config.Config
	repo      repositories.RepositoriesInterface
	encryptor *crypto.Encryptor
	issues    issue_service.IssueServiceInterface
	comments  comment_service.CommentServiceInterface
	labels    label_service.LabelServiceInterface
	projects  project_service.ProjectServiceInterface
	client    github.ClientInterface
	now       func() time.Time
	run       func(func())
}

func NewGithubService(cfg *config.Config, repo repositories.RepositoriesInterface, encryptor *crypto.Encryptor, deps Dependencies) *GithubService {
	return &GithubService{
		cfg: cfg, repo: repo, encryptor: encryptor,
		issues: deps.Issues, comments: deps.Comments, labels: deps.Labels, projects: deps.Projects, client: deps.Client,
		now: time.Now, run: func(job func()) { go job() },
	}
}

func (self *GithubService) checkAdmin(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleAdmin, WorkspaceID: workspaceID}})
}

func (self *GithubService) webhookURL(workspaceID uuid.UUID) string {
	return self.cfg.ExternalUIUrl + self.cfg.APIPathPrefix() + "/integrations/github/webhook?" + url.Values{"workspace_id": {workspaceID.String()}}.Encode()
}
