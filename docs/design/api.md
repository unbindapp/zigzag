# API

Conventions (from Unbind): every operation is `oapi.Register(grp, action, huma.Operation{OperationID, Summary, Description, Path, Method}, handler, opts...)`. Paths are `<group prefix>/<verb>`: `/list`, `/get`, `/create`, `/update`, `/delete`, plus named actions. Reads use `GET` with query parameters. Mutations use `POST` (create, invoke), `PUT` (update), `DELETE` (delete) with a JSON body. Responses are `{ "data": ... }`; lists are `{ "data": [...] }` or, when paginated, `{ "data": [...], "next_cursor": "..." | null }`. Deletes return `{ "data": { "id", "deleted": true } }`.

Response models tag every scalar pointer field (`*string`, `*int`, `*time.Time`, `*uuid.UUID`, ...) `nullable:"true"`, because Huma marks pointers non-null otherwise and the generated zod client then rejects `null`; struct pointers cannot take the tag (Huma panics) and use `omitempty` with `required:"false"` instead. Optional input fields stay non-nullable: omitted means unchanged. Response structs are never recursive, directly or through another struct: the web client generator types a recursive schema as `unknown`. Nest a separate shape instead (`IssueDetailResponse.children` holds `IssueResponse`, `CommentResponse.replies` holds `CommentReplyResponse`).

Operation ids are kebab case verbs and become MCP tool names. Summaries are title case. Descriptions are one or two sentences written for an agent: what it does, what it needs, what it returns.

Permission column: the workspace role required (`viewer|member|admin`), checked in the service. `system` means system admin. `public` means unauthenticated. `self` means the caller only.

MCP column: `yes` = `oapi.MCP`; `no` = `oapi.NoMCP(reason)`; `session` = `oapi.SessionOnly` (never a tool). Session-only groups: `/auth`, `/setup`, `/users` (except whoami), `/api-keys`, `/connected-apps`, `/system`, `/uploads`, `/oauth/*`.

All list operations accept `include_archived` (bool, default false) where the entity has `archived_at`, and `include_trashed` where it has `trashed_at`.

Pagination for `list-issues`, `list-notifications`, `list-issue-history`, `list-comments`: `limit` (1..250, default 50) and `cursor` (opaque, from the previous `next_cursor`).

Every response for an issue, project, cycle, team, comment includes the ids it references and the small embedded shapes the UI needs (`state`, `assignee`, `labels`, `project`, `cycle`, `parent {id, identifier, title}`) so a list renders without extra calls. Define these in `internal/models` once and reuse.

## Groups

### /setup (public)
| id | method path | input | output | perm | mcp |
| --- | --- | --- | --- | --- | --- |
| get-setup-status | GET /status | | `{is_first_user_created}` | public | session |
| create-bootstrap-user | POST /create-user | `{email, password, name, display_name, workspace_name, workspace_slug}` | `{user, workspace}` and session cookies | public | session |

### /auth (public)
| login | POST /login | `{email, password}` | session cookies, `{id, email}` | public | session |
| logout | POST /logout | | cleared cookies | self | session |
| request-password-reset | POST /request-password-reset | `{email}` | `{success}` always true | public | session |
| reset-password | POST /reset-password | `{token, password}` | `{success}` | public | session |

### /users
| whoami | GET /whoami | | `{user, system_admin, workspaces: [{id, name, slug, icon, color, role}], api_key?: {role, full_access, workspace_ids, privileges}}` | self | yes |
| update-me | PUT /me | `{name?, display_name?, timezone?, avatar_upload_id?}` (the nil uuid removes the avatar; it must be an image you uploaded) | user | self | session |
| change-password | PUT /password | `{current_password, new_password}` | `{success}` | self | session |
| get-notification-settings | GET /notification-settings | | settings | self | session |
| update-notification-settings | PUT /notification-settings | settings | settings | self | session |
| list-users | GET /list | `{query?, limit?}` | users (system admin sees all; others see members of their workspaces) | self | session |
| create-user | POST /create | `{email, password, name, display_name, system_admin}` | user | system | session |
| update-user | PUT /update | `{user_id, name?, display_name?, system_admin?, active?, password?}` | user | system | session |
| delete-user | DELETE /delete | `{user_id}` | deleted | system | session |

