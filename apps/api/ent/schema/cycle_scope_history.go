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

type CycleScopeHistory struct {
	ent.Schema
}

func (CycleScopeHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (CycleScopeHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("cycle_id", uuid.UUID{}),
		field.Time("date").SchemaType(dateType).Comment("Day of the snapshot in the team's time zone"),
		field.Float("scope"),
		field.Float("completed"),
		field.Float("started"),
	}
}

func (CycleScopeHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cycle", Cycle.Type).Ref("scope_history").Field("cycle_id").Unique().Required(),
	}
}

func (CycleScopeHistory) Indexes() []ent.Index {
	return []ent.Index{index.Fields("cycle_id", "date").Unique()}
}

func (CycleScopeHistory) Annotations() []schema.Annotation { return table("cycle_scope_history") }
