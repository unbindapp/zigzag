package notification_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i NotificationRepositoryInterface -p notification_repo -s NotificationRepository -o notification_repository_iface.go
type NotificationRepository struct {
	base *repository.BaseRepository
}

func NewNotificationRepository(db *ent.Client) *NotificationRepository {
	return &NotificationRepository{base: &repository.BaseRepository{DB: db}}
}
