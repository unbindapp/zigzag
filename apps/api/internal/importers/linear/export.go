package linear

import "time"

type LinearExport struct {
	_              struct{}              `additionalProperties:"true"`
	ExportedAt     string                `json:"exported_at,omitempty" required:"false" doc:"When the export was taken, RFC 3339"`
	Workspace      LinearWorkspace       `json:"workspace" required:"false" doc:"The Linear workspace the export came from; informational, the target is workspace_id"`
	Users          []LinearUser          `json:"users" required:"false" doc:"Every Linear user the export references"`
	Teams          []LinearTeam          `json:"teams" required:"false" doc:"Linear teams; each maps to the Zigzag team with the same key"`
	WorkflowStates []LinearWorkflowState `json:"workflow_states" required:"false" doc:"Workflow states of the exported teams"`
	Labels         []LinearLabel         `json:"labels" required:"false" doc:"Issue labels and label groups; teamId null means a workspace label"`
	Projects       []LinearProject       `json:"projects" required:"false" doc:"Projects"`
	Milestones     []LinearMilestone     `json:"milestones" required:"false" doc:"Project milestones"`
	Cycles         []LinearCycle         `json:"cycles" required:"false" doc:"Team cycles"`
	Issues         []LinearIssue         `json:"issues" required:"false" doc:"Issues with their comments, attachments, relations and state history"`
}

type LinearWorkspace struct {
	_    struct{} `additionalProperties:"true"`
	ID   string   `json:"id,omitempty" required:"false" doc:"Linear workspace id"`
	Name string   `json:"name,omitempty" required:"false" doc:"Linear workspace name"`
	URL  string   `json:"url,omitempty" required:"false" doc:"Linear workspace url"`
}

type LinearUser struct {
	_           struct{} `additionalProperties:"true"`
	ID          string   `json:"id" required:"true" doc:"Linear user id"`
	Name        *string  `json:"name,omitempty" required:"false" nullable:"true" doc:"Full name"`
	Email       string   `json:"email" required:"true" doc:"Email address; matched case insensitively to Zigzag accounts"`
	DisplayName *string  `json:"displayName,omitempty" required:"false" nullable:"true" doc:"Short handle"`
	AvatarURL   *string  `json:"avatarUrl,omitempty" required:"false" nullable:"true" doc:"Avatar url, not imported"`
	IsAdmin     bool     `json:"isAdmin,omitempty" required:"false" doc:"Admins join the workspace as admin, everyone else as member"`
	IsGuest     bool     `json:"isGuest,omitempty" required:"false" doc:"Guest account in Linear"`
	IsActive    *bool    `json:"isActive,omitempty" required:"false" nullable:"true" doc:"Whether the Linear account is active"`
}

type LinearTeam struct {
	_             struct{}   `additionalProperties:"true"`
	ID            string     `json:"id" required:"true" doc:"Linear team id"`
	Key           string     `json:"key" required:"true" doc:"Team key such as ENG; prefixes identifiers"`
	Name          string     `json:"name" required:"true" doc:"Team name"`
	Description   *string    `json:"description,omitempty" required:"false" nullable:"true" doc:"Team description"`
	Icon          *string    `json:"icon,omitempty" required:"false" nullable:"true" doc:"Icon name"`
	Color         *string    `json:"color,omitempty" required:"false" nullable:"true" doc:"Hex color"`
	CreatedAt     *time.Time `json:"createdAt,omitempty" required:"false" nullable:"true" doc:"Creation time"`
	CyclesEnabled *bool      `json:"cyclesEnabled,omitempty" required:"false" nullable:"true" doc:"Whether the team uses cycles"`
	Timezone      *string    `json:"timezone,omitempty" required:"false" nullable:"true" doc:"IANA time zone"`
}

type LinearWorkflowState struct {
	_        struct{} `additionalProperties:"true"`
	ID       string   `json:"id" required:"true" doc:"Linear state id"`
	TeamID   string   `json:"teamId" required:"true" doc:"Linear team id"`
	Name     string   `json:"name" required:"true" doc:"State name"`
	Type     string   `json:"type" required:"true" doc:"triage, backlog, unstarted, started, completed, canceled or duplicate (imported as canceled)"`
	Color    *string  `json:"color,omitempty" required:"false" nullable:"true" doc:"Hex color"`
	Position float64  `json:"position,omitempty" required:"false" doc:"Order within the type"`
}

