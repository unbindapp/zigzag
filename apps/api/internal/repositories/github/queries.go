package github_repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/githubintegration"
	"github.com/zigzagapp-org/zigzag-api/ent/githubrepository"
)

func (self *GithubRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID) (*ent.GithubIntegration, error) {
	return self.base.DB.GithubIntegration.Query().
		Where(githubintegration.WorkspaceID(workspaceID)).
		WithCreatedBy().
		WithRepositories(func(q *ent.GithubRepositoryQuery) {
			q.Order(ent.Desc(githubrepository.FieldLastEventAt), ent.Asc(githubrepository.FieldFullName))
		}).
		Only(ctx)
}
