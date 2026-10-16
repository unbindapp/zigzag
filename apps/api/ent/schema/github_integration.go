package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type GithubIntegration struct {
	ent.Schema
}

func (GithubIntegration) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (GithubIntegration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}).Unique(),
		field.Bytes("secret_encrypted").Sensitive().Comment("Webhook secret sealed with ENCRYPTION_KEY"),
		field.String("secret_prefix"),
		field.UUID("created_by_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("last_delivery_at").Optional().Nillable(),
		field.Bytes("token_encrypted").Optional().Sensitive().Comment("GitHub token sealed with ENCRYPTION_KEY, for Issues sync"),
		field.String("token_prefix").Default(""),
	}
}

func (GithubIntegration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("github_integration").Field("workspace_id").Unique().Required(),
		edge.From("created_by", User.Type).Ref("github_integrations").Field("created_by_id").Unique(),
		edge.To("repositories", GithubRepository.Type).Annotations(cascade()),
		edge.To("repo_syncs", GithubRepoSync.Type).Annotations(cascade()),
	}
}

func (GithubIntegration) Annotations() []schema.Annotation { return table("github_integrations") }
