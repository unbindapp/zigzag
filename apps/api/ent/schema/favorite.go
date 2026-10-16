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

type Favorite struct {
	ent.Schema
}

func (Favorite) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Favorite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("workspace_id", uuid.UUID{}),
		field.Enum("kind").GoType(FavoriteKind("")),
		field.UUID("entity_id", uuid.UUID{}),
		field.Float("sort_order").Default(0),
	}
}

func (Favorite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("favorites").Field("user_id").Unique().Required(),
		edge.From("workspace", Workspace.Type).Ref("favorites").Field("workspace_id").Unique().Required(),
	}
}

func (Favorite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "kind", "entity_id").Unique(),
		index.Fields("user_id", "workspace_id"),
	}
}

func (Favorite) Annotations() []schema.Annotation { return table("favorites") }
