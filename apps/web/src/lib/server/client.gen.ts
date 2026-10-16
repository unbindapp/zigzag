import { z } from 'zod';

export const KeyPrivilegeSchema = z.string();

export const WorkspaceRoleSchema = z.enum(['admin', 'member', 'viewer']);

export const APIKeyCreateInputSchema = z
  .object({
    expires_at: z.string().datetime({ offset: true }).nullable().optional(), // When the key stops working. Omit for a key that never expires.
    full_access: z.boolean(), // Reach every workspace you can, capped at role. workspace_ids must be empty.
    name: z.string(), // Name to recognize the key by, unique per user
    privileges: z.array(KeyPrivilegeSchema).optional(), // What the key may see beyond its role. None exist yet.
    role: WorkspaceRoleSchema, // Strongest role the key acts with. It never exceeds your own role in a workspace.
    workspace_ids: z.array(z.string()), // Workspaces the key is limited to. Required unless full_access.
  })
  .strip();

export const APIKeyCreatedResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    expires_at: z.string().datetime({ offset: true }).nullable().optional(),
    full_access: z.boolean(),
    id: z.string(),
    last_used_at: z.string().datetime({ offset: true }).nullable().optional(),
    name: z.string(),
    privileges: z.array(KeyPrivilegeSchema),
    role: WorkspaceRoleSchema,
    token: z.string(), // The full API key. Shown once, store it now.
    token_prefix: z.string(), // First characters of the token, for recognizing the key. Never the full token.
    user_id: z.string(),
    workspace_ids: z.array(z.string()),
  })
  .strip();

export const APIKeyDeleteInputSchema = z
  .object({
    id: z.string(), // The API key to revoke
  })
  .strip();

export const APIKeyResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    expires_at: z.string().datetime({ offset: true }).nullable().optional(),
    full_access: z.boolean(),
    id: z.string(),
    last_used_at: z.string().datetime({ offset: true }).nullable().optional(),
    name: z.string(),
    privileges: z.array(KeyPrivilegeSchema),
    role: WorkspaceRoleSchema,
    token_prefix: z.string(), // First characters of the token, for recognizing the key. Never the full token.
    user_id: z.string(),
    workspace_ids: z.array(z.string()),
  })
  .strip();

export const AcceptInvitationInputSchema = z
  .object({
    display_name: z.string().optional(), // Mention handle for a new account; derived from the email when omitted
    name: z.string().optional(), // Full name for a new account
    password: z.string().optional(), // The password of the existing account, or a new one of at least 8 characters when the account is created
    token: z.string(), // Token from the invitation link
  })
  .strip();

export const UserResponseSchema = z
  .object({
    active: z.boolean(), // False for placeholder accounts that have not joined yet
    avatar_url: z.string().nullable(),
    created_at: z.string().datetime({ offset: true }),
    display_name: z.string(), // Lowercase handle used in mentions
    email: z.string(),
    github_login: z.string().nullable(), // GitHub username used by GitHub Issues sync
    google_linked: z.boolean(), // Whether the account has signed in with Google
    has_password: z.boolean(), // Whether the account can sign in with a password
    id: z.string(),
    name: z.string(),
    system_admin: z.boolean(),
    timezone: z.string(),
  })
  .strip();

export const WorkspaceResponseSchema = z
  .object({
    color: z.string(),
    created_at: z.string().datetime({ offset: true }),
    default_team_id: z.string().optional(), // The team created together with the workspace. Only returned when the workspace is created.
    icon: z.string(),
    icon_url: z.string().nullable(), // Logo image, shown instead of icon when set
    id: z.string(),
    member_count: z.number(),
    name: z.string(),
    role: WorkspaceRoleSchema, // The caller's effective role in this workspace
    slug: z.string(), // URL key of the workspace
    team_count: z.number(),
  })
  .strip();

export const SessionWorkspaceResponseSchema = z
  .object({
    user: UserResponseSchema,
    workspace: WorkspaceResponseSchema,
  })
  .strip();

export const AcceptInvitationResponseBodySchema = z
  .object({
    data: SessionWorkspaceResponseSchema,
  })
  .strip();

export const AddGithubRepoSyncInputSchema = z
  .object({
    full_name: z.string(), // Repository as owner/name, like unbindapp/zigzag
    sync_closed: z.boolean().optional(), // Include closed issues when importing existing issues
    team_id: z.string(), // Team new issues from GitHub are created in
    workspace_id: z.string(), // Workspace to sync the repository into
  })
  .strip();

export const AddMemberInputSchema = z
  .object({
    email: z.string(), // Email of the person to add
    role: WorkspaceRoleSchema, // admin, member or viewer
    team_ids: z.array(z.string()).nullable().optional(), // Teams of this workspace to join
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const InvitationResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    email: z.string(),
    emailed: z.boolean(), // Whether the link was also emailed. False when mail is not configured.
    expires_at: z.string().datetime({ offset: true }),
    id: z.string(),
    link: z.string().optional(), // The invitation link. Only returned when the invitation is created or resent.
    role: WorkspaceRoleSchema,
    team_ids: z.array(z.string()), // Teams joined on accept
    workspace_id: z.string(),
  })
  .strip();

export const MemberUserSchema = z
  .object({
    active: z.boolean(),
    avatar_url: z.string().nullable(),
    display_name: z.string(),
    email: z.string(),
    id: z.string(),
    name: z.string(),
  })
  .strip();

export const WorkspaceMemberResponseSchema = z
  .object({
    joined_at: z.string().datetime({ offset: true }),
    role: WorkspaceRoleSchema,
    team_ids: z.array(z.string()),
    user: MemberUserSchema,
  })
  .strip();

export const AddWorkspaceMemberResponseSchema = z
  .object({
    invitation: InvitationResponseSchema.optional(), // Set when the address had no account and an invitation was created
    member: WorkspaceMemberResponseSchema.optional(), // Set when the address already had an account and was added at once
  })
  .strip();

export const AddMemberResponseBodySchema = z
  .object({
    data: AddWorkspaceMemberResponseSchema,
  })
  .strip();

export const AddTeamMemberInputSchema = z
  .object({
    owner: z.boolean().optional(), // Mark as team owner; admins only
    team_id: z.string(), // The team
    user_id: z.string(), // A member of the team's workspace; your own id to join
  })
  .strip();

export const AttachmentRefInputSchema = z
  .object({
    attachment_id: z.string(), // The attachment
  })
  .strip();

export const UserSummarySchema = z
  .object({
    avatar_url: z.string().nullable(),
    display_name: z.string(),
    id: z.string(),
    name: z.string(),
  })
  .strip();

export const AttachmentResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    creator: UserSummarySchema.optional(),
    id: z.string(),
    issue_id: z.string(),
    metadata: z.record(z.any()),
    sort_order: z.number(),
    source_type: z.string(), // upload, link, github (a pull request), github_branch, github_commit or import
    subtitle: z.string(),
    title: z.string(),
    updated_at: z.string().datetime({ offset: true }),
    upload_id: z.string().nullable(),
    url: z.string(),
  })
  .strip();

export const AttachmentResponseBodySchema = z
  .object({
    data: AttachmentResponseSchema,
  })
  .strip();

export const AuthProviderStatusSchema = z
  .object({
    enabled: z.boolean(),
  })
  .strip();

export const AuthProvidersResponseSchema = z
  .object({
    google: AuthProviderStatusSchema,
  })
  .strip();

export const AuthProvidersResponseBodySchema = z
  .object({
    data: AuthProvidersResponseSchema,
  })
  .strip();

export const BackfillGithubRepoSyncInputSchema = z
  .object({
    repo_sync_id: z.string(), // The synced repository to import existing issues from
    sync_closed: z.boolean().optional(), // Include closed issues; defaults to the repository's setting and becomes it
  })
  .strip();

export const BootstrapInputSchema = z
  .object({
    display_name: z.string(), // Short handle used in mentions
    email: z.string(), // Email of the first user, who becomes system admin
    name: z.string(), // Full name
    password: z.string(), // Password, at least 8 characters
    workspace_name: z.string(), // Name of the first workspace
    workspace_slug: z.string(), // URL key of the first workspace, lowercase letters, digits and dashes
  })
  .strip();

export const BulkUpdateIssuesInputSchema = z
  .object({
    add_label_ids: z.array(z.string()).nullable().optional(), // Labels to add
    assignee_id: z.string().optional(), // Assignee; the nil uuid unassigns
    cycle_id: z.string().optional(), // Cycle of every issue's team; the nil uuid removes it
    due_date: z.string().optional(), // Due date YYYY-MM-DD; empty clears it
    issue_ids: z.array(z.string()), // Issues to change together in one transaction
    parent_id: z.string().optional(), // Parent issue; the nil uuid makes them top level
    priority: z.number().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    project_id: z.string().optional(), // Project; the nil uuid removes it
    remove_label_ids: z.array(z.string()).nullable().optional(), // Labels to remove
    state_id: z.string().optional(), // Workflow state; must belong to every issue's team
    team_id: z.string().optional(), // Move every issue to this team
  })
  .strip();

export const ChangePasswordInputSchema = z
  .object({
    current_password: z.string(), // Your current password
    new_password: z.string(), // The new password, at least 8 characters
  })
  .strip();

export const ReactionGroupSchema = z
  .object({
    count: z.number(),
    emoji: z.string(),
    reacted: z.boolean(), // Whether the caller added this emoji
    users: z.array(UserSummarySchema),
  })
  .strip();

export const CommentReplyResponseSchema = z
  .object({
    author: UserSummarySchema.optional(),
    body: z.string(), // Markdown
    created_at: z.string().datetime({ offset: true }),
    edited_at: z.string().datetime({ offset: true }).nullable(),
    id: z.string(),
    issue_id: z.string(),
    parent_id: z.string().nullable(), // The comment this one replies to
    reactions: z.array(ReactionGroupSchema),
    resolved_at: z.string().datetime({ offset: true }).nullable(),
    resolved_by: UserSummarySchema.optional(),
    updated_at: z.string().datetime({ offset: true }),
  })
  .strip();

export const CommentResponseSchema = z
  .object({
    author: UserSummarySchema.optional(),
    body: z.string(), // Markdown
    created_at: z.string().datetime({ offset: true }),
    edited_at: z.string().datetime({ offset: true }).nullable(),
    id: z.string(),
    issue_id: z.string(),
    parent_id: z.string().nullable(), // The comment this one replies to
    reactions: z.array(ReactionGroupSchema),
    replies: z.array(CommentReplyResponseSchema), // Replies, oldest first
    resolved_at: z.string().datetime({ offset: true }).nullable(),
    resolved_by: UserSummarySchema.optional(),
    updated_at: z.string().datetime({ offset: true }),
  })
  .strip();

export const CommentListResponseBodySchema = z
  .object({
    data: z.array(CommentResponseSchema),
    next_cursor: z.string().nullable(), // Pass as cursor to get the next page of threads; null on the last page
  })
  .strip();

export const CommentRefInputSchema = z
  .object({
    comment_id: z.string(), // The comment
  })
  .strip();

export const CommentResponseBodySchema = z
  .object({
    data: CommentResponseSchema,
  })
  .strip();

export const ConnectedAppApproveInputSchema = z
  .object({
    client_id: z.string(), // OAuth client_id from the authorize request
    code_challenge: z.string(), // PKCE S256 code_challenge from the authorize request
    full_access: z.boolean(), // Reach every workspace you can, capped at role. workspace_ids must be empty.
    privileges: z.array(KeyPrivilegeSchema).optional(), // What the app may see beyond its role. None exist yet.
    redirect_uri: z.string(), // redirect_uri from the authorize request
    resource: z.string().optional(), // resource from the authorize request
    role: WorkspaceRoleSchema, // Strongest role the app acts with. It never exceeds your own role in a workspace.
    scope: z.string().optional(), // scope from the authorize request
    state: z.string().optional(), // state from the authorize request, echoed back to the app
    workspace_ids: z.array(z.string()), // Workspaces the app is limited to. Required unless full_access.
  })
  .strip();

export const OAuthClientKindSchema = z.enum(['dynamic', 'metadata_document']);

export const ConnectedAppClientResponseSchema = z
  .object({
    client_host: z.string().optional(), // Host that published the client metadata document, when the client has one.
    kind: OAuthClientKindSchema,
    loopback_only: z.boolean(), // The client only redirects to this device.
    name: z.string(), // Self reported by the client, unverified.
    redirect_host: z.string(), // Where the browser is sent after approval.
    verified_brand: z.enum(['claude', 'chatgpt']).optional(), // Set when the client is proven to be this first party. Its name can then be trusted.
  })
  .strip();

export const ConnectedAppDenyInputSchema = z
  .object({
    client_id: z.string(), // OAuth client_id from the authorize request
    redirect_uri: z.string(), // redirect_uri from the authorize request
    state: z.string().optional(), // state from the authorize request, echoed back to the app
  })
  .strip();

export const ConnectedAppRedirectResponseSchema = z
  .object({
    redirect_url: z.string(), // Send the browser here to finish the flow.
  })
  .strip();

export const ConnectedAppResponseSchema = z
  .object({
    client_host: z.string().optional(),
    client_id: z.string(),
    client_name: z.string(), // Self reported by the client, unverified.
    created_at: z.string().datetime({ offset: true }),
    full_access: z.boolean(),
    id: z.string(),
    kind: OAuthClientKindSchema,
    last_used_at: z.string().datetime({ offset: true }).nullable().optional(),
    privileges: z.array(KeyPrivilegeSchema),
    redirect_host: z.string(),
    role: WorkspaceRoleSchema,
    verified_brand: z.enum(['claude', 'chatgpt']).optional(), // Set when the client is proven to be this first party.
    workspace_ids: z.array(z.string()),
  })
  .strip();

export const ConnectedAppRevokeInputSchema = z
  .object({
    id: z.string(), // The connected app to revoke
  })
  .strip();

export const CookieSchema = z
  .object({
    Domain: z.string(),
    Expires: z.string().datetime({ offset: true }),
    HttpOnly: z.boolean(),
    MaxAge: z.number(),
    Name: z.string(),
    Partitioned: z.boolean(),
    Path: z.string(),
    Quoted: z.boolean(),
    Raw: z.string(),
    RawExpires: z.string(),
    SameSite: z.number(),
    Secure: z.boolean(),
    Unparsed: z.array(z.string()).nullable(),
    Value: z.string(),
  })
  .strip();

export const CreateAPIKeyResponseBodySchema = z
  .object({
    data: APIKeyCreatedResponseSchema,
  })
  .strip();

export const CreateAttachmentInputSchema = z
  .object({
    issue_id: z.string(), // The issue to attach to
    subtitle: z.string().optional(), // Secondary line
    title: z.string(), // Link text
    upload_id: z.string().optional(), // A file uploaded to the same workspace with create-upload; give this or url
    url: z.string().optional(), // An http or https link; give this or upload_id
  })
  .strip();

export const CreateCommentInputSchema = z
  .object({
    body: z.string(), // Markdown; mention people with @[Name](user:<id>)
    issue_id: z.string(), // The issue to comment on
    parent_id: z.string().optional(), // Comment to reply to; a reply to a reply joins the same thread
  })
  .strip();

export const CreateCycleInputSchema = z
  .object({
    description: z.string().optional(), // What the cycle is about
    ends_at: z.string().datetime({ offset: true }), // End, after starts_at; cycles of a team cannot overlap
    name: z.string().optional(), // Optional name; the cycle shows as Cycle <number> without one
    starts_at: z.string().datetime({ offset: true }), // Start, RFC 3339 such as 2026-10-19T00:00:00Z
    team_id: z.string(), // The team; cycles must be turned on for it
  })
  .strip();

