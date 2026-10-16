package issue_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i IssueRepositoryInterface -p issue_repo -s IssueRepository -o issue_repository_iface.go
type IssueRepository struct {
	base *repository.BaseRepository
}

func NewIssueRepository(db *ent.Client) *IssueRepository {
	return &IssueRepository{base: &repository.BaseRepository{DB: db}}
}
