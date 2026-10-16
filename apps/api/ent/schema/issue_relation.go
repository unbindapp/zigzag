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

type IssueRelation struct {
	ent.Schema
}

func (IssueRelation) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (IssueRelation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("issue_id", uuid.UUID{}),
		field.UUID("related_issue_id", uuid.UUID{}),
		field.Enum("type").GoType(IssueRelationType("")),
	}
}

func (IssueRelation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).Ref("relations").Field("issue_id").Unique().Required(),
		edge.From("related_issue", Issue.Type).Ref("inverse_relations").Field("related_issue_id").Unique().Required(),
	}
}

func (IssueRelation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("issue_id", "related_issue_id", "type").Unique(),
		index.Fields("related_issue_id"),
	}
}

func (IssueRelation) Annotations() []schema.Annotation { return table("issue_relations") }
