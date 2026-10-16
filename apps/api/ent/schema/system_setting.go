package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type SystemSetting struct {
	ent.Schema
}

func (SystemSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (SystemSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("instance_name").Default("Zigzag"),
		field.JSON("email_settings", &EmailSettings{}).Optional(),
		field.Bytes("email_password_encrypted").Optional().Sensitive(),
		field.Bool("allow_public_signup").Default(false),
	}
}

func (SystemSetting) Annotations() []schema.Annotation { return table("system_settings") }