### /workspaces
| list-workspaces | GET /list | | workspaces the caller can see with role | self | yes |
| get-workspace | GET /get | `{workspace_id}` or `{slug}` | workspace with role, counts | viewer | yes |
| create-workspace | POST /create | `{name, slug, icon?, color?}` | workspace | any authenticated user (creator becomes admin); system admin only when `allow_public_signup` is false and caller is not system admin? No: any user may create a workspace | yes |
| update-workspace | PUT /update | `{workspace_id, name?, slug?, icon?, color?}` | workspace | admin | yes |
| delete-workspace | DELETE /delete | `{workspace_id}` | deleted | admin | yes |
| list-workspace-members | GET /members/list | `{workspace_id}` | `[{user: {id, email, name, display_name, avatar_url, active}, role, joined_at, team_ids}]` | viewer | yes |
| add-workspace-member | POST /members/add | `{workspace_id, email, role, team_ids?}` | member (existing account) or invitation (no account) `{member?, invitation?}` | admin | session |
| update-workspace-member | PUT /members/update | `{workspace_id, user_id, role}` | member | admin | session |
| remove-workspace-member | DELETE /members/remove | `{workspace_id, user_id}` | success | admin (or self to leave) | session |
| list-invitations | GET /invitations/list | `{workspace_id}` | invitations (no tokens) | admin | session |
| create-invitation | POST /invitations/create | `{workspace_id, email, role, team_ids?}` | invitation with `link` (one time) and `emailed` bool | admin | session |
| revoke-invitation | DELETE /invitations/revoke | `{workspace_id, invitation_id}` | deleted | admin | session |
| resend-invitation | POST /invitations/resend | `{workspace_id, invitation_id}` | invitation with new link, `emailed` | admin | session |

### /invitations (public)
| get-invitation | GET /get | `{token}` | `{email, workspace_name, role, expires_at, account_exists}` | public | session |
| accept-invitation | POST /accept | `{token, password?, name?, display_name?}` | session cookies, `{user, workspace}`. A new or placeholder account sets `password` (8+ chars); an existing active account must send its current password, since the link is also returned to the inviting admin | public | session |

### /teams
| list-teams | GET /list | `{workspace_id, include_archived?}` | teams with member counts, caller `is_member` | viewer | yes |
| get-team | GET /get | `{team_id}` or `{workspace_id, key}` | team with states, settings | viewer | yes |
| create-team | POST /create | `{workspace_id, name, key, description?, icon?, color?, private?, timezone?}` | team (default states created) | admin | yes |
| update-team | PUT /update | `{team_id, name?, key?, description?, icon?, color?, private?, timezone?, cycles_enabled?, cycle_duration_weeks?, cycle_cooldown_weeks?, cycle_start_day?, cycle_auto_move?, cycle_auto_add_started?, estimate_type?, estimate_allow_zero?, estimate_extended?, triage_enabled?, default_state_id?, auto_close_months?, auto_archive_months?}` | team | admin | yes |
| delete-team | DELETE /delete | `{team_id}` | deleted (issues deleted) | admin | yes |
| list-team-members | GET /members/list | `{team_id}` | `[{user, owner, joined_at}]` | viewer | yes |
| add-team-member | POST /members/add | `{team_id, user_id, owner?}` | member | admin, or member adding self to a non-private team | yes |
| remove-team-member | DELETE /members/remove | `{team_id, user_id}` | success | admin or self | yes |
| list-workflow-states | GET /states/list | `{team_id}` | states ordered by type then position | viewer | yes |
| create-workflow-state | POST /states/create | `{team_id, name, type, color, description?, position?}` | state | admin | yes |
| update-workflow-state | PUT /states/update | `{state_id, name?, color?, description?, position?}` | state | admin | yes |
| delete-workflow-state | DELETE /states/delete | `{state_id, move_issues_to_state_id}` | deleted | admin | yes |

