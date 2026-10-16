package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type IssueMove struct {
	ent.Schema
}

func (IssueMove) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (IssueMove) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("old_team_id", uuid.UUID{}),
		field.Int("old_number"),
		field.UUID("issue_id", uuid.UUID{}),
	}
}

func (IssueMove) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("old_team", Team.Type).Ref("issue_moves").Field("old_team_id").Unique().Required(),
		edge.From("issue", Issue.Type).Ref("moves").Field("issue_id").Unique().Required(),
	}
}

func (IssueMove) Indexes() []ent.Index {
	return []ent.Index{index.Fields("old_team_id", "old_number").Unique()}
}

func (IssueMove) Annotations() []schema.Annotation { return table("issue_moves") }
