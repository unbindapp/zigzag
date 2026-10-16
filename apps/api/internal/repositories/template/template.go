package template_repo

import (
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i TemplateRepositoryInterface -p template_repo -s TemplateRepository -o template_repository_iface.go
type TemplateRepository struct {
	base *repository.BaseRepository
}

func NewTemplateRepository(db *ent.Client) *TemplateRepository {
	return &TemplateRepository{base: &repository.BaseRepository{DB: db}}
}
