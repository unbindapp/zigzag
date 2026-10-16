package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type InitiativePost struct {
	ent.Schema
}

func (InitiativePost) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (InitiativePost) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("initiative_id", uuid.UUID{}),
		field.UUID("author_id", uuid.UUID{}).Optional().Nillable(),
		field.String("body").Comment("Markdown"),
		field.Enum("health").GoType(Health("")),
		field.Time("edited_at").Optional().Nillable(),
	}
}

func (InitiativePost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("initiative", Initiative.Type).Ref("updates").Field("initiative_id").Unique().Required(),
		edge.From("author", User.Type).Ref("initiative_updates").Field("author_id").Unique(),
	}
}

func (InitiativePost) Annotations() []schema.Annotation { return table("initiative_updates") }
