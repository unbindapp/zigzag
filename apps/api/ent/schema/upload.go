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

type Upload struct {
	ent.Schema
}

func (Upload) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Upload) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("uploader_id", uuid.UUID{}).Optional().Nillable(),
		field.String("file_name"),
		field.String("content_type"),
		field.Int64("size_bytes"),
		field.String("storage_key").Unique(),
		field.Int("width").Optional().Nillable(),
		field.Int("height").Optional().Nillable(),
	}
}

func (Upload) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("uploads").Field("workspace_id").Unique().Required(),
		edge.From("uploader", User.Type).Ref("uploads").Field("uploader_id").Unique(),
		edge.To("avatar_users", User.Type).Annotations(setNull()),
		edge.To("icon_workspaces", Workspace.Type).Annotations(setNull()),
		edge.To("attachments", Attachment.Type).Annotations(setNull()),
		edge.To("blob", UploadBlob.Type).Unique().Annotations(cascade()),
	}
}

func (Upload) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id")}
}

func (Upload) Annotations() []schema.Annotation { return table("uploads") }
