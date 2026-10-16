package initiative_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i InitiativeRepositoryInterface -p initiative_repo -s InitiativeRepository -o initiative_repository_iface.go
type InitiativeRepository struct {
	base *repository.BaseRepository
}

func NewInitiativeRepository(db *ent.Client) *InitiativeRepository {
	return &InitiativeRepository{base: &repository.BaseRepository{DB: db}}
}
