package issue_repo

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/attachment"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/issuehistory"
	"github.com/unbindapp/zigzag-api/ent/issuerelation"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

const (
	importBatchSize      = 200
	AttachmentTypeImport = "import"
)

type ImportedIssue struct {
	ID                uuid.UUID
	Existing          bool
	ExternalID        string
	URL               string
	TeamID            uuid.UUID
	Number            int
	Title             string
	Description       string
	DescriptionText   string
	Priority          int
	Estimate          *float64
	StateID           uuid.UUID
	AssigneeID        *uuid.UUID
	CreatorID         *uuid.UUID
	ProjectID         *uuid.UUID
	MilestoneID       *uuid.UUID
	CycleID           *uuid.UUID
	DueDate           *time.Time
	SortOrder         float64
	SubIssueSortOrder float64
	StartedAt         *time.Time
	CompletedAt       *time.Time
	CanceledAt        *time.Time
	TriagedAt         *time.Time
	ArchivedAt        *time.Time
	TrashedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LabelIDs          []uuid.UUID
	SubscriberIDs     []uuid.UUID
}

type ImportedParent struct {
	IssueID  uuid.UUID
	ParentID *uuid.UUID
}

type ImportedAttachment struct {
	ID        uuid.UUID
	Existing  bool
	IssueID   uuid.UUID
	Title     string
	Subtitle  string
	URL       string
	SortOrder float64
	CreatedAt time.Time
}

type ImportedRelation struct {
	IssueID        uuid.UUID
	RelatedIssueID uuid.UUID
	Type           schema.IssueRelationType
}

type ImportedHistory struct {
	IssueID   uuid.UUID
	CreatedAt time.Time
	Changes   []schema.FieldChange
}

// IssueKeys loads only the identity columns of every issue in the workspace.
func (self *IssueRepository) IssueKeys(ctx context.Context, workspaceID uuid.UUID) ([]*ent.Issue, error) {
	return self.base.DB.Issue.Query().
		Where(issue.WorkspaceID(workspaceID)).
		Select(issue.FieldID, issue.FieldTeamID, issue.FieldNumber, issue.FieldParentID, issue.FieldImportSource, issue.FieldImportExternalID).
		All(ctx)
}

func (self *IssueRepository) ImportedAttachments(ctx context.Context, workspaceID uuid.UUID, source string) ([]*ent.Attachment, error) {
	return self.base.DB.Attachment.Query().
		Where(attachment.SourceType(AttachmentTypeImport), attachment.HasIssueWith(issue.WorkspaceID(workspaceID), issue.ImportSource(source))).
		All(ctx)
}

func (self *IssueRepository) ImportedRelations(ctx context.Context, workspaceID uuid.UUID, source string) ([]*ent.IssueRelation, error) {
	return self.base.DB.IssueRelation.Query().
		Where(issuerelation.Or(
			issuerelation.HasIssueWith(issue.WorkspaceID(workspaceID), issue.ImportSource(source)),
			issuerelation.HasRelatedIssueWith(issue.WorkspaceID(workspaceID), issue.ImportSource(source)),
		)).
		All(ctx)
}

func (self *IssueRepository) IssueIDsWithImportHistory(ctx context.Context, workspaceID uuid.UUID, source string) ([]uuid.UUID, error) {
	return self.base.DB.Issue.Query().
		Where(issue.WorkspaceID(workspaceID), issue.ImportSource(source), issue.HasHistoryWith(issuehistory.SourceEQ(schema.HistorySourceImport))).
		IDs(ctx)
}

