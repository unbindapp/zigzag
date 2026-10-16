package schema

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// https://github.com/danielgtaylor/huma/issues/621
func enumSchema(r huma.Registry, name string, values []string) *huma.Schema {
	if r.Map()[name] == nil {
		schemaRef := r.Schema(reflect.TypeOf(""), true, name)
		schemaRef.Title = name
		for _, v := range values {
			schemaRef.Enum = append(schemaRef.Enum, v)
		}
		r.Map()[name] = schemaRef
	}
	return &huma.Schema{Ref: "#/components/schemas/" + name}
}

func stringValues[T ~string](all []T) []string {
	values := make([]string, len(all))
	for i, v := range all {
		values[i] = string(v)
	}
	return values
}

type WorkspaceRole string

const (
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

var allWorkspaceRoles = []WorkspaceRole{WorkspaceRoleAdmin, WorkspaceRoleMember, WorkspaceRoleViewer}

func (WorkspaceRole) Values() []string { return stringValues(allWorkspaceRoles) }

func (WorkspaceRole) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "WorkspaceRole", stringValues(allWorkspaceRoles))
}

type KeyPrivilege string

var allPrivileges = []KeyPrivilege{}

func AllPrivileges() []KeyPrivilege { return allPrivileges }

func (KeyPrivilege) Values() []string { return stringValues(allPrivileges) }

func (KeyPrivilege) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "KeyPrivilege", stringValues(allPrivileges))
}

type OAuthClientKind string

const (
	OAuthClientKindDynamic          OAuthClientKind = "dynamic"
	OAuthClientKindMetadataDocument OAuthClientKind = "metadata_document"
)

var allOAuthClientKinds = []OAuthClientKind{OAuthClientKindDynamic, OAuthClientKindMetadataDocument}

func (OAuthClientKind) Values() []string { return stringValues(allOAuthClientKinds) }

func (OAuthClientKind) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "OAuthClientKind", stringValues(allOAuthClientKinds))
}

type OAuthTokenKind string

const (
	OAuthTokenKindAccess  OAuthTokenKind = "access"
	OAuthTokenKindRefresh OAuthTokenKind = "refresh"
)

func (OAuthTokenKind) Values() []string {
	return []string{string(OAuthTokenKindAccess), string(OAuthTokenKindRefresh)}
}

type StateType string

const (
	StateTypeTriage    StateType = "triage"
	StateTypeBacklog   StateType = "backlog"
	StateTypeUnstarted StateType = "unstarted"
	StateTypeStarted   StateType = "started"
	StateTypeCompleted StateType = "completed"
	StateTypeCanceled  StateType = "canceled"
)

var allStateTypes = []StateType{StateTypeTriage, StateTypeBacklog, StateTypeUnstarted, StateTypeStarted, StateTypeCompleted, StateTypeCanceled}

func AllStateTypes() []StateType { return allStateTypes }

func (StateType) Values() []string { return stringValues(allStateTypes) }

func (StateType) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "StateType", stringValues(allStateTypes))
}

type EstimateType string

const (
	EstimateTypeNone        EstimateType = "none"
	EstimateTypeExponential EstimateType = "exponential"
	EstimateTypeFibonacci   EstimateType = "fibonacci"
	EstimateTypeLinear      EstimateType = "linear"
	EstimateTypeTShirt      EstimateType = "tshirt"
)

var allEstimateTypes = []EstimateType{EstimateTypeNone, EstimateTypeExponential, EstimateTypeFibonacci, EstimateTypeLinear, EstimateTypeTShirt}

func (EstimateType) Values() []string { return stringValues(allEstimateTypes) }

func (EstimateType) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "EstimateType", stringValues(allEstimateTypes))
}

type ProjectStatusType string

const (
	ProjectStatusBacklog   ProjectStatusType = "backlog"
	ProjectStatusPlanned   ProjectStatusType = "planned"
	ProjectStatusStarted   ProjectStatusType = "started"
	ProjectStatusPaused    ProjectStatusType = "paused"
	ProjectStatusCompleted ProjectStatusType = "completed"
	ProjectStatusCanceled  ProjectStatusType = "canceled"
)

var allProjectStatusTypes = []ProjectStatusType{ProjectStatusBacklog, ProjectStatusPlanned, ProjectStatusStarted, ProjectStatusPaused, ProjectStatusCompleted, ProjectStatusCanceled}

func (ProjectStatusType) Values() []string { return stringValues(allProjectStatusTypes) }

func (ProjectStatusType) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "ProjectStatusType", stringValues(allProjectStatusTypes))
}

type DateResolution string

const (
	DateResolutionDay      DateResolution = "day"
	DateResolutionMonth    DateResolution = "month"
	DateResolutionQuarter  DateResolution = "quarter"
	DateResolutionHalfYear DateResolution = "half_year"
	DateResolutionYear     DateResolution = "year"
)

var allDateResolutions = []DateResolution{DateResolutionDay, DateResolutionMonth, DateResolutionQuarter, DateResolutionHalfYear, DateResolutionYear}

func (DateResolution) Values() []string { return stringValues(allDateResolutions) }

func (DateResolution) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "DateResolution", stringValues(allDateResolutions))
}

type Health string

const (
	HealthOnTrack  Health = "on_track"
	HealthAtRisk   Health = "at_risk"
	HealthOffTrack Health = "off_track"
)

var allHealths = []Health{HealthOnTrack, HealthAtRisk, HealthOffTrack}

func (Health) Values() []string { return stringValues(allHealths) }

func (Health) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "Health", stringValues(allHealths))
}

type InitiativeStatus string

const (
	InitiativeStatusPlanned   InitiativeStatus = "planned"
	InitiativeStatusActive    InitiativeStatus = "active"
	InitiativeStatusCompleted InitiativeStatus = "completed"
)

