package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique().Comment("Stored lowercased"),
		field.String("password_hash").Sensitive().Default("").Comment("Empty for placeholder accounts that never set a password"),
		field.String("name").Default(""),
		field.String("display_name").Unique().Comment("Lowercase handle used in mentions"),
		field.UUID("avatar_upload_id", uuid.UUID{}).Optional().Nillable(),
		field.Bool("system_admin").Default(false),
		field.Bool("active").Default(true),
		field.String("timezone").Default("UTC"),
		field.Time("last_seen_at").Optional().Nillable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("avatar_upload", Upload.Type).Ref("avatar_users").Field("avatar_upload_id").Unique(),
		edge.To("workspace_memberships", WorkspaceMember.Type).Annotations(cascade()),
		edge.To("team_memberships", TeamMember.Type).Annotations(cascade()),
		edge.To("api_keys", APIKey.Type).Annotations(cascade()),
		edge.To("oauth2_tokens", Oauth2Token.Type).Annotations(cascade()),
		edge.To("oauth_grants", OAuthGrant.Type).Annotations(cascade()),
		edge.To("oauth_authorization_codes", OAuthAuthorizationCode.Type).Annotations(cascade()),
		edge.To("notifications", Notification.Type).Annotations(cascade()),
		edge.To("acted_notifications", Notification.Type).Annotations(setNull()),
		edge.To("notification_setting", NotificationSetting.Type).Unique().Annotations(cascade()),
		edge.To("favorites", Favorite.Type).Annotations(cascade()),
		edge.To("sent_invitations", Invitation.Type).Annotations(setNull()),
		edge.To("password_reset_tokens", PasswordResetToken.Type).Annotations(cascade()),
		edge.To("uploads", Upload.Type).Annotations(setNull()),
		edge.To("assigned_issues", Issue.Type).Annotations(setNull()),
		edge.To("created_issues", Issue.Type).Annotations(setNull()),
		edge.From("issue_subscriptions", Issue.Type).Ref("subscribers"),
		edge.To("led_projects", Project.Type).Annotations(setNull()),
		edge.To("created_projects", Project.Type).Annotations(setNull()),
		edge.From("project_memberships", Project.Type).Ref("members"),
		edge.To("project_updates", ProjectPost.Type).Annotations(setNull()),
		edge.To("project_links", ProjectLink.Type).Annotations(setNull()),
		edge.To("owned_initiatives", Initiative.Type).Annotations(setNull()),
		edge.To("initiative_updates", InitiativePost.Type).Annotations(setNull()),
		edge.To("comments", Comment.Type).Annotations(setNull()),
		edge.To("resolved_comments", Comment.Type).Annotations(setNull()),
		edge.To("reactions", Reaction.Type).Annotations(cascade()),
		edge.To("attachments", Attachment.Type).Annotations(setNull()),
		edge.To("issue_history", IssueHistory.Type).Annotations(setNull()),
		edge.To("created_documents", Document.Type).Annotations(setNull()),
		edge.To("updated_documents", Document.Type).Annotations(setNull()),
		edge.To("views", View.Type).Annotations(cascade()),
		edge.To("templates", Template.Type).Annotations(setNull()),
		edge.To("imports", ImportJob.Type).Annotations(setNull()),
		edge.To("github_integrations", GithubIntegration.Type).Annotations(setNull()),
	}
}

func (User) Annotations() []schema.Annotation { return table("users") }
