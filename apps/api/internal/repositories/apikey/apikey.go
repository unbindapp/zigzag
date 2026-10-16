package apikey_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i APIKeyRepositoryInterface -p apikey_repo -s APIKeyRepository -o apikey_repository_iface.go
type APIKeyRepository struct {
	base *repository.BaseRepository
}

func NewAPIKeyRepository(db *ent.Client) *APIKeyRepository {
	return &APIKeyRepository{base: &repository.BaseRepository{DB: db}}
}
