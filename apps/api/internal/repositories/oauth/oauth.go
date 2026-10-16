package oauth_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i OauthRepositoryInterface -p oauth_repo -s OauthRepository -o oauth_repository_iface.go
type OauthRepository struct {
	base *repository.BaseRepository
}

func NewOauthRepository(db *ent.Client) *OauthRepository {
	return &OauthRepository{
		base: &repository.BaseRepository{DB: db},
	}
}
