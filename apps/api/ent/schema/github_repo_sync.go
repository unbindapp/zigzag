package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema/mixin"
)

type GithubRepoSync struct {
	ent.Schema
}

func (GithubRepoSync) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (GithubRepoSync) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("integration_id", uuid.UUID{}),
		field.String("full_name").NotEmpty().Unique().Comment("owner/name, lowercased"),
		field.UUID("team_id", uuid.UUID{}),
		field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
		field.Bool("sync_closed").Default(false).Comment("Backfill imports closed issues too"),
		field.String("last_error").Default(""),
		field.Time("last_synced_at").Optional().Nillable(),
		field.Enum("backfill_status").GoType(GithubBackfillStatus("")).Default(string(GithubBackfillNone)),
		field.JSON("backfill_report", &GithubBackfillReport{}).Optional(),
	}
}

func (GithubRepoSync) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("integration", GithubIntegration.Type).Ref("repo_syncs").Field("integration_id").Unique().Required(),
		edge.From("team", Team.Type).Ref("github_repo_syncs").Field("team_id").Unique().Required(),
		edge.From("project", Project.Type).Ref("github_repo_syncs").Field("project_id").Unique(),
		edge.To("issue_links", GithubIssueLink.Type).Annotations(cascade()),
	}
}

func (GithubRepoSync) Annotations() []schema.Annotation { return table("github_repo_syncs") }
