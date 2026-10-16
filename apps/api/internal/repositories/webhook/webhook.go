package webhook_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i WebhookRepositoryInterface -p webhook_repo -s WebhookRepository -o webhook_repository_iface.go
type WebhookRepository struct {
	base *repository.BaseRepository
}

func NewWebhookRepository(db *ent.Client) *WebhookRepository {
	return &WebhookRepository{base: &repository.BaseRepository{DB: db}}
}
