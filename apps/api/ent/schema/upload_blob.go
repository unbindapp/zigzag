package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type UploadBlob struct {
	ent.Schema
}

func (UploadBlob) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}}
}

func (UploadBlob) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("upload_id", uuid.UUID{}).Unique().Immutable(),
		field.Bytes("data"),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (UploadBlob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("upload", Upload.Type).Ref("blob").Field("upload_id").Unique().Required().Immutable(),
	}
}

func (UploadBlob) Annotations() []schema.Annotation { return table("upload_blobs") }
