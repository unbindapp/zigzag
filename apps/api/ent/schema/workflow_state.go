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

type WorkflowState struct {
	ent.Schema
}

func (WorkflowState) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (WorkflowState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("team_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.Enum("type").GoType(StateType("")),
		field.String("color"),
		field.String("description").Optional(),
		field.Float("position").Default(0).Comment("Order within its type"),
		field.Time("archived_at").Optional().Nillable(),
	}
}

func (WorkflowState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).Ref("states").Field("team_id").Unique().Required(),
		edge.To("issues", Issue.Type).Annotations(noAction()),
		edge.To("default_for_teams", Team.Type).Annotations(setNull()),
		edge.To("triage_for_teams", Team.Type).Annotations(setNull()),
	}
}

func (WorkflowState) Indexes() []ent.Index {
	return []ent.Index{index.Fields("team_id", "name").Unique()}
}

func (WorkflowState) Annotations() []schema.Annotation { return table("workflow_states") }
