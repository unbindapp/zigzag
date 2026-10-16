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

type Initiative struct {
	ent.Schema
}

func (Initiative) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Initiative) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.String("slug"),
		field.String("description").Optional(),
		field.String("icon").Optional(),
		field.String("color").Optional(),
		field.UUID("owner_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("status").GoType(InitiativeStatus("")).Default(string(InitiativeStatusPlanned)),
		field.Time("target_date").Optional().Nillable().SchemaType(dateType),
		field.Float("sort_order").Default(0),
		field.Time("archived_at").Optional().Nillable(),
		field.Time("trashed_at").Optional().Nillable(),
	}
}

func (Initiative) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("initiatives").Field("workspace_id").Unique().Required(),
		edge.From("owner", User.Type).Ref("owned_initiatives").Field("owner_id").Unique(),
		edge.To("projects", Project.Type).StorageKey(edge.Table("initiative_projects"), edge.Columns("initiative_id", "project_id")),
		edge.To("updates", InitiativePost.Type).Annotations(cascade()),
		edge.To("documents", Document.Type).Annotations(cascade()),
	}
}

func (Initiative) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id", "slug").Unique()}
}

func (Initiative) Annotations() []schema.Annotation { return table("initiatives") }
