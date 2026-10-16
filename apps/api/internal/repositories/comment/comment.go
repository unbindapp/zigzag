package comment_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i CommentRepositoryInterface -p comment_repo -s CommentRepository -o comment_repository_iface.go
type CommentRepository struct {
	base *repository.BaseRepository
}

func NewCommentRepository(db *ent.Client) *CommentRepository {
	return &CommentRepository{base: &repository.BaseRepository{DB: db}}
}
