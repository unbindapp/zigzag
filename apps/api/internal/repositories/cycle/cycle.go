package cycle_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i CycleRepositoryInterface -p cycle_repo -s CycleRepository -o cycle_repository_iface.go
type CycleRepository struct {
	base *repository.BaseRepository
}

func NewCycleRepository(db *ent.Client) *CycleRepository {
	return &CycleRepository{base: &repository.BaseRepository{DB: db}}
}
