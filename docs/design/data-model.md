# Data model

Every entity uses `mixin.PKMixin` (uuid `id`) and `mixin.TimeMixin` (`created_at`, `updated_at`) unless noted. Go type names that would collide with ent: **Import** is `ImportJob` (`import` is a Go keyword), **ProjectUpdate** is `ProjectPost` and the initiative update is `InitiativePost` (ent generates `ProjectUpdate`/`ProjectStatusUpdate` builders); table names are unchanged. Table names are plural snake case via `entsql.Annotation{Table: ...}`. Enums are Go string types with `Values()` and a huma `Schema` registration, like Unbind's `PermittedAction`. Foreign keys are explicit `field.UUID` fields with `edge.From(...).Field(...)`. Deletes cascade from parent to child unless a `SetNull` is listed.

Soft deletion: entities with `archived_at` are hidden by default from lists and shown with `include_archived`. Entities with `trashed_at` are hidden by default and shown with `include_trashed`; a trashed row is hard deleted by the scheduler 30 days after `trashed_at`.

## Accounts

**User** `users`
- `email` string unique (lowercased)
- `password_hash` string sensitive, empty for placeholder accounts that have never set a password
- `name` string (full name), `display_name` string (short handle, unique per instance, lowercase, used in `@mentions`, validated as `[a-z0-9][a-z0-9._-]{0,31}`)
- `avatar_upload_id` uuid optional nillable (SetNull)
- `system_admin` bool default false
- `active` bool default true. Placeholder users created by an import are `active=false` until they accept an invitation or an admin sets a password.
- `timezone` string default "UTC"
- `last_seen_at` time optional nillable
- edges: `workspace_memberships`, `team_memberships`, `api_keys`, `oauth2_tokens`, `oauth_grants`, `oauth_authorization_codes`, `notifications`, `favorites`, `sent_invitations` (SetNull), `password_reset_tokens`, `issue_subscriptions`

**Oauth2Token**, **JWTKey**, **OAuthClient**, **OAuthGrant**, **OAuthGrantToken**, **OAuthAuthorizationCode**, **PasswordResetToken**: port from Unbind unchanged, except that `oauth2_tokens.access_token` and `refresh_token` store the SHA-256 hex of the session tokens, never the tokens (`password_reset_tokens`: `user_id`, `token_hash` unique sensitive, `expires_at`, `used_at` nillable), except that grants and codes carry the API key access model: `role` WorkspaceRole, `full_access`, `workspace_ids` JSON `[]uuid.UUID`, `privileges`. There is no bootstrap table: the first user exists when `users` is not empty, and first-run setup serializes on a row lock of `system_settings`.

**APIKey** `api_keys`
- `name` string, `token_prefix` string, `token_hash` string unique sensitive
- `role` enum WorkspaceRole: `admin|member|viewer`
- `full_access` bool default false
- `workspace_ids` JSON `[]uuid.UUID` (empty when `full_access`)
- `privileges` JSON `[]KeyPrivilege` default empty
- `expires_at`, `last_used_at` time optional nillable
- `user_id` uuid
- index unique `(user_id, name)`

**SystemSettings** `system_settings` singleton
- `instance_name` string default "Zigzag"
- `email_settings` JSON `*EmailSettings{Host, Port, Username, FromAddress, FromName, Encryption}` optional
- `email_password_encrypted` bytes optional sensitive
- `allow_public_signup` bool default false

## Workspaces

**Workspace** `workspaces`
- `name` string, `slug` string unique (url key, lowercase `[a-z0-9-]{2,32}`, same validation as Unbind `names.Clean` then slugified)
- `icon` string optional (emoji or icon name), `color` string optional (hex)
- `issue_counter_mode` not needed: counters live on teams
- `archived_at` time optional nillable
- edges: `members`, `teams`, `projects`, `project_statuses`, `labels`, `initiatives`, `documents`, `views`, `invitations`, `webhooks`, `uploads`, `templates`, `imports`

**WorkspaceMember** `workspace_members`
- `workspace_id`, `user_id` uuid, unique together
- `role` enum WorkspaceRole
- `joined_at` time default now

