package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type WorkspaceMember struct {
	ent.Schema
}

func (WorkspaceMember) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (WorkspaceMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("role").GoType(WorkspaceRole("")),
		field.Time("joined_at").Default(time.Now),
	}
}

func (WorkspaceMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("members").Field("workspace_id").Unique().Required(),
		edge.From("user", User.Type).Ref("workspace_memberships").Field("user_id").Unique().Required(),
	}
}

func (WorkspaceMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}

func (WorkspaceMember) Annotations() []schema.Annotation { return table("workspace_members") }
