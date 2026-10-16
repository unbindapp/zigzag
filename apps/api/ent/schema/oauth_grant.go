package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type OAuthGrant struct {
	ent.Schema
}

func (OAuthGrant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PKMixin{},
		mixin.TimeMixin{},
	}
}

func (OAuthGrant) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id"),
		field.String("client_name"),
		field.Enum("client_kind").GoType(OAuthClientKind("")),
		field.String("client_uri").Optional(),
		field.String("redirect_uri"),
		field.Enum("role").GoType(WorkspaceRole("")),
		field.Bool("full_access").Default(false),
		field.JSON("workspace_ids", []uuid.UUID{}).Default([]uuid.UUID{}),
		field.JSON("privileges", []KeyPrivilege{}).Default([]KeyPrivilege{}),
		field.String("resource"),
		field.String("scope").Optional(),
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (OAuthGrant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("oauth_grants").Field("user_id").Unique().Required(),
		edge.To("tokens", OAuthGrantToken.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

func (OAuthGrant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("client_id"),
	}
}

func (OAuthGrant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_grants"},
	}
}
