package schema

import "github.com/google/uuid"

type EmailSettings struct {
	Host        string `json:"host" required:"true" minLength:"1" doc:"SMTP host, for example smtp.resend.com"`
	Port        int    `json:"port" required:"true" minimum:"1" maximum:"65535" doc:"SMTP port, usually 465 for tls or 587 for starttls"`
	Username    string `json:"username" required:"false" doc:"SMTP username, resend for Resend"`
	FromAddress string `json:"from_address" required:"true" format:"email" doc:"Address mail is sent from"`
	FromName    string `json:"from_name" required:"false" doc:"Display name mail is sent from"`
	Encryption  string `json:"encryption" required:"true" enum:"starttls,tls,none" doc:"starttls upgrades the connection (usually port 587), tls connects encrypted (usually port 465), none sends in the clear"`
}

type FieldChange struct {
	Field     string `json:"field" doc:"Changed field: title, description, state, priority, estimate, assignee, project, milestone, cycle, parent, due_date, labels_added, labels_removed, relation_added, relation_removed, attachment_added, attachment_removed, archived, unarchived, trashed, restored, team, created, imported"`
	From      any    `json:"from,omitempty" required:"false" doc:"Previous value, usually an id or a scalar"`
	To        any    `json:"to,omitempty" required:"false" doc:"New value, usually an id or a scalar"`
	FromLabel string `json:"from_label,omitempty" required:"false" doc:"Human readable previous value"`
	ToLabel   string `json:"to_label,omitempty" required:"false" doc:"Human readable new value"`
}

type ViewDisplay struct {
	Layout        string `json:"layout" enum:"list,board" doc:"list or board"`
	GroupBy       string `json:"group_by,omitempty" required:"false" doc:"Grouping, same values as list-issues group_by"`
	OrderBy       string `json:"order_by,omitempty" required:"false" doc:"Ordering, same values as list-issues order_by"`
	ShowSubIssues bool   `json:"show_sub_issues" doc:"Show sub-issues as rows"`
	ShowCompleted bool   `json:"show_completed" doc:"Show completed and canceled issues"`
}

type TemplateSubIssue struct {
	Title       string `json:"title" doc:"Sub-issue title"`
	Description string `json:"description,omitempty" required:"false" doc:"Sub-issue description in markdown"`
}

type TemplateData struct {
	Title       string             `json:"title,omitempty" required:"false" doc:"Prefilled title"`
	Description string             `json:"description,omitempty" required:"false" doc:"Prefilled description in markdown"`
	Priority    *int               `json:"priority,omitempty" nullable:"true" required:"false" minimum:"0" maximum:"4" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	Estimate    *float64           `json:"estimate,omitempty" nullable:"true" required:"false" doc:"Prefilled estimate"`
	LabelIDs    []uuid.UUID        `json:"label_ids,omitempty" required:"false" doc:"Labels to apply"`
	AssigneeID  *uuid.UUID         `json:"assignee_id,omitempty" nullable:"true" required:"false" format:"uuid" doc:"Assignee"`
	StateID     *uuid.UUID         `json:"state_id,omitempty" nullable:"true" required:"false" format:"uuid" doc:"Workflow state"`
	ProjectID   *uuid.UUID         `json:"project_id,omitempty" nullable:"true" required:"false" format:"uuid" doc:"Project"`
	SubIssues   []TemplateSubIssue `json:"sub_issues,omitempty" required:"false" doc:"Sub-issues created with the issue"`
}

type ImportEntityCount struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

type ImportWarning struct {
	Entity     string `json:"entity"`
	ExternalID string `json:"external_id"`
	Message    string `json:"message"`
}

type ImportUserMapping struct {
	LinearID string     `json:"linear_id"`
	Email    string     `json:"email"`
	UserID   *uuid.UUID `json:"user_id" nullable:"true" format:"uuid"`
	Action   string     `json:"action" enum:"matched,created,skipped"`
}

type ImportReport struct {
	DryRun      bool                         `json:"dry_run"`
	ImportID    uuid.UUID                    `json:"import_id" format:"uuid"`
	Counts      map[string]ImportEntityCount `json:"counts"`
	Warnings    []ImportWarning              `json:"warnings" nullable:"false"`
	UserMapping []ImportUserMapping          `json:"user_mapping" nullable:"false"`
}
