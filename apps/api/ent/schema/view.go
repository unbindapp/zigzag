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

type View struct {
	ent.Schema
}

func (View) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (View) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("team_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("owner_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("icon").Optional(),
		field.String("color").Optional(),
		field.Enum("kind").GoType(ViewKind("")),
		field.Bool("shared").Default(false),
		field.JSON("filters", map[string]any{}).Default(map[string]any{}).Comment("Same shape as the list-issues filter input"),
		field.JSON("display", ViewDisplay{}),
	}
}

func (View) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("views").Field("workspace_id").Unique().Required(),
		edge.From("team", Team.Type).Ref("views").Field("team_id").Unique(),
		edge.From("owner", User.Type).Ref("views").Field("owner_id").Unique().Required(),
	}
}

func (View) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id", "owner_id")}
}

func (View) Annotations() []schema.Annotation { return table("views") }
