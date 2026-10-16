package apikey_service

import (
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i APIKeyServiceInterface -p apikey_service -s APIKeyService -o apikey_service_iface.go
type APIKeyService struct {
	repo repositories.RepositoriesInterface
}

func NewAPIKeyService(repo repositories.RepositoriesInterface) *APIKeyService {
	return &APIKeyService{repo: repo}
}