type LinearLabel struct {
	_           struct{}   `additionalProperties:"true"`
	ID          string     `json:"id" required:"true" doc:"Linear label id"`
	Name        string     `json:"name" required:"true" doc:"Label name"`
	Color       *string    `json:"color,omitempty" required:"false" nullable:"true" doc:"Hex color"`
	Description *string    `json:"description,omitempty" required:"false" nullable:"true" doc:"Label description"`
	ParentID    *string    `json:"parentId,omitempty" required:"false" nullable:"true" doc:"Label group this label belongs to"`
	IsGroup     bool       `json:"isGroup,omitempty" required:"false" doc:"Whether this is a label group"`
	GroupType   *string    `json:"groupType,omitempty" required:"false" nullable:"true" doc:"singleSelect when only one child may be applied to an issue"`
	TeamID      *string    `json:"teamId,omitempty" required:"false" nullable:"true" doc:"Linear team id; null for a workspace label"`
	ArchivedAt  *time.Time `json:"archivedAt,omitempty" required:"false" nullable:"true" doc:"When the label was archived"`
}

type LinearProjectStatus struct {
	_    struct{} `additionalProperties:"true"`
	ID   string   `json:"id,omitempty" required:"false" doc:"Linear project status id"`
	Name string   `json:"name,omitempty" required:"false" doc:"Status name"`
	Type string   `json:"type,omitempty" required:"false" doc:"backlog, planned, started, paused, completed or canceled"`
}