### /labels
| list-labels | GET /list | `{workspace_id, team_id?, all_teams?, include_archived?}`; without `team_id` only workspace labels, `all_teams` adds the labels of every team the caller can see | `LabelTreeResponse` list: top-level labels and groups ordered by `sort_order`, groups with `children` | viewer | yes |
| create-label | POST /create | `{workspace_id, team_id?, name, color, description?, parent_id?, is_group?, group_single_select?}` | label | admin | yes |
| update-label | PUT /update | `{label_id, name?, color?, description?, parent_id? (nil uuid takes it out of its group), group_single_select?, sort_order?}` | label | admin | yes |
| archive-label | POST /archive | `{label_id}` | label | admin | yes |
| unarchive-label | POST /unarchive | `{label_id}` | label | admin | yes |
| delete-label | DELETE /delete | `{label_id}` | deleted (removed from issues) | admin | yes |

### /projects
| list-projects | GET /list | `{workspace_id, team_id?, status_type?, status_id?, lead_id?, member_id?, initiative_id?, label_id?, query?, include_archived?, include_trashed?, order_by? (name|priority|target_date|created_at|updated_at|sort_order)}` | projects with progress | viewer | yes |
| get-project | GET /get | `{project_id}` or `{workspace_id, slug}` | project with teams, members, milestones, links, progress, last update | viewer | yes |
| create-project | POST /create | `{workspace_id, name, team_ids, summary?, description?, icon?, color?, status_id?, lead_id?, member_ids?, priority?, start_date?, target_date?, start_date_resolution?, target_date_resolution?, label_ids?, initiative_ids?}` | project | member | yes |
| update-project | PUT /update | all of the above optional plus `project_id`, `sort_order?` | project | member | yes |
| delete-project | DELETE /delete | `{project_id}` | deleted | admin | yes |
| archive-project / unarchive-project | POST /archive, /unarchive | `{project_id}` | project | member | yes |
| trash-project / restore-project | POST /trash, /restore | `{project_id}` | project | member | yes |
| list-project-statuses | GET /statuses/list | `{workspace_id}` | statuses | viewer | yes |
| create-project-status / update-project-status / delete-project-status | POST /statuses/create, PUT /statuses/update, DELETE /statuses/delete | name, type, color, position; delete needs `move_to_status_id` | admin | yes |
| list-milestones | GET /milestones/list | `{project_id}` | milestones with progress | viewer | yes |
| create-milestone / update-milestone / delete-milestone | POST /milestones/create, PUT /milestones/update, DELETE /milestones/delete | `{project_id, name, description?, target_date?, sort_order?}` / `{milestone_id, ...}` | milestone | member | yes |
| list-project-updates | GET /updates/list | `{project_id}` | updates with author, reactions | viewer | yes |
| create-project-update / update-project-update / delete-project-update | POST /updates/create, PUT /updates/update, DELETE /updates/delete | `{project_id, body, health}` / `{update_id, body?, health?}` / `{update_id}` | update; the project's `health` follows the newest update and its lead and members get `project_update` | member (author or admin to edit or delete) | yes |
| create-project-link / update-project-link / delete-project-link | POST /links/create, PUT /links/update, DELETE /links/delete | `{project_id, url, label}` / `{link_id, url?, label?, sort_order?}` / `{link_id}`; http and https URLs only | link | member | yes |

Project inputs: dates are `YYYY-MM-DD` strings; on update an empty string clears a date and the nil uuid clears `lead_id`; list fields (`team_ids`, `member_ids`, `label_ids`, `initiative_ids`) replace the whole list and `team_ids` cannot be empty. `label_ids` take workspace labels that are not groups. Moving to a started or completed status stamps `started_at` once; completed stamps `completed_at` and canceled `canceled_at`, and leaving either clears it. `create-project` defaults the status to the first backlog status. Lead and members are notified (`project_lead`, `project_added_as_member`) when they are newly set; `update-project`, archive, unarchive, trash and restore emit the `project.updated` webhook with the project.

