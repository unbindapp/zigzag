package user_service

import (
	"time"

	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	email_service "github.com/unbindapp/zigzag-api/internal/services/email"
)

const (
	MinPasswordLength = 8
	passwordResetTTL  = time.Hour
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i UserServiceInterface -p user_service -s UserService -o user_service_iface.go
type UserService struct {
	cfg   *config.Config
	repo  repositories.RepositoriesInterface
	email email_service.EmailServiceInterface
	now   func() time.Time
}

func NewUserService(cfg *config.Config, repo repositories.RepositoriesInterface, email email_service.EmailServiceInterface) *UserService {
	return &UserService{cfg: cfg, repo: repo, email: email, now: time.Now}
}
