package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type GithubCommentLink struct {
	ent.Schema
}

func (GithubCommentLink) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (GithubCommentLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("comment_id", uuid.UUID{}).Unique().Comment("No foreign key: the link outlives a deleted comment until GitHub deletes its copy"),
		field.UUID("issue_link_id", uuid.UUID{}),
		field.Int64("github_comment_id").Unique(),
	}
}

func (GithubCommentLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue_link", GithubIssueLink.Type).Ref("comment_links").Field("issue_link_id").Unique().Required(),
	}
}

func (GithubCommentLink) Annotations() []schema.Annotation { return table("github_comment_links") }
