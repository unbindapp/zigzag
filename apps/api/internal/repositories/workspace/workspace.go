package workspace_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i WorkspaceRepositoryInterface -p workspace_repo -s WorkspaceRepository -o workspace_repository_iface.go
type WorkspaceRepository struct {
	base *repository.BaseRepository
}

func NewWorkspaceRepository(db *ent.Client) *WorkspaceRepository {
	return &WorkspaceRepository{base: &repository.BaseRepository{DB: db}}
}
