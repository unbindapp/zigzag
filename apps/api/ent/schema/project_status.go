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

type ProjectStatus struct {
	ent.Schema
}

func (ProjectStatus) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (ProjectStatus) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.Enum("type").GoType(ProjectStatusType("")),
		field.String("color"),
		field.Float("position").Default(0),
		field.String("description").Optional(),
	}
}

func (ProjectStatus) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("project_statuses").Field("workspace_id").Unique().Required(),
		edge.To("projects", Project.Type).Annotations(noAction()),
	}
}

func (ProjectStatus) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id")}
}

func (ProjectStatus) Annotations() []schema.Annotation { return table("project_statuses") }
