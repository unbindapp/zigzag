package comment_service

import (
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i CommentServiceInterface -p comment_service -s CommentService -o comment_service_iface.go
type CommentService struct {
	cfg  *config.Config
	repo repositories.RepositoriesInterface
}

func NewCommentService(cfg *config.Config, repo repositories.RepositoriesInterface) *CommentService {
	return &CommentService{cfg: cfg, repo: repo}
}
