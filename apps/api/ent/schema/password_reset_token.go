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

type PasswordResetToken struct {
	ent.Schema
}

func (PasswordResetToken) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (PasswordResetToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.String("token_hash").Unique().Sensitive(),
		field.Time("expires_at"),
		field.Time("used_at").Optional().Nillable(),
	}
}

func (PasswordResetToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("password_reset_tokens").Field("user_id").Unique().Required(),
	}
}

func (PasswordResetToken) Indexes() []ent.Index {
	return []ent.Index{index.Fields("expires_at")}
}

func (PasswordResetToken) Annotations() []schema.Annotation { return table("password_reset_tokens") }