### /initiatives
| list-initiatives, get-initiative, create-initiative, update-initiative, delete-initiative, archive-initiative, unarchive-initiative | standard | `{workspace_id, name, description?, icon?, color?, owner_id?, status?, target_date?, project_ids?}`; list takes `{workspace_id, status?, owner_id?, include_archived?, include_trashed?}`, get takes `{initiative_id}` or `{workspace_id, slug}`; update takes `{initiative_id, ..., sort_order?}` with the same clearing rules as projects (nil uuid clears `owner_id`, `project_ids` replaces) | initiative with projects and progress | viewer / member / admin for delete | yes |
| add-initiative-project / remove-initiative-project | POST /projects/add, DELETE /projects/remove | `{initiative_id, project_id}` | initiative | member | yes |
| create-initiative-update / list-initiative-updates / delete-initiative-update | POST /updates/create, GET /updates/list, DELETE /updates/delete | `{initiative_id, body, health}` / `{initiative_id}` / `{update_id}` | update; the initiative's `health` is the newest update's | member (author or admin to delete) | yes |

### /cycles
| list-cycles | GET /list | `{team_id, include_archived?}` | cycles with progress, ordered by number desc | viewer | yes |
| get-cycle | GET /get | `{cycle_id}` or `{team_id, number}` | cycle with progress and `scope_history` (daily snapshots, plus a live point for today while active) | viewer | yes |
| create-cycle | POST /create | `{team_id, name?, description?, starts_at, ends_at}` | cycle | admin | yes |
| update-cycle | PUT /update | `{cycle_id, name?, description?, starts_at?, ends_at?}` | cycle | admin | yes |
| delete-cycle | DELETE /delete | `{cycle_id}` | deleted (issues unassigned from the cycle) | admin | yes |
| archive-cycle | POST /archive | `{cycle_id}` | cycle | admin | yes |