type LinearProject struct {
	_           struct{}            `additionalProperties:"true"`
	ID          string              `json:"id" required:"true" doc:"Linear project id"`
	Identifier  *string             `json:"identifier,omitempty" required:"false" nullable:"true" doc:"Linear project identifier"`
	Name        string              `json:"name" required:"true" doc:"Project name"`
	Summary     *string             `json:"summary,omitempty" required:"false" nullable:"true" doc:"One line summary"`
	Description *string             `json:"description,omitempty" required:"false" nullable:"true" doc:"Markdown description"`
	URL         *string             `json:"url,omitempty" required:"false" nullable:"true" doc:"Link to the Linear project"`
	Status      LinearProjectStatus `json:"status" required:"false" doc:"Project status; matched to a Zigzag status by type, then name"`
	LeadID      *string             `json:"leadId,omitempty" required:"false" nullable:"true" doc:"Linear user id of the lead"`
	Priority    int                 `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	StartDate   *string             `json:"startDate,omitempty" required:"false" nullable:"true" doc:"Start date, YYYY-MM-DD"`
	TargetDate  *string             `json:"targetDate,omitempty" required:"false" nullable:"true" doc:"Target date, YYYY-MM-DD"`
	SortOrder   *float64            `json:"sortOrder,omitempty" required:"false" nullable:"true" doc:"Manual order"`
	CreatedAt   *time.Time          `json:"createdAt,omitempty" required:"false" nullable:"true" doc:"Creation time"`
	UpdatedAt   *time.Time          `json:"updatedAt,omitempty" required:"false" nullable:"true" doc:"Last update time"`
	StartedAt   *time.Time          `json:"startedAt,omitempty" required:"false" nullable:"true" doc:"When work started"`
	CompletedAt *time.Time          `json:"completedAt,omitempty" required:"false" nullable:"true" doc:"When the project completed"`
	CanceledAt  *time.Time          `json:"canceledAt,omitempty" required:"false" nullable:"true" doc:"When the project was canceled"`
	ArchivedAt  *time.Time          `json:"archivedAt,omitempty" required:"false" nullable:"true" doc:"When the project was archived"`
	Trashed     *bool               `json:"trashed,omitempty" required:"false" nullable:"true" doc:"Whether the project is in the trash"`
	TeamIDs     []string            `json:"teamIds,omitempty" required:"false" doc:"Linear team ids"`
	MemberIDs   []string            `json:"memberIds,omitempty" required:"false" doc:"Linear user ids of the members"`
	LabelIDs    []string            `json:"labelIds,omitempty" required:"false" doc:"Linear label ids; only workspace labels apply to projects"`
}

type LinearMilestone struct {
	_           struct{} `additionalProperties:"true"`
	ID          string   `json:"id" required:"true" doc:"Linear milestone id"`
	ProjectID   string   `json:"projectId" required:"true" doc:"Linear project id"`
	Name        string   `json:"name" required:"true" doc:"Milestone name"`
	Description *string  `json:"description,omitempty" required:"false" nullable:"true" doc:"Milestone description"`
	TargetDate  *string  `json:"targetDate,omitempty" required:"false" nullable:"true" doc:"Target date, YYYY-MM-DD"`
	SortOrder   *float64 `json:"sortOrder,omitempty" required:"false" nullable:"true" doc:"Order within the project"`
}

type LinearCycle struct {
	_           struct{}   `additionalProperties:"true"`
	ID          string     `json:"id" required:"true" doc:"Linear cycle id"`
	TeamID      string     `json:"teamId" required:"true" doc:"Linear team id"`
	Number      int        `json:"number" required:"true" doc:"Cycle number within the team"`
	Name        *string    `json:"name,omitempty" required:"false" nullable:"true" doc:"Cycle name"`
	Description *string    `json:"description,omitempty" required:"false" nullable:"true" doc:"Cycle description"`
	StartsAt    *time.Time `json:"startsAt,omitempty" required:"false" nullable:"true" doc:"Start time"`
	EndsAt      *time.Time `json:"endsAt,omitempty" required:"false" nullable:"true" doc:"End time"`
	CompletedAt *time.Time `json:"completedAt,omitempty" required:"false" nullable:"true" doc:"When the cycle was completed"`
}

type LinearIssue struct {
	_             struct{}           `additionalProperties:"true"`
	ID            string             `json:"id" required:"true" doc:"Linear issue id; a second import of the same id updates the issue"`
	Identifier    string             `json:"identifier,omitempty" required:"false" doc:"Identifier such as ENG-12"`
	Number        int                `json:"number,omitempty" required:"false" doc:"Number within the team; kept unless another issue holds it"`
	TeamID        string             `json:"teamId" required:"true" doc:"Linear team id"`
	Title         string             `json:"title" required:"true" doc:"Issue title"`
	Description   *string            `json:"description,omitempty" required:"false" nullable:"true" doc:"Markdown description"`
	Priority      int                `json:"priority,omitempty" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	Estimate      *float64           `json:"estimate,omitempty" required:"false" nullable:"true" doc:"Estimate"`
	URL           *string            `json:"url,omitempty" required:"false" nullable:"true" doc:"Link to the Linear issue, kept as import_url"`
	GitBranchName *string            `json:"gitBranchName,omitempty" required:"false" nullable:"true" doc:"Linear branch name, not imported; Zigzag computes its own"`
	CreatedAt     *time.Time         `json:"createdAt,omitempty" required:"false" nullable:"true" doc:"Creation time, preserved"`
	UpdatedAt     *time.Time         `json:"updatedAt,omitempty" required:"false" nullable:"true" doc:"Last update time, preserved"`
	ArchivedAt    *time.Time         `json:"archivedAt,omitempty" required:"false" nullable:"true" doc:"When the issue was archived; kept unless restore_archived"`
	CompletedAt   *time.Time         `json:"completedAt,omitempty" required:"false" nullable:"true" doc:"When the issue was completed"`
	StartedAt     *time.Time         `json:"startedAt,omitempty" required:"false" nullable:"true" doc:"When work started"`
	CanceledAt    *time.Time         `json:"canceledAt,omitempty" required:"false" nullable:"true" doc:"When the issue was canceled"`
	TriagedAt     *time.Time         `json:"triagedAt,omitempty" required:"false" nullable:"true" doc:"When the issue left triage"`
	Trashed       *bool              `json:"trashed,omitempty" required:"false" nullable:"true" doc:"Whether the issue is in the trash; only GraphQL exports carry it"`
	DueDate       *string            `json:"dueDate,omitempty" required:"false" nullable:"true" doc:"Due date, YYYY-MM-DD"`
	StateID       *string            `json:"stateId,omitempty" required:"false" nullable:"true" doc:"Linear state id"`
	StateName     *string            `json:"stateName,omitempty" required:"false" nullable:"true" doc:"State name, used when stateId is unknown"`
	StateType     *string            `json:"stateType,omitempty" required:"false" nullable:"true" doc:"State type, used with stateName"`
	LabelIDs      []string           `json:"labelIds,omitempty" required:"false" doc:"Linear label ids"`
	LabelNames    []string           `json:"labelNames,omitempty" required:"false" doc:"Label names, used for ids the export does not list"`
	CreatorID     *string            `json:"creatorId,omitempty" required:"false" nullable:"true" doc:"Linear user id of the creator"`
	AssigneeID    *string            `json:"assigneeId,omitempty" required:"false" nullable:"true" doc:"Linear user id of the assignee"`
	ProjectID     *string            `json:"projectId,omitempty" required:"false" nullable:"true" doc:"Linear project id"`
	MilestoneID   *string            `json:"milestoneId,omitempty" required:"false" nullable:"true" doc:"Linear milestone id"`
	ParentID      *string            `json:"parentId,omitempty" required:"false" nullable:"true" doc:"Linear id of the parent issue"`
	CycleID       *string            `json:"cycleId,omitempty" required:"false" nullable:"true" doc:"Linear cycle id"`
	SortOrder     *float64           `json:"sortOrder,omitempty" required:"false" nullable:"true" doc:"Manual order"`
	Attachments   []LinearAttachment `json:"attachments,omitempty" required:"false" doc:"Links attached to the issue"`
	Relations     LinearRelations    `json:"relations" required:"false" doc:"Relations to other issues by Linear id"`
	StateHistory  []LinearStateSpan  `json:"stateHistory,omitempty" required:"false" doc:"Each state the issue was in, oldest first; recorded as history"`
	Comments      []LinearComment    `json:"comments,omitempty" required:"false" doc:"Comments, oldest first"`
}

