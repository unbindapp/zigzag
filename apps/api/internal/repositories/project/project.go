package project_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ProjectRepositoryInterface -p project_repo -s ProjectRepository -o project_repository_iface.go
type ProjectRepository struct {
	base *repository.BaseRepository
}

func NewProjectRepository(db *ent.Client) *ProjectRepository {
	return &ProjectRepository{base: &repository.BaseRepository{DB: db}}
}