func (self *IssueRepository) ImportIssues(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, source string, issues []ImportedIssue) error {
	db := self.base.Client(tx)
	var builders []*ent.IssueCreate
	for _, i := range issues {
		if i.Existing {
			if err := updateImportedIssue(ctx, db, i); err != nil {
				return err
			}
			continue
		}
		builders = append(builders, db.Issue.Create().
			SetID(i.ID).
			SetWorkspaceID(workspaceID).
			SetTeamID(i.TeamID).
			SetNumber(i.Number).
			SetTitle(i.Title).
			SetDescription(i.Description).
			SetDescriptionText(i.DescriptionText).
			SetPriority(i.Priority).
			SetNillableEstimate(i.Estimate).
			SetStateID(i.StateID).
			SetNillableAssigneeID(i.AssigneeID).
			SetNillableCreatorID(i.CreatorID).
			SetNillableProjectID(i.ProjectID).
			SetNillableMilestoneID(i.MilestoneID).
			SetNillableCycleID(i.CycleID).
			SetNillableDueDate(i.DueDate).
			SetSortOrder(i.SortOrder).
			SetSubIssueSortOrder(i.SubIssueSortOrder).
			SetNillableStartedAt(i.StartedAt).
			SetNillableCompletedAt(i.CompletedAt).
			SetNillableCanceledAt(i.CanceledAt).
			SetNillableTriagedAt(i.TriagedAt).
			SetNillableArchivedAt(i.ArchivedAt).
			SetNillableTrashedAt(i.TrashedAt).
			SetImportSource(source).
			SetImportExternalID(i.ExternalID).
			SetImportURL(i.URL).
			SetCreatedAt(i.CreatedAt).
			SetUpdatedAt(i.UpdatedAt).
			AddLabelIDs(i.LabelIDs...).
			AddSubscriberIDs(i.SubscriberIDs...))
	}
	for batch := range slices.Chunk(builders, importBatchSize) {
		if err := db.Issue.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func updateImportedIssue(ctx context.Context, db *ent.Client, i ImportedIssue) error {
	update := db.Issue.UpdateOneID(i.ID).
		SetTitle(i.Title).
		SetDescription(i.Description).
		SetDescriptionText(i.DescriptionText).
		SetPriority(i.Priority).
		SetStateID(i.StateID).
		SetImportURL(i.URL).
		SetUpdatedAt(i.UpdatedAt).
		ClearLabels().
		AddLabelIDs(i.LabelIDs...)
	setOrClear(i.Estimate, update.SetEstimate, update.ClearEstimate)
	setOrClear(i.AssigneeID, update.SetAssigneeID, update.ClearAssigneeID)
	setOrClear(i.CreatorID, update.SetCreatorID, update.ClearCreatorID)
	setOrClear(i.ProjectID, update.SetProjectID, update.ClearProjectID)
	setOrClear(i.MilestoneID, update.SetMilestoneID, update.ClearMilestoneID)
	setOrClear(i.CycleID, update.SetCycleID, update.ClearCycleID)
	setOrClear(i.DueDate, update.SetDueDate, update.ClearDueDate)
	setOrClear(i.StartedAt, update.SetStartedAt, update.ClearStartedAt)
	setOrClear(i.CompletedAt, update.SetCompletedAt, update.ClearCompletedAt)
	setOrClear(i.CanceledAt, update.SetCanceledAt, update.ClearCanceledAt)
	setOrClear(i.TriagedAt, update.SetTriagedAt, update.ClearTriagedAt)
	setOrClear(i.ArchivedAt, update.SetArchivedAt, update.ClearArchivedAt)
	setOrClear(i.TrashedAt, update.SetTrashedAt, update.ClearTrashedAt)
	return update.Exec(ctx)
}

func (self *IssueRepository) SetParents(ctx context.Context, tx repository.TxInterface, parents []ImportedParent) error {
	db := self.base.Client(tx)
	for _, link := range parents {
		update := db.Issue.UpdateOneID(link.IssueID)
		setOrClear(link.ParentID, update.SetParentID, update.ClearParentID)
		if err := update.Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (self *IssueRepository) ImportAttachments(ctx context.Context, tx repository.TxInterface, attachments []ImportedAttachment) error {
	db := self.base.Client(tx)
	var builders []*ent.AttachmentCreate
	for _, a := range attachments {
		if a.Existing {
			if err := db.Attachment.UpdateOneID(a.ID).SetTitle(a.Title).SetSubtitle(a.Subtitle).SetSortOrder(a.SortOrder).Exec(ctx); err != nil {
				return err
			}
			continue
		}
		builders = append(builders, db.Attachment.Create().
			SetID(a.ID).
			SetIssueID(a.IssueID).
			SetTitle(a.Title).
			SetSubtitle(a.Subtitle).
			SetURL(a.URL).
			SetSourceType(AttachmentTypeImport).
			SetSortOrder(a.SortOrder).
			SetCreatedAt(a.CreatedAt).
			SetUpdatedAt(a.CreatedAt))
	}
	for batch := range slices.Chunk(builders, 500) {
		if err := db.Attachment.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (self *IssueRepository) ImportRelations(ctx context.Context, tx repository.TxInterface, relations []ImportedRelation) error {
	db := self.base.Client(tx)
	builders := make([]*ent.IssueRelationCreate, len(relations))
	for i, r := range relations {
		builders[i] = db.IssueRelation.Create().SetIssueID(r.IssueID).SetRelatedIssueID(r.RelatedIssueID).SetType(r.Type)
	}
	for batch := range slices.Chunk(builders, 500) {
		if err := db.IssueRelation.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (self *IssueRepository) ImportHistory(ctx context.Context, tx repository.TxInterface, history []ImportedHistory) error {
	db := self.base.Client(tx)
	builders := make([]*ent.IssueHistoryCreate, len(history))
	for i, h := range history {
		builders[i] = db.IssueHistory.Create().
			SetIssueID(h.IssueID).
			SetSource(schema.HistorySourceImport).
			SetChanges(h.Changes).
			SetCreatedAt(h.CreatedAt)
	}
	for batch := range slices.Chunk(builders, 500) {
		if err := db.IssueHistory.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
