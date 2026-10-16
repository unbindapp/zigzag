package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type Cycle struct {
	ent.Schema
}

func (Cycle) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Cycle) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("team_id", uuid.UUID{}),
		field.Int("number"),
		field.String("name").Optional(),
		field.String("description").Optional(),
		field.Time("starts_at"),
		field.Time("ends_at"),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("archived_at").Optional().Nillable(),
	}
}

func (Cycle) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).Ref("cycles").Field("team_id").Unique().Required(),
		edge.To("issues", Issue.Type).Annotations(setNull()),
		edge.To("scope_history", CycleScopeHistory.Type).Annotations(cascade()),
	}
}

func (Cycle) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "number").Unique(),
		index.Fields("ends_at"),
	}
}

func (Cycle) Annotations() []schema.Annotation { return table("cycles") }