export const CreateDocumentInputSchema = z
  .object({
    content: z.string().optional(), // Markdown content
    icon: z.string().optional(), // Emoji or icon name
    initiative_id: z.string().optional(), // Initiative the document belongs to
    issue_id: z.string().optional(), // Issue the document belongs to
    project_id: z.string().optional(), // Project the document belongs to
    title: z.string(), // Document title; empty becomes Untitled
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const InitiativeStatusSchema = z.enum(['planned', 'active', 'completed']);

export const CreateInitiativeInputSchema = z
  .object({
    color: z.string().optional(), // Hex color such as #5e6ad2
    description: z.string().optional(), // Markdown description
    icon: z.string().optional(), // Emoji or icon name
    name: z.string(), // Initiative name
    owner_id: z.string().optional(), // Workspace member who owns the initiative
    project_ids: z.array(z.string()).nullable().optional(), // Projects of this workspace that belong to the initiative
    status: InitiativeStatusSchema.optional(), // planned (default), active or completed
    target_date: z.string().optional(), // Target date, YYYY-MM-DD
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const HealthSchema = z.enum(['on_track', 'at_risk', 'off_track']);

export const CreateInitiativeUpdateInputSchema = z
  .object({
    body: z.string(), // Markdown text of the update
    health: HealthSchema, // on_track, at_risk or off_track; becomes the initiative's health
    initiative_id: z.string(), // The initiative
  })
  .strip();

export const CreateInvitationInputSchema = z
  .object({
    email: z.string(), // Email to invite
    role: WorkspaceRoleSchema, // admin, member or viewer
    team_ids: z.array(z.string()).nullable().optional(), // Teams of this workspace joined on accept
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const CreateIssueInputSchema = z
  .object({
    assignee_id: z.string().optional(), // Workspace member to assign
    cycle_id: z.string().optional(), // Cycle of the team
    description: z.string().optional(), // Markdown; mention people with @[Name](user:<id>)
    due_date: z.string().optional(), // Due date, YYYY-MM-DD
    estimate: z.number().optional(), // Estimate in the team's scale
    label_ids: z.array(z.string()).nullable().optional(), // Workspace labels or labels of this team; not label groups
    milestone_id: z.string().optional(), // Milestone of the project; sets project_id when omitted
    parent_id: z.string().optional(), // Parent issue in the same workspace, making this a sub-issue
    priority: z.number().optional(), // 0 none (default), 1 urgent, 2 high, 3 medium, 4 low
    project_id: z.string().optional(), // Project of the workspace
    sort_order: z.number().optional(), // Manual order; defaults to the top of the state
    state_id: z.string().optional(), // Workflow state of the team; defaults to the team's default state, or triage for people outside a team with triage on
    subscriber_ids: z.array(z.string()).nullable().optional(), // Workspace members to subscribe besides you, the assignee and mentioned people
    team_id: z.string(), // Team the issue belongs to; its key prefixes the identifier
    template_id: z.string().optional(), // Issue template whose values fill every field you leave out; its sub-issues are created too
    title: z.string(), // One line summary
  })
  .strip();

export const CreateLabelInputSchema = z
  .object({
    color: z.string(), // Hex color such as #eb5757
    description: z.string().optional(), // What the label means
    group_single_select: z.boolean().optional(), // For groups: an issue may carry only one label of the group
    is_group: z.boolean().optional(), // Create a label group that holds other labels; groups cannot be applied to issues
    name: z.string(), // Label name, unique among its siblings
    parent_id: z.string().optional(), // Put the label in this group; the group must have the same team
    team_id: z.string().optional(), // Make it a label of this team only; omit for a workspace label every team shares
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const CreateMilestoneInputSchema = z
  .object({
    description: z.string().optional(), // What the milestone delivers
    name: z.string(), // Milestone name
    project_id: z.string(), // The project
    sort_order: z.number().optional(), // Order within the project; appended last when omitted
    target_date: z.string().optional(), // Target date, YYYY-MM-DD
  })
  .strip();

export const DateResolutionSchema = z.enum(['day', 'month', 'quarter', 'half_year', 'year']);

export const CreateProjectInputSchema = z
  .object({
    color: z.string().optional(), // Hex color such as #5e6ad2
    description: z.string().optional(), // Markdown description
    icon: z.string().optional(), // Emoji or icon name
    initiative_ids: z.array(z.string()).nullable().optional(), // Initiatives the project belongs to
    label_ids: z.array(z.string()).nullable().optional(), // Workspace labels (not team labels, not groups)
    lead_id: z.string().optional(), // Workspace member who leads the project
    member_ids: z.array(z.string()).nullable().optional(), // Workspace members working on the project
    name: z.string(), // Project name
    priority: z.number().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    start_date: z.string().optional(), // Planned start, YYYY-MM-DD
    start_date_resolution: DateResolutionSchema.optional(), // Precision of start_date: day (default), month, quarter, half_year or year
    status_id: z.string().optional(), // Project status from list-project-statuses; defaults to the first backlog status
    summary: z.string().optional(), // One line summary
    target_date: z.string().optional(), // Target completion, YYYY-MM-DD
    target_date_resolution: DateResolutionSchema.optional(), // Precision of target_date: day (default), month, quarter, half_year or year
    team_ids: z.array(z.string()), // Teams working on the project, at least one, all from this workspace
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const CreateProjectLinkInputSchema = z
  .object({
    label: z.string(), // Text shown for the link
    project_id: z.string(), // The project
    url: z.string(), // http or https URL
  })
  .strip();

export const ProjectStatusTypeSchema = z.enum([
  'backlog',
  'planned',
  'started',
  'paused',
  'completed',
  'canceled',
]);

export const CreateProjectStatusInputSchema = z
  .object({
    color: z.string(), // Hex color such as #f2c94c
    description: z.string().optional(), // What the status means
    name: z.string(), // Status name, unique in the workspace
    position: z.number().optional(), // Order within its type; appended last when omitted
    type: ProjectStatusTypeSchema, // backlog, planned, started, paused, completed or canceled
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const CreateProjectUpdateInputSchema = z
  .object({
    body: z.string(), // Markdown text of the update
    health: HealthSchema, // on_track, at_risk or off_track; becomes the project's health
    project_id: z.string(), // The project
  })
  .strip();

export const IssueRelationTypeSchema = z.enum(['blocks', 'related', 'duplicate']);

export const CreateRelationInputSchema = z
  .object({
    issue_id: z.string(), // The issue the relation starts from
    related_issue_id: z.string(), // The other issue, in the same workspace
    type: IssueRelationTypeSchema, // blocks: issue_id blocks related_issue_id; related: a plain link; duplicate: issue_id duplicates related_issue_id and moves to the Duplicate state
  })
  .strip();

export const StateTypeSchema = z.enum([
  'triage',
  'backlog',
  'unstarted',
  'started',
  'completed',
  'canceled',
]);

export const CreateStateInputSchema = z
  .object({
    color: z.string(), // Hex color such as #f2c94c
    description: z.string().optional(), // What the state means
    name: z.string(), // State name, unique in the team
    position: z.number().optional(), // Order within its type; appended last when omitted
    team_id: z.string(), // The team
    type: StateTypeSchema, // triage, backlog, unstarted, started, completed or canceled
  })
  .strip();

export const CreateTeamInputSchema = z
  .object({
    color: z.string().optional(), // Hex color such as #5e6ad2
    description: z.string().optional(), // What the team works on
    icon: z.string().optional(), // Emoji or icon name
    key: z.string(), // Issue prefix, 1 to 7 uppercase letters or digits starting with a letter, unique in the workspace
    name: z.string(), // Team name, unique in the workspace
    private: z.boolean().optional(), // Private teams are visible to their members and workspace admins only
    timezone: z.string().optional(), // IANA time zone for cycles, defaults to UTC
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const TemplateSubIssueSchema = z
  .object({
    description: z.string().optional(), // Sub-issue description in markdown
    title: z.string(), // Sub-issue title
  })
  .strip();

export const TemplateDataSchema = z
  .object({
    assignee_id: z.string().nullable().optional(), // Assignee
    description: z.string().optional(), // Prefilled description in markdown
    estimate: z.number().nullable().optional(), // Prefilled estimate
    label_ids: z.array(z.string()).nullable().optional(), // Labels to apply
    priority: z.number().nullable().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    project_id: z.string().nullable().optional(), // Project
    state_id: z.string().nullable().optional(), // Workflow state
    sub_issues: z.array(TemplateSubIssueSchema).nullable().optional(), // Sub-issues created with the issue
    title: z.string().optional(), // Prefilled title
  })
  .strip();

export const TemplateKindSchema = z.enum(['issue', 'project', 'document']);

export const CreateTemplateInputSchema = z
  .object({
    data: TemplateDataSchema, // Values a new issue starts with: title, description, priority, estimate, label_ids, assignee_id, state_id, project_id, sub_issues
    description: z.string().optional(), // When to use the template
    kind: TemplateKindSchema, // issue, project or document
    name: z.string(), // Template name shown in the picker
    team_id: z.string().optional(), // Make it a template of this team; needed when data.state_id is set
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const CreateUserInputSchema = z
  .object({
    display_name: z.string(), // Short handle used in mentions, lowercase and unique
    email: z.string(), // Email address, used to sign in
    name: z.string(), // Full name
    password: z.string(), // Initial password, at least 8 characters
    system_admin: z.boolean().optional(), // Grant instance administration
  })
  .strip();

export const CreateUserResponseBodySchema = z
  .object({
    data: SessionWorkspaceResponseSchema,
  })
  .strip();

export const ViewDisplaySchema = z
  .object({
    group_by: z.string().optional(), // Grouping, same values as list-issues group_by
    layout: z.enum(['list', 'board']), // list or board
    order_by: z.string().optional(), // Ordering, same values as list-issues order_by
    show_completed: z.enum(['all', 'past_week', 'none']), // Which completed and canceled issues to show: all, past_week or none
    show_sub_issues: z.boolean(), // Nest sub-issues under their parent instead of flat rows
  })
  .strip();

export const ViewKindSchema = z.enum(['issue', 'project']);

export const CreateViewInputSchema = z
  .object({
    color: z.string().optional(), // Hex color such as #5e6ad2
    description: z.string().optional(), // What the view shows
    display: ViewDisplaySchema, // How the issues show: layout, grouping, ordering
    filters: z.record(z.any()), // list-issues parameters to apply, for example {"state_types": ["started"], "assignee_ids": ["<user id>"]}
    icon: z.string().optional(), // Emoji or icon name
    kind: ViewKindSchema, // issue for issue views; project views are reserved
    name: z.string(), // View name
    shared: z.boolean().optional(), // Share with everyone in the workspace; private views are visible to you only
    team_id: z.string().optional(), // Make it a view of this team; omit for a workspace view
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const CreateWebhookInputSchema = z
  .object({
    enabled: z.boolean(), // Whether events are sent
    events: z.array(
      z.enum([
        'issue.created',
        'issue.updated',
        'issue.deleted',
        'comment.created',
        'project.updated',
        'cycle.updated',
      ]),
    ), // Events to send: issue.created, issue.updated, issue.deleted, comment.created, project.updated, cycle.updated
    label: z.string(), // Name shown in settings, like Slack bridge
    team_id: z.string().optional(), // Only send events of this team; omit for the whole workspace
    url: z.string(), // Public http or https endpoint that receives a JSON POST per event
    workspace_id: z.string(), // Workspace the webhook belongs to
  })
  .strip();

export const CreateWorkspaceInputSchema = z
  .object({
    color: z.string().optional(), // Hex color such as #5e6ad2
    icon: z.string().optional(), // Emoji or icon name
    name: z.string(), // Workspace name
    slug: z.string(), // URL key, lowercase letters, digits and dashes, unique across the instance
  })
  .strip();

export const CycleProgressSchema = z
  .object({
    completed: z.number(), // Scope in completed states
    completed_issue_count: z.number(),
    issue_count: z.number(), // Issues in the cycle, canceled ones left out
    percent: z.number(), // Completed scope as a percentage of the scope, 0 to 100
    scope: z.number(), // Sum of estimates when the team estimates, else the issue count; canceled issues are left out
    started: z.number(), // Scope in started states
  })
  .strip();

export const CycleScopePointSchema = z
  .object({
    completed: z.number(),
    date: z.string(), // Day in the team's time zone, YYYY-MM-DD
    scope: z.number(),
    started: z.number(),
  })
  .strip();

export const CycleResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    completed_at: z.string().datetime({ offset: true }).nullable(),
    created_at: z.string().datetime({ offset: true }),
    description: z.string(),
    ends_at: z.string().datetime({ offset: true }),
    id: z.string(),
    is_active: z.boolean(),
    is_future: z.boolean(),
    is_past: z.boolean(),
    name: z.string(), // Optional name; show Cycle <number> when empty
    number: z.number(), // Number of the cycle within its team
    progress: CycleProgressSchema,
    scope_history: z.array(CycleScopePointSchema).nullable().optional(), // Daily scope snapshots, oldest first; get-cycle only
    starts_at: z.string().datetime({ offset: true }),
    team_id: z.string(),
  })
  .strip();

export const CycleListResponseBodySchema = z
  .object({
    data: z.array(CycleResponseSchema),
  })
  .strip();

export const CycleRefInputSchema = z
  .object({
    cycle_id: z.string(), // The cycle
  })
  .strip();

export const CycleResponseBodySchema = z
  .object({
    data: CycleResponseSchema,
  })
  .strip();

export const DeletedResponseSchema = z
  .object({
    deleted: z.boolean(),
    id: z.string(),
  })
  .strip();

export const DeleteAPIKeyResponseBodySchema = z
  .object({
    data: DeletedResponseSchema,
  })
  .strip();

export const DeleteGithubResponseBodySchema = z
  .object({
    data: DeletedResponseSchema,
  })
  .strip();

export const DeleteProjectStatusInputSchema = z
  .object({
    move_to_status_id: z.string(), // Another status of the same workspace that receives the deleted status's projects
    status_id: z.string(), // The project status to delete
  })
  .strip();

export const DeleteStateInputSchema = z
  .object({
    move_issues_to_state_id: z.string(), // Another state of the same team that receives the deleted state's issues
    state_id: z.string(), // The state to delete
  })
  .strip();

export const DeleteTeamInputSchema = z
  .object({
    team_id: z.string(), // The team to delete together with all its issues, states and cycles
  })
  .strip();

export const DeleteTemplateInputSchema = z
  .object({
    template_id: z.string(), // The template to delete; issues created from it are not touched
  })
  .strip();

export const DeleteUserInputSchema = z
  .object({
    user_id: z.string(), // The user to delete. Their issues and comments stay, without an author.
  })
  .strip();

export const DeleteViewInputSchema = z
  .object({
    view_id: z.string(), // The view to delete; you must own it, or be an admin for a shared view
  })
  .strip();

export const DeleteWorkspaceInputSchema = z
  .object({
    confirm_slug: z.string(), // The workspace slug, repeated to confirm
    workspace_id: z.string(), // The workspace to delete with all its teams, issues, projects and documents
  })
  .strip();

export const DeletedResponseBodySchema = z
  .object({
    data: DeletedResponseSchema,
  })
  .strip();

export const InitiativeSummarySchema = z
  .object({
    color: z.string(),
    icon: z.string(),
    id: z.string(),
    name: z.string(),
    slug: z.string(),
  })
  .strip();

export const ProjectSummarySchema = z
  .object({
    color: z.string(),
    icon: z.string(),
    id: z.string(),
    name: z.string(),
    slug: z.string(),
  })
  .strip();

export const DocumentResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    content: z.string(), // Markdown
    created_at: z.string().datetime({ offset: true }),
    creator: UserSummarySchema.optional(),
    icon: z.string(),
    id: z.string(),
    initiative: InitiativeSummarySchema.optional(),
    issue_id: z.string().nullable(),
    project: ProjectSummarySchema.optional(),
    slug: z.string(), // URL key, unique in the workspace
    title: z.string(),
    trashed_at: z.string().datetime({ offset: true }).nullable(),
    updated_at: z.string().datetime({ offset: true }),
    updated_by: UserSummarySchema.optional(),
    workspace_id: z.string(),
  })
  .strip();

export const DocumentListResponseBodySchema = z
  .object({
    data: z.array(DocumentResponseSchema),
  })
  .strip();

export const DocumentRefInputSchema = z
  .object({
    document_id: z.string(), // The document
  })
  .strip();

export const DocumentResponseBodySchema = z
  .object({
    data: DocumentResponseSchema,
  })
  .strip();

export const EditProjectUpdateInputSchema = z
  .object({
    body: z.string().optional(), // New markdown text
    health: HealthSchema.optional(), // on_track, at_risk or off_track
    update_id: z.string(), // The project update to change
  })
  .strip();

export const EmailSettingsSchema = z
  .object({
    encryption: z.enum(['starttls', 'tls', 'none']), // starttls upgrades the connection (usually port 587), tls connects encrypted (usually port 465), none sends in the clear
    from_address: z.string(), // Address mail is sent from
    from_name: z.string().optional(), // Display name mail is sent from
    host: z.string(), // SMTP host, for example smtp.resend.com
    port: z.number(), // SMTP port, usually 465 for tls or 587 for starttls
    username: z.string().optional(), // SMTP username, resend for Resend
  })
  .strip();

export const EmailSettingsResponseSchema = z
  .object({
    configured: z.boolean(), // True when a relay is set up and mail can be sent
    password_set: z.boolean(), // Whether an SMTP password is stored; it is never returned
    settings: EmailSettingsSchema.optional(),
  })
  .strip();

export const EmailSettingsResponseBodySchema = z
  .object({
    data: EmailSettingsResponseSchema,
  })
  .strip();

export const EmailSettingsUpdateInputSchema = z
  .object({
    encryption: z.enum(['starttls', 'tls', 'none']), // starttls upgrades the connection (usually port 587), tls connects encrypted (usually port 465), none sends in the clear
    from_address: z.string(), // Address mail is sent from
    from_name: z.string().optional(), // Display name mail is sent from
    host: z.string(), // SMTP host, for example smtp.resend.com
    password: z.string().optional(), // SMTP password, for Resend the API key. Omit to keep the stored one; send an empty string to clear it.
    port: z.number(), // SMTP port, usually 465 for tls or 587 for starttls
    username: z.string().optional(), // SMTP username, resend for Resend
  })
  .strip();

export const EstimateTypeSchema = z.enum(['none', 'exponential', 'fibonacci', 'linear', 'tshirt']);

export const FavoriteKindSchema = z.enum([
  'issue',
  'project',
  'cycle',
  'view',
  'document',
  'label',
  'team',
  'initiative',
]);

export const FavoriteResponseSchema = z
  .object({
    color: z.string(),
    created_at: z.string().datetime({ offset: true }),
    entity_id: z.string(),
    icon: z.string(),
    id: z.string(),
    kind: FavoriteKindSchema,
    name: z.string(), // Current name or title of the favorited entity
    sort_order: z.number(),
    url_params: z.record(z.string()), // Route parameters of the entity's page: identifier for issues, slug for projects, documents and initiatives, key and number for cycles, view_id for views, key for teams, label_id for labels
  })
  .strip();

export const FavoriteListResponseBodySchema = z
  .object({
    data: z.array(FavoriteResponseSchema),
  })
  .strip();

export const FavoriteResponseBodySchema = z
  .object({
    data: FavoriteResponseSchema,
  })
  .strip();

export const FavoriteTargetInputSchema = z
  .object({
    entity_id: z.string(), // Id of the issue, project, cycle, view, document, label, team or initiative
    kind: FavoriteKindSchema, // issue, project, cycle, view, document, label, team or initiative
    workspace_id: z.string(), // The workspace the entity belongs to
  })
  .strip();

export const FieldChangeSchema = z
  .object({
    field: z.string(), // Changed field: title, description, state, priority, estimate, assignee, project, milestone, cycle, parent, due_date, labels_added, labels_removed, relation_added, relation_removed, attachment_added, attachment_removed, github, archived, unarchived, trashed, restored, team, created, imported
    from: z.any().optional(), // Previous value, usually an id or a scalar
    from_label: z.string().optional(), // Human readable previous value
    to: z.any().optional(), // New value, usually an id or a scalar
    to_label: z.string().optional(), // Human readable new value
  })
  .strip();

export const GetClientResponseBodySchema = z
  .object({
    data: ConnectedAppClientResponseSchema,
  })
  .strip();

export const GithubRepositoryResponseSchema = z
  .object({
    full_name: z.string(), // owner/name on GitHub
    id: z.string(),
    last_event: z.string(), // Last event GitHub delivered for it, like pull_request.opened or push
    last_event_at: z.string().datetime({ offset: true }),
  })
  .strip();

export const GithubIntegrationResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    created_by: UserSummarySchema.optional(),
    has_token: z.boolean(), // Whether a GitHub token is stored for Issues sync
    id: z.string(),
    last_delivery_at: z.string().datetime({ offset: true }).nullable(), // When GitHub last delivered an event with a valid signature
    repositories: z.array(GithubRepositoryResponseSchema), // Repositories that delivered events, most recent first
    secret_prefix: z.string(), // First characters of the webhook secret, to tell secrets apart
    token_prefix: z.string(), // First characters of the GitHub token, empty without one
    webhook_url: z.string(), // Payload URL for the GitHub webhook
    workspace_id: z.string(),
  })
  .strip();

export const GithubIntegrationStatusResponseSchema = z
  .object({
    connected: z.boolean(), // Whether the workspace has a GitHub webhook secret
    integration: GithubIntegrationResponseSchema.optional(), // Present when connected
    webhook_url: z.string(), // Payload URL for the GitHub webhook
  })
  .strip();

export const GetGithubResponseBodySchema = z
  .object({
    data: GithubIntegrationStatusResponseSchema,
  })
  .strip();

export const InvitationPreviewResponseSchema = z
  .object({
    account_exists: z.boolean(), // True when an active account with this email exists, so accepting needs no password
    email: z.string(),
    expires_at: z.string().datetime({ offset: true }),
    role: WorkspaceRoleSchema,
    workspace_name: z.string(),
  })
  .strip();

export const GetInvitationResponseBodySchema = z
  .object({
    data: InvitationPreviewResponseSchema,
  })
  .strip();

export const GithubBackfillReportSchema = z
  .object({
    comments_created: z.number(), // GitHub comments copied to Zigzag
    issues_created: z.number(), // GitHub issues that became new Zigzag issues
    issues_updated: z.number(), // Linked issues that changed to match GitHub
    skipped: z.number(), // Pull requests and linked issues that were already up to date
  })
  .strip();

export const GithubBackfillStatusSchema = z.enum(['none', 'running', 'completed', 'failed']);

export const GithubDeliveryResponseSchema = z
  .object({
    event: z.string(), // The X-GitHub-Event header
    handled: z.boolean(), // False when Zigzag ignores this event or action
  })
  .strip();

export const GithubIntegrationRefInputSchema = z
  .object({
    workspace_id: z.string(), // Workspace whose GitHub connection to change
  })
  .strip();

export const GithubIntegrationResponseBodySchema = z
  .object({
    data: GithubIntegrationResponseSchema,
  })
  .strip();

export const GithubIntegrationSecretResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    created_by: UserSummarySchema.optional(),
    has_token: z.boolean(), // Whether a GitHub token is stored for Issues sync
    id: z.string(),
    last_delivery_at: z.string().datetime({ offset: true }).nullable(), // When GitHub last delivered an event with a valid signature
    repositories: z.array(GithubRepositoryResponseSchema), // Repositories that delivered events, most recent first
    secret: z.string(), // Webhook secret to paste into GitHub. Returned only here, store it now
    secret_prefix: z.string(), // First characters of the webhook secret, to tell secrets apart
    token_prefix: z.string(), // First characters of the GitHub token, empty without one
    webhook_url: z.string(), // Payload URL for the GitHub webhook
    workspace_id: z.string(),
  })
  .strip();

export const GithubIssueLinkResponseSchema = z
  .object({
    html_url: z.string(), // The issue on GitHub
    number: z.number(), // Issue number on GitHub
    repository: z.string(), // owner/name on GitHub
  })
  .strip();

export const GithubRepoSyncRefInputSchema = z
  .object({
    repo_sync_id: z.string(), // The synced repository
  })
  .strip();

export const TeamSummarySchema = z
  .object({
    color: z.string(),
    icon: z.string(),
    id: z.string(),
    key: z.string(),
    name: z.string(),
  })
  .strip();

export const GithubRepoSyncResponseSchema = z
  .object({
    backfill_report: GithubBackfillReportSchema.optional(), // Counts of the running or last import
    backfill_status: GithubBackfillStatusSchema, // none, running, completed or failed
    created_at: z.string().datetime({ offset: true }),
    full_name: z.string(), // owner/name on GitHub
    html_url: z.string(), // The repository on GitHub
    id: z.string(),
    last_error: z.string(), // Last error talking to GitHub, empty after a success
    last_synced_at: z.string().datetime({ offset: true }).nullable(), // Last successful import or change sent to GitHub
    linked_issues: z.number(), // Zigzag issues linked to an issue of this repository
    project: ProjectSummarySchema.optional(), // Project new issues from GitHub join; missing when it was deleted
    sync_closed: z.boolean(), // Whether imports include closed issues
    team: TeamSummarySchema.optional(), // Team new issues from GitHub land in
    workspace_id: z.string(),
  })
  .strip();

export const GithubRepoSyncResponseBodySchema = z
  .object({
    data: GithubRepoSyncResponseSchema,
  })
  .strip();

export const GithubSecretResponseBodySchema = z
  .object({
    data: GithubIntegrationSecretResponseSchema,
  })
  .strip();

export const GithubWebhookResponseBodySchema = z
  .object({
    data: GithubDeliveryResponseSchema,
  })
  .strip();

export const GoogleAuthSettingsResponseSchema = z
  .object({
    client_id: z.string(), // OAuth client id from Google Cloud
    configured: z.boolean(), // True when a client id and secret are stored
    enabled: z.boolean(), // Whether Sign in with Google is offered
    hosted_domain: z.string(), // Only Google Workspace accounts of this domain may sign in; empty allows any Google account
    redirect_uri: z.string(), // Authorized redirect URI to register on the Google OAuth client
    secret_prefix: z.string(), // First characters of the client secret, empty without one. Never the full secret.
  })
  .strip();

export const GoogleAuthSettingsResponseBodySchema = z
  .object({
    data: GoogleAuthSettingsResponseSchema,
  })
  .strip();

export const HistorySourceSchema = z.enum(['user', 'api', 'mcp', 'import', 'system']);

export const HistoryResponseSchema = z
  .object({
    actor: UserSummarySchema.optional(), // Missing for system changes and imports
    changes: z.array(FieldChangeSchema),
    created_at: z.string().datetime({ offset: true }),
    id: z.string(),
    issue_id: z.string(),
    source: HistorySourceSchema,
  })
  .strip();

export const HistoryListResponseBodySchema = z
  .object({
    data: z.array(HistoryResponseSchema),
    next_cursor: z.string().nullable(), // Pass as cursor to get the next page; null on the last page
  })
  .strip();

export const ImportEntityCountSchema = z
  .object({
    created: z.number(),
    skipped: z.number(),
    updated: z.number(),
  })
  .strip();

export const LinearCycleSchema = z
  .object({
    completedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the cycle was completed
    description: z.string().nullable().optional(), // Cycle description
    endsAt: z.string().datetime({ offset: true }).nullable().optional(), // End time
    id: z.string(), // Linear cycle id
    name: z.string().nullable().optional(), // Cycle name
    number: z.number(), // Cycle number within the team
    startsAt: z.string().datetime({ offset: true }).nullable().optional(), // Start time
    teamId: z.string(), // Linear team id
  })
  .passthrough();

export const LinearAttachmentSchema = z
  .object({
    createdAt: z.string().datetime({ offset: true }).nullable().optional(), // Creation time
    id: z.string().optional(), // Linear attachment id
    subtitle: z.string().nullable().optional(), // Subtitle
    title: z.string().nullable().optional(), // Title
    url: z.string(), // Link target
  })
  .passthrough();

export const LinearCommentSchema = z
  .object({
    authorId: z.string().nullable().optional(), // Linear user id of the author
    authorName: z.string().nullable().optional(), // Author name
    body: z.string().optional(), // Markdown body
    createdAt: z.string().datetime({ offset: true }).nullable().optional(), // Creation time, preserved
    id: z.string(), // Linear comment id; a second import of the same id updates the comment
    parentId: z.string().nullable().optional(), // Linear id of the comment this one replies to
    quotedText: z.string().nullable().optional(), // Text of the issue the comment quotes, not imported
    updatedAt: z.string().datetime({ offset: true }).nullable().optional(), // Last edit time
  })
  .passthrough();

export const LinearRelationsSchema = z
  .object({
    blockedBy: z.array(z.string()).nullable().optional(), // Linear ids of issues blocking this issue
    blocks: z.array(z.string()).nullable().optional(), // Linear ids of issues this issue blocks
    duplicateOf: z.string().nullable().optional(), // Linear id of the issue this one duplicates
    relatedTo: z.array(z.string()).nullable().optional(), // Linear ids of related issues
  })
  .passthrough();

export const LinearStateSpanSchema = z
  .object({
    endedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the issue left the state; null for the current one
    startedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the issue entered the state
    stateId: z.string().nullable().optional(), // Linear state id
    stateName: z.string().nullable().optional(), // State name
    stateType: z.string().nullable().optional(), // State type
  })
  .passthrough();

export const LinearIssueSchema = z
  .object({
    archivedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the issue was archived; kept unless restore_archived
    assigneeId: z.string().nullable().optional(), // Linear user id of the assignee
    attachments: z.array(LinearAttachmentSchema).nullable().optional(), // Links attached to the issue
    canceledAt: z.string().datetime({ offset: true }).nullable().optional(), // When the issue was canceled
    comments: z.array(LinearCommentSchema).nullable().optional(), // Comments, oldest first
    completedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the issue was completed
    createdAt: z.string().datetime({ offset: true }).nullable().optional(), // Creation time, preserved
    creatorId: z.string().nullable().optional(), // Linear user id of the creator
    cycleId: z.string().nullable().optional(), // Linear cycle id
    description: z.string().nullable().optional(), // Markdown description
    dueDate: z.string().nullable().optional(), // Due date, YYYY-MM-DD
    estimate: z.number().nullable().optional(), // Estimate
    gitBranchName: z.string().nullable().optional(), // Linear branch name, not imported; Zigzag computes its own
    id: z.string(), // Linear issue id; a second import of the same id updates the issue
    identifier: z.string().optional(), // Identifier such as ENG-12
    labelIds: z.array(z.string()).nullable().optional(), // Linear label ids
    labelNames: z.array(z.string()).nullable().optional(), // Label names, used for ids the export does not list
    milestoneId: z.string().nullable().optional(), // Linear milestone id
    number: z.number().optional(), // Number within the team; kept unless another issue holds it
    parentId: z.string().nullable().optional(), // Linear id of the parent issue
    priority: z.number().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    projectId: z.string().nullable().optional(), // Linear project id
    relations: LinearRelationsSchema.optional(), // Relations to other issues by Linear id
    sortOrder: z.number().nullable().optional(), // Manual order
    startedAt: z.string().datetime({ offset: true }).nullable().optional(), // When work started
    stateHistory: z.array(LinearStateSpanSchema).nullable().optional(), // Each state the issue was in, oldest first; recorded as history
    stateId: z.string().nullable().optional(), // Linear state id
    stateName: z.string().nullable().optional(), // State name, used when stateId is unknown
    stateType: z.string().nullable().optional(), // State type, used with stateName
    teamId: z.string(), // Linear team id
    title: z.string(), // Issue title
    trashed: z.boolean().nullable().optional(), // Whether the issue is in the trash; only GraphQL exports carry it
    triagedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the issue left triage
    updatedAt: z.string().datetime({ offset: true }).nullable().optional(), // Last update time, preserved
    url: z.string().nullable().optional(), // Link to the Linear issue, kept as import_url
  })
  .passthrough();

export const LinearLabelSchema = z
  .object({
    archivedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the label was archived
    color: z.string().nullable().optional(), // Hex color
    description: z.string().nullable().optional(), // Label description
    groupType: z.string().nullable().optional(), // singleSelect when only one child may be applied to an issue
    id: z.string(), // Linear label id
    isGroup: z.boolean().optional(), // Whether this is a label group
    name: z.string(), // Label name
    parentId: z.string().nullable().optional(), // Label group this label belongs to
    teamId: z.string().nullable().optional(), // Linear team id; null for a workspace label
  })
  .passthrough();

export const LinearMilestoneSchema = z
  .object({
    description: z.string().nullable().optional(), // Milestone description
    id: z.string(), // Linear milestone id
    name: z.string(), // Milestone name
    projectId: z.string(), // Linear project id
    sortOrder: z.number().nullable().optional(), // Order within the project
    targetDate: z.string().nullable().optional(), // Target date, YYYY-MM-DD
  })
  .passthrough();

export const LinearProjectStatusSchema = z
  .object({
    id: z.string().optional(), // Linear project status id
    name: z.string().optional(), // Status name
    type: z.string().optional(), // backlog, planned, started, paused, completed or canceled
  })
  .passthrough();

export const LinearProjectSchema = z
  .object({
    archivedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the project was archived
    canceledAt: z.string().datetime({ offset: true }).nullable().optional(), // When the project was canceled
    completedAt: z.string().datetime({ offset: true }).nullable().optional(), // When the project completed
    createdAt: z.string().datetime({ offset: true }).nullable().optional(), // Creation time
    description: z.string().nullable().optional(), // Markdown description
    id: z.string(), // Linear project id
    identifier: z.string().nullable().optional(), // Linear project identifier
    labelIds: z.array(z.string()).nullable().optional(), // Linear label ids; only workspace labels apply to projects
    leadId: z.string().nullable().optional(), // Linear user id of the lead
    memberIds: z.array(z.string()).nullable().optional(), // Linear user ids of the members
    name: z.string(), // Project name
    priority: z.number().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    sortOrder: z.number().nullable().optional(), // Manual order
    startDate: z.string().nullable().optional(), // Start date, YYYY-MM-DD
    startedAt: z.string().datetime({ offset: true }).nullable().optional(), // When work started
    status: LinearProjectStatusSchema.optional(), // Project status; matched to a Zigzag status by type, then name
    summary: z.string().nullable().optional(), // One line summary
    targetDate: z.string().nullable().optional(), // Target date, YYYY-MM-DD
    teamIds: z.array(z.string()).nullable().optional(), // Linear team ids
    trashed: z.boolean().nullable().optional(), // Whether the project is in the trash
    updatedAt: z.string().datetime({ offset: true }).nullable().optional(), // Last update time
    url: z.string().nullable().optional(), // Link to the Linear project
  })
  .passthrough();

export const LinearTeamSchema = z
  .object({
    color: z.string().nullable().optional(), // Hex color
    createdAt: z.string().datetime({ offset: true }).nullable().optional(), // Creation time
    cyclesEnabled: z.boolean().nullable().optional(), // Whether the team uses cycles
    description: z.string().nullable().optional(), // Team description
    icon: z.string().nullable().optional(), // Icon name
    id: z.string(), // Linear team id
    key: z.string(), // Team key such as ENG; prefixes identifiers
    name: z.string(), // Team name
    timezone: z.string().nullable().optional(), // IANA time zone
  })
  .passthrough();

export const LinearUserSchema = z
  .object({
    avatarUrl: z.string().nullable().optional(), // Avatar url, not imported
    displayName: z.string().nullable().optional(), // Short handle
    email: z.string(), // Email address; matched case insensitively to Zigzag accounts
    id: z.string(), // Linear user id
    isActive: z.boolean().nullable().optional(), // Whether the Linear account is active
    isAdmin: z.boolean().optional(), // Admins join the workspace as admin, everyone else as member
    isGuest: z.boolean().optional(), // Guest account in Linear
    name: z.string().nullable().optional(), // Full name
  })
  .passthrough();

export const LinearWorkflowStateSchema = z
  .object({
    color: z.string().nullable().optional(), // Hex color
    id: z.string(), // Linear state id
    name: z.string(), // State name
    position: z.number().optional(), // Order within the type
    teamId: z.string(), // Linear team id
    type: z.string(), // triage, backlog, unstarted, started, completed, canceled or duplicate (imported as canceled)
  })
  .passthrough();

export const LinearWorkspaceSchema = z
  .object({
    id: z.string().optional(), // Linear workspace id
    name: z.string().optional(), // Linear workspace name
    url: z.string().optional(), // Linear workspace url
  })
  .passthrough();

export const LinearExportSchema = z
  .object({
    cycles: z.array(LinearCycleSchema).nullable().optional(), // Team cycles
    exported_at: z.string().optional(), // When the export was taken, RFC 3339
    issues: z.array(LinearIssueSchema).nullable().optional(), // Issues with their comments, attachments, relations and state history
    labels: z.array(LinearLabelSchema).nullable().optional(), // Issue labels and label groups; teamId null means a workspace label
    milestones: z.array(LinearMilestoneSchema).nullable().optional(), // Project milestones
    projects: z.array(LinearProjectSchema).nullable().optional(), // Projects
    teams: z.array(LinearTeamSchema).nullable().optional(), // Linear teams; each maps to the Zigzag team with the same key
    users: z.array(LinearUserSchema).nullable().optional(), // Every Linear user the export references
    workflow_states: z.array(LinearWorkflowStateSchema).nullable().optional(), // Workflow states of the exported teams
    workspace: LinearWorkspaceSchema.optional(), // The Linear workspace the export came from; informational, the target is workspace_id
  })
  .passthrough();

export const ImportLinearInputSchema = z
  .object({
    create_missing_users: z.boolean().optional(), // Default true: users without a Zigzag account get an inactive placeholder account; false leaves their issues and comments without a person
    dry_run: z.boolean().optional(), // Default true: runs the whole import in a transaction that is rolled back, so the report is exact and nothing is kept. Send false to import for real
    export: LinearExportSchema, // The Linear export: users, teams, workflow_states, labels, projects, milestones, cycles and issues with their comments, attachments, relations and stateHistory
    restore_archived: z.boolean().optional(), // Import archived issues and projects as active instead of archived
    restore_trashed: z.boolean().optional(), // Import trashed issues and projects as active instead of trashed
    workspace_id: z.string(), // Workspace to import into; needs admin
  })
  .strip();

export const ImportUserMappingSchema = z
  .object({
    action: z.enum(['matched', 'created', 'skipped']),
    email: z.string(),
    linear_id: z.string(),
    user_id: z.string().nullable(),
  })
  .strip();

export const ImportWarningSchema = z
  .object({
    entity: z.string(),
    external_id: z.string(),
    message: z.string(),
  })
  .strip();

export const ImportReportSchema = z
  .object({
    counts: z.record(ImportEntityCountSchema),
    dry_run: z.boolean(),
    import_id: z.string(),
    user_mapping: z.array(ImportUserMappingSchema),
    warnings: z.array(ImportWarningSchema),
  })
  .strip();

export const ImportSourceSchema = z.enum(['linear']);

export const ImportStatusSchema = z.enum(['running', 'completed', 'failed']);

export const ImportResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    dry_run: z.boolean(),
    finished_at: z.string().datetime({ offset: true }).nullable(),
    id: z.string(),
    report: ImportReportSchema, // Counts, warnings and user mapping; empty while the import runs
    source: ImportSourceSchema,
    started_by: UserSummarySchema.optional(), // Who ran the import; absent when the account is gone
    status: ImportStatusSchema,
    workspace_id: z.string(),
  })
  .strip();

export const ImportListResponseBodySchema = z
  .object({
    data: z.array(ImportResponseSchema),
  })
  .strip();

export const ImportReportResponseBodySchema = z
  .object({
    data: ImportReportSchema,
  })
  .strip();

export const ProgressResponseSchema = z
  .object({
    completed: z.number(), // Scope of completed issues
    percent: z.number(), // Completed scope as a percentage of the scope, 0 to 100
    scope: z.number(), // Sum of estimates of issues that are not canceled; an issue without an estimate counts 1
    started: z.number(), // Scope of started issues
  })
  .strip();

export const InitiativeResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    color: z.string(),
    created_at: z.string().datetime({ offset: true }),
    description: z.string(), // Markdown
    health: HealthSchema.nullable(), // Health of the latest initiative update
    icon: z.string(),
    id: z.string(),
    name: z.string(),
    owner: UserSummarySchema.optional(),
    progress: ProgressResponseSchema, // Progress over the issues of all its projects
    projects: z.array(ProjectSummarySchema), // Projects in the initiative, trashed ones left out
    slug: z.string(), // URL key, unique in the workspace
    sort_order: z.number(),
    status: InitiativeStatusSchema,
    target_date: z.string().nullable(),
    trashed_at: z.string().datetime({ offset: true }).nullable(),
    updated_at: z.string().datetime({ offset: true }),
    workspace_id: z.string(),
  })
  .strip();

export const InitiativeListResponseBodySchema = z
  .object({
    data: z.array(InitiativeResponseSchema),
  })
  .strip();

export const InitiativeProjectInputSchema = z
  .object({
    initiative_id: z.string(), // The initiative
    project_id: z.string(), // A project of the same workspace
  })
  .strip();

export const InitiativeRefInputSchema = z
  .object({
    initiative_id: z.string(), // The initiative
  })
  .strip();

export const InitiativeResponseBodySchema = z
  .object({
    data: InitiativeResponseSchema,
  })
  .strip();

export const InitiativeUpdateResponseSchema = z
  .object({
    author: UserSummarySchema.optional(),
    body: z.string(), // Markdown
    created_at: z.string().datetime({ offset: true }),
    edited_at: z.string().datetime({ offset: true }).nullable(),
    health: HealthSchema,
    id: z.string(),
    initiative_id: z.string(),
    updated_at: z.string().datetime({ offset: true }),
  })
  .strip();

export const InitiativeUpdateListResponseBodySchema = z
  .object({
    data: z.array(InitiativeUpdateResponseSchema),
  })
  .strip();

export const InitiativeUpdateRefInputSchema = z
  .object({
    update_id: z.string(), // The initiative update to delete
  })
  .strip();

export const InitiativeUpdateResponseBodySchema = z
  .object({
    data: InitiativeUpdateResponseSchema,
  })
  .strip();

export const InvitationListResponseBodySchema = z
  .object({
    data: z.array(InvitationResponseSchema),
  })
  .strip();

export const InvitationRefInputSchema = z
  .object({
    invitation_id: z.string(), // The invitation
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const InvitationResponseBodySchema = z
  .object({
    data: InvitationResponseSchema,
  })
  .strip();

export const IssueCountsByStateTypeSchema = z
  .object({
    backlog: z.number(),
    canceled: z.number(),
    completed: z.number(),
    started: z.number(),
    triage: z.number(),
    unstarted: z.number(),
  })
  .strip();

export const IssueCountsResponseSchema = z
  .object({
    by_state_type: IssueCountsByStateTypeSchema,
    total: z.number(),
  })
  .strip();

export const IssueCycleSummarySchema = z
  .object({
    id: z.string(),
    name: z.string(),
    number: z.number(),
  })
  .strip();

export const LabelSummarySchema = z
  .object({
    color: z.string(),
    id: z.string(),
    name: z.string(),
    parent_id: z.string().nullable(), // The label group this label belongs to
  })
  .strip();

export const IssueMilestoneSummarySchema = z
  .object({
    id: z.string(),
    name: z.string(),
  })
  .strip();

export const WorkflowStateResponseSchema = z
  .object({
    color: z.string(),
    description: z.string(),
    id: z.string(),
    name: z.string(),
    position: z.number(),
    team_id: z.string(),
    type: StateTypeSchema,
  })
  .strip();

export const IssueSummarySchema = z
  .object({
    id: z.string(),
    identifier: z.string(), // Team key and number, like ENG-123
    state: WorkflowStateResponseSchema.optional(),
    title: z.string(),
  })
  .strip();

export const IssueResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    assignee: UserSummarySchema.optional(),
    attachment_count: z.number(),
    canceled_at: z.string().datetime({ offset: true }).nullable(),
    child_count: z.number(),
    comment_count: z.number(),
    completed_at: z.string().datetime({ offset: true }).nullable(),
    completed_child_count: z.number(),
    created_at: z.string().datetime({ offset: true }),
    creator: UserSummarySchema.optional(),
    cycle: IssueCycleSummarySchema.optional(),
    description: z.string(), // Markdown
    due_date: z.string().nullable(),
    estimate: z.number().nullable(),
    git_branch_name: z.string(),
    id: z.string(),
    identifier: z.string(), // Team key and number, like ENG-123
    import_url: z.string(),
    labels: z.array(LabelSummarySchema),
    milestone: IssueMilestoneSummarySchema.optional(),
    number: z.number(),
    parent: IssueSummarySchema.optional(),
    priority: z.number(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    priority_label: z.string(),
    project: ProjectSummarySchema.optional(),
    snoozed_until: z.string().datetime({ offset: true }).nullable(),
    sort_order: z.number(),
    started_at: z.string().datetime({ offset: true }).nullable(),
    state: WorkflowStateResponseSchema,
    sub_issue_sort_order: z.number(),
    team_id: z.string(),
    team_key: z.string(),
    title: z.string(),
    trashed_at: z.string().datetime({ offset: true }).nullable(),
    triaged_at: z.string().datetime({ offset: true }).nullable(),
    updated_at: z.string().datetime({ offset: true }),
    url: z.string(),
    workspace_id: z.string(),
  })
  .strip();

export const RelationResponseSchema = z
  .object({
    issue: IssueSummarySchema, // The other issue of the relation
    relation_id: z.string(),
    type: IssueRelationTypeSchema,
  })
  .strip();

export const IssueRelationsSchema = z
  .object({
    blocked_by: z.array(RelationResponseSchema), // Issues blocking this issue
    blocks: z.array(RelationResponseSchema), // Issues this issue blocks
    duplicate_of: z.array(RelationResponseSchema), // The canonical issue this one duplicates
    duplicates: z.array(RelationResponseSchema), // Issues marked as duplicates of this one
    related: z.array(RelationResponseSchema),
  })
  .strip();

export const IssueDetailResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    assignee: UserSummarySchema.optional(),
    attachment_count: z.number(),
    attachments: z.array(AttachmentResponseSchema),
    canceled_at: z.string().datetime({ offset: true }).nullable(),
    child_count: z.number(),
    children: z.array(IssueResponseSchema), // Sub-issues
    comment_count: z.number(),
    completed_at: z.string().datetime({ offset: true }).nullable(),
    completed_child_count: z.number(),
    created_at: z.string().datetime({ offset: true }),
    creator: UserSummarySchema.optional(),
    cycle: IssueCycleSummarySchema.optional(),
    description: z.string(), // Markdown
    due_date: z.string().nullable(),
    estimate: z.number().nullable(),
    git_branch_name: z.string(),
    github_issue: GithubIssueLinkResponseSchema.optional(), // The GitHub issue this issue syncs with, when its repository is synced
    id: z.string(),
    identifier: z.string(), // Team key and number, like ENG-123
    import_url: z.string(),
    labels: z.array(LabelSummarySchema),
    milestone: IssueMilestoneSummarySchema.optional(),
    number: z.number(),
    parent: IssueSummarySchema.optional(),
    priority: z.number(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    priority_label: z.string(),
    project: ProjectSummarySchema.optional(),
    reactions: z.array(ReactionGroupSchema),
    relations: IssueRelationsSchema,
    snoozed_until: z.string().datetime({ offset: true }).nullable(),
    sort_order: z.number(),
    started_at: z.string().datetime({ offset: true }).nullable(),
    state: WorkflowStateResponseSchema,
    sub_issue_sort_order: z.number(),
    subscribers: z.array(UserSummarySchema),
    team_id: z.string(),
    team_key: z.string(),
    title: z.string(),
    trashed_at: z.string().datetime({ offset: true }).nullable(),
    triaged_at: z.string().datetime({ offset: true }).nullable(),
    updated_at: z.string().datetime({ offset: true }),
    url: z.string(),
    workspace_id: z.string(),
  })
  .strip();

export const IssueDetailResponseBodySchema = z
  .object({
    data: IssueDetailResponseSchema,
  })
  .strip();

export const IssueListResponseBodySchema = z
  .object({
    data: z.array(IssueResponseSchema),
    next_cursor: z.string().nullable(), // Pass as cursor to get the next page; null on the last page
    total: z.number(), // Issues matching the filters across all pages
  })
  .strip();

export const IssueRefInputSchema = z
  .object({
    issue_id: z.string(), // The issue
  })
  .strip();

export const IssueResponseBodySchema = z
  .object({
    data: IssueResponseSchema,
  })
  .strip();

export const IssuesResponseBodySchema = z
  .object({
    data: z.array(IssueResponseSchema),
  })
  .strip();

export const LabelResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    color: z.string(),
    description: z.string(),
    group_single_select: z.boolean(), // Only one label of this group may be on an issue at a time
    id: z.string(),
    is_group: z.boolean(), // Groups hold labels and are never applied to issues themselves
    name: z.string(),
    parent_id: z.string().nullable(), // The label group this label belongs to
    sort_order: z.number(),
    team_id: z.string().nullable(), // Null for workspace labels shared by every team
    workspace_id: z.string(),
  })
  .strip();

export const LabelTreeResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    children: z.array(LabelResponseSchema), // Labels of this group, empty for plain labels
    color: z.string(),
    description: z.string(),
    group_single_select: z.boolean(), // Only one label of this group may be on an issue at a time
    id: z.string(),
    is_group: z.boolean(), // Groups hold labels and are never applied to issues themselves
    name: z.string(),
    parent_id: z.string().nullable(), // The label group this label belongs to
    sort_order: z.number(),
    team_id: z.string().nullable(), // Null for workspace labels shared by every team
    workspace_id: z.string(),
  })
  .strip();

export const LabelListResponseBodySchema = z
  .object({
    data: z.array(LabelTreeResponseSchema),
  })
  .strip();

export const LabelRefInputSchema = z
  .object({
    label_id: z.string(), // The label
  })
  .strip();

export const LabelResponseBodySchema = z
  .object({
    data: LabelResponseSchema,
  })
  .strip();

export const ProjectLinkResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    id: z.string(),
    label: z.string(),
    project_id: z.string(),
    sort_order: z.number(),
    url: z.string(),
  })
  .strip();

