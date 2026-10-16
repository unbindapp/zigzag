package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type ProjectMilestone struct {
	ent.Schema
}

func (ProjectMilestone) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (ProjectMilestone) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("project_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.Time("target_date").Optional().Nillable().SchemaType(dateType),
		field.Float("sort_order").Default(0),
		field.Time("archived_at").Optional().Nillable(),
	}
}

func (ProjectMilestone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("milestones").Field("project_id").Unique().Required(),
		edge.To("issues", Issue.Type).Annotations(setNull()),
	}
}

func (ProjectMilestone) Annotations() []schema.Annotation { return table("project_milestones") }
