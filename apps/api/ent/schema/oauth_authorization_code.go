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

type OAuthAuthorizationCode struct {
	ent.Schema
}

func (OAuthAuthorizationCode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PKMixin{},
		mixin.TimeMixin{},
	}
}

func (OAuthAuthorizationCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code_hash").Unique().Sensitive(),
		field.String("client_id"),
		field.String("client_name"),
		field.Enum("client_kind").GoType(OAuthClientKind("")),
		field.String("client_uri").Optional(),
		field.String("redirect_uri"),
		field.String("code_challenge"),
		field.String("resource"),
		field.String("scope").Optional(),
		field.Enum("role").GoType(WorkspaceRole("")),
		field.Bool("full_access").Default(false),
		field.JSON("workspace_ids", []uuid.UUID{}).Default([]uuid.UUID{}),
		field.JSON("privileges", []KeyPrivilege{}).Default([]KeyPrivilege{}),
		field.Time("expires_at"),
		field.Time("used_at").Optional().Nillable(),
		field.UUID("grant_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (OAuthAuthorizationCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("oauth_authorization_codes").Field("user_id").Unique().Required(),
	}
}

func (OAuthAuthorizationCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
	}
}

func (OAuthAuthorizationCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_authorization_codes"},
	}
}
