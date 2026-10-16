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

type Template struct {
	ent.Schema
}

func (Template) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Template) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("team_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("kind").GoType(TemplateKind("")),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.JSON("data", TemplateData{}),
		field.Float("sort_order").Default(0),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (Template) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("templates").Field("workspace_id").Unique().Required(),
		edge.From("team", Team.Type).Ref("templates").Field("team_id").Unique(),
		edge.From("creator", User.Type).Ref("templates").Field("creator_id").Unique(),
	}
}

func (Template) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id", "kind")}
}

func (Template) Annotations() []schema.Annotation { return table("templates") }
