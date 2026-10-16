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

type GithubIssueLink struct {
	ent.Schema
}

func (GithubIssueLink) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (GithubIssueLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("issue_id", uuid.UUID{}).Unique(),
		field.UUID("repo_sync_id", uuid.UUID{}),
		field.Int("number"),
		field.String("node_id").NotEmpty().Unique(),
		field.String("html_url"),
		field.Time("github_updated_at").Comment("GitHub updated_at of the newest state applied in either direction"),
		field.Time("last_outbound_at").Optional().Nillable(),
	}
}

func (GithubIssueLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).Ref("github_link").Field("issue_id").Unique().Required(),
		edge.From("repo_sync", GithubRepoSync.Type).Ref("issue_links").Field("repo_sync_id").Unique().Required(),
		edge.To("comment_links", GithubCommentLink.Type).Annotations(cascade()),
	}
}

func (GithubIssueLink) Indexes() []ent.Index {
	return []ent.Index{index.Fields("repo_sync_id", "number")}
}

func (GithubIssueLink) Annotations() []schema.Annotation { return table("github_issue_links") }