var allInitiativeStatuses = []InitiativeStatus{InitiativeStatusPlanned, InitiativeStatusActive, InitiativeStatusCompleted}

func (InitiativeStatus) Values() []string { return stringValues(allInitiativeStatuses) }

func (InitiativeStatus) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "InitiativeStatus", stringValues(allInitiativeStatuses))
}

type IssueRelationType string

const (
	IssueRelationBlocks    IssueRelationType = "blocks"
	IssueRelationRelated   IssueRelationType = "related"
	IssueRelationDuplicate IssueRelationType = "duplicate"
)

var allIssueRelationTypes = []IssueRelationType{IssueRelationBlocks, IssueRelationRelated, IssueRelationDuplicate}

func (IssueRelationType) Values() []string { return stringValues(allIssueRelationTypes) }

func (IssueRelationType) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "IssueRelationType", stringValues(allIssueRelationTypes))
}

type HistorySource string

const (
	HistorySourceUser   HistorySource = "user"
	HistorySourceAPI    HistorySource = "api"
	HistorySourceMCP    HistorySource = "mcp"
	HistorySourceImport HistorySource = "import"
	HistorySourceSystem HistorySource = "system"
)

var allHistorySources = []HistorySource{HistorySourceUser, HistorySourceAPI, HistorySourceMCP, HistorySourceImport, HistorySourceSystem}

func (HistorySource) Values() []string { return stringValues(allHistorySources) }

func (HistorySource) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "HistorySource", stringValues(allHistorySources))
}

type NotificationType string

const (
	NotificationIssueAssigned        NotificationType = "issue_assigned"
	NotificationIssueUnassigned      NotificationType = "issue_unassigned"
	NotificationIssueMentioned       NotificationType = "issue_mentioned"
	NotificationCommentMentioned     NotificationType = "comment_mentioned"
	NotificationIssueCommented       NotificationType = "issue_commented"
	NotificationIssueStateChanged    NotificationType = "issue_state_changed"
	NotificationIssuePriorityChanged NotificationType = "issue_priority_changed"
	NotificationIssueDueSoon         NotificationType = "issue_due_soon"
	NotificationIssueBlocking        NotificationType = "issue_blocking"
	NotificationProjectUpdate        NotificationType = "project_update"
	NotificationProjectMember        NotificationType = "project_added_as_member"
	NotificationProjectLead          NotificationType = "project_lead"
	NotificationReaction             NotificationType = "reaction"
)

var allNotificationTypes = []NotificationType{
	NotificationIssueAssigned, NotificationIssueUnassigned, NotificationIssueMentioned, NotificationCommentMentioned,
	NotificationIssueCommented, NotificationIssueStateChanged, NotificationIssuePriorityChanged, NotificationIssueDueSoon,
	NotificationIssueBlocking, NotificationProjectUpdate, NotificationProjectMember, NotificationProjectLead, NotificationReaction,
}

func (NotificationType) Values() []string { return stringValues(allNotificationTypes) }

func (NotificationType) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "NotificationType", stringValues(allNotificationTypes))
}

type ViewKind string

const (
	ViewKindIssue   ViewKind = "issue"
	ViewKindProject ViewKind = "project"
)

var allViewKinds = []ViewKind{ViewKindIssue, ViewKindProject}

func (ViewKind) Values() []string { return stringValues(allViewKinds) }

func (ViewKind) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "ViewKind", stringValues(allViewKinds))
}

type TemplateKind string

const (
	TemplateKindIssue    TemplateKind = "issue"
	TemplateKindProject  TemplateKind = "project"
	TemplateKindDocument TemplateKind = "document"
)

var allTemplateKinds = []TemplateKind{TemplateKindIssue, TemplateKindProject, TemplateKindDocument}

func (TemplateKind) Values() []string { return stringValues(allTemplateKinds) }

func (TemplateKind) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "TemplateKind", stringValues(allTemplateKinds))
}

type FavoriteKind string

const (
	FavoriteKindIssue      FavoriteKind = "issue"
	FavoriteKindProject    FavoriteKind = "project"
	FavoriteKindCycle      FavoriteKind = "cycle"
	FavoriteKindView       FavoriteKind = "view"
	FavoriteKindDocument   FavoriteKind = "document"
	FavoriteKindLabel      FavoriteKind = "label"
	FavoriteKindTeam       FavoriteKind = "team"
	FavoriteKindInitiative FavoriteKind = "initiative"
)

var allFavoriteKinds = []FavoriteKind{FavoriteKindIssue, FavoriteKindProject, FavoriteKindCycle, FavoriteKindView, FavoriteKindDocument, FavoriteKindLabel, FavoriteKindTeam, FavoriteKindInitiative}

func (FavoriteKind) Values() []string { return stringValues(allFavoriteKinds) }

func (FavoriteKind) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "FavoriteKind", stringValues(allFavoriteKinds))
}

type ImportSource string

const ImportSourceLinear ImportSource = "linear"

var allImportSources = []ImportSource{ImportSourceLinear}

func (ImportSource) Values() []string { return stringValues(allImportSources) }

func (ImportSource) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "ImportSource", stringValues(allImportSources))
}

type ImportStatus string

const (
	ImportStatusRunning   ImportStatus = "running"
	ImportStatusCompleted ImportStatus = "completed"
	ImportStatusFailed    ImportStatus = "failed"
)

var allImportStatuses = []ImportStatus{ImportStatusRunning, ImportStatusCompleted, ImportStatusFailed}

func (ImportStatus) Values() []string { return stringValues(allImportStatuses) }

func (ImportStatus) Schema(r huma.Registry) *huma.Schema {
	return enumSchema(r, "ImportStatus", stringValues(allImportStatuses))
}
