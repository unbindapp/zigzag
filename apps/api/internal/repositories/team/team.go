package team_repo

import (
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i TeamRepositoryInterface -p team_repo -s TeamRepository -o team_repository_iface.go
type TeamRepository struct {
	base *repository.BaseRepository
}

func NewTeamRepository(db *ent.Client) *TeamRepository {
	return &TeamRepository{base: &repository.BaseRepository{DB: db}}
}
