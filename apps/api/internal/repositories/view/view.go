package view_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ViewRepositoryInterface -p view_repo -s ViewRepository -o view_repository_iface.go
type ViewRepository struct {
	base *repository.BaseRepository
}

func NewViewRepository(db *ent.Client) *ViewRepository {
	return &ViewRepository{base: &repository.BaseRepository{DB: db}}
}
