package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type ProjectLink struct {
	ent.Schema
}

func (ProjectLink) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (ProjectLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("project_id", uuid.UUID{}),
		field.String("url"),
		field.String("label"),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
		field.Float("sort_order").Default(0),
	}
}

func (ProjectLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("links").Field("project_id").Unique().Required(),
		edge.From("creator", User.Type).Ref("project_links").Field("creator_id").Unique(),
	}
}

func (ProjectLink) Annotations() []schema.Annotation { return table("project_links") }