export const LinkResponseBodySchema = z
  .object({
    data: ProjectLinkResponseSchema,
  })
  .strip();

export const ListAPIKeysResponseBodySchema = z
  .object({
    data: z.array(APIKeyResponseSchema),
  })
  .strip();

export const ListGithubRepoSyncsResponseBodySchema = z
  .object({
    data: z.array(GithubRepoSyncResponseSchema),
  })
  .strip();

export const ListResponseBodySchema = z
  .object({
    data: z.array(ConnectedAppResponseSchema),
  })
  .strip();

export const ListUsersResponseBodySchema = z
  .object({
    data: z.array(UserResponseSchema),
  })
  .strip();

export const LoginInputBodySchema = z
  .object({
    email: z.string(), // Account email
    password: z.string(), // Account password
  })
  .strip();

export const MarkAllReadInputSchema = z
  .object({
    workspace_id: z.string(), // Workspace whose inbox to mark read
  })
  .strip();

export const MilestoneResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    description: z.string(),
    id: z.string(),
    issue_count: z.number(),
    name: z.string(),
    progress: ProgressResponseSchema,
    project_id: z.string(),
    sort_order: z.number(),
    target_date: z.string().nullable(),
  })
  .strip();

export const MilestoneListResponseBodySchema = z
  .object({
    data: z.array(MilestoneResponseSchema),
  })
  .strip();

export const MilestoneRefInputSchema = z
  .object({
    milestone_id: z.string(), // The milestone to delete; its issues stay in the project without a milestone
  })
  .strip();

export const MilestoneResponseBodySchema = z
  .object({
    data: MilestoneResponseSchema,
  })
  .strip();

export const NotificationCountsResponseSchema = z
  .object({
    unread: z.number(), // Unread notifications in the inbox, not counting archived or snoozed ones
  })
  .strip();

export const NotificationCountsResultBodySchema = z
  .object({
    data: NotificationCountsResponseSchema,
  })
  .strip();

export const NotificationIDsInputSchema = z
  .object({
    notification_ids: z.array(z.string()), // Ids of your notifications, from list-notifications
  })
  .strip();

