package oauthserver_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i OAuthServerRepositoryInterface -p oauthserver_repo -s OAuthServerRepository -o oauthserver_repository_iface.go
type OAuthServerRepository struct {
	base *repository.BaseRepository
}

func NewOAuthServerRepository(db *ent.Client) *OAuthServerRepository {
	return &OAuthServerRepository{base: &repository.BaseRepository{DB: db}}
}
