package project_repo

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/project"
)

const searchCandidates = 200

func (self *ProjectRepository) Search(ctx context.Context, workspaceIDs []uuid.UUID, query string, limit int) ([]*ent.Project, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(workspaceIDs) == 0 || limit <= 0 {
		return []*ent.Project{}, nil
	}
	found, err := self.base.DB.Project.Query().
		Where(
			project.WorkspaceIDIn(workspaceIDs...),
			project.ArchivedAtIsNil(),
			project.TrashedAtIsNil(),
			project.Or(project.NameContainsFold(query), project.SummaryContainsFold(query)),
		).
		Order(ent.Desc(project.FieldUpdatedAt)).
		Limit(searchCandidates).
		All(ctx)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(query)
	slices.SortStableFunc(found, func(a, b *ent.Project) int {
		return searchRank(a.Name, needle) - searchRank(b.Name, needle)
	})
	return found[:min(limit, len(found))], nil
}

func searchRank(text, needle string) int {
	lowered := strings.ToLower(text)
	switch {
	case lowered == needle:
		return 0
	case strings.HasPrefix(lowered, needle):
		return 1
	case strings.Contains(lowered, needle):
		return 2
	default:
		return 3
	}
}
