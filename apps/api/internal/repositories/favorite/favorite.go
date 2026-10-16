package favorite_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i FavoriteRepositoryInterface -p favorite_repo -s FavoriteRepository -o favorite_repository_iface.go
type FavoriteRepository struct {
	base *repository.BaseRepository
}

func NewFavoriteRepository(db *ent.Client) *FavoriteRepository {
	return &FavoriteRepository{base: &repository.BaseRepository{DB: db}}
}
