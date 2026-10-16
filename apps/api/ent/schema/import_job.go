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

type ImportJob struct {
	ent.Schema
}

func (ImportJob) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (ImportJob) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.Enum("source").GoType(ImportSource("")),
		field.UUID("started_by_id", uuid.UUID{}).Optional().Nillable(),
		field.Bool("dry_run").Default(true),
		field.Enum("status").GoType(ImportStatus("")).Default(string(ImportStatusRunning)),
		field.JSON("report", &ImportReport{}).Optional(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (ImportJob) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("imports").Field("workspace_id").Unique().Required(),
		edge.From("started_by", User.Type).Ref("imports").Field("started_by_id").Unique(),
	}
}

func (ImportJob) Indexes() []ent.Index {
	return []ent.Index{index.Fields("workspace_id", "created_at")}
}

func (ImportJob) Annotations() []schema.Annotation { return table("imports") }
