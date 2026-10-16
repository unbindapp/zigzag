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

type APIKey struct {
	ent.Schema
}

func (APIKey) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (APIKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("token_prefix"),
		field.String("token_hash").Unique().Sensitive(),
		field.Enum("role").GoType(WorkspaceRole("")).Comment("Strongest role the key can act with anywhere"),
		field.Bool("full_access").Default(false).Comment("Reach every workspace the owner can, capped at role"),
		field.JSON("workspace_ids", []uuid.UUID{}).Default([]uuid.UUID{}).Comment("Workspaces the key is limited to; empty when full_access"),
		field.JSON("privileges", []KeyPrivilege{}).Default([]KeyPrivilege{}),
		field.Time("expires_at").Optional().Nillable(),
		field.Time("last_used_at").Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (APIKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("api_keys").Field("user_id").Unique().Required(),
	}
}

func (APIKey) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "name").Unique()}
}

func (APIKey) Annotations() []schema.Annotation { return table("api_keys") }
