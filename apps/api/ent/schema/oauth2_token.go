package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type Oauth2Token struct {
	ent.Schema
}

func (Oauth2Token) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PKMixin{},
		mixin.TimeMixin{},
	}
}

func (Oauth2Token) Fields() []ent.Field {
	return []ent.Field{
		field.String("access_token").Unique().Sensitive(),
		field.String("refresh_token").Unique().Sensitive(),
		field.String("client_id"),
		field.Time("expires_at"),
		field.Bool("revoked").Default(false),
		field.String("scope"),
		field.String("device_info").Optional(),
	}
}

func (Oauth2Token) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("oauth2_tokens").
			Unique().
			Required(),
	}
}

func (Oauth2Token) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "oauth2_tokens",
		},
	}
}
