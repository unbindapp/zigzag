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

type Invitation struct {
	ent.Schema
}

func (Invitation) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Invitation) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty(),
		field.String("token_hash").Unique().Sensitive(),
		field.Enum("role").GoType(WorkspaceRole("")),
		field.Time("expires_at"),
		field.Time("accepted_at").Optional().Nillable(),
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("invited_by", uuid.UUID{}).Optional().Nillable(),
		field.JSON("team_ids", []uuid.UUID{}).Default([]uuid.UUID{}).Comment("Teams joined on accept"),
	}
}

func (Invitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("invitations").Field("workspace_id").Unique().Required(),
		edge.From("inviter", User.Type).Ref("sent_invitations").Field("invited_by").Unique(),
	}
}

func (Invitation) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id", "email").Unique()}
}

func (Invitation) Annotations() []schema.Annotation { return table("invitations") }
