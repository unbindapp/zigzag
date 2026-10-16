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

type TeamMember struct {
	ent.Schema
}

func (TeamMember) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (TeamMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("team_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Bool("owner").Default(false),
	}
}

func (TeamMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).Ref("members").Field("team_id").Unique().Required(),
		edge.From("user", User.Type).Ref("team_memberships").Field("user_id").Unique().Required(),
	}
}

func (TeamMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}

func (TeamMember) Annotations() []schema.Annotation { return table("team_members") }