**Invitation** `invitations`: port Unbind with `workspace_id` instead of `team_id`. Fields: `email`, `token_hash` unique sensitive, `role` WorkspaceRole, `expires_at`, `accepted_at` nillable, `workspace_id`, `invited_by` nillable, `team_ids` JSON `[]uuid.UUID` (teams to join on accept). Index `(workspace_id, email)` unique.

**Upload** `uploads`
- `workspace_id` uuid, `uploader_id` uuid nillable (SetNull)
- `file_name` string, `content_type` string, `size_bytes` int64, `storage_key` string unique
- `width`, `height` int optional nillable (images)

## Teams

**Team** `teams`
- `workspace_id` uuid
- `name` string, `key` string (uppercase `[A-Z][A-Z0-9]{0,6}`), unique `(workspace_id, key)` and `(workspace_id, name)`
- `description` string optional, `icon` string optional, `color` string optional
- `issue_counter` int default 0 (last issue number handed out; incremented atomically with `UPDATE ... SET issue_counter = issue_counter + 1 RETURNING`)
- `timezone` string default "UTC"
- `private` bool default false (private teams are listed only to their members and workspace admins)
- `cycles_enabled` bool default false, `cycle_duration_weeks` int default 2, `cycle_cooldown_weeks` int default 0, `cycle_start_day` int default 1 (Monday=1), `cycle_auto_move` bool default true (move unfinished issues to the next cycle), `cycle_auto_add_started` bool default true (an issue that starts joins the active cycle)
- `estimate_type` enum `none|exponential|fibonacci|linear|tshirt` default `none`, `estimate_allow_zero` bool default false, `estimate_extended` bool default false
- `triage_enabled` bool default false
- `default_state_id` uuid nillable (state new issues get), `triage_state_id` uuid nillable
- `auto_close_months` int nillable (auto close stale issues; nil disables), `auto_archive_months` int default 6 (completed and canceled issues archive after this)
- `archived_at` nillable
- edges: `members`, `states`, `issues`, `cycles`, `labels` (team labels), `projects` (M2M through `project_teams`), `templates`, `views`

**TeamMember** `team_members`: `team_id`, `user_id` unique together, `owner` bool default false.

**WorkflowState** `workflow_states`
- `team_id` uuid
- `name` string, unique `(team_id, name)`
- `type` enum StateType `triage|backlog|unstarted|started|completed|canceled`
- `color` string, `description` string optional
- `position` float64 (order within type)
- `archived_at` nillable
- `issues.state_id` is ON DELETE NO ACTION: `delete-workflow-state` moves the issues first. Same for `projects.status_id` and `delete-project-status`.
- Every team is created with: Triage (triage, only when triage enabled, color `#fc7840`; also created when triage is turned on later), Backlog (backlog), Todo (unstarted), In Progress (started), In Review (started), Done (completed), Canceled (canceled), Duplicate (canceled). Colors: backlog `#95a2b3`, unstarted `#e2e2e2`, started `#f2c94c`, review `#0f783c`, completed `#5e6ad2`, canceled `#95a2b3`.

**Cycle** `cycles`
- `team_id` uuid, `number` int, unique `(team_id, number)`
- `name` string optional, `description` string optional
- `starts_at`, `ends_at` time; `completed_at` nillable
- `archived_at` nillable
- Computed in responses: `progress` (completed scope / total scope), `issue_count`, `completed_issue_count`, `scope` (sum of estimates or count), `is_active`, `is_past`, `is_future`. Scope counts issues when the team's `estimate_type` is `none`, else sums estimates with an unestimated issue counting 1. Trashed and canceled issues are left out of scope and counts.
- `cycle_service.ProcessCycles` (scheduler, every 10 minutes): for teams with cycles on, keeps two unfinished cycles ending in the future (the first starts at local midnight of the latest `cycle_start_day` in the team time zone, later ones chain after the previous end plus `cycle_cooldown_weeks`); completes cycles whose `ends_at` passed and, with `cycle_auto_move`, moves their issues not in a completed or canceled state to the next open cycle, recording a `cycle` history entry with source `system`; upserts today's `cycle_scope_history` row for every active cycle and a last one when a cycle closes.

