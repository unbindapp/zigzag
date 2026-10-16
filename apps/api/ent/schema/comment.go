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

type Comment struct {
	ent.Schema
}

func (Comment) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("issue_id", uuid.UUID{}),
		field.UUID("author_id", uuid.UUID{}).Optional().Nillable(),
		field.String("body").Comment("Markdown"),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable().Comment("Reply thread, one level"),
		field.Time("edited_at").Optional().Nillable(),
		field.Time("resolved_at").Optional().Nillable(),
		field.UUID("resolved_by_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).Ref("comments").Field("issue_id").Unique().Required(),
		edge.From("author", User.Type).Ref("comments").Field("author_id").Unique(),
		edge.From("resolved_by", User.Type).Ref("resolved_comments").Field("resolved_by_id").Unique(),
		edge.To("replies", Comment.Type).Annotations(cascade()).From("parent").Field("parent_id").Unique(),
		edge.To("reactions", Reaction.Type).Annotations(cascade()),
		edge.To("notifications", Notification.Type).Annotations(setNull()),
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{index.Fields("issue_id", "created_at")}
}

func (Comment) Annotations() []schema.Annotation { return table("comments") }
