package issue_service

import (
	"context"

	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i IssueServiceInterface -p issue_service -s IssueService -o issue_service_iface.go
type IssueService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewIssueService(cfg *config.Config, repo repositories.RepositoriesInterface) *IssueService {
	return &IssueService{cfg: cfg, repo: repo}
}

func (self *IssueService) PurgeTrashed(ctx context.Context) error {
	return nil
}
