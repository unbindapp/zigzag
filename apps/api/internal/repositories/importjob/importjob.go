package importjob_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i ImportJobRepositoryInterface -p importjob_repo -s ImportJobRepository -o importjob_repository_iface.go
type ImportJobRepository struct {
	base *repository.BaseRepository
}

func NewImportJobRepository(db *ent.Client) *ImportJobRepository {
	return &ImportJobRepository{base: &repository.BaseRepository{DB: db}}
}
