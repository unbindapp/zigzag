package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type Project struct {
	ent.Schema
}

func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.String("slug").Comment("Unique per workspace, name plus 8 random hex chars"),
		field.String("summary").Optional(),
		field.String("description").Optional(),
		field.String("icon").Optional(),
		field.String("color").Optional(),
		field.UUID("status_id", uuid.UUID{}),
		field.UUID("lead_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("priority").Default(0).Comment("0 none, 1 urgent, 2 high, 3 medium, 4 low"),
		field.Time("start_date").Optional().Nillable().SchemaType(dateType),
		field.Time("target_date").Optional().Nillable().SchemaType(dateType),
		field.Enum("start_date_resolution").GoType(DateResolution("")).Default(string(DateResolutionDay)),
		field.Enum("target_date_resolution").GoType(DateResolution("")).Default(string(DateResolutionDay)),
		field.Float("sort_order").Default(0),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("canceled_at").Optional().Nillable(),
		field.Time("archived_at").Optional().Nillable(),
		field.Time("trashed_at").Optional().Nillable(),
		field.Enum("health").GoType(Health("")).Optional().Nillable(),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
		field.String("import_source").Optional(),
		field.String("import_external_id").Optional(),
	}
}

func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("projects").Field("workspace_id").Unique().Required(),
		edge.From("status", ProjectStatus.Type).Ref("projects").Field("status_id").Unique().Required(),
		edge.From("lead", User.Type).Ref("led_projects").Field("lead_id").Unique(),
		edge.From("creator", User.Type).Ref("created_projects").Field("creator_id").Unique(),
		edge.To("teams", Team.Type).StorageKey(edge.Table("project_teams"), edge.Columns("project_id", "team_id")),
		edge.To("members", User.Type).StorageKey(edge.Table("project_members"), edge.Columns("project_id", "user_id")),
		edge.To("labels", Label.Type).StorageKey(edge.Table("project_labels"), edge.Columns("project_id", "label_id")),
		edge.From("initiatives", Initiative.Type).Ref("projects"),
		edge.To("milestones", ProjectMilestone.Type).Annotations(cascade()),
		edge.To("issues", Issue.Type).Annotations(setNull()),
		edge.To("updates", ProjectPost.Type).Annotations(cascade()),
		edge.To("documents", Document.Type).Annotations(cascade()),
		edge.To("links", ProjectLink.Type).Annotations(cascade()),
		edge.To("github_repo_syncs", GithubRepoSync.Type).Annotations(setNull()),
		edge.To("notifications", Notification.Type).Annotations(setNull()),
	}
}

func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "slug").Unique(),
		index.Fields("status_id"),
		index.Fields("workspace_id", "import_source", "import_external_id").
			Unique().
			Annotations(entsql.IndexWhere("import_source IS NOT NULL AND import_external_id IS NOT NULL")),
	}
}

func (Project) Annotations() []schema.Annotation { return table("projects") }
