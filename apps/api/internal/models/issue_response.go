package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
)

const DateFormat = "2006-01-02"

func FormatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(DateFormat)
	return &formatted
}

var priorityLabels = []string{"No priority", "Urgent", "High", "Medium", "Low"}

func PriorityLabel(priority int) string {
	if priority < 0 || priority >= len(priorityLabels) {
		return priorityLabels[0]
	}
	return priorityLabels[priority]
}

func Identifier(teamKey string, number int) string {
	return teamKey + "-" + strconv.Itoa(number)
}

func IssueURL(uiURL, workspaceSlug, identifier string) string {
	return strings.TrimRight(uiURL, "/") + "/" + workspaceSlug + "/issue/" + identifier
}

func GitBranchName(identifier, title, prefix string) string {
	slug := utils.TruncateSlug(utils.Slugify(title), 50)
	branch := strings.ToLower(identifier)
	if slug != "" {
		branch += "-" + slug
	}
	if prefix == "" {
		return branch
	}
	return prefix + "/" + branch
}

type IssueMilestoneSummary struct {
	ID   uuid.UUID `json:"id" format:"uuid"`
	Name string    `json:"name"`
}

type IssueCycleSummary struct {
	ID     uuid.UUID `json:"id" format:"uuid"`
	Number int       `json:"number"`
	Name   string    `json:"name"`
}

type IssueSummary struct {
	ID         uuid.UUID              `json:"id" format:"uuid"`
	Identifier string                 `json:"identifier" doc:"Team key and number, like ENG-123"`
	Title      string                 `json:"title"`
	State      *WorkflowStateResponse `json:"state,omitempty" required:"false"`
}

type RelationResponse struct {
	RelationID uuid.UUID                `json:"relation_id" format:"uuid"`
	Type       schema.IssueRelationType `json:"type"`
	Issue      *IssueSummary            `json:"issue" doc:"The other issue of the relation"`
}

type IssueRelations struct {
	Blocks      []*RelationResponse `json:"blocks" nullable:"false" doc:"Issues this issue blocks"`
	BlockedBy   []*RelationResponse `json:"blocked_by" nullable:"false" doc:"Issues blocking this issue"`
	Related     []*RelationResponse `json:"related" nullable:"false"`
	DuplicateOf []*RelationResponse `json:"duplicate_of" nullable:"false" doc:"The canonical issue this one duplicates"`
	Duplicates  []*RelationResponse `json:"duplicates" nullable:"false" doc:"Issues marked as duplicates of this one"`
}

type IssueCounts struct {
	Children          int
	CompletedChildren int
	Comments          int
	Attachments       int
}

type IssueView struct {
	UIURL    string
	CallerID uuid.UUID
	Counts   IssueCounts
}