export const NotificationTypeSchema = z.enum([
  'issue_assigned',
  'issue_unassigned',
  'issue_mentioned',
  'comment_mentioned',
  'issue_commented',
  'issue_state_changed',
  'issue_priority_changed',
  'issue_due_soon',
  'issue_blocking',
  'project_update',
  'project_added_as_member',
  'project_lead',
  'reaction',
]);

export const NotificationResponseSchema = z
  .object({
    actor: UserSummarySchema.optional(), // Who caused the notification; absent for system notifications
    archived_at: z.string().datetime({ offset: true }).nullable(),
    body: z.string(), // Short plain text snippet, may be empty
    comment_id: z.string().nullable(),
    created_at: z.string().datetime({ offset: true }),
    id: z.string(),
    issue: IssueSummarySchema.optional(),
    project: ProjectSummarySchema.optional(),
    project_update_id: z.string().nullable(),
    read_at: z.string().datetime({ offset: true }).nullable(),
    snoozed_until: z.string().datetime({ offset: true }).nullable(),
    title: z.string(),
    type: NotificationTypeSchema,
    workspace_id: z.string(),
  })
  .strip();

export const NotificationListResponseBodySchema = z
  .object({
    data: z.array(NotificationResponseSchema),
    next_cursor: z.string().nullable(), // Pass as cursor for the next page; null on the last page
  })
  .strip();

export const NotificationSettingsResponseSchema = z
  .object({
    assigned: z.boolean(), // Issues assigned to or unassigned from you
    commented: z.boolean(), // Comments on issues you subscribe to
    due_soon: z.boolean(), // Issues due soon
    email_enabled: z.boolean(), // Send email for notifications at all
    mentioned: z.boolean(), // Mentions in issues and comments
    project_updates: z.boolean(), // Status updates on projects you follow
    state_changed: z.boolean(), // State changes on issues you subscribe to
  })
  .strip();

export const NotificationSettingsResponseBodySchema = z
  .object({
    data: NotificationSettingsResponseSchema,
  })
  .strip();

export const SuccessResponseSchema = z
  .object({
    success: z.boolean(),
  })
  .strip();

export const PlainSuccessResponseBodySchema = z
  .object({
    data: SuccessResponseSchema,
  })
  .strip();

export const ProjectLinkRefInputSchema = z
  .object({
    link_id: z.string(), // The link to delete
  })
  .strip();

export const ProjectUpdateResponseSchema = z
  .object({
    author: UserSummarySchema.optional(),
    body: z.string(), // Markdown
    created_at: z.string().datetime({ offset: true }),
    edited_at: z.string().datetime({ offset: true }).nullable(),
    health: HealthSchema,
    id: z.string(),
    project_id: z.string(),
    reactions: z.array(ReactionGroupSchema),
    updated_at: z.string().datetime({ offset: true }),
  })
  .strip();

export const ProjectStatusResponseSchema = z
  .object({
    color: z.string(),
    description: z.string(),
    id: z.string(),
    name: z.string(),
    position: z.number(), // Order within its type
    type: ProjectStatusTypeSchema,
    workspace_id: z.string(),
  })
  .strip();

