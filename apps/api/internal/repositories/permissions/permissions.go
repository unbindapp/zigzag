package permissions_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i PermissionsRepositoryInterface -p permissions_repo -s PermissionsRepository -o permissions_repository_iface.go
type PermissionsRepository struct {
	base *repository.BaseRepository
}

func NewPermissionsRepository(db *ent.Client) *PermissionsRepository {
	return &PermissionsRepository{base: &repository.BaseRepository{DB: db}}
}
