package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type NotificationSetting struct {
	ent.Schema
}

func (NotificationSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (NotificationSetting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Unique(),
		field.Bool("email_enabled").Default(true),
		field.Bool("assigned").Default(true),
		field.Bool("mentioned").Default(true),
		field.Bool("commented").Default(true),
		field.Bool("state_changed").Default(true),
		field.Bool("due_soon").Default(true),
		field.Bool("project_updates").Default(true),
	}
}

func (NotificationSetting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("notification_setting").Field("user_id").Unique().Required(),
	}
}

func (NotificationSetting) Annotations() []schema.Annotation { return table("notification_settings") }
