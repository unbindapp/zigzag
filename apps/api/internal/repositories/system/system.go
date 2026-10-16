package system_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i SystemRepositoryInterface -p system_repo -s SystemRepository -o system_repository_iface.go
type SystemRepository struct {
	base *repository.BaseRepository
}

func NewSystemRepository(db *ent.Client) *SystemRepository {
	return &SystemRepository{
		base: &repository.BaseRepository{DB: db},
	}
}
