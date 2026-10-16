package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type IssueHistory struct {
	ent.Schema
}

func (IssueHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}}
}

func (IssueHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("issue_id", uuid.UUID{}),
		field.UUID("actor_id", uuid.UUID{}).Optional().Nillable().Comment("Nil for system and imports"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.JSON("changes", []FieldChange{}),
		field.Enum("source").GoType(HistorySource("")),
	}
}

func (IssueHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).Ref("history").Field("issue_id").Unique().Required(),
		edge.From("actor", User.Type).Ref("issue_history").Field("actor_id").Unique(),
	}
}

func (IssueHistory) Indexes() []ent.Index {
	return []ent.Index{index.Fields("issue_id", "created_at")}
}

func (IssueHistory) Annotations() []schema.Annotation { return table("issue_history") }