type LinearAttachment struct {
	_         struct{}   `additionalProperties:"true"`
	ID        string     `json:"id,omitempty" required:"false" doc:"Linear attachment id"`
	Title     *string    `json:"title,omitempty" required:"false" nullable:"true" doc:"Title"`
	Subtitle  *string    `json:"subtitle,omitempty" required:"false" nullable:"true" doc:"Subtitle"`
	URL       string     `json:"url" required:"true" doc:"Link target"`
	CreatedAt *time.Time `json:"createdAt,omitempty" required:"false" nullable:"true" doc:"Creation time"`
}

type LinearRelations struct {
	_           struct{} `additionalProperties:"true"`
	Blocks      []string `json:"blocks,omitempty" required:"false" doc:"Linear ids of issues this issue blocks"`
	BlockedBy   []string `json:"blockedBy,omitempty" required:"false" doc:"Linear ids of issues blocking this issue"`
	RelatedTo   []string `json:"relatedTo,omitempty" required:"false" doc:"Linear ids of related issues"`
	DuplicateOf *string  `json:"duplicateOf,omitempty" required:"false" nullable:"true" doc:"Linear id of the issue this one duplicates"`
}

type LinearStateSpan struct {
	_         struct{}   `additionalProperties:"true"`
	StateID   *string    `json:"stateId,omitempty" required:"false" nullable:"true" doc:"Linear state id"`
	StateName *string    `json:"stateName,omitempty" required:"false" nullable:"true" doc:"State name"`
	StateType *string    `json:"stateType,omitempty" required:"false" nullable:"true" doc:"State type"`
	StartedAt *time.Time `json:"startedAt,omitempty" required:"false" nullable:"true" doc:"When the issue entered the state"`
	EndedAt   *time.Time `json:"endedAt,omitempty" required:"false" nullable:"true" doc:"When the issue left the state; null for the current one"`
}

type LinearComment struct {
	_          struct{}   `additionalProperties:"true"`
	ID         string     `json:"id" required:"true" doc:"Linear comment id; a second import of the same id updates the comment"`
	Body       string     `json:"body" required:"false" doc:"Markdown body"`
	AuthorID   *string    `json:"authorId,omitempty" required:"false" nullable:"true" doc:"Linear user id of the author"`
	AuthorName *string    `json:"authorName,omitempty" required:"false" nullable:"true" doc:"Author name"`
	CreatedAt  *time.Time `json:"createdAt,omitempty" required:"false" nullable:"true" doc:"Creation time, preserved"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty" required:"false" nullable:"true" doc:"Last edit time"`
	ParentID   *string    `json:"parentId,omitempty" required:"false" nullable:"true" doc:"Linear id of the comment this one replies to"`
	QuotedText *string    `json:"quotedText,omitempty" required:"false" nullable:"true" doc:"Text of the issue the comment quotes, not imported"`
}
