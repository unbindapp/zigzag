package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type OAuthClient struct {
	ent.Schema
}

func (OAuthClient) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PKMixin{},
		mixin.TimeMixin{},
	}
}

func (OAuthClient) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id").Unique().NotEmpty(),
		field.String("name"),
		field.JSON("redirect_uris", []string{}),
		field.String("client_uri").Optional(),
		field.Time("last_used_at").Optional().Nillable(),
	}
}

func (OAuthClient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}

func (OAuthClient) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_clients"},
	}
}
