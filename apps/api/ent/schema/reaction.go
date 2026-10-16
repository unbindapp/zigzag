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

type Reaction struct {
	ent.Schema
}

func (Reaction) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Reaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("emoji").NotEmpty(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("issue_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("comment_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("project_update_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (Reaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("reactions").Field("user_id").Unique().Required(),
		edge.From("issue", Issue.Type).Ref("reactions").Field("issue_id").Unique(),
		edge.From("comment", Comment.Type).Ref("reactions").Field("comment_id").Unique(),
		edge.From("project_update", ProjectPost.Type).Ref("reactions").Field("project_update_id").Unique(),
	}
}

func (Reaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "emoji", "issue_id").Unique(),
		index.Fields("user_id", "emoji", "comment_id").Unique(),
		index.Fields("user_id", "emoji", "project_update_id").Unique(),
	}
}

func (Reaction) Annotations() []schema.Annotation { return table("reactions") }
