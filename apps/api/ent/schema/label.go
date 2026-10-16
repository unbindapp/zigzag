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

type Label struct {
	ent.Schema
}

func (Label) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Label) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("team_id", uuid.UUID{}).Optional().Nillable().Comment("Nil for workspace labels"),
		field.String("name").NotEmpty(),
		field.String("color"),
		field.String("description").Optional(),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable().Comment("Label group"),
		field.Bool("is_group").Default(false),
		field.Bool("group_single_select").Default(false),
		field.Float("sort_order").Default(0),
		field.Time("archived_at").Optional().Nillable(),
	}
}

func (Label) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("labels").Field("workspace_id").Unique().Required(),
		edge.From("team", Team.Type).Ref("labels").Field("team_id").Unique(),
		edge.To("children", Label.Type).Annotations(cascade()).From("parent").Field("parent_id").Unique(),
		edge.From("issues", Issue.Type).Ref("labels"),
		edge.From("projects", Project.Type).Ref("labels"),
	}
}

func (Label) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "team_id"),
		index.Fields("parent_id"),
	}
}

func (Label) Annotations() []schema.Annotation { return table("labels") }
