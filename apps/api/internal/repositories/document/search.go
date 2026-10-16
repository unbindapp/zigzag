package document_repo

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/document"
)

const searchCandidates = 200

func (self *DocumentRepository) Search(ctx context.Context, workspaceIDs []uuid.UUID, visibleTeamIDs []uuid.UUID, query string, limit int) ([]*ent.Document, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(workspaceIDs) == 0 || limit <= 0 {
		return []*ent.Document{}, nil
	}
	found, err := self.base.DB.Document.Query().
		Where(
			document.WorkspaceIDIn(workspaceIDs...),
			onVisibleIssue(visibleTeamIDs),
			document.ArchivedAtIsNil(),
			document.TrashedAtIsNil(),
			document.Or(document.TitleContainsFold(query), document.ContentContainsFold(query)),
		).
		Order(ent.Desc(document.FieldUpdatedAt)).
		Limit(searchCandidates).
		WithProject().
		WithInitiative().
		WithCreator().
		WithUpdatedBy().
		All(ctx)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(query)
	slices.SortStableFunc(found, func(a, b *ent.Document) int {
		return searchRank(a.Title, needle) - searchRank(b.Title, needle)
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
