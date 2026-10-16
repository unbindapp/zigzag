package issue_repo

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const maxSearchWords = 8

func (self *IssueRepository) Search(ctx context.Context, workspaceID uuid.UUID, teamIDs []uuid.UUID, query string, limit int) ([]uuid.UUID, error) {
	words := strings.Fields(query)
	if len(words) == 0 || len(teamIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	if len(words) > maxSearchWords {
		words = words[:maxSearchWords]
	}
	teams := make([]string, len(teamIDs))
	for i, id := range teamIDs {
		teams[i] = id.String()
	}
	args := []any{workspaceID, teams, strings.Join(words, " "), limit}
	where := []string{}
	titleHits := []string{}
	descriptionHits := []string{}
	for _, word := range words {
		args = append(args, word, "%"+escapeLike(word)+"%")
		w, like := "$"+strconv.Itoa(len(args)-1), "$"+strconv.Itoa(len(args))
		where = append(where, "(title ILIKE "+like+" OR description_text ILIKE "+like+" OR "+w+" <% title OR "+w+" <% description_text)")
		titleHits = append(titleHits, "(title ILIKE "+like+")::int")
		descriptionHits = append(descriptionHits, "(coalesce(description_text, '') ILIKE "+like+")::int")
	}
	statement := `SELECT id FROM issues
WHERE workspace_id = $1 AND team_id = ANY($2::uuid[]) AND trashed_at IS NULL AND ` + strings.Join(where, " AND ") + `
ORDER BY ` + strings.Join(titleHits, " + ") + ` DESC,
	word_similarity($3, title) DESC,
	` + strings.Join(descriptionHits, " + ") + ` DESC,
	word_similarity($3, coalesce(description_text, '')) DESC,
	updated_at DESC
LIMIT $4`
	rows, err := self.base.DB.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
