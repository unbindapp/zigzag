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
    icon: z.string(),
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

export const ChangePasswordInputSchema = z
  .object({
    current_password: z.string(), // Your current password
    new_password: z.string(), // The new password, at least 8 characters
  })
  .strip();

export const ConnectedAppApproveInputSchema = z
  .object({
    client_id: z.string(),
    code_challenge: z.string(),
    full_access: z.boolean(), // Reach every workspace you can, capped at role. workspace_ids must be empty.
    privileges: z.array(KeyPrivilegeSchema).optional(), // What the app may see beyond its role. None exist yet.
    redirect_uri: z.string(),
    resource: z.string().optional(),
    role: WorkspaceRoleSchema, // Strongest role the app acts with. It never exceeds your own role in a workspace.
    scope: z.string().optional(),
    state: z.string().optional(),
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
    client_id: z.string(),
    redirect_uri: z.string(),
    state: z.string().optional(),
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

export const CreateInvitationInputSchema = z
  .object({
    email: z.string(), // Email to invite
    role: WorkspaceRoleSchema, // admin, member or viewer
    team_ids: z.array(z.string()).nullable().optional(), // Teams of this workspace joined on accept
    workspace_id: z.string(), // The workspace
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

export const CreateWorkspaceInputSchema = z
  .object({
    color: z.string().optional(), // Hex color such as #5e6ad2
    icon: z.string().optional(), // Emoji or icon name
    name: z.string(), // Workspace name
    slug: z.string(), // URL key, lowercase letters, digits and dashes, unique across the instance
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

export const DeleteUserInputSchema = z
  .object({
    user_id: z.string(), // The user to delete. Their issues and comments stay, without an author.
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

export const GetClientResponseBodySchema = z
  .object({
    data: ConnectedAppClientResponseSchema,
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

export const ListAPIKeysResponseBodySchema = z
  .object({
    data: z.array(APIKeyResponseSchema),
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

export const MemberListResponseBodySchema = z
  .object({
    data: z.array(WorkspaceMemberResponseSchema),
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

export const TeamMemberResponseSchema = z
  .object({
    joined_at: z.string().datetime({ offset: true }),
    owner: z.boolean(),
    user: UserSummarySchema,
  })
  .strip();

export const MemberListResponseBody1Schema = z
  .object({
    data: z.array(TeamMemberResponseSchema),
  })
  .strip();

export const MemberResponseBodySchema = z
  .object({
    data: WorkspaceMemberResponseSchema,
  })
  .strip();

export const MemberResponseBody1Schema = z
  .object({
    data: TeamMemberResponseSchema,
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

export const RedirectResponseBodySchema = z
  .object({
    data: ConnectedAppRedirectResponseSchema,
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

export const TeamResponseBodySchema = z
  .object({
    data: TeamResponseSchema,
  })
  .strip();

export const UpdateMeInputSchema = z
  .object({
    avatar_upload_id: z.string().optional(), // An image you uploaded. Send 00000000-0000-0000-0000-000000000000 to remove the avatar.
    display_name: z.string().optional(), // Short handle used in mentions, lowercase and unique
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
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    key: z.string().optional(), // New issue prefix; existing identifiers change with it
    name: z.string().optional(), // New name
    private: z.boolean().optional(), // Visible to members and workspace admins only
    team_id: z.string(), // The team to change
    timezone: z.string().optional(), // IANA time zone
    triage_enabled: z.boolean().optional(), // New issues from outside the team land in triage
  })
  .strip();

export const UpdateUserInputSchema = z
  .object({
    active: z.boolean().optional(), // Inactive users cannot sign in
    display_name: z.string().optional(), // Short handle used in mentions
    name: z.string().optional(), // Full name
    password: z.string().optional(), // Set a new password; also activates placeholder accounts when active is sent
    system_admin: z.boolean().optional(), // Grant or revoke instance administration
    user_id: z.string(), // The user to change
  })
  .strip();

export const UpdateWorkspaceInputSchema = z
  .object({
    color: z.string().optional(), // Hex color, empty to clear
    icon: z.string().optional(), // Emoji or icon name, empty to clear
    name: z.string().optional(), // New name
    slug: z.string().optional(), // New URL key; old links stop working
    workspace_id: z.string(), // The workspace to change
  })
  .strip();

export const UploadResponseSchema = z
  .object({
    content_type: z.string(),
    file_name: z.string(),
    height: z.number().nullable().optional(), // Pixel height for png, jpeg, gif and webp images
    id: z.string(),
    size_bytes: z.number(),
    url: z.string(), // Same-origin path that serves the file; use it in markdown
    width: z.number().nullable().optional(), // Pixel width for png, jpeg, gif and webp images
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
export type AddMemberInput = z.infer<typeof AddMemberInputSchema>;
export type InvitationResponse = z.infer<typeof InvitationResponseSchema>;
export type MemberUser = z.infer<typeof MemberUserSchema>;
export type WorkspaceMemberResponse = z.infer<typeof WorkspaceMemberResponseSchema>;
export type AddWorkspaceMemberResponse = z.infer<typeof AddWorkspaceMemberResponseSchema>;
export type AddMemberResponseBody = z.infer<typeof AddMemberResponseBodySchema>;
export type AddTeamMemberInput = z.infer<typeof AddTeamMemberInputSchema>;
export type BootstrapInput = z.infer<typeof BootstrapInputSchema>;
export type ChangePasswordInput = z.infer<typeof ChangePasswordInputSchema>;
export type ConnectedAppApproveInput = z.infer<typeof ConnectedAppApproveInputSchema>;
export type OAuthClientKind = z.infer<typeof OAuthClientKindSchema>;
export type ConnectedAppClientResponse = z.infer<typeof ConnectedAppClientResponseSchema>;
export type ConnectedAppDenyInput = z.infer<typeof ConnectedAppDenyInputSchema>;
export type ConnectedAppRedirectResponse = z.infer<typeof ConnectedAppRedirectResponseSchema>;
export type ConnectedAppResponse = z.infer<typeof ConnectedAppResponseSchema>;
export type ConnectedAppRevokeInput = z.infer<typeof ConnectedAppRevokeInputSchema>;
export type Cookie = z.infer<typeof CookieSchema>;
export type CreateAPIKeyResponseBody = z.infer<typeof CreateAPIKeyResponseBodySchema>;
export type CreateInvitationInput = z.infer<typeof CreateInvitationInputSchema>;
export type StateType = z.infer<typeof StateTypeSchema>;
export type CreateStateInput = z.infer<typeof CreateStateInputSchema>;
export type CreateTeamInput = z.infer<typeof CreateTeamInputSchema>;
export type CreateUserInput = z.infer<typeof CreateUserInputSchema>;
export type CreateUserResponseBody = z.infer<typeof CreateUserResponseBodySchema>;
export type CreateWorkspaceInput = z.infer<typeof CreateWorkspaceInputSchema>;
export type DeletedResponse = z.infer<typeof DeletedResponseSchema>;
export type DeleteAPIKeyResponseBody = z.infer<typeof DeleteAPIKeyResponseBodySchema>;
export type DeleteStateInput = z.infer<typeof DeleteStateInputSchema>;
export type DeleteTeamInput = z.infer<typeof DeleteTeamInputSchema>;
export type DeleteUserInput = z.infer<typeof DeleteUserInputSchema>;
export type DeleteWorkspaceInput = z.infer<typeof DeleteWorkspaceInputSchema>;
export type DeletedResponseBody = z.infer<typeof DeletedResponseBodySchema>;
export type EmailSettings = z.infer<typeof EmailSettingsSchema>;
export type EmailSettingsResponse = z.infer<typeof EmailSettingsResponseSchema>;
export type EmailSettingsResponseBody = z.infer<typeof EmailSettingsResponseBodySchema>;
export type EmailSettingsUpdateInput = z.infer<typeof EmailSettingsUpdateInputSchema>;
export type EstimateType = z.infer<typeof EstimateTypeSchema>;
export type GetClientResponseBody = z.infer<typeof GetClientResponseBodySchema>;
export type InvitationPreviewResponse = z.infer<typeof InvitationPreviewResponseSchema>;
export type GetInvitationResponseBody = z.infer<typeof GetInvitationResponseBodySchema>;
export type InvitationListResponseBody = z.infer<typeof InvitationListResponseBodySchema>;
export type InvitationRefInput = z.infer<typeof InvitationRefInputSchema>;
export type InvitationResponseBody = z.infer<typeof InvitationResponseBodySchema>;
export type ListAPIKeysResponseBody = z.infer<typeof ListAPIKeysResponseBodySchema>;
export type ListResponseBody = z.infer<typeof ListResponseBodySchema>;
export type ListUsersResponseBody = z.infer<typeof ListUsersResponseBodySchema>;
export type LoginInputBody = z.infer<typeof LoginInputBodySchema>;
export type MemberListResponseBody = z.infer<typeof MemberListResponseBodySchema>;
export type UserSummary = z.infer<typeof UserSummarySchema>;
export type TeamMemberResponse = z.infer<typeof TeamMemberResponseSchema>;
export type MemberListResponseBody1 = z.infer<typeof MemberListResponseBody1Schema>;
export type MemberResponseBody = z.infer<typeof MemberResponseBodySchema>;
export type MemberResponseBody1 = z.infer<typeof MemberResponseBody1Schema>;
export type NotificationSettingsResponse = z.infer<typeof NotificationSettingsResponseSchema>;
export type NotificationSettingsResponseBody = z.infer<
  typeof NotificationSettingsResponseBodySchema
>;
export type SuccessResponse = z.infer<typeof SuccessResponseSchema>;
export type PlainSuccessResponseBody = z.infer<typeof PlainSuccessResponseBodySchema>;
export type RedirectResponseBody = z.infer<typeof RedirectResponseBodySchema>;
export type RemoveMemberInput = z.infer<typeof RemoveMemberInputSchema>;
export type RemoveTeamMemberInput = z.infer<typeof RemoveTeamMemberInputSchema>;
export type RequestPasswordResetInputBody = z.infer<typeof RequestPasswordResetInputBodySchema>;
export type ResetPasswordInput = z.infer<typeof ResetPasswordInputSchema>;
export type ResponseError = z.infer<typeof ResponseErrorSchema>;
export type RevokeResponseBody = z.infer<typeof RevokeResponseBodySchema>;
export type SessionUser = z.infer<typeof SessionUserSchema>;
export type SessionResponseBody = z.infer<typeof SessionResponseBodySchema>;
export type SystemSettingsResponse = z.infer<typeof SystemSettingsResponseSchema>;
export type SettingsResponseBody = z.infer<typeof SettingsResponseBodySchema>;
export type SetupStatus = z.infer<typeof SetupStatusSchema>;
export type SetupStatusResponseBody = z.infer<typeof SetupStatusResponseBodySchema>;
export type WorkflowStateResponse = z.infer<typeof WorkflowStateResponseSchema>;
export type StateListResponseBody = z.infer<typeof StateListResponseBodySchema>;
export type StateResponseBody = z.infer<typeof StateResponseBodySchema>;
export type SuccessResponseBody = z.infer<typeof SuccessResponseBodySchema>;
export type TeamResponse = z.infer<typeof TeamResponseSchema>;
export type TeamListResponseBody = z.infer<typeof TeamListResponseBodySchema>;
export type TeamResponseBody = z.infer<typeof TeamResponseBodySchema>;
export type UpdateMeInput = z.infer<typeof UpdateMeInputSchema>;
export type UpdateMemberInput = z.infer<typeof UpdateMemberInputSchema>;
export type UpdateNotificationSettingsInput = z.infer<typeof UpdateNotificationSettingsInputSchema>;
export type UpdateSettingsInput = z.infer<typeof UpdateSettingsInputSchema>;
export type UpdateStateInput = z.infer<typeof UpdateStateInputSchema>;
export type UpdateTeamInput = z.infer<typeof UpdateTeamInputSchema>;
export type UpdateUserInput = z.infer<typeof UpdateUserInputSchema>;
export type UpdateWorkspaceInput = z.infer<typeof UpdateWorkspaceInputSchema>;
export type UploadResponse = z.infer<typeof UploadResponseSchema>;
export type UploadResponseBody = z.infer<typeof UploadResponseBodySchema>;
export type UserResponseBody = z.infer<typeof UserResponseBodySchema>;
export type WhoamiAPIKey = z.infer<typeof WhoamiAPIKeySchema>;
export type WorkspaceMembership = z.infer<typeof WorkspaceMembershipSchema>;
export type WhoamiResponse = z.infer<typeof WhoamiResponseSchema>;
export type WhoamiResponseBody = z.infer<typeof WhoamiResponseBodySchema>;
export type WorkspaceListResponseBody = z.infer<typeof WorkspaceListResponseBodySchema>;
export type WorkspaceResponseBody = z.infer<typeof WorkspaceResponseBodySchema>;

export const list_api_keysQuerySchema = z
  .object({
    user_id: z.string().optional(), // List another user's keys. Requires system admin. Defaults to your own.
  })
  .passthrough();

export const get_connected_app_clientQuerySchema = z
  .object({
    client_id: z.string(), // OAuth client_id from the authorize request
    redirect_uri: z.string(), // redirect_uri from the authorize request
  })
  .passthrough();

export const get_invitationQuerySchema = z
  .object({
    token: z.string(), // Token from the invitation link
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
        ): Promise<MemberResponseBody1> => {
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
            const { data: parsedData, error } = MemberResponseBody1Schema.safeParse(data);
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
        ): Promise<MemberListResponseBody1> => {
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
            const { data: parsedData, error } = MemberListResponseBody1Schema.safeParse(data);
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
        ): Promise<MemberListResponseBody> => {
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
            const { data: parsedData, error } = MemberListResponseBodySchema.safeParse(data);
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
        ): Promise<MemberResponseBody> => {
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
            const { data: parsedData, error } = MemberResponseBodySchema.safeParse(data);
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