### /issues
| list-issues | GET /list | `{workspace_id, team_id?, team_ids?, state_ids?, state_types?, assignee_ids? ("null" for unassigned), creator_ids?, label_ids?, project_ids?, milestone_ids?, cycle_ids?, parent_id?, has_parent? ("true"|"false"), priorities?, due_before?, due_after?, created_after?, updated_after?, query?, subscriber_id?, include_archived?, include_trashed?, order_by? (manual|priority|created_at|updated_at|due_date|title), order? (asc|desc), group_by? (none|state|assignee|priority|project|cycle|label|parent), limit?, cursor?}`. `group_by` sorts by the group first (states in workflow order, priority urgent to none, empty groups last; `label` by the alphabetically first label), then by `order_by`, so pages arrive group by group. `order` defaults to desc for created_at and updated_at, asc otherwise. Teams the caller cannot see are dropped from the filter | `IssueResponse` list with embedded shapes, `next_cursor`, `total` | viewer | yes |
| get-issue | GET /get | `{issue_id}` or `{identifier, workspace_id?}` (`workspace_id` only disambiguates a team key shared by several of the caller's workspaces) | `IssueDetailResponse`: issue with labels, state, assignee, creator, project, milestone, cycle, parent, children summary, relations, attachments, subscribers, reactions, counts | viewer | yes |
| create-issue | POST /create | `{team_id, title, description?, state_id?, priority?, estimate?, assignee_id?, label_ids?, project_id?, milestone_id?, cycle_id?, parent_id?, due_date?, subscriber_ids?, template_id?, sort_order?}` | issue | member | yes |
| update-issue | PUT /update | `{issue_id, title?, description?, state_id?, priority?, estimate? (-1 clears), assignee_id?, label_ids?, add_label_ids?, remove_label_ids?, project_id?, milestone_id?, cycle_id?, parent_id?, due_date?, sort_order?, sub_issue_sort_order?, snoozed_until?, team_id? (move)}`. Optional inputs are never nullable: the nil uuid `00000000-0000-0000-0000-000000000000` clears `assignee_id`, `project_id`, `milestone_id`, `cycle_id`, `parent_id`; an empty string clears `due_date` and `snoozed_until` | issue; records history and notifications | member | yes |
| delete-issue | DELETE /delete | `{issue_id}` | deleted permanently (children re-parented to nil) | admin | yes |
| trash-issue / restore-issue | POST /trash, /restore | `{issue_id}` | issue | member | yes |
| archive-issue / unarchive-issue | POST /archive, /unarchive | `{issue_id}` | issue | member | yes |
| bulk-update-issues | PUT /bulk-update | `{issue_ids, state_id?, priority?, assignee_id?, add_label_ids?, remove_label_ids?, project_id?, cycle_id?, parent_id?, due_date?, team_id?}` | issues | member | yes |
| list-issue-history | GET /history/list | `{issue_id, limit?, cursor?}` | history entries with actor | viewer | yes |
| list-comments | GET /comments/list | `{issue_id, limit?, cursor?}` (pages by thread) | comments threaded (`replies`), with author and reactions | viewer | yes |
| create-comment | POST /comments/create | `{issue_id, body, parent_id?}` | comment | member | yes |
| update-comment | PUT /comments/update | `{comment_id, body}` | comment | author, or admin | yes |
| delete-comment | DELETE /comments/delete | `{comment_id}` | deleted | author or admin | yes |
| resolve-comment / unresolve-comment | POST /comments/resolve, /comments/unresolve | `{comment_id}` | comment | member | yes |
| create-issue-relation | POST /relations/create | `{issue_id, related_issue_id, type}` | relation | member | yes |
| delete-issue-relation | DELETE /relations/delete | `{relation_id}` | deleted | member | yes |
| create-attachment | POST /attachments/create | `{issue_id, title, url?, upload_id?, subtitle?}` (one of url, upload_id) | attachment | member | yes |
| update-attachment / delete-attachment | PUT /attachments/update, DELETE /attachments/delete | `{attachment_id, title?, subtitle?}` | | member | yes |
| subscribe-issue / unsubscribe-issue | POST /subscribe, /unsubscribe | `{issue_id, user_id?}` (default self) | `IssueDetailResponse` | member | yes |
| add-reaction / remove-reaction | POST /reactions/add, DELETE /reactions/remove | `{emoji, issue_id? | comment_id? | project_update_id?}` | reactions for the target | member | yes |
| search-issues | GET /search | `{workspace_id, query, limit? (1..50, default 20)}` | issues containing every word (substring or pg_trgm word similarity), ranked by title match then description; an identifier like ENG-123 comes first; trashed issues and private teams the caller is not on are left out | viewer | yes |

### /documents
| list-documents, get-document (by id or slug), create-document, update-document, delete-document, archive-document, unarchive-document, trash-document, restore-document | standard | `{workspace_id, title, content?, icon?, project_id?, initiative_id?, issue_id?}` ; list takes `{workspace_id, project_id?, initiative_id?, issue_id?, query?, include_archived?, include_trashed?}`; update takes `{document_id, title?, content?, icon?, project_id?, initiative_id?, issue_id?}` where the nil uuid detaches a parent; an empty title becomes Untitled | document | viewer / member / admin delete | yes |

### /views
| list-views | GET /list | `{workspace_id, team_id?, kind?}` | views the caller owns plus shared views | viewer | yes |
| get-view, create-view, update-view, delete-view | standard | `{workspace_id, team_id?, name, description?, icon?, color?, kind, shared, filters, display}` | view | member (owner to edit, admin for shared views they do not own) | yes |

### /templates
| list-templates, get-template, create-template, update-template, delete-template | standard | `{workspace_id, team_id?, kind, name, description?, data}`; list takes `{workspace_id, team_id?, kind?}` and returns workspace templates plus the team's; update takes `{template_id, name?, description?, data?, sort_order?}`; `data.state_id` needs a team template | template | viewer / admin | yes |

### /favorites
| list-favorites, add-favorite, remove-favorite, reorder-favorites | GET /list, POST /add, DELETE /remove, PUT /reorder | list `{workspace_id}`; add and remove `{workspace_id, kind, entity_id}` (adding an existing favorite returns it); reorder `{workspace_id, favorite_ids}` | `FavoriteResponse {id, kind, entity_id, name, icon, color, url_params, sort_order, created_at}`; entities that are gone or hidden from the caller are left out | self | yes |

### /notifications
| list-notifications | GET /list | `{workspace_id, filter? (inbox|unread|read|archived|snoozed|all, default inbox), limit?, cursor?}`. `inbox` is read and unread, not archived, not snoozed; `unread` and `read` split it; `snoozed` is hidden until `snoozed_until`; `all` is everything | notifications with issue, actor, newest first, `next_cursor` | self | yes |
| get-notification-counts | GET /counts | `{workspace_id}` | `{unread}` | self | yes |
| mark-notification-read / mark-notification-unread / archive-notification / unarchive-notification / snooze-notification | POST /read, /unread, /archive, /unarchive, /snooze | `{notification_ids}` (`snooze` adds `until`) | success | self | yes |
| mark-all-notifications-read | POST /read-all | `{workspace_id}` | success | self | yes |

### /uploads (session)
| create-upload | POST /create | multipart `workspace_id`, `file` | `{id, url, file_name, content_type, size_bytes, width?, height?}`; a file over `MAX_UPLOAD_BYTES` answers 400 (413 when the request body itself is too large) | member | session |
| get-upload | GET /get | `{upload_id}` | the file, `inline` for safe types and `attachment` otherwise (architecture.md, Uploads) | viewer | session |

### /api-keys (session) port Unbind: list-api-keys (`{user_id?}`, another user's keys for system admins), create-api-key (`{name, role, full_access, workspace_ids, expires_at?, privileges?}` returns token once), delete-api-key (`{id}`).

### /connected-apps (session) port Unbind: list-connected-apps, revoke-connected-app (`{id}`), plus the consent page operations get-connected-app-client (`GET /client`), approve-connected-app (`POST /approve`, `{client_id, redirect_uri, state?, code_challenge, resource?, scope?, role, full_access, workspace_ids, privileges?}`) and deny-connected-app (`POST /deny`).

### /oauth/* and /.well-known/* port Unbind's oauthserver handler at the root: authorize, token, register, consent info, resource metadata.

### /system (session, system admin)
| get-system-settings | GET /settings | | `{instance_name, allow_public_signup, email: {configured, settings}}` | system | session |
| update-system-settings | PUT /settings | `{instance_name?, allow_public_signup?}` | settings | system | session |
| get-email-settings / update-email-settings / clear-email-settings / send-test-email | GET /email, PUT /email (`{host, port, username, password?, from_address, from_name, encryption}`), DELETE /email, POST /email/test | | system | session |

### /webhooks
| list-webhooks, create-webhook, update-webhook, delete-webhook, test-webhook | standard | `{workspace_id, team_id?, url, label, events, enabled}`; update takes `webhook_id` and optional fields, `team_id` set to the nil uuid sends the whole workspace again; delete and test take `{webhook_id}`. The url must be public http(s): localhost, private, link-local and other reserved ranges are refused at create and again at delivery time (oauthserver `SafeDialContext`) | webhook (`secret` only in the create response); test answers `{delivered, status_code, error}` | admin | yes except test |

### /imports
| import-linear | POST /linear | see linear-import.md | ImportReport | admin | yes |
| list-imports | GET /list | `{workspace_id}` | the 50 most recent imports, dry runs included, newest first: `ImportResponse {id, workspace_id, source, dry_run, status, started_by, report, created_at, finished_at}` | admin | yes |

### /search
| search | GET /search (the group root) | `{workspace_id, query, limit? (1..50, default 10, per kind)}` | `{issues: [IssueSummary], projects: [ProjectSummary], documents: [DocumentResponse]}`; issues of private teams only for their members and admins | viewer | yes |

## Response shapes (internal/models)

- `UserResponse {id, email, name, display_name, avatar_url, active, system_admin, timezone, created_at}`; `UserSummary {id, name, display_name, avatar_url}` embedded everywhere else.
- `WorkspaceResponse {id, name, slug, icon, color, role, member_count, team_count, created_at}`.
- `TeamResponse {id, workspace_id, name, key, description, icon, color, private, timezone, issue_counter, member_count, is_member, cycles_enabled, cycle_duration_weeks, cycle_cooldown_weeks, cycle_start_day, cycle_auto_move, cycle_auto_add_started, estimate_type, estimate_allow_zero, estimate_extended, triage_enabled, default_state_id, triage_state_id, auto_close_months, auto_archive_months, active_cycle_id, states: [WorkflowStateResponse], created_at}`.
- `WorkflowStateResponse {id, team_id, name, type, color, description, position}`.
- `LabelResponse {id, workspace_id, team_id, name, color, description, parent_id, is_group, group_single_select, sort_order, archived_at}`; `LabelTreeResponse` (list-labels) is `LabelResponse` plus `children: [LabelResponse]`.
- `IssueResponse {id, identifier, number, workspace_id, team_id, team_key, title, description, priority, priority_label, estimate, state: WorkflowStateResponse, assignee: UserSummary|null, creator: UserSummary|null, labels: [LabelSummary {id, name, color, parent_id}], project: {id, name, slug, icon, color}|null, milestone: {id, name}|null, cycle: {id, number, name}|null, parent: {id, identifier, title}|null, child_count, completed_child_count, comment_count, attachment_count, sort_order, sub_issue_sort_order, due_date, started_at, completed_at, canceled_at, triaged_at, snoozed_until, archived_at, trashed_at, git_branch_name, url, import_url, created_at, updated_at}`; `parent` is an `IssueSummary {id, identifier, title, state}`. `IssueDetailResponse` (get-issue, subscribe-issue, unsubscribe-issue) is `IssueResponse` plus `children: [IssueResponse]`, `relations: {blocks, blocked_by, related, duplicate_of, duplicates}` each `[RelationResponse {relation_id, type, issue: IssueSummary}]`, `attachments`, `subscribers: [UserSummary]`, `reactions: [ReactionGroup {emoji, count, users: [UserSummary], reacted}]`. Every other issue operation returns `IssueResponse`.
- `ProjectResponse {id, workspace_id, name, slug, summary, description, icon, color, status: ProjectStatusResponse, lead?: UserSummary, priority, start_date, target_date, start_date_resolution, target_date_resolution, sort_order, health, teams: [{id, key, name, icon, color}], members: [UserSummary], labels: [LabelSummary], milestones: [MilestoneResponse], links: [{id, project_id, url, label, sort_order, created_at}], initiative_ids, progress: {scope, completed, started, percent}, issue_counts: {total, by_state_type: {triage, backlog, unstarted, started, completed, canceled}}, last_update? (get-project only), started_at, completed_at, canceled_at, archived_at, trashed_at, created_at, updated_at}`. `scope` is the sum of estimates of issues that are not canceled or trashed, an unestimated issue counting 1; `completed` and `started` are the scope in those state types; `percent` is completed over scope, rounded.
- `ProjectStatusResponse {id, workspace_id, name, type, color, description, position}`; `MilestoneResponse {id, project_id, name, description, target_date, sort_order, progress, issue_count, created_at}`; `ProjectUpdateResponse {id, project_id, author?, body, health, edited_at, reactions, created_at, updated_at}`.
- `InitiativeResponse {id, workspace_id, name, slug, description, icon, color, owner?, status, target_date, sort_order, health, projects: [{id, name, slug, icon, color}], progress (over all its projects' issues), archived_at, trashed_at, created_at, updated_at}`; `InitiativeUpdateResponse {id, initiative_id, author?, body, health, edited_at, created_at, updated_at}`.
- `DocumentResponse {id, workspace_id, title, slug, icon, content, project?: {id, name, slug, icon, color}, initiative?: {id, name, slug, icon, color}, issue_id, creator?, updated_by?, archived_at, trashed_at, created_at, updated_at}`.
- `CycleResponse {id, team_id, number, name, description, starts_at, ends_at, completed_at, archived_at, is_active, is_past, is_future, progress: {scope, completed, started, percent, issue_count, completed_issue_count}, scope_history?: [{date, scope, completed, started}] (get-cycle only), created_at}`.
- `CommentResponse {id, issue_id, author: UserSummary|null, body, parent_id, edited_at, resolved_at, resolved_by, reactions, replies: [CommentReplyResponse], created_at, updated_at}`; `CommentReplyResponse` is the same without `replies`.
- `NotificationResponse {id, type, workspace_id, title, body, actor, issue: IssueSummary {id, identifier, title, state}|null, comment_id, project: ProjectSummary|null, project_update_id, read_at, archived_at, snoozed_until, created_at}`.
- `HistoryResponse {id, actor: UserSummary|null, source, changes: [FieldChange], created_at}`.

`git_branch_name` is `<display_name>/<identifier lowercased>-<slug of title, max 50 chars>` for the caller when assigned, else `<identifier lowercased>-<slug>`.

`url` is `<EXTERNAL_UI_URL>/<workspace_slug>/issue/<identifier>`.
