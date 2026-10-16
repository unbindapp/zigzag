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

type Attachment struct {
	ent.Schema
}

func (Attachment) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("issue_id", uuid.UUID{}),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
		field.String("title"),
		field.String("subtitle").Optional(),
		field.String("url"),
		field.String("source_type").Optional().Comment("upload, link, github, github_branch, github_commit or import"),
		field.UUID("upload_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.Float("sort_order").Default(0),
	}
}

func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).Ref("attachments").Field("issue_id").Unique().Required(),
		edge.From("creator", User.Type).Ref("attachments").Field("creator_id").Unique(),
		edge.From("upload", Upload.Type).Ref("attachments").Field("upload_id").Unique(),
	}
}

func (Attachment) Indexes() []ent.Index {
	return []ent.Index{index.Fields("issue_id")}
}

func (Attachment) Annotations() []schema.Annotation { return table("attachments") }
