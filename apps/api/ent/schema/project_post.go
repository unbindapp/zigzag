package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type ProjectPost struct {
	ent.Schema
}

func (ProjectPost) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (ProjectPost) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("project_id", uuid.UUID{}),
		field.UUID("author_id", uuid.UUID{}).Optional().Nillable(),
		field.String("body").Comment("Markdown"),
		field.Enum("health").GoType(Health("")),
		field.Time("edited_at").Optional().Nillable(),
	}
}

func (ProjectPost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("updates").Field("project_id").Unique().Required(),
		edge.From("author", User.Type).Ref("project_updates").Field("author_id").Unique(),
		edge.To("reactions", Reaction.Type).Annotations(cascade()),
		edge.To("notifications", Notification.Type).Annotations(setNull()),
	}
}

func (ProjectPost) Annotations() []schema.Annotation { return table("project_updates") }
