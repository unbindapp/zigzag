package linear

import (
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	comment_repo "github.com/unbindapp/zigzag-api/internal/repositories/comment"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

func (self *planner) planComments() {
	existing := map[string]uuid.UUID{}
	for _, c := range self.snapshot.Comments {
		existing[c.ImportExternalID] = c.ID
	}
	var roots, replies []comment_repo.ImportedComment
	for _, issue := range self.export.Issues {
		target, ok := self.issues[issue.ID]
		if !ok {
			for _, c := range issue.Comments {
				self.warn(EntityComments, c.ID, "comment on %s was skipped with its issue", issueRef(issue))
				self.count(EntityComments, skipped)
			}
			continue
		}
		byID := map[string]LinearComment{}
		for _, c := range issue.Comments {
			byID[c.ID] = c
		}
		ids := map[string]uuid.UUID{}
		for _, c := range issue.Comments {
			if id, ok := existing[c.ID]; ok {
				ids[c.ID] = id
				continue
			}
			ids[c.ID] = uuid.New()
		}
		for _, c := range issue.Comments {
			createdAt := timeOr(c.CreatedAt, timeOr(issue.CreatedAt, self.options.Now))
			row := comment_repo.ImportedComment{
				ID:         ids[c.ID],
				ExternalID: c.ID,
				IssueID:    target.id,
				AuthorID:   self.user(EntityComments, c.ID, c.AuthorID),
				Body:       clip(c.Body, maxBodyLength),
				CreatedAt:  createdAt,
				UpdatedAt:  timeOr(c.UpdatedAt, createdAt),
			}
			if row.UpdatedAt.Before(row.CreatedAt) {
				row.UpdatedAt = row.CreatedAt
			}
			self.appear(target.team.id, row.AuthorID)
			if root := threadRoot(c, byID); root != c.ID {
				parentID := ids[root]
				row.ParentID = &parentID
			}
			if _, ok := existing[c.ID]; ok {
				row.Existing = true
				self.count(EntityComments, updated)
			} else {
				self.count(EntityComments, created)
			}
			if row.ParentID == nil {
				roots = append(roots, row)
			} else {
				replies = append(replies, row)
			}
		}
	}
	self.plan.Comments = append(roots, replies...)
}

func threadRoot(c LinearComment, byID map[string]LinearComment) string {
	current := c
	for range len(byID) {
		parent, ok := byID[text(current.ParentID)]
		if !ok {
			return current.ID
		}
		current = parent
	}
	return current.ID
}

type relationKey struct {
	from, to uuid.UUID
	kind     schema.IssueRelationType
}

func (self *planner) planRelations() {
	known := map[relationKey]bool{}
	for _, r := range self.snapshot.Relations {
		known[relationKey{r.IssueID, r.RelatedIssueID, r.Type}] = true
	}
	exists := func(key relationKey) bool {
		reversed := relationKey{key.to, key.from, key.kind}
		return known[key] || (key.kind == schema.IssueRelationRelated && known[reversed])
	}
	planned := map[relationKey]bool{}
	add := func(issue LinearIssue, from, to string, kind schema.IssueRelationType) {
		source, sourceOK := self.issues[from]
		target, targetOK := self.issues[to]
		if !sourceOK || !targetOK {
			missing := to
			if !sourceOK {
				missing = from
			}
			self.warn(EntityRelations, issueRef(issue), "related issue %s is not imported; the %s relation was skipped", missing, kind)
			self.count(EntityRelations, skipped)
			return
		}
		if source.id == target.id {
			return
		}
		key := relationKey{source.id, target.id, kind}
		reversed := relationKey{target.id, source.id, kind}
		if planned[key] || (kind == schema.IssueRelationRelated && planned[reversed]) {
			return
		}
		planned[key] = true
		if exists(key) {
			self.count(EntityRelations, skipped)
			return
		}
		self.plan.Relations = append(self.plan.Relations, issue_repo.ImportedRelation{IssueID: source.id, RelatedIssueID: target.id, Type: kind})
		self.count(EntityRelations, created)
	}
	for _, issue := range self.export.Issues {
		if _, ok := self.issues[issue.ID]; !ok {
			continue
		}
		for _, id := range issue.Relations.Blocks {
			add(issue, issue.ID, id, schema.IssueRelationBlocks)
		}
		for _, id := range issue.Relations.BlockedBy {
			add(issue, id, issue.ID, schema.IssueRelationBlocks)
		}
		for _, id := range issue.Relations.RelatedTo {
			add(issue, issue.ID, id, schema.IssueRelationRelated)
		}
		if id := text(issue.Relations.DuplicateOf); id != "" {
			add(issue, issue.ID, id, schema.IssueRelationDuplicate)
		}
	}
}