export const ProjectResponseSchema = z
  .object({
    archived_at: z.string().datetime({ offset: true }).nullable(),
    canceled_at: z.string().datetime({ offset: true }).nullable(),
    color: z.string(),
    completed_at: z.string().datetime({ offset: true }).nullable(),
    created_at: z.string().datetime({ offset: true }),
    description: z.string(), // Markdown
    health: HealthSchema.nullable(), // Health of the latest project update
    icon: z.string(),
    id: z.string(),
    initiative_ids: z.array(z.string()),
    issue_counts: IssueCountsResponseSchema,
    labels: z.array(LabelSummarySchema),
    last_update: ProjectUpdateResponseSchema.optional(), // Latest status update, on get-project only
    lead: UserSummarySchema.optional(),
    links: z.array(ProjectLinkResponseSchema),
    members: z.array(UserSummarySchema),
    milestones: z.array(MilestoneResponseSchema),
    name: z.string(),
    priority: z.number(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    progress: ProgressResponseSchema,
    slug: z.string(), // URL key, unique in the workspace
    sort_order: z.number(),
    start_date: z.string().nullable(),
    start_date_resolution: DateResolutionSchema, // How precise start_date is: day, month, quarter, half_year or year
    started_at: z.string().datetime({ offset: true }).nullable(),
    status: ProjectStatusResponseSchema,
    summary: z.string(), // One line summary
    target_date: z.string().nullable(),
    target_date_resolution: DateResolutionSchema, // How precise target_date is: day, month, quarter, half_year or year
    teams: z.array(TeamSummarySchema),
    trashed_at: z.string().datetime({ offset: true }).nullable(),
    updated_at: z.string().datetime({ offset: true }),
    workspace_id: z.string(),
  })
  .strip();

export const ProjectListResponseBodySchema = z
  .object({
    data: z.array(ProjectResponseSchema),
  })
  .strip();

export const ProjectRefInputSchema = z
  .object({
    project_id: z.string(), // The project
  })
  .strip();

export const ProjectResponseBodySchema = z
  .object({
    data: ProjectResponseSchema,
  })
  .strip();

export const ProjectUpdateListResponseBodySchema = z
  .object({
    data: z.array(ProjectUpdateResponseSchema),
  })
  .strip();

export const ProjectUpdateRefInputSchema = z
  .object({
    update_id: z.string(), // The project update to delete
  })
  .strip();

export const ProjectUpdateResponseBodySchema = z
  .object({
    data: ProjectUpdateResponseSchema,
  })
  .strip();

export const ReactionInputSchema = z
  .object({
    comment_id: z.string().optional(), // React to this comment
    emoji: z.string(), // The emoji character, like 👍
    issue_id: z.string().optional(), // React to this issue; give exactly one of issue_id, comment_id and project_update_id
    project_update_id: z.string().optional(), // React to this project update
  })
  .strip();

export const ReactionsResponseBodySchema = z
  .object({
    data: z.array(ReactionGroupSchema),
  })
  .strip();

export const RedirectResponseBodySchema = z
  .object({
    data: ConnectedAppRedirectResponseSchema,
  })
  .strip();

export const RelationRefInputSchema = z
  .object({
    relation_id: z.string(), // The relation, from relation_id in get-issue's relations
  })
  .strip();

export const RelationResponseBodySchema = z
  .object({
    data: RelationResponseSchema,
  })
  .strip();

export const RemoveGithubRepoSyncResponseBodySchema = z
  .object({
    data: DeletedResponseSchema,
  })
  .strip();

export const RemoveMemberInputSchema = z
  .object({
    user_id: z.string(), // The member to remove; your own id to leave
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const RemoveTeamMemberInputSchema = z
  .object({
    team_id: z.string(), // The team
    user_id: z.string(), // The member to remove; your own id to leave
  })
  .strip();

export const ReorderFavoritesInputSchema = z
  .object({
    favorite_ids: z.array(z.string()), // Favorite ids (not entity ids) in the new order, first on top
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const RequestPasswordResetInputBodySchema = z
  .object({
    email: z.string(), // Email of the account
  })
  .strip();

export const ResetPasswordInputSchema = z
  .object({
    password: z.string(), // The new password, at least 8 characters
    token: z.string(), // Token from the reset link
  })
  .strip();

export const ResponseErrorSchema = z
  .object({
    details: z.array(z.string()).nullable().optional(), // Optional actionable details, e.g. which field failed validation
    message: z.string(), // Human-readable summary of what went wrong
    request_id: z.string().optional(), // Identifier of the request, quote it when reporting the failure so it can be found in the server logs
    status: z.number(), // HTTP status code
    type: z.enum([
      'bad_request',
      'unauthorized',
      'forbidden',
      'not_found',
      'conflict',
      'validation_error',
      'rate_limited',
      'internal_error',
      'error',
    ]), // Stable, machine-readable error code
  })
  .strip();

export const RevokeResponseBodySchema = z
  .object({
    data: DeletedResponseSchema,
  })
  .strip();

export const SearchResponseSchema = z
  .object({
    documents: z.array(DocumentResponseSchema),
    issues: z.array(IssueSummarySchema),
    projects: z.array(ProjectSummarySchema),
  })
  .strip();

export const SearchResultBodySchema = z
  .object({
    data: SearchResponseSchema,
  })
  .strip();

export const SessionUserSchema = z
  .object({
    email: z.string(),
    id: z.string(),
  })
  .strip();

export const SessionResponseBodySchema = z
  .object({
    data: SessionUserSchema,
  })
  .strip();

export const SystemSettingsResponseSchema = z
  .object({
    allow_public_signup: z.boolean(),
    email: EmailSettingsResponseSchema,
    instance_name: z.string(),
  })
  .strip();

export const SettingsResponseBodySchema = z
  .object({
    data: SystemSettingsResponseSchema,
  })
  .strip();

export const SetupStatusSchema = z
  .object({
    is_first_user_created: z.boolean(),
  })
  .strip();

export const SetupStatusResponseBodySchema = z
  .object({
    data: SetupStatusSchema,
  })
  .strip();

export const SnoozeNotificationsInputSchema = z
  .object({
    notification_ids: z.array(z.string()), // Ids of your notifications, from list-notifications
    until: z.string().datetime({ offset: true }), // RFC 3339 time in the future when the notifications return to the inbox as unread
  })
  .strip();

export const StateListResponseBodySchema = z
  .object({
    data: z.array(WorkflowStateResponseSchema),
  })
  .strip();

export const StateResponseBodySchema = z
  .object({
    data: WorkflowStateResponseSchema,
  })
  .strip();

export const StatusListResponseBodySchema = z
  .object({
    data: z.array(ProjectStatusResponseSchema),
  })
  .strip();

export const StatusResponseBodySchema = z
  .object({
    data: ProjectStatusResponseSchema,
  })
  .strip();

export const SubscriptionInputSchema = z
  .object({
    issue_id: z.string(), // The issue
    user_id: z.string().optional(), // Workspace member to subscribe or unsubscribe; defaults to you
  })
  .strip();

export const SuccessResponseBodySchema = z
  .object({
    data: SuccessResponseSchema,
  })
  .strip();

export const TeamResponseSchema = z
  .object({
    active_cycle_id: z.string().nullable(),
    auto_archive_months: z.number(),
    auto_close_months: z.number().nullable(),
    color: z.string(),
    created_at: z.string().datetime({ offset: true }),
    cycle_auto_add_started: z.boolean(),
    cycle_auto_move: z.boolean(),
    cycle_cooldown_weeks: z.number(),
    cycle_duration_weeks: z.number(),
    cycle_start_day: z.number(), // 1 Monday through 7 Sunday
    cycles_enabled: z.boolean(),
    default_state_id: z.string().nullable(), // State new issues get
    description: z.string(),
    estimate_allow_zero: z.boolean(),
    estimate_extended: z.boolean(),
    estimate_type: EstimateTypeSchema,
    github_pr_merged_state_id: z.string().nullable(), // State a merged pull request moves the issues it fixes to; null means the first completed state
    github_pr_opened_state_id: z.string().nullable(), // State a linked pull request opening moves issues to; null means the first started state
    icon: z.string(),
    id: z.string(),
    is_member: z.boolean(), // Whether the caller is on this team
    issue_counter: z.number(), // Last issue number handed out
    key: z.string(), // Prefix of issue identifiers, like ENG in ENG-123
    member_count: z.number(),
    name: z.string(),
    private: z.boolean(),
    states: z.array(WorkflowStateResponseSchema),
    timezone: z.string(),
    triage_enabled: z.boolean(),
    triage_state_id: z.string().nullable(),
    workspace_id: z.string(),
  })
  .strip();

export const TeamListResponseBodySchema = z
  .object({
    data: z.array(TeamResponseSchema),
  })
  .strip();

export const TeamMemberResponseSchema = z
  .object({
    joined_at: z.string().datetime({ offset: true }),
    owner: z.boolean(),
    user: UserSummarySchema,
  })
  .strip();

export const TeamMemberListResponseBodySchema = z
  .object({
    data: z.array(TeamMemberResponseSchema),
  })
  .strip();

export const TeamMemberResponseBodySchema = z
  .object({
    data: TeamMemberResponseSchema,
  })
  .strip();

export const TeamResponseBodySchema = z
  .object({
    data: TeamResponseSchema,
  })
  .strip();

export const TemplateResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    creator: UserSummarySchema.optional(),
    data: TemplateDataSchema, // Values a new issue starts with
    description: z.string(),
    id: z.string(),
    kind: TemplateKindSchema,
    name: z.string(),
    sort_order: z.number(),
    team_id: z.string().nullable(), // Team the template belongs to, null for a workspace template
    updated_at: z.string().datetime({ offset: true }),
    workspace_id: z.string(),
  })
  .strip();

export const TemplateListResponseBodySchema = z
  .object({
    data: z.array(TemplateResponseSchema),
  })
  .strip();

export const TemplateResponseBodySchema = z
  .object({
    data: TemplateResponseSchema,
  })
  .strip();

export const UpdateAttachmentInputSchema = z
  .object({
    attachment_id: z.string(), // The attachment
    subtitle: z.string().optional(), // New secondary line; empty clears it
    title: z.string().optional(), // New link text
  })
  .strip();

export const UpdateCommentInputSchema = z
  .object({
    body: z.string(), // New markdown body
    comment_id: z.string(), // The comment to edit; only its author or an admin may
  })
  .strip();

export const UpdateCycleInputSchema = z
  .object({
    cycle_id: z.string(), // The cycle to change
    description: z.string().optional(), // Description, empty to clear
    ends_at: z.string().datetime({ offset: true }).optional(), // New end, RFC 3339, after the start
    name: z.string().optional(), // Name, empty to clear
    starts_at: z.string().datetime({ offset: true }).optional(), // New start, RFC 3339
  })
  .strip();

export const UpdateDocumentInputSchema = z
  .object({
    content: z.string().optional(), // New markdown content, replacing the old
    document_id: z.string(), // The document to change
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    initiative_id: z.string().optional(), // Move to this initiative; the nil uuid detaches it
    issue_id: z.string().optional(), // Move to this issue; the nil uuid detaches it
    project_id: z.string().optional(), // Move to this project; the nil uuid detaches it
    title: z.string().optional(), // New title; empty becomes Untitled
  })
  .strip();

export const UpdateGithubIntegrationInputSchema = z
  .object({
    token: z.string(), // Fine-grained GitHub token with Issues read and write on the synced repositories; empty clears it
    workspace_id: z.string(), // Workspace whose GitHub connection to change
  })
  .strip();

export const UpdateGoogleAuthSettingsInputSchema = z
  .object({
    client_id: z.string(), // OAuth client id of a Google Cloud web application client
    client_secret: z.string().optional(), // OAuth client secret. Omit to keep the stored one.
    enabled: z.boolean(), // Offer Sign in with Google on the sign-in page
    hosted_domain: z.string().optional(), // Only accept Google Workspace accounts of this domain, for example example.com. Empty or omitted accepts any Google account.
  })
  .strip();

export const UpdateInitiativeInputSchema = z
  .object({
    color: z.string().optional(), // Hex color, empty to clear
    description: z.string().optional(), // Markdown description, empty to clear
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    initiative_id: z.string(), // The initiative to change
    name: z.string().optional(), // New name
    owner_id: z.string().optional(), // New owner; the nil uuid 00000000-0000-0000-0000-000000000000 removes the owner
    project_ids: z.array(z.string()).nullable().optional(), // Replaces the projects; an empty list removes all
    sort_order: z.number().optional(), // Manual order in lists
    status: InitiativeStatusSchema.optional(), // planned, active or completed
    target_date: z.string().optional(), // Target date YYYY-MM-DD, empty to clear
  })
  .strip();

export const UpdateIssueInputSchema = z
  .object({
    add_label_ids: z.array(z.string()).nullable().optional(), // Labels to add; a label of a single-select group replaces its sibling
    assignee_id: z.string().optional(), // Workspace member to assign; the nil uuid 00000000-0000-0000-0000-000000000000 unassigns
    cycle_id: z.string().optional(), // Cycle of the issue's team; the nil uuid removes it
    description: z.string().optional(), // New markdown description; empty clears it
    due_date: z.string().optional(), // Due date YYYY-MM-DD; empty clears it
    estimate: z.number().optional(), // Estimate in the team's scale; -1 clears it
    issue_id: z.string(), // The issue to change; omitted fields keep their value
    label_ids: z.array(z.string()).nullable().optional(), // Replaces every label; an empty list removes all
    milestone_id: z.string().optional(), // Milestone of the issue's project; the nil uuid removes it
    parent_id: z.string().optional(), // Parent issue; the nil uuid makes it top level
    priority: z.number().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    project_id: z.string().optional(), // Project of the workspace; the nil uuid removes it and its milestone
    remove_label_ids: z.array(z.string()).nullable().optional(), // Labels to remove
    snoozed_until: z.string().optional(), // RFC 3339 time to snooze the issue until; empty clears it
    sort_order: z.number().optional(), // Manual order in lists and boards, ascending
    state_id: z.string().optional(), // Workflow state of the issue's team
    sub_issue_sort_order: z.number().optional(), // Order among the parent's sub-issues, ascending
    team_id: z.string().optional(), // Move the issue to another team of the workspace; it gets a new identifier and the old one keeps resolving
    title: z.string().optional(), // New title
  })
  .strip();

export const UpdateLabelInputSchema = z
  .object({
    color: z.string().optional(), // Hex color
    description: z.string().optional(), // Description, empty to clear
    group_single_select: z.boolean().optional(), // For groups: an issue may carry only one label of the group
    label_id: z.string(), // The label to change
    name: z.string().optional(), // New name
    parent_id: z.string().optional(), // Move the label into this group; the nil uuid 00000000-0000-0000-0000-000000000000 takes it out of its group
    sort_order: z.number().optional(), // Position among its siblings, lower first
  })
  .strip();

export const UpdateMeInputSchema = z
  .object({
    avatar_upload_id: z.string().optional(), // An image you uploaded. Send 00000000-0000-0000-0000-000000000000 to remove the avatar.
    display_name: z.string().optional(), // Short handle used in mentions, lowercase and unique
    github_login: z.string().optional(), // Your GitHub username, so GitHub Issues sync credits and assigns you; empty clears it
    name: z.string().optional(), // Full name
    timezone: z.string().optional(), // IANA time zone such as Europe/Berlin
  })
  .strip();

export const UpdateMemberInputSchema = z
  .object({
    role: WorkspaceRoleSchema, // admin, member or viewer
    user_id: z.string(), // The member
    workspace_id: z.string(), // The workspace
  })
  .strip();

export const UpdateMilestoneInputSchema = z
  .object({
    description: z.string().optional(), // Description, empty to clear
    milestone_id: z.string(), // The milestone to change
    name: z.string().optional(), // New name
    sort_order: z.number().optional(), // Order within the project
    target_date: z.string().optional(), // Target date YYYY-MM-DD, empty to clear
  })
  .strip();

export const UpdateNotificationSettingsInputSchema = z
  .object({
    assigned: z.boolean().optional(), // Issues assigned to or unassigned from you
    commented: z.boolean().optional(), // Comments on issues you subscribe to
    due_soon: z.boolean().optional(), // Issues due soon
    email_enabled: z.boolean().optional(), // Send email for notifications at all
    mentioned: z.boolean().optional(), // Mentions in issues and comments
    project_updates: z.boolean().optional(), // Status updates on projects you follow
    state_changed: z.boolean().optional(), // State changes on issues you subscribe to
  })
  .strip();

export const UpdateProjectInputSchema = z
  .object({
    color: z.string().optional(), // Hex color, empty to clear
    description: z.string().optional(), // Markdown description, empty to clear
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    initiative_ids: z.array(z.string()).nullable().optional(), // Replaces the initiatives; an empty list removes all
    label_ids: z.array(z.string()).nullable().optional(), // Replaces the labels; an empty list removes all
    lead_id: z.string().optional(), // New lead; the nil uuid 00000000-0000-0000-0000-000000000000 removes the lead
    member_ids: z.array(z.string()).nullable().optional(), // Replaces the members; an empty list removes everyone
    name: z.string().optional(), // New name
    priority: z.number().optional(), // 0 none, 1 urgent, 2 high, 3 medium, 4 low
    project_id: z.string(), // The project to change
    sort_order: z.number().optional(), // Manual order in lists and boards
    start_date: z.string().optional(), // Planned start YYYY-MM-DD, empty to clear
    start_date_resolution: DateResolutionSchema.optional(), // Precision of start_date: day, month, quarter, half_year or year
    status_id: z.string().optional(), // New project status; moving to a started, completed or canceled status stamps started_at, completed_at or canceled_at
    summary: z.string().optional(), // One line summary, empty to clear
    target_date: z.string().optional(), // Target completion YYYY-MM-DD, empty to clear
    target_date_resolution: DateResolutionSchema.optional(), // Precision of target_date: day, month, quarter, half_year or year
    team_ids: z.array(z.string()).nullable().optional(), // Replaces the teams; at least one
  })
  .strip();

export const UpdateProjectLinkInputSchema = z
  .object({
    label: z.string().optional(), // New label
    link_id: z.string(), // The link to change
    sort_order: z.number().optional(), // Order among the project's links
    url: z.string().optional(), // New http or https URL
  })
  .strip();

export const UpdateProjectStatusInputSchema = z
  .object({
    color: z.string().optional(), // Hex color
    description: z.string().optional(), // Description, empty to clear
    name: z.string().optional(), // New name
    position: z.number().optional(), // Order within its type
    status_id: z.string(), // The project status to change
  })
  .strip();

export const UpdateSettingsInputSchema = z
  .object({
    allow_public_signup: z.boolean().optional(), // Let anyone create an account
    instance_name: z.string().optional(), // Name shown in the app and in emails
  })
  .strip();

export const UpdateStateInputSchema = z
  .object({
    color: z.string().optional(), // Hex color
    description: z.string().optional(), // Description, empty to clear
    name: z.string().optional(), // New name
    position: z.number().optional(), // Order within its type
    state_id: z.string(), // The state to change
  })
  .strip();

export const UpdateTeamInputSchema = z
  .object({
    auto_archive_months: z.number().optional(), // Archive completed and canceled issues after this many months
    auto_close_months: z.number().optional(), // Close stale issues after this many months, 0 turns it off
    color: z.string().optional(), // Hex color, empty to clear
    cycle_auto_add_started: z.boolean().optional(), // An issue that starts joins the active cycle
    cycle_auto_move: z.boolean().optional(), // Move unfinished issues to the next cycle
    cycle_cooldown_weeks: z.number().optional(), // Weeks between cycles
    cycle_duration_weeks: z.number().optional(), // Weeks per cycle
    cycle_start_day: z.number().optional(), // 1 Monday through 7 Sunday
    cycles_enabled: z.boolean().optional(), // Turn cycles on or off
    default_state_id: z.string().optional(), // State new issues get, one of this team's states
    description: z.string().optional(), // Description, empty to clear
    estimate_allow_zero: z.boolean().optional(), // Allow a zero estimate
    estimate_extended: z.boolean().optional(), // Offer the extended estimate scale
    estimate_type: EstimateTypeSchema.optional(), // none, exponential, fibonacci, linear or tshirt
    github_pr_merged_state_id: z.string().optional(), // State an issue moves to when a pull request that fixes, closes or resolves it merges; one of this team's states, the nil uuid restores the default, the first completed state
    github_pr_opened_state_id: z.string().optional(), // State an issue moves to from triage, backlog or unstarted when a linked pull request opens or its branch is pushed; one of this team's states, the nil uuid restores the default, the first started state
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    key: z.string().optional(), // New issue prefix; existing identifiers change with it
    name: z.string().optional(), // New name
    private: z.boolean().optional(), // Visible to members and workspace admins only
    team_id: z.string(), // The team to change
    timezone: z.string().optional(), // IANA time zone
    triage_enabled: z.boolean().optional(), // New issues from outside the team land in triage
  })
  .strip();

export const UpdateTemplateInputSchema = z
  .object({
    data: TemplateDataSchema.optional(), // Replaces all template values
    description: z.string().optional(), // Description, empty to clear
    name: z.string().optional(), // New name
    sort_order: z.number().optional(), // Position in the picker, lower first
    template_id: z.string(), // The template to change
  })
  .strip();

export const UpdateUserInputSchema = z
  .object({
    active: z.boolean().optional(), // Inactive users cannot sign in
    display_name: z.string().optional(), // Short handle used in mentions
    github_login: z.string().optional(), // GitHub username used by GitHub Issues sync; empty clears it
    name: z.string().optional(), // Full name
    password: z.string().optional(), // Set a new password; also activates placeholder accounts when active is sent
    system_admin: z.boolean().optional(), // Grant or revoke instance administration
    user_id: z.string(), // The user to change
  })
  .strip();

export const UpdateViewInputSchema = z
  .object({
    color: z.string().optional(), // Hex color, empty to clear
    description: z.string().optional(), // Description, empty to clear
    display: ViewDisplaySchema.optional(), // Replaces the display settings
    filters: z.record(z.any()).optional(), // Replaces all filters; list-issues parameter names
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    name: z.string().optional(), // New name
    shared: z.boolean().optional(), // Share with the workspace or make private
    view_id: z.string(), // The view to change; you must own it, or be an admin for a shared view
  })
  .strip();

export const UpdateWebhookInputSchema = z
  .object({
    enabled: z.boolean().optional(), // Turn delivery on or off
    events: z
      .array(
        z.enum([
          'issue.created',
          'issue.updated',
          'issue.deleted',
          'comment.created',
          'project.updated',
          'cycle.updated',
        ]),
      )
      .optional(), // Replaces the event list
    label: z.string().optional(), // New name
    team_id: z.string().optional(), // Only send events of this team; the nil uuid 00000000-0000-0000-0000-000000000000 sends the whole workspace
    url: z.string().optional(), // New endpoint
    webhook_id: z.string(), // Webhook to change
  })
  .strip();

export const UpdateWorkspaceInputSchema = z
  .object({
    color: z.string().optional(), // Hex color, empty to clear
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    icon_upload_id: z.string().optional(), // Logo image uploaded to this workspace with create-upload. Send 00000000-0000-0000-0000-000000000000 to remove the logo.
    name: z.string().optional(), // New name
    slug: z.string().optional(), // New URL key; old links stop working
    workspace_id: z.string(), // The workspace to change
  })
  .strip();

export const UploadResponseSchema = z
  .object({
    content_type: z.string(),
    file_name: z.string(),
    height: z.number().nullable().optional(), // Pixel height for png, jpeg, gif and webp images; for svg, the height from its height attribute or viewBox when given
    id: z.string(),
    size_bytes: z.number(),
    url: z.string(), // Same-origin path that serves the file; use it in markdown
    width: z.number().nullable().optional(), // Pixel width for png, jpeg, gif and webp images; for svg, the width from its width attribute or viewBox when given
  })
  .strip();

export const UploadResponseBodySchema = z
  .object({
    data: UploadResponseSchema,
  })
  .strip();

export const UserResponseBodySchema = z
  .object({
    data: UserResponseSchema,
  })
  .strip();

export const ViewResponseSchema = z
  .object({
    color: z.string(),
    created_at: z.string().datetime({ offset: true }),
    description: z.string(),
    display: ViewDisplaySchema,
    filters: z.record(z.any()), // list-issues parameters the view applies, such as state_types or label_ids
    icon: z.string(),
    id: z.string(),
    is_owner: z.boolean(), // Whether the caller owns the view
    kind: ViewKindSchema,
    name: z.string(),
    owner: UserSummarySchema.optional(),
    shared: z.boolean(), // Shared views are visible to everyone in the workspace
    team_id: z.string().nullable(), // Team the view belongs to, null for a workspace view
    updated_at: z.string().datetime({ offset: true }),
    workspace_id: z.string(),
  })
  .strip();

export const ViewListResponseBodySchema = z
  .object({
    data: z.array(ViewResponseSchema),
  })
  .strip();

export const ViewResponseBodySchema = z
  .object({
    data: ViewResponseSchema,
  })
  .strip();

export const WebhookCreatedResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    enabled: z.boolean(),
    events: z.array(
      z.enum([
        'issue.created',
        'issue.updated',
        'issue.deleted',
        'comment.created',
        'project.updated',
        'cycle.updated',
      ]),
    ),
    id: z.string(),
    label: z.string(),
    secret: z.string(), // Signing secret for X-Zigzag-Signature. Returned only here, store it now
    team_id: z.string().nullable(), // Only events of this team are sent; null for the whole workspace
    updated_at: z.string().datetime({ offset: true }),
    url: z.string(),
    workspace_id: z.string(),
  })
  .strip();

export const WebhookCreatedResultBodySchema = z
  .object({
    data: WebhookCreatedResponseSchema,
  })
  .strip();

export const WebhookResponseSchema = z
  .object({
    created_at: z.string().datetime({ offset: true }),
    enabled: z.boolean(),
    events: z.array(
      z.enum([
        'issue.created',
        'issue.updated',
        'issue.deleted',
        'comment.created',
        'project.updated',
        'cycle.updated',
      ]),
    ),
    id: z.string(),
    label: z.string(),
    team_id: z.string().nullable(), // Only events of this team are sent; null for the whole workspace
    updated_at: z.string().datetime({ offset: true }),
    url: z.string(),
    workspace_id: z.string(),
  })
  .strip();

export const WebhookListResponseBodySchema = z
  .object({
    data: z.array(WebhookResponseSchema),
  })
  .strip();

export const WebhookRefInputSchema = z
  .object({
    webhook_id: z.string(), // Webhook id from list-webhooks
  })
  .strip();

export const WebhookResultBodySchema = z
  .object({
    data: WebhookResponseSchema,
  })
  .strip();

export const WebhookTestResponseSchema = z
  .object({
    delivered: z.boolean(),
    error: z.string(), // Why the delivery failed, empty on success
    status_code: z.number(), // HTTP status the endpoint answered, 0 when it could not be reached
  })
  .strip();

export const WebhookTestResultBodySchema = z
  .object({
    data: WebhookTestResponseSchema,
  })
  .strip();

export const WhoamiAPIKeySchema = z
  .object({
    full_access: z.boolean(), // True when it reaches every workspace the user can
    privileges: z.array(KeyPrivilegeSchema),
    role: WorkspaceRoleSchema, // Strongest role this connection acts with
    workspace_ids: z.array(z.string()), // Workspaces this connection is limited to when not full access
  })
  .strip();

export const WorkspaceMembershipSchema = z
  .object({
    color: z.string(),
    icon: z.string(),
    icon_url: z.string().nullable(), // Logo image, shown instead of icon when set
    id: z.string(),
    name: z.string(),
    role: WorkspaceRoleSchema,
    slug: z.string(),
  })
  .strip();

export const WhoamiResponseSchema = z
  .object({
    api_key: WhoamiAPIKeySchema.optional(), // Present for API keys and connected apps: the limit the credential puts on this user
    system_admin: z.boolean(), // Effective system admin, false for narrowed credentials unless they hold full admin access
    user: UserResponseSchema,
    workspaces: z.array(WorkspaceMembershipSchema), // Workspaces this connection can see, with the effective role in each
  })
  .strip();

export const WhoamiResponseBodySchema = z
  .object({
    data: WhoamiResponseSchema,
  })
  .strip();

export const WorkspaceListResponseBodySchema = z
  .object({
    data: z.array(WorkspaceResponseSchema),
  })
  .strip();

export const WorkspaceMemberListResponseBodySchema = z
  .object({
    data: z.array(WorkspaceMemberResponseSchema),
  })
  .strip();

export const WorkspaceMemberResponseBodySchema = z
  .object({
    data: WorkspaceMemberResponseSchema,
  })
  .strip();

export const WorkspaceResponseBodySchema = z
  .object({
    data: WorkspaceResponseSchema,
  })
  .strip();

export type KeyPrivilege = z.infer<typeof KeyPrivilegeSchema>;
export type WorkspaceRole = z.infer<typeof WorkspaceRoleSchema>;
export type APIKeyCreateInput = z.infer<typeof APIKeyCreateInputSchema>;
export type APIKeyCreatedResponse = z.infer<typeof APIKeyCreatedResponseSchema>;
export type APIKeyDeleteInput = z.infer<typeof APIKeyDeleteInputSchema>;
export type APIKeyResponse = z.infer<typeof APIKeyResponseSchema>;
export type AcceptInvitationInput = z.infer<typeof AcceptInvitationInputSchema>;
export type UserResponse = z.infer<typeof UserResponseSchema>;
export type WorkspaceResponse = z.infer<typeof WorkspaceResponseSchema>;
export type SessionWorkspaceResponse = z.infer<typeof SessionWorkspaceResponseSchema>;
export type AcceptInvitationResponseBody = z.infer<typeof AcceptInvitationResponseBodySchema>;
export type AddGithubRepoSyncInput = z.infer<typeof AddGithubRepoSyncInputSchema>;
export type AddMemberInput = z.infer<typeof AddMemberInputSchema>;
export type InvitationResponse = z.infer<typeof InvitationResponseSchema>;
export type MemberUser = z.infer<typeof MemberUserSchema>;
export type WorkspaceMemberResponse = z.infer<typeof WorkspaceMemberResponseSchema>;
export type AddWorkspaceMemberResponse = z.infer<typeof AddWorkspaceMemberResponseSchema>;
export type AddMemberResponseBody = z.infer<typeof AddMemberResponseBodySchema>;
export type AddTeamMemberInput = z.infer<typeof AddTeamMemberInputSchema>;
export type AttachmentRefInput = z.infer<typeof AttachmentRefInputSchema>;
export type UserSummary = z.infer<typeof UserSummarySchema>;
export type AttachmentResponse = z.infer<typeof AttachmentResponseSchema>;
export type AttachmentResponseBody = z.infer<typeof AttachmentResponseBodySchema>;
export type AuthProviderStatus = z.infer<typeof AuthProviderStatusSchema>;
export type AuthProvidersResponse = z.infer<typeof AuthProvidersResponseSchema>;
export type AuthProvidersResponseBody = z.infer<typeof AuthProvidersResponseBodySchema>;
export type BackfillGithubRepoSyncInput = z.infer<typeof BackfillGithubRepoSyncInputSchema>;
export type BootstrapInput = z.infer<typeof BootstrapInputSchema>;
export type BulkUpdateIssuesInput = z.infer<typeof BulkUpdateIssuesInputSchema>;
export type ChangePasswordInput = z.infer<typeof ChangePasswordInputSchema>;
export type ReactionGroup = z.infer<typeof ReactionGroupSchema>;
export type CommentReplyResponse = z.infer<typeof CommentReplyResponseSchema>;
export type CommentResponse = z.infer<typeof CommentResponseSchema>;
export type CommentListResponseBody = z.infer<typeof CommentListResponseBodySchema>;
export type CommentRefInput = z.infer<typeof CommentRefInputSchema>;
export type CommentResponseBody = z.infer<typeof CommentResponseBodySchema>;
export type ConnectedAppApproveInput = z.infer<typeof ConnectedAppApproveInputSchema>;
export type OAuthClientKind = z.infer<typeof OAuthClientKindSchema>;
export type ConnectedAppClientResponse = z.infer<typeof ConnectedAppClientResponseSchema>;
export type ConnectedAppDenyInput = z.infer<typeof ConnectedAppDenyInputSchema>;
export type ConnectedAppRedirectResponse = z.infer<typeof ConnectedAppRedirectResponseSchema>;
export type ConnectedAppResponse = z.infer<typeof ConnectedAppResponseSchema>;
export type ConnectedAppRevokeInput = z.infer<typeof ConnectedAppRevokeInputSchema>;
export type Cookie = z.infer<typeof CookieSchema>;
export type CreateAPIKeyResponseBody = z.infer<typeof CreateAPIKeyResponseBodySchema>;
export type CreateAttachmentInput = z.infer<typeof CreateAttachmentInputSchema>;
export type CreateCommentInput = z.infer<typeof CreateCommentInputSchema>;
export type CreateCycleInput = z.infer<typeof CreateCycleInputSchema>;
export type CreateDocumentInput = z.infer<typeof CreateDocumentInputSchema>;
export type InitiativeStatus = z.infer<typeof InitiativeStatusSchema>;
export type CreateInitiativeInput = z.infer<typeof CreateInitiativeInputSchema>;
export type Health = z.infer<typeof HealthSchema>;
export type CreateInitiativeUpdateInput = z.infer<typeof CreateInitiativeUpdateInputSchema>;
export type CreateInvitationInput = z.infer<typeof CreateInvitationInputSchema>;
export type CreateIssueInput = z.infer<typeof CreateIssueInputSchema>;
export type CreateLabelInput = z.infer<typeof CreateLabelInputSchema>;
export type CreateMilestoneInput = z.infer<typeof CreateMilestoneInputSchema>;
export type DateResolution = z.infer<typeof DateResolutionSchema>;
export type CreateProjectInput = z.infer<typeof CreateProjectInputSchema>;
export type CreateProjectLinkInput = z.infer<typeof CreateProjectLinkInputSchema>;
export type ProjectStatusType = z.infer<typeof ProjectStatusTypeSchema>;
export type CreateProjectStatusInput = z.infer<typeof CreateProjectStatusInputSchema>;
export type CreateProjectUpdateInput = z.infer<typeof CreateProjectUpdateInputSchema>;
export type IssueRelationType = z.infer<typeof IssueRelationTypeSchema>;
export type CreateRelationInput = z.infer<typeof CreateRelationInputSchema>;
export type StateType = z.infer<typeof StateTypeSchema>;
export type CreateStateInput = z.infer<typeof CreateStateInputSchema>;
export type CreateTeamInput = z.infer<typeof CreateTeamInputSchema>;
export type TemplateSubIssue = z.infer<typeof TemplateSubIssueSchema>;
export type TemplateData = z.infer<typeof TemplateDataSchema>;
export type TemplateKind = z.infer<typeof TemplateKindSchema>;
export type CreateTemplateInput = z.infer<typeof CreateTemplateInputSchema>;
export type CreateUserInput = z.infer<typeof CreateUserInputSchema>;
export type CreateUserResponseBody = z.infer<typeof CreateUserResponseBodySchema>;
export type ViewDisplay = z.infer<typeof ViewDisplaySchema>;
export type ViewKind = z.infer<typeof ViewKindSchema>;
export type CreateViewInput = z.infer<typeof CreateViewInputSchema>;
export type CreateWebhookInput = z.infer<typeof CreateWebhookInputSchema>;
export type CreateWorkspaceInput = z.infer<typeof CreateWorkspaceInputSchema>;
export type CycleProgress = z.infer<typeof CycleProgressSchema>;
export type CycleScopePoint = z.infer<typeof CycleScopePointSchema>;
export type CycleResponse = z.infer<typeof CycleResponseSchema>;
export type CycleListResponseBody = z.infer<typeof CycleListResponseBodySchema>;
export type CycleRefInput = z.infer<typeof CycleRefInputSchema>;
export type CycleResponseBody = z.infer<typeof CycleResponseBodySchema>;
export type DeletedResponse = z.infer<typeof DeletedResponseSchema>;
export type DeleteAPIKeyResponseBody = z.infer<typeof DeleteAPIKeyResponseBodySchema>;
export type DeleteGithubResponseBody = z.infer<typeof DeleteGithubResponseBodySchema>;
export type DeleteProjectStatusInput = z.infer<typeof DeleteProjectStatusInputSchema>;
export type DeleteStateInput = z.infer<typeof DeleteStateInputSchema>;
export type DeleteTeamInput = z.infer<typeof DeleteTeamInputSchema>;
export type DeleteTemplateInput = z.infer<typeof DeleteTemplateInputSchema>;
export type DeleteUserInput = z.infer<typeof DeleteUserInputSchema>;
export type DeleteViewInput = z.infer<typeof DeleteViewInputSchema>;
export type DeleteWorkspaceInput = z.infer<typeof DeleteWorkspaceInputSchema>;
export type DeletedResponseBody = z.infer<typeof DeletedResponseBodySchema>;
export type InitiativeSummary = z.infer<typeof InitiativeSummarySchema>;
export type ProjectSummary = z.infer<typeof ProjectSummarySchema>;
export type DocumentResponse = z.infer<typeof DocumentResponseSchema>;
export type DocumentListResponseBody = z.infer<typeof DocumentListResponseBodySchema>;
export type DocumentRefInput = z.infer<typeof DocumentRefInputSchema>;
export type DocumentResponseBody = z.infer<typeof DocumentResponseBodySchema>;
export type EditProjectUpdateInput = z.infer<typeof EditProjectUpdateInputSchema>;
export type EmailSettings = z.infer<typeof EmailSettingsSchema>;
export type EmailSettingsResponse = z.infer<typeof EmailSettingsResponseSchema>;
export type EmailSettingsResponseBody = z.infer<typeof EmailSettingsResponseBodySchema>;
export type EmailSettingsUpdateInput = z.infer<typeof EmailSettingsUpdateInputSchema>;
export type EstimateType = z.infer<typeof EstimateTypeSchema>;
export type FavoriteKind = z.infer<typeof FavoriteKindSchema>;
export type FavoriteResponse = z.infer<typeof FavoriteResponseSchema>;
export type FavoriteListResponseBody = z.infer<typeof FavoriteListResponseBodySchema>;
export type FavoriteResponseBody = z.infer<typeof FavoriteResponseBodySchema>;
export type FavoriteTargetInput = z.infer<typeof FavoriteTargetInputSchema>;
export type FieldChange = z.infer<typeof FieldChangeSchema>;
export type GetClientResponseBody = z.infer<typeof GetClientResponseBodySchema>;
export type GithubRepositoryResponse = z.infer<typeof GithubRepositoryResponseSchema>;
export type GithubIntegrationResponse = z.infer<typeof GithubIntegrationResponseSchema>;
export type GithubIntegrationStatusResponse = z.infer<typeof GithubIntegrationStatusResponseSchema>;
export type GetGithubResponseBody = z.infer<typeof GetGithubResponseBodySchema>;
export type InvitationPreviewResponse = z.infer<typeof InvitationPreviewResponseSchema>;
export type GetInvitationResponseBody = z.infer<typeof GetInvitationResponseBodySchema>;
export type GithubBackfillReport = z.infer<typeof GithubBackfillReportSchema>;
export type GithubBackfillStatus = z.infer<typeof GithubBackfillStatusSchema>;
export type GithubDeliveryResponse = z.infer<typeof GithubDeliveryResponseSchema>;
export type GithubIntegrationRefInput = z.infer<typeof GithubIntegrationRefInputSchema>;
export type GithubIntegrationResponseBody = z.infer<typeof GithubIntegrationResponseBodySchema>;
export type GithubIntegrationSecretResponse = z.infer<typeof GithubIntegrationSecretResponseSchema>;
export type GithubIssueLinkResponse = z.infer<typeof GithubIssueLinkResponseSchema>;
export type GithubRepoSyncRefInput = z.infer<typeof GithubRepoSyncRefInputSchema>;
export type TeamSummary = z.infer<typeof TeamSummarySchema>;
export type GithubRepoSyncResponse = z.infer<typeof GithubRepoSyncResponseSchema>;
export type GithubRepoSyncResponseBody = z.infer<typeof GithubRepoSyncResponseBodySchema>;
export type GithubSecretResponseBody = z.infer<typeof GithubSecretResponseBodySchema>;
export type GithubWebhookResponseBody = z.infer<typeof GithubWebhookResponseBodySchema>;
export type GoogleAuthSettingsResponse = z.infer<typeof GoogleAuthSettingsResponseSchema>;
export type GoogleAuthSettingsResponseBody = z.infer<typeof GoogleAuthSettingsResponseBodySchema>;
export type HistorySource = z.infer<typeof HistorySourceSchema>;
export type HistoryResponse = z.infer<typeof HistoryResponseSchema>;
export type HistoryListResponseBody = z.infer<typeof HistoryListResponseBodySchema>;
export type ImportEntityCount = z.infer<typeof ImportEntityCountSchema>;
export type LinearCycle = z.infer<typeof LinearCycleSchema>;
export type LinearAttachment = z.infer<typeof LinearAttachmentSchema>;
export type LinearComment = z.infer<typeof LinearCommentSchema>;
export type LinearRelations = z.infer<typeof LinearRelationsSchema>;
export type LinearStateSpan = z.infer<typeof LinearStateSpanSchema>;
export type LinearIssue = z.infer<typeof LinearIssueSchema>;
export type LinearLabel = z.infer<typeof LinearLabelSchema>;
export type LinearMilestone = z.infer<typeof LinearMilestoneSchema>;
export type LinearProjectStatus = z.infer<typeof LinearProjectStatusSchema>;
export type LinearProject = z.infer<typeof LinearProjectSchema>;
export type LinearTeam = z.infer<typeof LinearTeamSchema>;
export type LinearUser = z.infer<typeof LinearUserSchema>;
export type LinearWorkflowState = z.infer<typeof LinearWorkflowStateSchema>;
export type LinearWorkspace = z.infer<typeof LinearWorkspaceSchema>;
export type LinearExport = z.infer<typeof LinearExportSchema>;
export type ImportLinearInput = z.infer<typeof ImportLinearInputSchema>;
export type ImportUserMapping = z.infer<typeof ImportUserMappingSchema>;
export type ImportWarning = z.infer<typeof ImportWarningSchema>;
export type ImportReport = z.infer<typeof ImportReportSchema>;
export type ImportSource = z.infer<typeof ImportSourceSchema>;
export type ImportStatus = z.infer<typeof ImportStatusSchema>;
export type ImportResponse = z.infer<typeof ImportResponseSchema>;
export type ImportListResponseBody = z.infer<typeof ImportListResponseBodySchema>;
export type ImportReportResponseBody = z.infer<typeof ImportReportResponseBodySchema>;
export type ProgressResponse = z.infer<typeof ProgressResponseSchema>;
export type InitiativeResponse = z.infer<typeof InitiativeResponseSchema>;
export type InitiativeListResponseBody = z.infer<typeof InitiativeListResponseBodySchema>;
export type InitiativeProjectInput = z.infer<typeof InitiativeProjectInputSchema>;
export type InitiativeRefInput = z.infer<typeof InitiativeRefInputSchema>;
export type InitiativeResponseBody = z.infer<typeof InitiativeResponseBodySchema>;
export type InitiativeUpdateResponse = z.infer<typeof InitiativeUpdateResponseSchema>;
export type InitiativeUpdateListResponseBody = z.infer<
  typeof InitiativeUpdateListResponseBodySchema
>;
export type InitiativeUpdateRefInput = z.infer<typeof InitiativeUpdateRefInputSchema>;
export type InitiativeUpdateResponseBody = z.infer<typeof InitiativeUpdateResponseBodySchema>;
export type InvitationListResponseBody = z.infer<typeof InvitationListResponseBodySchema>;
export type InvitationRefInput = z.infer<typeof InvitationRefInputSchema>;
export type InvitationResponseBody = z.infer<typeof InvitationResponseBodySchema>;
export type IssueCountsByStateType = z.infer<typeof IssueCountsByStateTypeSchema>;
export type IssueCountsResponse = z.infer<typeof IssueCountsResponseSchema>;
export type IssueCycleSummary = z.infer<typeof IssueCycleSummarySchema>;
export type LabelSummary = z.infer<typeof LabelSummarySchema>;
export type IssueMilestoneSummary = z.infer<typeof IssueMilestoneSummarySchema>;
export type WorkflowStateResponse = z.infer<typeof WorkflowStateResponseSchema>;
export type IssueSummary = z.infer<typeof IssueSummarySchema>;
export type IssueResponse = z.infer<typeof IssueResponseSchema>;
export type RelationResponse = z.infer<typeof RelationResponseSchema>;
export type IssueRelations = z.infer<typeof IssueRelationsSchema>;
export type IssueDetailResponse = z.infer<typeof IssueDetailResponseSchema>;
export type IssueDetailResponseBody = z.infer<typeof IssueDetailResponseBodySchema>;
export type IssueListResponseBody = z.infer<typeof IssueListResponseBodySchema>;
export type IssueRefInput = z.infer<typeof IssueRefInputSchema>;
export type IssueResponseBody = z.infer<typeof IssueResponseBodySchema>;
export type IssuesResponseBody = z.infer<typeof IssuesResponseBodySchema>;
export type LabelResponse = z.infer<typeof LabelResponseSchema>;
export type LabelTreeResponse = z.infer<typeof LabelTreeResponseSchema>;
export type LabelListResponseBody = z.infer<typeof LabelListResponseBodySchema>;
export type LabelRefInput = z.infer<typeof LabelRefInputSchema>;
export type LabelResponseBody = z.infer<typeof LabelResponseBodySchema>;
export type ProjectLinkResponse = z.infer<typeof ProjectLinkResponseSchema>;
export type LinkResponseBody = z.infer<typeof LinkResponseBodySchema>;
export type ListAPIKeysResponseBody = z.infer<typeof ListAPIKeysResponseBodySchema>;
export type ListGithubRepoSyncsResponseBody = z.infer<typeof ListGithubRepoSyncsResponseBodySchema>;
export type ListResponseBody = z.infer<typeof ListResponseBodySchema>;
export type ListUsersResponseBody = z.infer<typeof ListUsersResponseBodySchema>;
export type LoginInputBody = z.infer<typeof LoginInputBodySchema>;
export type MarkAllReadInput = z.infer<typeof MarkAllReadInputSchema>;
export type MilestoneResponse = z.infer<typeof MilestoneResponseSchema>;
export type MilestoneListResponseBody = z.infer<typeof MilestoneListResponseBodySchema>;
export type MilestoneRefInput = z.infer<typeof MilestoneRefInputSchema>;
export type MilestoneResponseBody = z.infer<typeof MilestoneResponseBodySchema>;
export type NotificationCountsResponse = z.infer<typeof NotificationCountsResponseSchema>;
export type NotificationCountsResultBody = z.infer<typeof NotificationCountsResultBodySchema>;
export type NotificationIDsInput = z.infer<typeof NotificationIDsInputSchema>;
export type NotificationType = z.infer<typeof NotificationTypeSchema>;
export type NotificationResponse = z.infer<typeof NotificationResponseSchema>;
export type NotificationListResponseBody = z.infer<typeof NotificationListResponseBodySchema>;
export type NotificationSettingsResponse = z.infer<typeof NotificationSettingsResponseSchema>;
export type NotificationSettingsResponseBody = z.infer<
  typeof NotificationSettingsResponseBodySchema
>;
export type SuccessResponse = z.infer<typeof SuccessResponseSchema>;
export type PlainSuccessResponseBody = z.infer<typeof PlainSuccessResponseBodySchema>;
export type ProjectLinkRefInput = z.infer<typeof ProjectLinkRefInputSchema>;
export type ProjectUpdateResponse = z.infer<typeof ProjectUpdateResponseSchema>;
export type ProjectStatusResponse = z.infer<typeof ProjectStatusResponseSchema>;
export type ProjectResponse = z.infer<typeof ProjectResponseSchema>;
export type ProjectListResponseBody = z.infer<typeof ProjectListResponseBodySchema>;
export type ProjectRefInput = z.infer<typeof ProjectRefInputSchema>;
export type ProjectResponseBody = z.infer<typeof ProjectResponseBodySchema>;
export type ProjectUpdateListResponseBody = z.infer<typeof ProjectUpdateListResponseBodySchema>;
export type ProjectUpdateRefInput = z.infer<typeof ProjectUpdateRefInputSchema>;
export type ProjectUpdateResponseBody = z.infer<typeof ProjectUpdateResponseBodySchema>;
export type ReactionInput = z.infer<typeof ReactionInputSchema>;
export type ReactionsResponseBody = z.infer<typeof ReactionsResponseBodySchema>;
export type RedirectResponseBody = z.infer<typeof RedirectResponseBodySchema>;
export type RelationRefInput = z.infer<typeof RelationRefInputSchema>;
export type RelationResponseBody = z.infer<typeof RelationResponseBodySchema>;
export type RemoveGithubRepoSyncResponseBody = z.infer<
  typeof RemoveGithubRepoSyncResponseBodySchema
>;
export type RemoveMemberInput = z.infer<typeof RemoveMemberInputSchema>;
export type RemoveTeamMemberInput = z.infer<typeof RemoveTeamMemberInputSchema>;
export type ReorderFavoritesInput = z.infer<typeof ReorderFavoritesInputSchema>;
export type RequestPasswordResetInputBody = z.infer<typeof RequestPasswordResetInputBodySchema>;
export type ResetPasswordInput = z.infer<typeof ResetPasswordInputSchema>;
export type ResponseError = z.infer<typeof ResponseErrorSchema>;
export type RevokeResponseBody = z.infer<typeof RevokeResponseBodySchema>;
export type SearchResponse = z.infer<typeof SearchResponseSchema>;
export type SearchResultBody = z.infer<typeof SearchResultBodySchema>;
export type SessionUser = z.infer<typeof SessionUserSchema>;
export type SessionResponseBody = z.infer<typeof SessionResponseBodySchema>;
export type SystemSettingsResponse = z.infer<typeof SystemSettingsResponseSchema>;
export type SettingsResponseBody = z.infer<typeof SettingsResponseBodySchema>;
export type SetupStatus = z.infer<typeof SetupStatusSchema>;
export type SetupStatusResponseBody = z.infer<typeof SetupStatusResponseBodySchema>;
export type SnoozeNotificationsInput = z.infer<typeof SnoozeNotificationsInputSchema>;
export type StateListResponseBody = z.infer<typeof StateListResponseBodySchema>;
export type StateResponseBody = z.infer<typeof StateResponseBodySchema>;
export type StatusListResponseBody = z.infer<typeof StatusListResponseBodySchema>;
export type StatusResponseBody = z.infer<typeof StatusResponseBodySchema>;
export type SubscriptionInput = z.infer<typeof SubscriptionInputSchema>;
export type SuccessResponseBody = z.infer<typeof SuccessResponseBodySchema>;
export type TeamResponse = z.infer<typeof TeamResponseSchema>;
export type TeamListResponseBody = z.infer<typeof TeamListResponseBodySchema>;
export type TeamMemberResponse = z.infer<typeof TeamMemberResponseSchema>;
export type TeamMemberListResponseBody = z.infer<typeof TeamMemberListResponseBodySchema>;
export type TeamMemberResponseBody = z.infer<typeof TeamMemberResponseBodySchema>;
export type TeamResponseBody = z.infer<typeof TeamResponseBodySchema>;
export type TemplateResponse = z.infer<typeof TemplateResponseSchema>;
export type TemplateListResponseBody = z.infer<typeof TemplateListResponseBodySchema>;
export type TemplateResponseBody = z.infer<typeof TemplateResponseBodySchema>;
export type UpdateAttachmentInput = z.infer<typeof UpdateAttachmentInputSchema>;
export type UpdateCommentInput = z.infer<typeof UpdateCommentInputSchema>;
export type UpdateCycleInput = z.infer<typeof UpdateCycleInputSchema>;
export type UpdateDocumentInput = z.infer<typeof UpdateDocumentInputSchema>;
export type UpdateGithubIntegrationInput = z.infer<typeof UpdateGithubIntegrationInputSchema>;
export type UpdateGoogleAuthSettingsInput = z.infer<typeof UpdateGoogleAuthSettingsInputSchema>;
export type UpdateInitiativeInput = z.infer<typeof UpdateInitiativeInputSchema>;
export type UpdateIssueInput = z.infer<typeof UpdateIssueInputSchema>;
export type UpdateLabelInput = z.infer<typeof UpdateLabelInputSchema>;
export type UpdateMeInput = z.infer<typeof UpdateMeInputSchema>;
export type UpdateMemberInput = z.infer<typeof UpdateMemberInputSchema>;
export type UpdateMilestoneInput = z.infer<typeof UpdateMilestoneInputSchema>;
export type UpdateNotificationSettingsInput = z.infer<typeof UpdateNotificationSettingsInputSchema>;
export type UpdateProjectInput = z.infer<typeof UpdateProjectInputSchema>;
export type UpdateProjectLinkInput = z.infer<typeof UpdateProjectLinkInputSchema>;
export type UpdateProjectStatusInput = z.infer<typeof UpdateProjectStatusInputSchema>;
export type UpdateSettingsInput = z.infer<typeof UpdateSettingsInputSchema>;
export type UpdateStateInput = z.infer<typeof UpdateStateInputSchema>;
export type UpdateTeamInput = z.infer<typeof UpdateTeamInputSchema>;
export type UpdateTemplateInput = z.infer<typeof UpdateTemplateInputSchema>;
export type UpdateUserInput = z.infer<typeof UpdateUserInputSchema>;
export type UpdateViewInput = z.infer<typeof UpdateViewInputSchema>;
export type UpdateWebhookInput = z.infer<typeof UpdateWebhookInputSchema>;
export type UpdateWorkspaceInput = z.infer<typeof UpdateWorkspaceInputSchema>;
export type UploadResponse = z.infer<typeof UploadResponseSchema>;
export type UploadResponseBody = z.infer<typeof UploadResponseBodySchema>;
export type UserResponseBody = z.infer<typeof UserResponseBodySchema>;
export type ViewResponse = z.infer<typeof ViewResponseSchema>;
export type ViewListResponseBody = z.infer<typeof ViewListResponseBodySchema>;
export type ViewResponseBody = z.infer<typeof ViewResponseBodySchema>;
export type WebhookCreatedResponse = z.infer<typeof WebhookCreatedResponseSchema>;
export type WebhookCreatedResultBody = z.infer<typeof WebhookCreatedResultBodySchema>;
export type WebhookResponse = z.infer<typeof WebhookResponseSchema>;
export type WebhookListResponseBody = z.infer<typeof WebhookListResponseBodySchema>;
export type WebhookRefInput = z.infer<typeof WebhookRefInputSchema>;
export type WebhookResultBody = z.infer<typeof WebhookResultBodySchema>;
export type WebhookTestResponse = z.infer<typeof WebhookTestResponseSchema>;
export type WebhookTestResultBody = z.infer<typeof WebhookTestResultBodySchema>;
export type WhoamiAPIKey = z.infer<typeof WhoamiAPIKeySchema>;
export type WorkspaceMembership = z.infer<typeof WorkspaceMembershipSchema>;
export type WhoamiResponse = z.infer<typeof WhoamiResponseSchema>;
export type WhoamiResponseBody = z.infer<typeof WhoamiResponseBodySchema>;
export type WorkspaceListResponseBody = z.infer<typeof WorkspaceListResponseBodySchema>;
export type WorkspaceMemberListResponseBody = z.infer<typeof WorkspaceMemberListResponseBodySchema>;
export type WorkspaceMemberResponseBody = z.infer<typeof WorkspaceMemberResponseBodySchema>;
export type WorkspaceResponseBody = z.infer<typeof WorkspaceResponseBodySchema>;

export const list_api_keysQuerySchema = z
  .object({
    user_id: z.string().optional(), // List another user's keys. Requires system admin. Defaults to your own.
  })
  .passthrough();

export const finish_google_loginQuerySchema = z
  .object({
    code: z.string().optional(), // Authorization code from Google
    state: z.string().optional(), // State echoed back by Google
    error: z.string().optional(), // Error from Google, for example access_denied
  })
  .passthrough();

export const start_google_loginQuerySchema = z
  .object({
    redirect: z.string().optional(), // App path to return to after signing in, for example /acme/my-issues
  })
  .passthrough();

export const get_connected_app_clientQuerySchema = z
  .object({
    client_id: z.string(), // OAuth client_id from the authorize request
    redirect_uri: z.string(), // redirect_uri from the authorize request
  })
  .passthrough();

export const get_cycleQuerySchema = z
  .object({
    cycle_id: z.string().optional(), // Cycle id; give this, or team_id and number
    team_id: z.string().optional(), // Team of the cycle when looking up by number
    number: z.number().optional(), // Cycle number within the team, with team_id
  })
  .passthrough();

export const list_cyclesQuerySchema = z
  .object({
    team_id: z.string(), // The team
    include_archived: z.boolean().optional(), // Include archived cycles
  })
  .passthrough();

export const get_documentQuerySchema = z
  .object({
    document_id: z.string().optional(), // Document id; give this, or workspace_id and slug
    workspace_id: z.string().optional(), // Workspace of the document when looking up by slug
    slug: z.string().optional(), // Document slug from its URL, with workspace_id
  })
  .passthrough();

export const list_documentsQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    project_id: z.string().optional(), // Only documents of this project
    initiative_id: z.string().optional(), // Only documents of this initiative
    issue_id: z.string().optional(), // Only documents of this issue
    query: z.string().optional(), // Case insensitive text matched against the title
    include_archived: z.boolean().optional(), // Include archived documents
    include_trashed: z.boolean().optional(), // Include trashed documents
  })
  .passthrough();

