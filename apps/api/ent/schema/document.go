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

type Document struct {
	ent.Schema
}

func (Document) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Document) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("initiative_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("issue_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("updated_by_id", uuid.UUID{}).Optional().Nillable(),
		field.String("title"),
		field.String("slug"),
		field.String("icon").Optional(),
		field.String("content").Default("").Comment("Markdown"),
		field.Time("archived_at").Optional().Nillable(),
		field.Time("trashed_at").Optional().Nillable(),
	}
}

func (Document) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("documents").Field("workspace_id").Unique().Required(),
		edge.From("project", Project.Type).Ref("documents").Field("project_id").Unique(),
		edge.From("initiative", Initiative.Type).Ref("documents").Field("initiative_id").Unique(),
		edge.From("issue", Issue.Type).Ref("documents").Field("issue_id").Unique(),
		edge.From("creator", User.Type).Ref("created_documents").Field("creator_id").Unique(),
		edge.From("updated_by", User.Type).Ref("updated_documents").Field("updated_by_id").Unique(),
	}
}

func (Document) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id", "slug").Unique()}
}

func (Document) Annotations() []schema.Annotation { return table("documents") }