type IssueResponse struct {
	ID                  uuid.UUID              `json:"id" format:"uuid"`
	Identifier          string                 `json:"identifier" doc:"Team key and number, like ENG-123"`
	Number              int                    `json:"number"`
	WorkspaceID         uuid.UUID              `json:"workspace_id" format:"uuid"`
	TeamID              uuid.UUID              `json:"team_id" format:"uuid"`
	TeamKey             string                 `json:"team_key"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description" doc:"Markdown"`
	Priority            int                    `json:"priority" doc:"0 none, 1 urgent, 2 high, 3 medium, 4 low"`
	PriorityLabel       string                 `json:"priority_label"`
	Estimate            *float64               `json:"estimate" nullable:"true"`
	State               *WorkflowStateResponse `json:"state"`
	Assignee            *UserSummary           `json:"assignee,omitempty" required:"false"`
	Creator             *UserSummary           `json:"creator,omitempty" required:"false"`
	Labels              []*LabelSummary        `json:"labels" nullable:"false"`
	Project             *ProjectSummary        `json:"project,omitempty" required:"false"`
	Milestone           *IssueMilestoneSummary `json:"milestone,omitempty" required:"false"`
	Cycle               *IssueCycleSummary     `json:"cycle,omitempty" required:"false"`
	Parent              *IssueSummary          `json:"parent,omitempty" required:"false"`
	ChildCount          int                    `json:"child_count"`
	CompletedChildCount int                    `json:"completed_child_count"`
	CommentCount        int                    `json:"comment_count"`
	AttachmentCount     int                    `json:"attachment_count"`
	SortOrder           float64                `json:"sort_order"`
	SubIssueSortOrder   float64                `json:"sub_issue_sort_order"`
	DueDate             *string                `json:"due_date" nullable:"true" format:"date"`
	StartedAt           *time.Time             `json:"started_at" nullable:"true"`
	CompletedAt         *time.Time             `json:"completed_at" nullable:"true"`
	CanceledAt          *time.Time             `json:"canceled_at" nullable:"true"`
	TriagedAt           *time.Time             `json:"triaged_at" nullable:"true"`
	SnoozedUntil        *time.Time             `json:"snoozed_until" nullable:"true"`
	ArchivedAt          *time.Time             `json:"archived_at" nullable:"true"`
	TrashedAt           *time.Time             `json:"trashed_at" nullable:"true"`
	GitBranchName       string                 `json:"git_branch_name"`
	URL                 string                 `json:"url"`
	ImportURL           string                 `json:"import_url"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type IssueDetailResponse struct {
	IssueResponse
	Children    []*IssueResponse         `json:"children" nullable:"false" doc:"Sub-issues"`
	Relations   *IssueRelations          `json:"relations"`
	Attachments []*AttachmentResponse    `json:"attachments" nullable:"false"`
	Subscribers []*UserSummary           `json:"subscribers" nullable:"false"`
	Reactions   []*ReactionGroup         `json:"reactions" nullable:"false"`
	GithubIssue *GithubIssueLinkResponse `json:"github_issue,omitempty" required:"false" doc:"The GitHub issue this issue syncs with, when its repository is synced"`
}

func issueIdentifier(entity *ent.Issue) string {
	if entity.Edges.Team == nil {
		return ""
	}
	return Identifier(entity.Edges.Team.Key, entity.Number)
}

func TransformIssueSummary(entity *ent.Issue) *IssueSummary {
	if entity == nil {
		return nil
	}
	return &IssueSummary{
		ID:         entity.ID,
		Identifier: issueIdentifier(entity),
		Title:      entity.Title,
		State:      TransformWorkflowState(entity.Edges.State),
	}
}

func transformIssueMilestone(entity *ent.ProjectMilestone) *IssueMilestoneSummary {
	if entity == nil {
		return nil
	}
	return &IssueMilestoneSummary{ID: entity.ID, Name: entity.Name}
}

func transformIssueCycle(entity *ent.Cycle) *IssueCycleSummary {
	if entity == nil {
		return nil
	}
	return &IssueCycleSummary{ID: entity.ID, Number: entity.Number, Name: entity.Name}
}

func branchPrefix(entity *ent.Issue, callerID uuid.UUID) string {
	assignee := entity.Edges.Assignee
	if assignee == nil || assignee.ID != callerID {
		return ""
	}
	return assignee.DisplayName
}

func TransformIssueEntity(entity *ent.Issue, view IssueView) *IssueResponse {
	identifier := issueIdentifier(entity)
	teamKey := ""
	if entity.Edges.Team != nil {
		teamKey = entity.Edges.Team.Key
	}
	workspaceSlug := ""
	if entity.Edges.Workspace != nil {
		workspaceSlug = entity.Edges.Workspace.Slug
	}
	return &IssueResponse{
		ID:                  entity.ID,
		Identifier:          identifier,
		Number:              entity.Number,
		WorkspaceID:         entity.WorkspaceID,
		TeamID:              entity.TeamID,
		TeamKey:             teamKey,
		Title:               entity.Title,
		Description:         entity.Description,
		Priority:            entity.Priority,
		PriorityLabel:       PriorityLabel(entity.Priority),
		Estimate:            entity.Estimate,
		State:               TransformWorkflowState(entity.Edges.State),
		Assignee:            TransformUserSummary(entity.Edges.Assignee),
		Creator:             TransformUserSummary(entity.Edges.Creator),
		Labels:              TransformLabelSummaries(entity.Edges.Labels),
		Project:             TransformProjectSummary(entity.Edges.Project),
		Milestone:           transformIssueMilestone(entity.Edges.Milestone),
		Cycle:               transformIssueCycle(entity.Edges.Cycle),
		Parent:              TransformIssueSummary(entity.Edges.Parent),
		ChildCount:          view.Counts.Children,
		CompletedChildCount: view.Counts.CompletedChildren,
		CommentCount:        view.Counts.Comments,
		AttachmentCount:     view.Counts.Attachments,
		SortOrder:           entity.SortOrder,
		SubIssueSortOrder:   entity.SubIssueSortOrder,
		DueDate:             FormatDate(entity.DueDate),
		StartedAt:           entity.StartedAt,
		CompletedAt:         entity.CompletedAt,
		CanceledAt:          entity.CanceledAt,
		TriagedAt:           entity.TriagedAt,
		SnoozedUntil:        entity.SnoozedUntil,
		ArchivedAt:          entity.ArchivedAt,
		TrashedAt:           entity.TrashedAt,
		GitBranchName:       GitBranchName(identifier, entity.Title, branchPrefix(entity, view.CallerID)),
		URL:                 IssueURL(view.UIURL, workspaceSlug, identifier),
		ImportURL:           entity.ImportURL,
		CreatedAt:           entity.CreatedAt,
		UpdatedAt:           entity.UpdatedAt,
	}
}

func TransformRelations(issueID uuid.UUID, outgoing, incoming []*ent.IssueRelation) *IssueRelations {
	relations := &IssueRelations{
		Blocks:      []*RelationResponse{},
		BlockedBy:   []*RelationResponse{},
		Related:     []*RelationResponse{},
		DuplicateOf: []*RelationResponse{},
		Duplicates:  []*RelationResponse{},
	}
	for _, relation := range outgoing {
		entry := &RelationResponse{RelationID: relation.ID, Type: relation.Type, Issue: TransformIssueSummary(relation.Edges.RelatedIssue)}
		switch relation.Type {
		case schema.IssueRelationBlocks:
			relations.Blocks = append(relations.Blocks, entry)
		case schema.IssueRelationRelated:
			relations.Related = append(relations.Related, entry)
		case schema.IssueRelationDuplicate:
			relations.DuplicateOf = append(relations.DuplicateOf, entry)
		}
	}
	for _, relation := range incoming {
		entry := &RelationResponse{RelationID: relation.ID, Type: relation.Type, Issue: TransformIssueSummary(relation.Edges.Issue)}
		switch relation.Type {
		case schema.IssueRelationBlocks:
			relations.BlockedBy = append(relations.BlockedBy, entry)
		case schema.IssueRelationRelated:
			relations.Related = append(relations.Related, entry)
		case schema.IssueRelationDuplicate:
			relations.Duplicates = append(relations.Duplicates, entry)
		}
	}
	return relations
}

func TransformRelation(relation *ent.IssueRelation, other *ent.Issue) *RelationResponse {
	return &RelationResponse{RelationID: relation.ID, Type: relation.Type, Issue: TransformIssueSummary(other)}
}
