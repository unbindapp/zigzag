package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema/mixin"
)

type Issue struct {
	ent.Schema
}

func (Issue) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Issue) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}).Comment("Denormalized from the team"),
		field.UUID("team_id", uuid.UUID{}),
		field.Int("number"),
		field.String("title").NotEmpty().MaxRuneLen(255),
		field.String("description").Optional().Comment("Markdown"),
		field.String("description_text").Optional().Comment("Plain text derived from description, for search"),
		field.Int("priority").Default(0).Comment("0 none, 1 urgent, 2 high, 3 medium, 4 low"),
		field.Float("estimate").Optional().Nillable(),
		field.UUID("state_id", uuid.UUID{}),
		field.UUID("assignee_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("creator_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("milestone_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("cycle_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.Float("sort_order").Default(0),
		field.Float("sub_issue_sort_order").Default(0),
		field.Time("due_date").Optional().Nillable().SchemaType(dateType),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("canceled_at").Optional().Nillable(),
		field.Time("triaged_at").Optional().Nillable(),
		field.Time("snoozed_until").Optional().Nillable(),
		field.Time("archived_at").Optional().Nillable(),
		field.Time("trashed_at").Optional().Nillable(),
		field.Time("auto_closed_at").Optional().Nillable(),
		field.Time("auto_archived_at").Optional().Nillable(),
		field.String("import_source").Optional(),
		field.String("import_external_id").Optional(),
		field.String("import_url").Optional(),
	}
}

func (Issue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("issues").Field("workspace_id").Unique().Required(),
		edge.From("team", Team.Type).Ref("issues").Field("team_id").Unique().Required(),
		edge.From("state", WorkflowState.Type).Ref("issues").Field("state_id").Unique().Required(),
		edge.From("assignee", User.Type).Ref("assigned_issues").Field("assignee_id").Unique(),
		edge.From("creator", User.Type).Ref("created_issues").Field("creator_id").Unique(),
		edge.From("project", Project.Type).Ref("issues").Field("project_id").Unique(),
		edge.From("milestone", ProjectMilestone.Type).Ref("issues").Field("milestone_id").Unique(),
		edge.From("cycle", Cycle.Type).Ref("issues").Field("cycle_id").Unique(),
		edge.To("children", Issue.Type).Annotations(setNull()).From("parent").Field("parent_id").Unique(),
		edge.To("labels", Label.Type).StorageKey(edge.Table("issue_labels"), edge.Columns("issue_id", "label_id")),
		edge.To("subscribers", User.Type).StorageKey(edge.Table("issue_subscriptions"), edge.Columns("issue_id", "user_id")),
		edge.To("comments", Comment.Type).Annotations(cascade()),
		edge.To("attachments", Attachment.Type).Annotations(cascade()),
		edge.To("relations", IssueRelation.Type).Annotations(cascade()),
		edge.To("inverse_relations", IssueRelation.Type).Annotations(cascade()),
		edge.To("history", IssueHistory.Type).Annotations(cascade()),
		edge.To("reactions", Reaction.Type).Annotations(cascade()),
		edge.To("moves", IssueMove.Type).Annotations(cascade()),
		edge.To("github_link", GithubIssueLink.Type).Unique().Annotations(cascade()),
		edge.To("documents", Document.Type).Annotations(cascade()),
		edge.To("notifications", Notification.Type).Annotations(setNull()),
	}
}

func (Issue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "number").Unique(),
		index.Fields("team_id", "state_id"),
		index.Fields("workspace_id", "assignee_id"),
		index.Fields("project_id"),
		index.Fields("cycle_id"),
		index.Fields("parent_id"),
		index.Fields("workspace_id", "updated_at"),
		index.Fields("workspace_id", "sort_order"),
		index.Fields("workspace_id", "import_source", "import_external_id").
			Unique().
			Annotations(entsql.IndexWhere("import_source IS NOT NULL AND import_external_id IS NOT NULL")),
		index.Fields("title").
			StorageKey("issues_title_trgm").
			Annotations(entsql.IndexTypes(map[string]string{dialect.Postgres: "GIN"}), entsql.OpClass("gin_trgm_ops")),
		index.Fields("description_text").
			StorageKey("issues_description_text_trgm").
			Annotations(entsql.IndexTypes(map[string]string{dialect.Postgres: "GIN"}), entsql.OpClass("gin_trgm_ops")),
	}
}

func (Issue) Annotations() []schema.Annotation { return table("issues") }
