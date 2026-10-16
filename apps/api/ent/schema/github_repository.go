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

type GithubRepository struct {
	ent.Schema
}

func (GithubRepository) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (GithubRepository) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("integration_id", uuid.UUID{}),
		field.String("full_name").NotEmpty().Comment("owner/name"),
		field.Time("last_event_at"),
		field.String("last_event").Comment("GitHub event and action, like pull_request.opened"),
	}
}

func (GithubRepository) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("integration", GithubIntegration.Type).Ref("repositories").Field("integration_id").Unique().Required(),
	}
}

func (GithubRepository) Indexes() []ent.Index {
	return []ent.Index{index.Fields("integration_id", "full_name").Unique()}
}

func (GithubRepository) Annotations() []schema.Annotation { return table("github_repositories") }
