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

type Webhook struct {
	ent.Schema
}

func (Webhook) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Webhook) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("team_id", uuid.UUID{}).Optional().Nillable(),
		field.String("url"),
		field.String("secret").Sensitive(),
		field.Bool("enabled").Default(true),
		field.JSON("events", []string{}).Default([]string{}).Comment("issue.created, issue.updated, issue.deleted, comment.created, project.updated, cycle.updated"),
		field.String("label").Default(""),
	}
}

func (Webhook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("webhooks").Field("workspace_id").Unique().Required(),
		edge.From("team", Team.Type).Ref("webhooks").Field("team_id").Unique(),
	}
}

func (Webhook) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id")}
}

func (Webhook) Annotations() []schema.Annotation { return table("webhooks") }