**CycleScopeHistory** `cycle_scope_history`
- `cycle_id` uuid (cascade), `date` date (day in the team time zone), `scope`, `completed`, `started` float64 (same measure as the cycle's progress)
- unique `(cycle_id, date)`; one row per cycle per day, overwritten during the day

## Labels

**Label** `labels`
- `workspace_id` uuid, `team_id` uuid nillable (nil = workspace label)
- `name` string, `color` string, `description` string optional
- `parent_id` uuid nillable (label group), `is_group` bool default false
- `group_single_select` bool default false (only one child of the group may be applied to an issue)
- `sort_order` float64 default 0 (order among siblings, set by drag and drop)
- `archived_at` nillable
- unique `(workspace_id, team_id, parent_id, name)` enforced in the service (null-safe)
- A group cannot be applied to issues. Applying a child of a single-select group replaces a sibling already on the issue.

## Projects

**ProjectStatus** `project_statuses`
- `workspace_id` uuid, `name` string, `type` enum `backlog|planned|started|paused|completed|canceled`, `color` string, `position` float64, `description` optional
- Every workspace gets: Backlog, Planned, In Progress, Paused, Completed, Canceled (colors `#95a2b3`, `#e2e2e2`, `#f2c94c`, `#f2994a`, `#5e6ad2`, `#95a2b3`).

**Project** `projects`
- `workspace_id` uuid
- `name` string, `slug` string unique per workspace (derived from name, 8 random hex chars appended like Linear `name-1e0e8889`), `summary` string optional (one line), `description` string optional (markdown), `icon` optional, `color` optional
- `status_id` uuid (ProjectStatus), `lead_id` uuid nillable (SetNull)
- `priority` int default 0 (0 none, 1 urgent, 2 high, 3 medium, 4 low; same scale as issues)
- `start_date`, `target_date` date optional nillable; `start_date_resolution`, `target_date_resolution` enum `day|month|quarter|half_year|year` default `day`
- `sort_order` float64
- `started_at`, `completed_at`, `canceled_at`, `archived_at`, `trashed_at` nillable
- `health` enum `on_track|at_risk|off_track` nillable (last update's health)
- `creator_id` uuid nillable
- `import_source` string optional, `import_external_id` string optional; unique `(workspace_id, import_source, import_external_id)` when both set
- edges: `teams` M2M (`project_teams`), `members` M2M (`project_members`), `milestones`, `issues`, `updates`, `labels` M2M (`project_labels`, uses workspace labels), `initiatives` M2M (`initiative_projects`), `documents`, `links`
- Computed in responses: `progress`, `scope`, `completed_scope`, `started_scope`, `issue_counts_by_state_type`.

**ProjectMilestone** `project_milestones`
- `project_id`, `name`, `description` optional, `target_date` nillable, `sort_order` float64, `archived_at` nillable
- Computed: progress.

**ProjectUpdate** `project_updates` (status updates)
- `project_id`, `author_id` nillable, `body` markdown, `health` enum, `edited_at` nillable

**ProjectLink** `project_links`: `project_id`, `url`, `label`, `creator_id` nillable, `sort_order`.

**Initiative** `initiatives`
- `workspace_id`, `name`, `slug` unique per workspace, `description` markdown optional, `icon`, `color`, `owner_id` nillable, `status` enum `planned|active|completed`, `target_date` nillable, `sort_order`, `archived_at`, `trashed_at`
- edges: `projects` M2M, `updates` (`initiative_updates`: body, health, author), `documents`

## Issues

**Issue** `issues`
- `workspace_id` uuid (denormalized from team for cheap workspace-wide queries), `team_id` uuid
- `number` int, unique `(team_id, number)`; identifier is computed `team.key + "-" + number` and returned on every response
- `title` string (max 255), `description` string optional (markdown), `description_text` string optional (plain text for search, derived)
- `priority` int default 0 (0 none, 1 urgent, 2 high, 3 medium, 4 low)
- `estimate` float64 nillable
- `state_id` uuid, `assignee_id` nillable (SetNull), `creator_id` nillable (SetNull), `project_id` nillable (SetNull), `milestone_id` nillable (SetNull), `cycle_id` nillable (SetNull), `parent_id` nillable (SetNull; cascade is not wanted: children survive a parent delete)
- `sort_order` float64 (manual order in lists and boards), `sub_issue_sort_order` float64 (order under its parent)
- `due_date` date nillable
- `started_at`, `completed_at`, `canceled_at`, `triaged_at`, `snoozed_until`, `archived_at`, `trashed_at` nillable
- `auto_closed_at`, `auto_archived_at` nillable
- `import_source` string optional, `import_external_id` string optional; unique `(workspace_id, import_source, import_external_id)` when both set
- `import_url` string optional (link back to the source issue)
- edges: `labels` M2M (`issue_labels`), `subscribers` M2M (`issue_subscriptions`), `comments`, `attachments`, `relations`, `history`, `children`, `reactions`
- Indexes: `(team_id, state_id)`, `(workspace_id, assignee_id)`, `(project_id)`, `(cycle_id)`, `(parent_id)`, `(workspace_id, updated_at)`, `(workspace_id, sort_order)` (the default manual order), GIN trigram on `title` (`pg_trgm`), plus `description_text`. The trigram indexes are declared in the ent schema (`entsql.IndexTypes` GIN + `entsql.OpClass("gin_trgm_ops")`) so atlas generates them and never drops them in a later diff.
- Manual order: `sort_order` and `sub_issue_sort_order` are floats set by drag and drop. When a new value lands within 1e-9 (relative) of a sibling, the whole scope is renumbered 1024 apart in list order in the same transaction (issues per workspace, sub-issues per parent; labels, projects, milestones, workflow state and project status positions the same way).
- Rules enforced in `issue_service`: state change sets or clears `started_at`, `completed_at`, `canceled_at` by state type; moving to a `started` state with `cycle_auto_add_started` adds the active cycle when the issue has none; a parent cannot be its own descendant; team change renumbers the issue in the new team and keeps a redirect row in `issue_moves` (`old_team_id`, `old_number`, `issue_id`) so old identifiers resolve; assignee is auto-subscribed; creator is auto-subscribed; mentioned users are subscribed.

**IssueRelation** `issue_relations`
- `issue_id`, `related_issue_id`, `type` enum `blocks|related|duplicate`, unique `(issue_id, related_issue_id, type)`
- `blocks` is stored once from the blocker; responses show `blocked_by` on the other side. `duplicate` is stored from the duplicate to the canonical issue; marking a duplicate also moves the issue to the team's `Duplicate` state when it exists, else the first canceled state.

**Comment** `comments`
- `issue_id`, `author_id` nillable (SetNull), `body` markdown, `parent_id` nillable (reply thread, one level), `edited_at`, `resolved_at` nillable, `resolved_by_id` nillable
- `import_source` string optional, `import_external_id` string optional; unique `(issue_id, import_source, import_external_id)` when both set
- edges: `reactions`

**Reaction** `reactions`: `emoji` string, `user_id`, exactly one of `issue_id`, `comment_id`, `project_update_id`. Unique `(user_id, emoji, issue_id)`, `(user_id, emoji, comment_id)`, `(user_id, emoji, project_update_id)`.

**Attachment** `attachments`
- `issue_id`, `creator_id` nillable, `title` string, `subtitle` string optional, `url` string, `source_type` string optional (`upload|link|github|import`), `upload_id` nillable, `metadata` JSON map optional, `sort_order`

**IssueHistory** `issue_history` (no `updated_at`)
- `issue_id`, `actor_id` nillable (nil for system and imports), `created_at`
- `changes` JSON `[]FieldChange{Field, From, To, FromLabel, ToLabel}`; field names: `title, description, state, priority, estimate, assignee, project, milestone, cycle, parent, due_date, labels_added, labels_removed, relation_added, relation_removed, attachment_added, attachment_removed, archived, unarchived, trashed, restored, team, created, imported`
- `source` enum `user|api|mcp|import|system`
- Besides issue mutations, these write history rows on every affected issue in the same transaction: deleting a project (`project`, and `milestone` when set), a milestone (`milestone`), a cycle (`cycle`), a label (`labels_removed`), a workflow state (`state`, with the target state's timestamps applied), and hard deleting an issue (`parent` on its children).

**Notification** `notifications`
- `user_id`, `actor_id` nillable, `workspace_id`
- `type` enum `issue_assigned|issue_unassigned|issue_mentioned|comment_mentioned|issue_commented|issue_state_changed|issue_priority_changed|issue_due_soon|issue_blocking|project_update|project_added_as_member|project_lead|reaction`
- `issue_id`, `comment_id`, `project_id`, `project_update_id` nillable (SetNull)
- `title` string, `body` string optional (snippet)
- `read_at`, `archived_at`, `snoozed_until`, `emailed_at` nillable
- Index `(user_id, archived_at, read_at)`

**NotificationSetting** `notification_settings`: `user_id` unique, `email_enabled` bool default true, per type bools `assigned`, `mentioned`, `commented`, `state_changed`, `due_soon`, `project_updates` default true. A user without a row gets the defaults. The type bools decide whether `notification_service.Notify` writes the row at all: `assigned` covers `issue_assigned|issue_unassigned`, `mentioned` covers `issue_mentioned|comment_mentioned`, `commented` covers `issue_commented`, `state_changed` covers `issue_state_changed|issue_priority_changed|issue_blocking`, `due_soon` covers `issue_due_soon`, `project_updates` covers `project_update`; `project_added_as_member`, `project_lead` and `reaction` are always written. `Notify` skips the actor and anyone who is not an active member of the workspace; for an issue of a private team it also skips members who are neither on the team nor workspace admins. `email_enabled` only gates the email digest.

## Documents, views, templates, favorites

**Document** `documents`
- `workspace_id`, `project_id` nillable, `initiative_id` nillable, `issue_id` nillable, `creator_id` nillable, `updated_by_id` nillable
- `title`, `slug` unique per workspace, `icon`, `content` markdown, `archived_at`, `trashed_at`

**View** `views` (custom views)
- `workspace_id`, `team_id` nillable, `owner_id`, `name`, `description` optional, `icon`, `color`
- `kind` enum `issue|project`, `shared` bool default false
- `filters` JSON (same shape as `list-issues` filter input), `display` JSON `{layout: list|board, group_by, order_by, show_sub_issues, show_completed}`

**Template** `templates`: `workspace_id`, `team_id` nillable, `kind` enum `issue|project|document`, `name`, `description` optional, `data` JSON (for issues: title, description, priority, estimate, label_ids, assignee_id, state_id, project_id, sub_issues `[{title, description}]`), `sort_order`, `creator_id` nillable.

**Favorite** `favorites`: `user_id`, `workspace_id`, `kind` enum `issue|project|cycle|view|document|label|team|initiative`, `entity_id` uuid, `sort_order`. Unique `(user_id, kind, entity_id)`.

**Webhook** `webhooks`: `workspace_id`, `team_id` nillable, `url`, `secret` sensitive, `enabled` bool, `events` JSON `[]string` (`issue.created|issue.updated|issue.deleted|comment.created|project.updated|cycle.updated`), `label` string. Deliveries are fire and forget: a JSON POST of `{id, event, workspace_id, created_at, data}` with headers `X-Zigzag-Event`, `X-Zigzag-Delivery` (the envelope id, the same on every retry) and `X-Zigzag-Signature: sha256=<hex HMAC-SHA256 of the body keyed with the secret>`; any non-2xx answer or network error is retried 3 times (after 2s, 10s, 30s); redirects are not followed. A team-scoped webhook receives an event when its payload's `team_id`, `issue.team_id`, one of `teams[].id`, or the team of its `issue_id` matches. `cycle.updated` carries `{action: created|updated|archived|deleted|completed, team_id, cycle: CycleResponse}` and is sent by create, update, archive and delete, by the scheduler when it plans or completes a cycle, and when a team turns cycles on. Secrets are `whsec_` plus 48 hex chars.

**Import** `imports` (Go `ImportJob`): `workspace_id`, `source` enum `linear`, `started_by_id` nillable (SetNull), `dry_run` bool, `status` enum `running|completed|failed`, `report` JSON, `finished_at` nillable.

**IssueMove** `issue_moves`: `old_team_id`, `old_number`, `issue_id`, unique `(old_team_id, old_number)`.