export const list_favoritesQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
  })
  .passthrough();

export const list_importsQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace whose imports to list; needs admin
  })
  .passthrough();

export const get_initiativeQuerySchema = z
  .object({
    initiative_id: z.string().optional(), // Initiative id; give this, or workspace_id and slug
    workspace_id: z.string().optional(), // Workspace of the initiative when looking up by slug
    slug: z.string().optional(), // Initiative slug from its URL, with workspace_id
  })
  .passthrough();

export const list_initiativesQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    status: InitiativeStatusSchema.optional(), // Only initiatives with this status: planned, active or completed
    owner_id: z.string().optional(), // Only initiatives owned by this user
    include_archived: z.boolean().optional(), // Include archived initiatives
    include_trashed: z.boolean().optional(), // Include trashed initiatives
  })
  .passthrough();

export const list_initiative_updatesQuerySchema = z
  .object({
    initiative_id: z.string(), // The initiative
  })
  .passthrough();

export const get_github_integrationQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace whose GitHub connection to read
  })
  .passthrough();

export const get_github_repo_syncQuerySchema = z
  .object({
    repo_sync_id: z.string(), // The synced repository
  })
  .passthrough();

export const list_github_repo_syncsQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace whose synced repositories to list
  })
  .passthrough();

export const github_webhookQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace the delivery is for, part of the webhook URL shown in Zigzag
  })
  .passthrough();

export const get_invitationQuerySchema = z
  .object({
    token: z.string(), // Token from the invitation link
  })
  .passthrough();

export const list_commentsQuerySchema = z
  .object({
    issue_id: z.string(), // The issue
    limit: z.number().optional(), // Page size, 1 to 250
    cursor: z.string().optional(), // Opaque cursor from the previous page's next_cursor
  })
  .passthrough();

export const get_issueQuerySchema = z
  .object({
    issue_id: z.string().optional(), // Issue id; give this or identifier
    identifier: z.string().optional(), // Identifier like ENG-123; identifiers from before a team move still resolve
    workspace_id: z.string().optional(), // Workspace to look the identifier up in; needed only when several of your workspaces have a team with that key
  })
  .passthrough();

export const list_issue_historyQuerySchema = z
  .object({
    issue_id: z.string(), // The issue
    limit: z.number().optional(), // Page size, 1 to 250
    cursor: z.string().optional(), // Opaque cursor from the previous page's next_cursor
  })
  .passthrough();

export const list_issuesQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace to list issues of
    team_id: z.string().optional(), // Only issues of this team
    team_ids: z.array(z.string()).nullable().optional(), // Only issues of these teams
    state_ids: z.array(z.string()).nullable().optional(), // Only issues in these workflow states
    state_types: z.array(StateTypeSchema).nullable().optional(), // Only issues whose state has one of these types: triage, backlog, unstarted, started, completed, canceled
    assignee_ids: z.array(z.string()).nullable().optional(), // Only issues assigned to these users; include the word null for unassigned issues
    creator_ids: z.array(z.string()).nullable().optional(), // Only issues created by these users
    label_ids: z.array(z.string()).nullable().optional(), // Only issues carrying any of these labels; a label group matches all of its labels
    project_ids: z.array(z.string()).nullable().optional(), // Only issues in these projects
    milestone_ids: z.array(z.string()).nullable().optional(), // Only issues in these project milestones
    cycle_ids: z.array(z.string()).nullable().optional(), // Only issues in these cycles
    parent_id: z.string().optional(), // Only sub-issues of this issue
    has_parent: z.enum(['true', 'false']).optional(), // true for sub-issues only, false for top-level issues only
    priorities: z.array(z.number()).nullable().optional(), // Only these priorities: 0 none, 1 urgent, 2 high, 3 medium, 4 low
    due_before: z.string().optional(), // Only issues due on or before this date, YYYY-MM-DD
    due_after: z.string().optional(), // Only issues due on or after this date, YYYY-MM-DD
    created_after: z.string().datetime({ offset: true }).optional(), // Only issues created at or after this RFC 3339 time
    updated_after: z.string().datetime({ offset: true }).optional(), // Only issues updated at or after this RFC 3339 time
    query: z.string().optional(), // Text in the title or description, or an identifier like ENG-123
    subscriber_id: z.string().optional(), // Only issues this user subscribes to
    include_archived: z.boolean().optional(), // Include archived issues
    include_trashed: z.boolean().optional(), // Include issues in the trash
    order_by: z
      .enum(['manual', 'priority', 'created_at', 'updated_at', 'due_date', 'title'])
      .optional(), // Sort key; manual is the drag and drop order
    order: z.enum(['asc', 'desc']).optional(), // Sort direction; defaults to desc for created_at and updated_at, asc otherwise
    group_by: z
      .enum(['none', 'state', 'assignee', 'priority', 'project', 'cycle', 'label', 'parent'])
      .optional(), // Sort by this grouping first so pages arrive group by group; label groups by the alphabetically first label
    limit: z.number().optional(), // Page size, 1 to 250
    cursor: z.string().optional(), // Opaque cursor from the previous page's next_cursor
  })
  .passthrough();

export const search_issuesQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace to search
    query: z.string(), // Words to look for in titles and descriptions, or an identifier like ENG-123
    limit: z.number().optional(), // Most results to return, 1 to 50
  })
  .passthrough();

export const list_labelsQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    team_id: z.string().optional(), // Also return the labels of this team. Without it only workspace labels come back
    all_teams: z.boolean().optional(), // Also return the labels of every team you can see; ignored when team_id is set
    include_archived: z.boolean().optional(), // Include archived labels
  })
  .passthrough();

export const get_notification_countsQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace to count notifications in
  })
  .passthrough();

export const list_notificationsQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace whose notifications to list
    filter: z.enum(['inbox', 'unread', 'read', 'archived', 'snoozed', 'all']).optional(), // inbox (default): read and unread, not archived, not snoozed; unread or read: that part of the inbox; archived; snoozed: hidden until snoozed_until; all: everything
    limit: z.number().optional(), // Page size, 1 to 250
    cursor: z.string().optional(), // Opaque cursor from the previous page's next_cursor
  })
  .passthrough();

export const get_projectQuerySchema = z
  .object({
    project_id: z.string().optional(), // Project id; give this, or workspace_id and slug
    workspace_id: z.string().optional(), // Workspace of the project when looking up by slug
    slug: z.string().optional(), // Project slug from its URL, such as launch-1e0e8889, with workspace_id
  })
  .passthrough();

export const list_projectsQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    team_id: z.string().optional(), // Only projects this team takes part in
    status_type: ProjectStatusTypeSchema.optional(), // Only projects whose status has this type: backlog, planned, started, paused, completed or canceled
    status_id: z.string().optional(), // Only projects in this status
    lead_id: z.string().optional(), // Only projects led by this user
    member_id: z.string().optional(), // Only projects this user is a member of
    initiative_id: z.string().optional(), // Only projects in this initiative
    label_id: z.string().optional(), // Only projects with this label
    query: z.string().optional(), // Case insensitive text matched against name and summary
    include_archived: z.boolean().optional(), // Include archived projects
    include_trashed: z.boolean().optional(), // Include trashed projects
    order_by: z
      .enum(['sort_order', 'name', 'priority', 'target_date', 'created_at', 'updated_at'])
      .optional(), // Sort order; sort_order (manual) by default, created_at and updated_at newest first
  })
  .passthrough();

