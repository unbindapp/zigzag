package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type Workspace struct {
	ent.Schema
}

func (Workspace) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("slug").Unique().Comment("URL key, lowercase [a-z0-9-]{2,32}"),
		field.String("icon").Optional(),
		field.String("color").Optional(),
		field.Time("archived_at").Optional().Nillable(),
	}
}

func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", WorkspaceMember.Type).Annotations(cascade()),
		edge.To("teams", Team.Type).Annotations(cascade()),
		edge.To("projects", Project.Type).Annotations(cascade()),
		edge.To("project_statuses", ProjectStatus.Type).Annotations(cascade()),
		edge.To("labels", Label.Type).Annotations(cascade()),
		edge.To("initiatives", Initiative.Type).Annotations(cascade()),
		edge.To("documents", Document.Type).Annotations(cascade()),
		edge.To("views", View.Type).Annotations(cascade()),
		edge.To("invitations", Invitation.Type).Annotations(cascade()),
		edge.To("webhooks", Webhook.Type).Annotations(cascade()),
		edge.To("uploads", Upload.Type).Annotations(cascade()),
		edge.To("templates", Template.Type).Annotations(cascade()),
		edge.To("imports", ImportJob.Type).Annotations(cascade()),
		edge.To("issues", Issue.Type).Annotations(cascade()),
		edge.To("notifications", Notification.Type).Annotations(cascade()),
		edge.To("favorites", Favorite.Type).Annotations(cascade()),
	}
}

func (Workspace) Annotations() []schema.Annotation { return table("workspaces") }
