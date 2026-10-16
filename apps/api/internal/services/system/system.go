package system_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	system_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/system"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i SystemServiceInterface -p system_service -s SystemService -o system_service_iface.go
type SystemService struct {
	repo repositories.RepositoriesInterface
}

func NewSystemService(repo repositories.RepositoriesInterface) *SystemService {
	return &SystemService{repo: repo}
}

func (self *SystemService) GetSettings(ctx context.Context, userID uuid.UUID) (*models.SystemSettingsResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, userID); err != nil {
		return nil, err
	}
	settings, err := self.repo.System().GetSystemSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	return models.TransformSystemSettings(settings), nil
}

type UpdateSettingsInput struct {
	InstanceName      *string `json:"instance_name,omitempty" required:"false" maxLength:"64" doc:"Name shown in the app and in emails"`
	AllowPublicSignup *bool   `json:"allow_public_signup,omitempty" required:"false" doc:"Let anyone create an account"`
}

func (self *SystemService) UpdateSettings(ctx context.Context, userID uuid.UUID, input *UpdateSettingsInput) (*models.SystemSettingsResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, userID); err != nil {
		return nil, err
	}
	update := &system_repo.SystemSettingUpdateInput{AllowPublicSignup: input.AllowPublicSignup}
	if input.InstanceName != nil {
		name := strings.TrimSpace(*input.InstanceName)
		if name == "" {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Instance name cannot be empty")
		}
		update.InstanceName = &name
	}
	settings, err := self.repo.System().UpdateSystemSettings(ctx, update)
	if err != nil {
		return nil, err
	}
	return models.TransformSystemSettings(settings), nil
}