export const list_milestonesQuerySchema = z
  .object({
    project_id: z.string(), // The project
  })
  .passthrough();

export const list_project_statusesQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
  })
  .passthrough();

export const list_project_updatesQuerySchema = z
  .object({
    project_id: z.string(), // The project
  })
  .passthrough();

export const searchQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace to search in
    query: z.string(), // Words to look for in issue titles and descriptions, project names and summaries, document titles and content
    limit: z.number().optional(), // Most results per kind, 1 to 50
  })
  .passthrough();

export const get_teamQuerySchema = z
  .object({
    team_id: z.string().optional(), // Team id; give this, or workspace_id and key
    workspace_id: z.string().optional(), // Workspace of the team when looking up by key
    key: z.string().optional(), // Team key such as ENG, with workspace_id
  })
  .passthrough();

export const list_teamsQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    include_archived: z.boolean().optional(), // Include archived teams
  })
  .passthrough();

export const list_team_membersQuerySchema = z
  .object({
    team_id: z.string(), // The team
  })
  .passthrough();

export const list_workflow_statesQuerySchema = z
  .object({
    team_id: z.string(), // The team
  })
  .passthrough();

export const get_templateQuerySchema = z
  .object({
    template_id: z.string(), // The template
  })
  .passthrough();

export const list_templatesQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    team_id: z.string().optional(), // Also return this team's templates. Without it only workspace templates come back
    kind: TemplateKindSchema.optional(), // issue, project or document
  })
  .passthrough();

export const get_uploadQuerySchema = z
  .object({
    upload_id: z.string(), // The upload
  })
  .passthrough();

export const list_usersQuerySchema = z
  .object({
    query: z.string().optional(), // Filter by email, name or display name
    limit: z.number().optional(), // Maximum number of users to return
  })
  .passthrough();

export const get_viewQuerySchema = z
  .object({
    view_id: z.string(), // The view
  })
  .passthrough();

export const list_viewsQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
    team_id: z.string().optional(), // Only views of this team
    kind: ViewKindSchema.optional(), // issue or project
  })
  .passthrough();

export const list_webhooksQuerySchema = z
  .object({
    workspace_id: z.string(), // Workspace whose webhooks to list
  })
  .passthrough();

export const get_workspaceQuerySchema = z
  .object({
    workspace_id: z.string().optional(), // Workspace id; give this or slug
    slug: z.string().optional(), // Workspace URL key; give this or workspace_id
  })
  .passthrough();

export const list_invitationsQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
  })
  .passthrough();

export const list_workspace_membersQuerySchema = z
  .object({
    workspace_id: z.string(), // The workspace
  })
  .passthrough();

/**
 * Error thrown for any non-2xx API response. Extends Error so the
 * existing `error.message` call sites keep working, while exposing the
 * structured fields of the API's error envelope for callers that want to
 * branch on them.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly type: string;
  readonly details: string[];
  /** Id of the request that failed, to look the failure up in the server logs. */
  readonly requestId: string | undefined;
  /** Parsed JSON body, or undefined when the response wasn't JSON. */
  readonly body: unknown;
  /** Unparsed response body, kept for non-JSON failures (proxies, gateways). */
  readonly raw: string;

  constructor(
    message: string,
    init: {
      status: number;
      type: string;
      details: string[];
      requestId?: string;
      body: unknown;
      raw: string;
    },
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = init.status;
    this.type = init.type;
    this.details = init.details;
    this.requestId = init.requestId;
    this.body = init.body;
    this.raw = init.raw;
  }
}

/**
 * Non-JSON bodies come from proxies and gateways rather than the API. A
 * short one-liner ("upstream connect error...") is worth surfacing; an
 * HTML page or a truncated JSON fragment in a toast is not, and an
 * unbounded body has no place in the DOM.
 */
