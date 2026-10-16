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

type Notification struct {
	ent.Schema
}

func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("actor_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("workspace_id", uuid.UUID{}),
		field.Enum("type").GoType(NotificationType("")),
		field.UUID("issue_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("comment_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("project_update_id", uuid.UUID{}).Optional().Nillable(),
		field.String("title"),
		field.String("body").Optional(),
		field.Time("read_at").Optional().Nillable(),
		field.Time("archived_at").Optional().Nillable(),
		field.Time("snoozed_until").Optional().Nillable(),
		field.Time("emailed_at").Optional().Nillable(),
	}
}

func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("notifications").Field("user_id").Unique().Required(),
		edge.From("actor", User.Type).Ref("acted_notifications").Field("actor_id").Unique(),
		edge.From("workspace", Workspace.Type).Ref("notifications").Field("workspace_id").Unique().Required(),
		edge.From("issue", Issue.Type).Ref("notifications").Field("issue_id").Unique(),
		edge.From("comment", Comment.Type).Ref("notifications").Field("comment_id").Unique(),
		edge.From("project", Project.Type).Ref("notifications").Field("project_id").Unique(),
		edge.From("project_update", ProjectPost.Type).Ref("notifications").Field("project_update_id").Unique(),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "archived_at", "read_at"),
		index.Fields("emailed_at"),
	}
}

func (Notification) Annotations() []schema.Annotation { return table("notifications") }
