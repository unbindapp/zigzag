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

type Team struct {
	ent.Schema
}

func (Team) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.PKMixin{}, mixin.TimeMixin{}}
}

func (Team) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("workspace_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.String("key").NotEmpty().Comment("Uppercase [A-Z][A-Z0-9]{0,6}"),
		field.String("description").Optional(),
		field.String("icon").Optional(),
		field.String("color").Optional(),
		field.Int("issue_counter").Default(0).Comment("Last issue number handed out"),
		field.String("timezone").Default("UTC"),
		field.Bool("private").Default(false),
		field.Bool("cycles_enabled").Default(false),
		field.Int("cycle_duration_weeks").Default(2),
		field.Int("cycle_cooldown_weeks").Default(0),
		field.Int("cycle_start_day").Default(1).Comment("Monday=1"),
		field.Bool("cycle_auto_move").Default(true),
		field.Bool("cycle_auto_add_started").Default(true),
		field.Enum("estimate_type").GoType(EstimateType("")).Default(string(EstimateTypeNone)),
		field.Bool("estimate_allow_zero").Default(false),
		field.Bool("estimate_extended").Default(false),
		field.Bool("triage_enabled").Default(false),
		field.UUID("default_state_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("triage_state_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("github_pr_opened_state_id", uuid.UUID{}).Optional().Nillable().Comment("Nil means the first started state"),
		field.UUID("github_pr_merged_state_id", uuid.UUID{}).Optional().Nillable().Comment("Nil means the first completed state"),
		field.Int("auto_close_months").Optional().Nillable(),
		field.Int("auto_archive_months").Default(6),
		field.Time("archived_at").Optional().Nillable(),
	}
}

func (Team) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).Ref("teams").Field("workspace_id").Unique().Required(),
		edge.From("default_state", WorkflowState.Type).Ref("default_for_teams").Field("default_state_id").Unique(),
		edge.From("triage_state", WorkflowState.Type).Ref("triage_for_teams").Field("triage_state_id").Unique(),
		edge.From("github_pr_opened_state", WorkflowState.Type).Ref("github_pr_opened_for_teams").Field("github_pr_opened_state_id").Unique(),
		edge.From("github_pr_merged_state", WorkflowState.Type).Ref("github_pr_merged_for_teams").Field("github_pr_merged_state_id").Unique(),
		edge.To("members", TeamMember.Type).Annotations(cascade()),
		edge.To("states", WorkflowState.Type).Annotations(cascade()),
		edge.To("issues", Issue.Type).Annotations(cascade()),
		edge.To("cycles", Cycle.Type).Annotations(cascade()),
		edge.To("labels", Label.Type).Annotations(cascade()),
		edge.From("projects", Project.Type).Ref("teams"),
		edge.To("templates", Template.Type).Annotations(cascade()),
		edge.To("views", View.Type).Annotations(cascade()),
		edge.To("webhooks", Webhook.Type).Annotations(cascade()),
		edge.To("issue_moves", IssueMove.Type).Annotations(cascade()),
		edge.To("github_repo_syncs", GithubRepoSync.Type).Annotations(cascade()),
	}
}

func (Team) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "key").Unique(),
		index.Fields("workspace_id", "name").Unique(),
	}
}

func (Team) Annotations() []schema.Annotation { return table("teams") }