function plainTextMessage(raw: string): string | undefined {
  const text = raw.trim().replace(/\s+/g, ' ');
  if (!text || text.length > 200) return undefined;
  if (/^[<{[]/.test(text)) return undefined;
  return text;
}

/**
 * Builds an ApiError from a failed response. The body is best-effort:
 * gateways and proxies can return HTML or nothing at all, so a body that
 * fails to parse must not mask the underlying HTTP failure.
 */
async function parseApiError(response: Response, url: string): Promise<ApiError> {
  // Read as text first and parse by hand: response.json() consumes the
  // stream, so a parse failure would leave the body unrecoverable.
  let raw = '';
  try {
    raw = await response.text();
  } catch {
    raw = '';
  }

  let body: unknown;
  try {
    body = raw ? JSON.parse(raw) : undefined;
  } catch {
    body = undefined;
  }

  // The documented envelope is ResponseError, but gateways and proxies can
  // return anything, so fall back to a lenient read before giving up on
  // the body entirely.
  const parsed = ResponseErrorSchema.safeParse(body);
  const envelope = parsed.success
    ? parsed.data
    : (body as { message?: unknown; type?: unknown; details?: unknown } | undefined);
  const details = Array.isArray(envelope?.details)
    ? envelope.details.filter((d): d is string => typeof d === 'string')
    : [];
  const summary = (typeof envelope?.message === 'string' && envelope.message) || undefined;
  // The summary is often a generic title ("Invalid input", "Not found") while the actionable
  // text sits in `details`. Join them so the `error.message` call sites show the reason too.
  const detail = details.find((d) => d && d !== summary);
  const message =
    (summary && detail ? `${summary}: ${detail}` : summary) ||
    detail ||
    // Only echo the body when it wasn't JSON at all: a body that parsed
    // but carried no message (null, {}, a bare array) is machine output,
    // not something to show a user.
    (body === undefined ? plainTextMessage(raw) : undefined) ||
    `Request failed with status ${response.status}`;
  const type = typeof envelope?.type === 'string' ? envelope.type : 'error';
  // The header is set for every response, so a failure that never reached a
  // handler (or never produced JSON) is still traceable.
  const requestId =
    (typeof (envelope as { request_id?: unknown } | undefined)?.request_id === 'string'
      ? ((envelope as { request_id?: string }).request_id as string)
      : undefined) ||
    response.headers.get('X-Request-Id') ||
    undefined;
  // Only server-side failures need the reference: it is what ties the red
  // banner to one line in the API log. A rejected input needs no ticket.
  const messageWithRef =
    requestId && response.status >= 500 ? `${message} (ref: ${requestId})` : message;

  if (import.meta.env.DEV) {
    // Never log the request body: it can contain passwords and secrets.
    console.error(`API ${response.status} ${url}`, body ?? (raw || '(empty body)'));
  }

  return new ApiError(messageWithRef, {
    status: response.status,
    type,
    details,
    requestId,
    body,
    raw,
  });
}

export type ClientOptions = {
  apiUrl: string;
  fetchFn?: typeof fetch;
};

export function createClient({ apiUrl, fetchFn = fetch }: ClientOptions) {
  return {
    apiKeys: {
      create: async (
        params: APIKeyCreateInput,
        fetchOptions?: RequestInit,
      ): Promise<CreateAPIKeyResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/api-keys/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = APIKeyCreateInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CreateAPIKeyResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: APIKeyDeleteInput,
        fetchOptions?: RequestInit,
      ): Promise<DeleteAPIKeyResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/api-keys/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = APIKeyDeleteInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeleteAPIKeyResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_api_keysQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ListAPIKeysResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/api-keys/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_api_keysQuerySchema.parse(params);
          const queryKeys = ['user_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ListAPIKeysResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    auth: {
      google: {
        callback: async (
          params: z.infer<typeof finish_google_loginQuerySchema>,
          fetchOptions?: RequestInit,
        ) => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/auth/google/callback`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = finish_google_loginQuerySchema.parse(params);
            const queryKeys = ['code', 'state', 'error'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            return data;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        start: async (
          params: z.infer<typeof start_google_loginQuerySchema>,
          fetchOptions?: RequestInit,
        ) => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/auth/google/start`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = start_google_loginQuerySchema.parse(params);
            const queryKeys = ['redirect'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            return data;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      login: async (
        params: LoginInputBody,
        fetchOptions?: RequestInit,
      ): Promise<SessionResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/auth/login`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = LoginInputBodySchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SessionResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      logout: async (
        params?: undefined,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/auth/logout`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      providers: async (
        params?: undefined,
        fetchOptions?: RequestInit,
      ): Promise<AuthProvidersResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/auth/providers`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = AuthProvidersResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      requestPasswordReset: async (
        params: RequestPasswordResetInputBody,
        fetchOptions?: RequestInit,
      ): Promise<PlainSuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/auth/request-password-reset`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = RequestPasswordResetInputBodySchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = PlainSuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      resetPassword: async (
        params: ResetPasswordInput,
        fetchOptions?: RequestInit,
      ): Promise<PlainSuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/auth/reset-password`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ResetPasswordInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = PlainSuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    connectedApps: {
      approve: async (
        params: ConnectedAppApproveInput,
        fetchOptions?: RequestInit,
      ): Promise<RedirectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/connected-apps/approve`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ConnectedAppApproveInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = RedirectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      client: async (
        params: z.infer<typeof get_connected_app_clientQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<GetClientResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/connected-apps/client`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_connected_app_clientQuerySchema.parse(params);
          const queryKeys = ['client_id', 'redirect_uri'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = GetClientResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      deny: async (
        params: ConnectedAppDenyInput,
        fetchOptions?: RequestInit,
      ): Promise<RedirectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/connected-apps/deny`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ConnectedAppDenyInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = RedirectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (params?: undefined, fetchOptions?: RequestInit): Promise<ListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/connected-apps/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      revoke: async (
        params: ConnectedAppRevokeInput,
        fetchOptions?: RequestInit,
      ): Promise<RevokeResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/connected-apps/revoke`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ConnectedAppRevokeInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = RevokeResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    cycles: {
      archive: async (
        params: CycleRefInput,
        fetchOptions?: RequestInit,
      ): Promise<CycleResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/cycles/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CycleRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CycleResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      create: async (
        params: CreateCycleInput,
        fetchOptions?: RequestInit,
      ): Promise<CycleResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/cycles/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateCycleInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CycleResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: CycleRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/cycles/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CycleRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_cycleQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<CycleResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/cycles/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_cycleQuerySchema.parse(params);
          const queryKeys = ['cycle_id', 'team_id', 'number'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CycleResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_cyclesQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<CycleListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/cycles/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_cyclesQuerySchema.parse(params);
          const queryKeys = ['team_id', 'include_archived'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CycleListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateCycleInput,
        fetchOptions?: RequestInit,
      ): Promise<CycleResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/cycles/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateCycleInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CycleResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    documents: {
      archive: async (
        params: DocumentRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DocumentRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      create: async (
        params: CreateDocumentInput,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateDocumentInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: DocumentRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DocumentRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_documentQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_documentQuerySchema.parse(params);
          const queryKeys = ['document_id', 'workspace_id', 'slug'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_documentsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<DocumentListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_documentsQuerySchema.parse(params);
          const queryKeys = [
            'workspace_id',
            'project_id',
            'initiative_id',
            'issue_id',
            'query',
            'include_archived',
            'include_trashed',
          ];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      restore: async (
        params: DocumentRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/restore`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DocumentRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      trash: async (
        params: DocumentRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/trash`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DocumentRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unarchive: async (
        params: DocumentRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/unarchive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DocumentRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateDocumentInput,
        fetchOptions?: RequestInit,
      ): Promise<DocumentResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/documents/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateDocumentInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DocumentResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    favorites: {
      add: async (
        params: FavoriteTargetInput,
        fetchOptions?: RequestInit,
      ): Promise<FavoriteResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/favorites/add`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = FavoriteTargetInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = FavoriteResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_favoritesQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<FavoriteListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/favorites/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_favoritesQuerySchema.parse(params);
          const queryKeys = ['workspace_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = FavoriteListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      remove: async (
        params: FavoriteTargetInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/favorites/remove`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = FavoriteTargetInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      reorder: async (
        params: ReorderFavoritesInput,
        fetchOptions?: RequestInit,
      ): Promise<FavoriteListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/favorites/reorder`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ReorderFavoritesInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = FavoriteListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    imports: {
      linear: async (
        params: ImportLinearInput,
        fetchOptions?: RequestInit,
      ): Promise<ImportReportResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/imports/linear`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ImportLinearInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ImportReportResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_importsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ImportListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/imports/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_importsQuerySchema.parse(params);
          const queryKeys = ['workspace_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ImportListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    initiatives: {
      archive: async (
        params: InitiativeRefInput,
        fetchOptions?: RequestInit,
      ): Promise<InitiativeResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = InitiativeRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      create: async (
        params: CreateInitiativeInput,
        fetchOptions?: RequestInit,
      ): Promise<InitiativeResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateInitiativeInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: InitiativeRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = InitiativeRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_initiativeQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<InitiativeResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_initiativeQuerySchema.parse(params);
          const queryKeys = ['initiative_id', 'workspace_id', 'slug'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_initiativesQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<InitiativeListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_initiativesQuerySchema.parse(params);
          const queryKeys = [
            'workspace_id',
            'status',
            'owner_id',
            'include_archived',
            'include_trashed',
          ];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = InitiativeListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      projects: {
        add: async (
          params: InitiativeProjectInput,
          fetchOptions?: RequestInit,
        ): Promise<InitiativeResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/initiatives/projects/add`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = InitiativeProjectInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        remove: async (
          params: InitiativeProjectInput,
          fetchOptions?: RequestInit,
        ): Promise<InitiativeResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/initiatives/projects/remove`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = InitiativeProjectInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      unarchive: async (
        params: InitiativeRefInput,
        fetchOptions?: RequestInit,
      ): Promise<InitiativeResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/unarchive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = InitiativeRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateInitiativeInput,
        fetchOptions?: RequestInit,
      ): Promise<InitiativeResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/initiatives/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateInitiativeInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = InitiativeResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      updates: {
        create: async (
          params: CreateInitiativeUpdateInput,
          fetchOptions?: RequestInit,
        ): Promise<InitiativeUpdateResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/initiatives/updates/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateInitiativeUpdateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = InitiativeUpdateResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: InitiativeUpdateRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/initiatives/updates/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = InitiativeUpdateRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_initiative_updatesQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<InitiativeUpdateListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/initiatives/updates/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_initiative_updatesQuerySchema.parse(params);
            const queryKeys = ['initiative_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } =
              InitiativeUpdateListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
    },
    integrations: {
      github: {
        create: async (
          params: GithubIntegrationRefInput,
          fetchOptions?: RequestInit,
        ): Promise<GithubSecretResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/integrations/github/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = GithubIntegrationRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = GithubSecretResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: GithubIntegrationRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeleteGithubResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/integrations/github/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = GithubIntegrationRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeleteGithubResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        get: async (
          params: z.infer<typeof get_github_integrationQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<GetGithubResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/integrations/github/get`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = get_github_integrationQuerySchema.parse(params);
            const queryKeys = ['workspace_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = GetGithubResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        regenerateSecret: async (
          params: GithubIntegrationRefInput,
          fetchOptions?: RequestInit,
        ): Promise<GithubSecretResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/integrations/github/regenerate-secret`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = GithubIntegrationRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = GithubSecretResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        repoSyncs: {
          add: async (
            params: AddGithubRepoSyncInput,
            fetchOptions?: RequestInit,
          ): Promise<GithubRepoSyncResponseBody> => {
            try {
              if (!apiUrl || typeof apiUrl !== 'string') {
                throw new Error('API URL is undefined or not a string');
              }
              const url = new URL(
                `${apiUrl}/integrations/github/repo-syncs/add`,
                typeof window !== 'undefined' ? window.location.origin : undefined,
              );

              const options: RequestInit = {
                method: 'POST',
                credentials: 'include',
                headers: {
                  'Content-Type': 'application/json',
                },
                ...fetchOptions,
              };
              const validatedBody = AddGithubRepoSyncInputSchema.parse(params);
              options.body = JSON.stringify(validatedBody);
              const response = await fetchFn(url.toString(), options);
              if (!response.ok) {
                throw await parseApiError(response, url.toString());
              }
              const data = await response.json();
              const { data: parsedData, error } = GithubRepoSyncResponseBodySchema.safeParse(data);
              if (error) {
                console.error('Response validation error:', error);
                console.error('Response data:', data);
                throw new Error(error.message);
              }
              return parsedData;
            } catch (error) {
              if (import.meta.env.DEV) {
                console.error('Error in API request:', error);
              }
              throw error;
            }
          },
          backfill: async (
            params: BackfillGithubRepoSyncInput,
            fetchOptions?: RequestInit,
          ): Promise<GithubRepoSyncResponseBody> => {
            try {
              if (!apiUrl || typeof apiUrl !== 'string') {
                throw new Error('API URL is undefined or not a string');
              }
              const url = new URL(
                `${apiUrl}/integrations/github/repo-syncs/backfill`,
                typeof window !== 'undefined' ? window.location.origin : undefined,
              );

              const options: RequestInit = {
                method: 'POST',
                credentials: 'include',
                headers: {
                  'Content-Type': 'application/json',
                },
                ...fetchOptions,
              };
              const validatedBody = BackfillGithubRepoSyncInputSchema.parse(params);
              options.body = JSON.stringify(validatedBody);
              const response = await fetchFn(url.toString(), options);
              if (!response.ok) {
                throw await parseApiError(response, url.toString());
              }
              const data = await response.json();
              const { data: parsedData, error } = GithubRepoSyncResponseBodySchema.safeParse(data);
              if (error) {
                console.error('Response validation error:', error);
                console.error('Response data:', data);
                throw new Error(error.message);
              }
              return parsedData;
            } catch (error) {
              if (import.meta.env.DEV) {
                console.error('Error in API request:', error);
              }
              throw error;
            }
          },
          get: async (
            params: z.infer<typeof get_github_repo_syncQuerySchema>,
            fetchOptions?: RequestInit,
          ): Promise<GithubRepoSyncResponseBody> => {
            try {
              if (!apiUrl || typeof apiUrl !== 'string') {
                throw new Error('API URL is undefined or not a string');
              }
              const url = new URL(
                `${apiUrl}/integrations/github/repo-syncs/get`,
                typeof window !== 'undefined' ? window.location.origin : undefined,
              );
              const validatedQuery = get_github_repo_syncQuerySchema.parse(params);
              const queryKeys = ['repo_sync_id'];
              queryKeys.forEach((key) => {
                const value = validatedQuery[key as keyof typeof validatedQuery];
                if (value !== undefined && value !== null) {
                  url.searchParams.append(key, String(value));
                }
              });
              const options: RequestInit = {
                method: 'GET',
                credentials: 'include',
                headers: {
                  'Content-Type': 'application/json',
                },
                ...fetchOptions,
              };

              const response = await fetchFn(url.toString(), options);
              if (!response.ok) {
                throw await parseApiError(response, url.toString());
              }
              const data = await response.json();
              const { data: parsedData, error } = GithubRepoSyncResponseBodySchema.safeParse(data);
              if (error) {
                console.error('Response validation error:', error);
                console.error('Response data:', data);
                throw new Error(error.message);
              }
              return parsedData;
            } catch (error) {
              if (import.meta.env.DEV) {
                console.error('Error in API request:', error);
              }
              throw error;
            }
          },
          list: async (
            params: z.infer<typeof list_github_repo_syncsQuerySchema>,
            fetchOptions?: RequestInit,
          ): Promise<ListGithubRepoSyncsResponseBody> => {
            try {
              if (!apiUrl || typeof apiUrl !== 'string') {
                throw new Error('API URL is undefined or not a string');
              }
              const url = new URL(
                `${apiUrl}/integrations/github/repo-syncs/list`,
                typeof window !== 'undefined' ? window.location.origin : undefined,
              );
              const validatedQuery = list_github_repo_syncsQuerySchema.parse(params);
              const queryKeys = ['workspace_id'];
              queryKeys.forEach((key) => {
                const value = validatedQuery[key as keyof typeof validatedQuery];
                if (value !== undefined && value !== null) {
                  url.searchParams.append(key, String(value));
                }
              });
              const options: RequestInit = {
                method: 'GET',
                credentials: 'include',
                headers: {
                  'Content-Type': 'application/json',
                },
                ...fetchOptions,
              };

              const response = await fetchFn(url.toString(), options);
              if (!response.ok) {
                throw await parseApiError(response, url.toString());
              }
              const data = await response.json();
              const { data: parsedData, error } =
                ListGithubRepoSyncsResponseBodySchema.safeParse(data);
              if (error) {
                console.error('Response validation error:', error);
                console.error('Response data:', data);
                throw new Error(error.message);
              }
              return parsedData;
            } catch (error) {
              if (import.meta.env.DEV) {
                console.error('Error in API request:', error);
              }
              throw error;
            }
          },
          remove: async (
            params: GithubRepoSyncRefInput,
            fetchOptions?: RequestInit,
          ): Promise<RemoveGithubRepoSyncResponseBody> => {
            try {
              if (!apiUrl || typeof apiUrl !== 'string') {
                throw new Error('API URL is undefined or not a string');
              }
              const url = new URL(
                `${apiUrl}/integrations/github/repo-syncs/remove`,
                typeof window !== 'undefined' ? window.location.origin : undefined,
              );

              const options: RequestInit = {
                method: 'DELETE',
                credentials: 'include',
                headers: {
                  'Content-Type': 'application/json',
                },
                ...fetchOptions,
              };
              const validatedBody = GithubRepoSyncRefInputSchema.parse(params);
              options.body = JSON.stringify(validatedBody);
              const response = await fetchFn(url.toString(), options);
              if (!response.ok) {
                throw await parseApiError(response, url.toString());
              }
              const data = await response.json();
              const { data: parsedData, error } =
                RemoveGithubRepoSyncResponseBodySchema.safeParse(data);
              if (error) {
                console.error('Response validation error:', error);
                console.error('Response data:', data);
                throw new Error(error.message);
              }
              return parsedData;
            } catch (error) {
              if (import.meta.env.DEV) {
                console.error('Error in API request:', error);
              }
              throw error;
            }
          },
        },
        update: async (
          params: UpdateGithubIntegrationInput,
          fetchOptions?: RequestInit,
        ): Promise<GithubIntegrationResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/integrations/github/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateGithubIntegrationInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = GithubIntegrationResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        webhook: async (
          params: z.infer<typeof github_webhookQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<GithubWebhookResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/integrations/github/webhook`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = github_webhookQuerySchema.parse(params);
            const queryKeys = ['workspace_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = GithubWebhookResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
    },
    invitations: {
      accept: async (
        params: AcceptInvitationInput,
        fetchOptions?: RequestInit,
      ): Promise<AcceptInvitationResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/invitations/accept`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = AcceptInvitationInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = AcceptInvitationResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_invitationQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<GetInvitationResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/invitations/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_invitationQuerySchema.parse(params);
          const queryKeys = ['token'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = GetInvitationResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    issues: {
      archive: async (
        params: IssueRefInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = IssueRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      attachments: {
        create: async (
          params: CreateAttachmentInput,
          fetchOptions?: RequestInit,
        ): Promise<AttachmentResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/attachments/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateAttachmentInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = AttachmentResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: AttachmentRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/attachments/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = AttachmentRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateAttachmentInput,
          fetchOptions?: RequestInit,
        ): Promise<AttachmentResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/attachments/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateAttachmentInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = AttachmentResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      bulkUpdate: async (
        params: BulkUpdateIssuesInput,
        fetchOptions?: RequestInit,
      ): Promise<IssuesResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/bulk-update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = BulkUpdateIssuesInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssuesResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      comments: {
        create: async (
          params: CreateCommentInput,
          fetchOptions?: RequestInit,
        ): Promise<CommentResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/comments/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateCommentInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = CommentResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: CommentRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/comments/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CommentRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_commentsQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<CommentListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/comments/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_commentsQuerySchema.parse(params);
            const queryKeys = ['issue_id', 'limit', 'cursor'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = CommentListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        resolve: async (
          params: CommentRefInput,
          fetchOptions?: RequestInit,
        ): Promise<CommentResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/comments/resolve`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CommentRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = CommentResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        unresolve: async (
          params: CommentRefInput,
          fetchOptions?: RequestInit,
        ): Promise<CommentResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/comments/unresolve`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CommentRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = CommentResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateCommentInput,
          fetchOptions?: RequestInit,
        ): Promise<CommentResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/comments/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateCommentInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = CommentResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      create: async (
        params: CreateIssueInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateIssueInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: IssueRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = IssueRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_issueQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<IssueDetailResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_issueQuerySchema.parse(params);
          const queryKeys = ['issue_id', 'identifier', 'workspace_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueDetailResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      history: {
        list: async (
          params: z.infer<typeof list_issue_historyQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<HistoryListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/history/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_issue_historyQuerySchema.parse(params);
            const queryKeys = ['issue_id', 'limit', 'cursor'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = HistoryListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      list: async (
        params: z.infer<typeof list_issuesQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<IssueListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_issuesQuerySchema.parse(params);
          const queryKeys = [
            'workspace_id',
            'team_id',
            'team_ids',
            'state_ids',
            'state_types',
            'assignee_ids',
            'creator_ids',
            'label_ids',
            'project_ids',
            'milestone_ids',
            'cycle_ids',
            'parent_id',
            'has_parent',
            'priorities',
            'due_before',
            'due_after',
            'created_after',
            'updated_after',
            'query',
            'subscriber_id',
            'include_archived',
            'include_trashed',
            'order_by',
            'order',
            'group_by',
            'limit',
            'cursor',
          ];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      reactions: {
        add: async (
          params: ReactionInput,
          fetchOptions?: RequestInit,
        ): Promise<ReactionsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/reactions/add`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = ReactionInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = ReactionsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        remove: async (
          params: ReactionInput,
          fetchOptions?: RequestInit,
        ): Promise<ReactionsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/reactions/remove`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = ReactionInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = ReactionsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      relations: {
        create: async (
          params: CreateRelationInput,
          fetchOptions?: RequestInit,
        ): Promise<RelationResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/relations/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateRelationInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = RelationResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: RelationRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/issues/relations/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = RelationRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      restore: async (
        params: IssueRefInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/restore`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = IssueRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      search: async (
        params: z.infer<typeof search_issuesQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<IssuesResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/search`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = search_issuesQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'query', 'limit'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssuesResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      subscribe: async (
        params: SubscriptionInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueDetailResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/subscribe`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = SubscriptionInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueDetailResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      trash: async (
        params: IssueRefInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/trash`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = IssueRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unarchive: async (
        params: IssueRefInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/unarchive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = IssueRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unsubscribe: async (
        params: SubscriptionInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueDetailResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/unsubscribe`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = SubscriptionInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueDetailResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateIssueInput,
        fetchOptions?: RequestInit,
      ): Promise<IssueResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/issues/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateIssueInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = IssueResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    labels: {
      archive: async (
        params: LabelRefInput,
        fetchOptions?: RequestInit,
      ): Promise<LabelResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/labels/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = LabelRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = LabelResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      create: async (
        params: CreateLabelInput,
        fetchOptions?: RequestInit,
      ): Promise<LabelResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/labels/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateLabelInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = LabelResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: LabelRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/labels/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = LabelRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_labelsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<LabelListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/labels/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_labelsQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'team_id', 'all_teams', 'include_archived'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = LabelListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unarchive: async (
        params: LabelRefInput,
        fetchOptions?: RequestInit,
      ): Promise<LabelResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/labels/unarchive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = LabelRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = LabelResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateLabelInput,
        fetchOptions?: RequestInit,
      ): Promise<LabelResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/labels/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateLabelInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = LabelResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    notifications: {
      archive: async (
        params: NotificationIDsInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = NotificationIDsInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      counts: async (
        params: z.infer<typeof get_notification_countsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<NotificationCountsResultBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/counts`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_notification_countsQuerySchema.parse(params);
          const queryKeys = ['workspace_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = NotificationCountsResultBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_notificationsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<NotificationListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_notificationsQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'filter', 'limit', 'cursor'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = NotificationListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      read: async (
        params: NotificationIDsInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/read`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = NotificationIDsInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      readAll: async (
        params: MarkAllReadInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/read-all`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = MarkAllReadInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      snooze: async (
        params: SnoozeNotificationsInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/snooze`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = SnoozeNotificationsInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unarchive: async (
        params: NotificationIDsInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/unarchive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = NotificationIDsInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unread: async (
        params: NotificationIDsInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/notifications/unread`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = NotificationIDsInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    projects: {
      archive: async (
        params: ProjectRefInput,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/archive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ProjectRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      create: async (
        params: CreateProjectInput,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateProjectInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: ProjectRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ProjectRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_projectQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_projectQuerySchema.parse(params);
          const queryKeys = ['project_id', 'workspace_id', 'slug'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      links: {
        create: async (
          params: CreateProjectLinkInput,
          fetchOptions?: RequestInit,
        ): Promise<LinkResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/links/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateProjectLinkInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = LinkResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: ProjectLinkRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/links/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = ProjectLinkRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateProjectLinkInput,
          fetchOptions?: RequestInit,
        ): Promise<LinkResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/links/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateProjectLinkInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = LinkResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      list: async (
        params: z.infer<typeof list_projectsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ProjectListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_projectsQuerySchema.parse(params);
          const queryKeys = [
            'workspace_id',
            'team_id',
            'status_type',
            'status_id',
            'lead_id',
            'member_id',
            'initiative_id',
            'label_id',
            'query',
            'include_archived',
            'include_trashed',
            'order_by',
          ];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      milestones: {
        create: async (
          params: CreateMilestoneInput,
          fetchOptions?: RequestInit,
        ): Promise<MilestoneResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/milestones/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateMilestoneInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = MilestoneResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: MilestoneRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/milestones/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = MilestoneRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_milestonesQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<MilestoneListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/milestones/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_milestonesQuerySchema.parse(params);
            const queryKeys = ['project_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = MilestoneListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateMilestoneInput,
          fetchOptions?: RequestInit,
        ): Promise<MilestoneResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/milestones/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateMilestoneInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = MilestoneResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      restore: async (
        params: ProjectRefInput,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/restore`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ProjectRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      statuses: {
        create: async (
          params: CreateProjectStatusInput,
          fetchOptions?: RequestInit,
        ): Promise<StatusResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/statuses/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateProjectStatusInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = StatusResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: DeleteProjectStatusInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/statuses/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = DeleteProjectStatusInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_project_statusesQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<StatusListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/statuses/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_project_statusesQuerySchema.parse(params);
            const queryKeys = ['workspace_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = StatusListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateProjectStatusInput,
          fetchOptions?: RequestInit,
        ): Promise<StatusResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/statuses/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateProjectStatusInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = StatusResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      trash: async (
        params: ProjectRefInput,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/trash`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ProjectRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      unarchive: async (
        params: ProjectRefInput,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/unarchive`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ProjectRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateProjectInput,
        fetchOptions?: RequestInit,
      ): Promise<ProjectResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/projects/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateProjectInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ProjectResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      updates: {
        create: async (
          params: CreateProjectUpdateInput,
          fetchOptions?: RequestInit,
        ): Promise<ProjectUpdateResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/updates/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateProjectUpdateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = ProjectUpdateResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: ProjectUpdateRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/updates/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = ProjectUpdateRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_project_updatesQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<ProjectUpdateListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/updates/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_project_updatesQuerySchema.parse(params);
            const queryKeys = ['project_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = ProjectUpdateListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: EditProjectUpdateInput,
          fetchOptions?: RequestInit,
        ): Promise<ProjectUpdateResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/projects/updates/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = EditProjectUpdateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = ProjectUpdateResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
    },
    search: async (
      params: z.infer<typeof searchQuerySchema>,
      fetchOptions?: RequestInit,
    ): Promise<SearchResultBody> => {
      try {
        if (!apiUrl || typeof apiUrl !== 'string') {
          throw new Error('API URL is undefined or not a string');
        }
        const url = new URL(
          `${apiUrl}/search`,
          typeof window !== 'undefined' ? window.location.origin : undefined,
        );
        const validatedQuery = searchQuerySchema.parse(params);
        const queryKeys = ['workspace_id', 'query', 'limit'];
        queryKeys.forEach((key) => {
          const value = validatedQuery[key as keyof typeof validatedQuery];
          if (value !== undefined && value !== null) {
            url.searchParams.append(key, String(value));
          }
        });
        const options: RequestInit = {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          ...fetchOptions,
        };

        const response = await fetchFn(url.toString(), options);
        if (!response.ok) {
          throw await parseApiError(response, url.toString());
        }
        const data = await response.json();
        const { data: parsedData, error } = SearchResultBodySchema.safeParse(data);
        if (error) {
          console.error('Response validation error:', error);
          console.error('Response data:', data);
          throw new Error(error.message);
        }
        return parsedData;
      } catch (error) {
        if (import.meta.env.DEV) {
          console.error('Error in API request:', error);
        }
        throw error;
      }
    },
    setup: {
      createUser: async (
        params: BootstrapInput,
        fetchOptions?: RequestInit,
      ): Promise<CreateUserResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/setup/create-user`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = BootstrapInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = CreateUserResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      status: async (
        params?: undefined,
        fetchOptions?: RequestInit,
      ): Promise<SetupStatusResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/setup/status`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SetupStatusResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    system: {
      email: {
        delete: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<SuccessResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/email`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        get: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<EmailSettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/email`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = EmailSettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        put: async (
          params: EmailSettingsUpdateInput,
          fetchOptions?: RequestInit,
        ): Promise<EmailSettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/email`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = EmailSettingsUpdateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = EmailSettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        test: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<SuccessResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/email/test`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      googleAuth: {
        delete: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<SuccessResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/google-auth`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        get: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<GoogleAuthSettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/google-auth`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } =
              GoogleAuthSettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        put: async (
          params: UpdateGoogleAuthSettingsInput,
          fetchOptions?: RequestInit,
        ): Promise<GoogleAuthSettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/google-auth`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateGoogleAuthSettingsInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } =
              GoogleAuthSettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      settings: {
        get: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<SettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/settings`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        put: async (
          params: UpdateSettingsInput,
          fetchOptions?: RequestInit,
        ): Promise<SettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/system/settings`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateSettingsInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
    },
    teams: {
      create: async (
        params: CreateTeamInput,
        fetchOptions?: RequestInit,
      ): Promise<TeamResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/teams/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateTeamInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TeamResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: DeleteTeamInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/teams/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DeleteTeamInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_teamQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<TeamResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/teams/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_teamQuerySchema.parse(params);
          const queryKeys = ['team_id', 'workspace_id', 'key'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TeamResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_teamsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<TeamListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/teams/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_teamsQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'include_archived'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TeamListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      members: {
        add: async (
          params: AddTeamMemberInput,
          fetchOptions?: RequestInit,
        ): Promise<TeamMemberResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/members/add`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = AddTeamMemberInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = TeamMemberResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_team_membersQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<TeamMemberListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/members/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_team_membersQuerySchema.parse(params);
            const queryKeys = ['team_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = TeamMemberListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        remove: async (
          params: RemoveTeamMemberInput,
          fetchOptions?: RequestInit,
        ): Promise<SuccessResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/members/remove`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = RemoveTeamMemberInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      states: {
        create: async (
          params: CreateStateInput,
          fetchOptions?: RequestInit,
        ): Promise<StateResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/states/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateStateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = StateResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        delete: async (
          params: DeleteStateInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/states/delete`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = DeleteStateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_workflow_statesQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<StateListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/states/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_workflow_statesQuerySchema.parse(params);
            const queryKeys = ['team_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = StateListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateStateInput,
          fetchOptions?: RequestInit,
        ): Promise<StateResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/teams/states/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateStateInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = StateResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      update: async (
        params: UpdateTeamInput,
        fetchOptions?: RequestInit,
      ): Promise<TeamResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/teams/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateTeamInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TeamResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    templates: {
      create: async (
        params: CreateTemplateInput,
        fetchOptions?: RequestInit,
      ): Promise<TemplateResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/templates/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateTemplateInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TemplateResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: DeleteTemplateInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/templates/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DeleteTemplateInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_templateQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<TemplateResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/templates/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_templateQuerySchema.parse(params);
          const queryKeys = ['template_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TemplateResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_templatesQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<TemplateListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/templates/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_templatesQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'team_id', 'kind'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TemplateListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateTemplateInput,
        fetchOptions?: RequestInit,
      ): Promise<TemplateResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/templates/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateTemplateInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = TemplateResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    uploads: {
      create: async (params: FormData, fetchOptions?: RequestInit): Promise<UploadResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/uploads/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',

            ...fetchOptions,
          };
          options.body = params;
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = UploadResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (params: z.infer<typeof get_uploadQuerySchema>, fetchOptions?: RequestInit) => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/uploads/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_uploadQuerySchema.parse(params);
          const queryKeys = ['upload_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          return data;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    users: {
      create: async (
        params: CreateUserInput,
        fetchOptions?: RequestInit,
      ): Promise<UserResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateUserInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = UserResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: DeleteUserInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DeleteUserInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_usersQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ListUsersResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_usersQuerySchema.parse(params);
          const queryKeys = ['query', 'limit'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ListUsersResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      me: async (params: UpdateMeInput, fetchOptions?: RequestInit): Promise<UserResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/me`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateMeInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = UserResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      notificationSettings: {
        get: async (
          params?: undefined,
          fetchOptions?: RequestInit,
        ): Promise<NotificationSettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/users/notification-settings`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } =
              NotificationSettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        put: async (
          params: UpdateNotificationSettingsInput,
          fetchOptions?: RequestInit,
        ): Promise<NotificationSettingsResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/users/notification-settings`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateNotificationSettingsInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } =
              NotificationSettingsResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      password: async (
        params: ChangePasswordInput,
        fetchOptions?: RequestInit,
      ): Promise<SuccessResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/password`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = ChangePasswordInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateUserInput,
        fetchOptions?: RequestInit,
      ): Promise<UserResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateUserInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = UserResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      whoami: async (
        params?: undefined,
        fetchOptions?: RequestInit,
      ): Promise<WhoamiResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/users/whoami`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WhoamiResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    views: {
      create: async (
        params: CreateViewInput,
        fetchOptions?: RequestInit,
      ): Promise<ViewResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/views/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateViewInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ViewResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: DeleteViewInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/views/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DeleteViewInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_viewQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ViewResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/views/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_viewQuerySchema.parse(params);
          const queryKeys = ['view_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ViewResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_viewsQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<ViewListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/views/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_viewsQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'team_id', 'kind'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ViewListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateViewInput,
        fetchOptions?: RequestInit,
      ): Promise<ViewResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/views/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateViewInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = ViewResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    webhooks: {
      create: async (
        params: CreateWebhookInput,
        fetchOptions?: RequestInit,
      ): Promise<WebhookCreatedResultBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/webhooks/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateWebhookInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WebhookCreatedResultBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: WebhookRefInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/webhooks/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = WebhookRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      list: async (
        params: z.infer<typeof list_webhooksQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<WebhookListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/webhooks/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = list_webhooksQuerySchema.parse(params);
          const queryKeys = ['workspace_id'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WebhookListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      test: async (
        params: WebhookRefInput,
        fetchOptions?: RequestInit,
      ): Promise<WebhookTestResultBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/webhooks/test`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = WebhookRefInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WebhookTestResultBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      update: async (
        params: UpdateWebhookInput,
        fetchOptions?: RequestInit,
      ): Promise<WebhookResultBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/webhooks/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateWebhookInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WebhookResultBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
    workspaces: {
      create: async (
        params: CreateWorkspaceInput,
        fetchOptions?: RequestInit,
      ): Promise<WorkspaceResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/workspaces/create`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = CreateWorkspaceInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WorkspaceResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      delete: async (
        params: DeleteWorkspaceInput,
        fetchOptions?: RequestInit,
      ): Promise<DeletedResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/workspaces/delete`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'DELETE',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = DeleteWorkspaceInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      get: async (
        params: z.infer<typeof get_workspaceQuerySchema>,
        fetchOptions?: RequestInit,
      ): Promise<WorkspaceResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/workspaces/get`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
          const validatedQuery = get_workspaceQuerySchema.parse(params);
          const queryKeys = ['workspace_id', 'slug'];
          queryKeys.forEach((key) => {
            const value = validatedQuery[key as keyof typeof validatedQuery];
            if (value !== undefined && value !== null) {
              url.searchParams.append(key, String(value));
            }
          });
          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WorkspaceResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      invitations: {
        create: async (
          params: CreateInvitationInput,
          fetchOptions?: RequestInit,
        ): Promise<InvitationResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/invitations/create`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = CreateInvitationInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = InvitationResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_invitationsQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<InvitationListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/invitations/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_invitationsQuerySchema.parse(params);
            const queryKeys = ['workspace_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = InvitationListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        resend: async (
          params: InvitationRefInput,
          fetchOptions?: RequestInit,
        ): Promise<InvitationResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/invitations/resend`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = InvitationRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = InvitationResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        revoke: async (
          params: InvitationRefInput,
          fetchOptions?: RequestInit,
        ): Promise<DeletedResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/invitations/revoke`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = InvitationRefInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = DeletedResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      list: async (
        params?: undefined,
        fetchOptions?: RequestInit,
      ): Promise<WorkspaceListResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/workspaces/list`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'GET',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };

          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WorkspaceListResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
      members: {
        add: async (
          params: AddMemberInput,
          fetchOptions?: RequestInit,
        ): Promise<AddMemberResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/members/add`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = AddMemberInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = AddMemberResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        list: async (
          params: z.infer<typeof list_workspace_membersQuerySchema>,
          fetchOptions?: RequestInit,
        ): Promise<WorkspaceMemberListResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/members/list`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );
            const validatedQuery = list_workspace_membersQuerySchema.parse(params);
            const queryKeys = ['workspace_id'];
            queryKeys.forEach((key) => {
              const value = validatedQuery[key as keyof typeof validatedQuery];
              if (value !== undefined && value !== null) {
                url.searchParams.append(key, String(value));
              }
            });
            const options: RequestInit = {
              method: 'GET',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };

            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } =
              WorkspaceMemberListResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        remove: async (
          params: RemoveMemberInput,
          fetchOptions?: RequestInit,
        ): Promise<SuccessResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/members/remove`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'DELETE',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = RemoveMemberInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = SuccessResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
        update: async (
          params: UpdateMemberInput,
          fetchOptions?: RequestInit,
        ): Promise<WorkspaceMemberResponseBody> => {
          try {
            if (!apiUrl || typeof apiUrl !== 'string') {
              throw new Error('API URL is undefined or not a string');
            }
            const url = new URL(
              `${apiUrl}/workspaces/members/update`,
              typeof window !== 'undefined' ? window.location.origin : undefined,
            );

            const options: RequestInit = {
              method: 'PUT',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
              },
              ...fetchOptions,
            };
            const validatedBody = UpdateMemberInputSchema.parse(params);
            options.body = JSON.stringify(validatedBody);
            const response = await fetchFn(url.toString(), options);
            if (!response.ok) {
              throw await parseApiError(response, url.toString());
            }
            const data = await response.json();
            const { data: parsedData, error } = WorkspaceMemberResponseBodySchema.safeParse(data);
            if (error) {
              console.error('Response validation error:', error);
              console.error('Response data:', data);
              throw new Error(error.message);
            }
            return parsedData;
          } catch (error) {
            if (import.meta.env.DEV) {
              console.error('Error in API request:', error);
            }
            throw error;
          }
        },
      },
      update: async (
        params: UpdateWorkspaceInput,
        fetchOptions?: RequestInit,
      ): Promise<WorkspaceResponseBody> => {
        try {
          if (!apiUrl || typeof apiUrl !== 'string') {
            throw new Error('API URL is undefined or not a string');
          }
          const url = new URL(
            `${apiUrl}/workspaces/update`,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );

          const options: RequestInit = {
            method: 'PUT',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            ...fetchOptions,
          };
          const validatedBody = UpdateWorkspaceInputSchema.parse(params);
          options.body = JSON.stringify(validatedBody);
          const response = await fetchFn(url.toString(), options);
          if (!response.ok) {
            throw await parseApiError(response, url.toString());
          }
          const data = await response.json();
          const { data: parsedData, error } = WorkspaceResponseBodySchema.safeParse(data);
          if (error) {
            console.error('Response validation error:', error);
            console.error('Response data:', data);
            throw new Error(error.message);
          }
          return parsedData;
        } catch (error) {
          if (import.meta.env.DEV) {
            console.error('Error in API request:', error);
          }
          throw error;
        }
      },
    },
  };
}
