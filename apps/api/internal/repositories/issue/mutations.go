package issue_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
)

type CreateIssueInput struct {
	WorkspaceID       uuid.UUID
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
	ParentID          *uuid.UUID
	DueDate           *time.Time
	SortOrder         float64
	SubIssueSortOrder float64
	StartedAt         *time.Time
	CompletedAt       *time.Time
	CanceledAt        *time.Time
	LabelIDs          []uuid.UUID
	SubscriberIDs     []uuid.UUID
}

func (self *IssueRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateIssueInput) (*ent.Issue, error) {
	return self.base.Client(tx).Issue.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetTeamID(input.TeamID).
		SetNumber(input.Number).
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetDescriptionText(input.DescriptionText).
		SetPriority(input.Priority).
		SetNillableEstimate(input.Estimate).
		SetStateID(input.StateID).
		SetNillableAssigneeID(input.AssigneeID).
		SetNillableCreatorID(input.CreatorID).
		SetNillableProjectID(input.ProjectID).
		SetNillableMilestoneID(input.MilestoneID).
		SetNillableCycleID(input.CycleID).
		SetNillableParentID(input.ParentID).
		SetNillableDueDate(input.DueDate).
		SetSortOrder(input.SortOrder).
		SetSubIssueSortOrder(input.SubIssueSortOrder).
		SetNillableStartedAt(input.StartedAt).
		SetNillableCompletedAt(input.CompletedAt).
		SetNillableCanceledAt(input.CanceledAt).
		AddLabelIDs(input.LabelIDs...).
		AddSubscriberIDs(input.SubscriberIDs...).
		Save(ctx)
}

// Update writes every field of IssueFields; a nil pointer clears the column.
type IssueFields struct {
	TeamID            uuid.UUID
	Number            int
	Title             string
	Description       string
	DescriptionText   string
	Priority          int
	Estimate          *float64
	StateID           uuid.UUID
	AssigneeID        *uuid.UUID
	ProjectID         *uuid.UUID
	MilestoneID       *uuid.UUID
	CycleID           *uuid.UUID
	ParentID          *uuid.UUID
	DueDate           *time.Time
	SortOrder         float64
	SubIssueSortOrder float64
	StartedAt         *time.Time
	CompletedAt       *time.Time
	CanceledAt        *time.Time
	TriagedAt         *time.Time
	SnoozedUntil      *time.Time
	ArchivedAt        *time.Time
	TrashedAt         *time.Time
	AutoClosedAt      *time.Time
	AutoArchivedAt    *time.Time
}

func FieldsOf(entity *ent.Issue) IssueFields {
	return IssueFields{
		TeamID:            entity.TeamID,
		Number:            entity.Number,
		Title:             entity.Title,
		Description:       entity.Description,
		DescriptionText:   entity.DescriptionText,
		Priority:          entity.Priority,
		Estimate:          entity.Estimate,
		StateID:           entity.StateID,
		AssigneeID:        entity.AssigneeID,
		ProjectID:         entity.ProjectID,
		MilestoneID:       entity.MilestoneID,
		CycleID:           entity.CycleID,
		ParentID:          entity.ParentID,
		DueDate:           entity.DueDate,
		SortOrder:         entity.SortOrder,
		SubIssueSortOrder: entity.SubIssueSortOrder,
		StartedAt:         entity.StartedAt,
		CompletedAt:       entity.CompletedAt,
		CanceledAt:        entity.CanceledAt,
		TriagedAt:         entity.TriagedAt,
		SnoozedUntil:      entity.SnoozedUntil,
		ArchivedAt:        entity.ArchivedAt,
		TrashedAt:         entity.TrashedAt,
		AutoClosedAt:      entity.AutoClosedAt,
		AutoArchivedAt:    entity.AutoArchivedAt,
	}
}

func (self *IssueRepository) Update(ctx context.Context, tx repository.TxInterface, id uuid.UUID, fields *IssueFields) error {
	update := self.base.Client(tx).Issue.UpdateOneID(id).
		SetTeamID(fields.TeamID).
		SetNumber(fields.Number).
		SetTitle(fields.Title).
		SetDescription(fields.Description).
		SetDescriptionText(fields.DescriptionText).
		SetPriority(fields.Priority).
		SetStateID(fields.StateID).
		SetSortOrder(fields.SortOrder).
		SetSubIssueSortOrder(fields.SubIssueSortOrder)
	setOrClear(fields.Estimate, update.SetEstimate, update.ClearEstimate)
	setOrClear(fields.AssigneeID, update.SetAssigneeID, update.ClearAssigneeID)
	setOrClear(fields.ProjectID, update.SetProjectID, update.ClearProjectID)
	setOrClear(fields.MilestoneID, update.SetMilestoneID, update.ClearMilestoneID)
	setOrClear(fields.CycleID, update.SetCycleID, update.ClearCycleID)
	setOrClear(fields.ParentID, update.SetParentID, update.ClearParentID)
	setOrClear(fields.DueDate, update.SetDueDate, update.ClearDueDate)
	setOrClear(fields.StartedAt, update.SetStartedAt, update.ClearStartedAt)
	setOrClear(fields.CompletedAt, update.SetCompletedAt, update.ClearCompletedAt)
	setOrClear(fields.CanceledAt, update.SetCanceledAt, update.ClearCanceledAt)
	setOrClear(fields.TriagedAt, update.SetTriagedAt, update.ClearTriagedAt)
	setOrClear(fields.SnoozedUntil, update.SetSnoozedUntil, update.ClearSnoozedUntil)
	setOrClear(fields.ArchivedAt, update.SetArchivedAt, update.ClearArchivedAt)
	setOrClear(fields.TrashedAt, update.SetTrashedAt, update.ClearTrashedAt)
	setOrClear(fields.AutoClosedAt, update.SetAutoClosedAt, update.ClearAutoClosedAt)
	setOrClear(fields.AutoArchivedAt, update.SetAutoArchivedAt, update.ClearAutoArchivedAt)
	return update.Exec(ctx)
}

func (self *IssueRepository) RespaceIfCrowded(ctx context.Context, tx repository.TxInterface, workspaceID, id uuid.UUID, sortOrder float64) error {
	_, err := self.base.RespaceIfCrowded(ctx, tx, repository.IssueSortOrder, workspaceID, id, sortOrder)
	return err
}

func (self *IssueRepository) RespaceSubIssuesIfCrowded(ctx context.Context, tx repository.TxInterface, parentID, id uuid.UUID, subIssueSortOrder float64) error {
	_, err := self.base.RespaceIfCrowded(ctx, tx, repository.SubIssueSortOrder, parentID, id, subIssueSortOrder)
	return err
}

func setOrClear[T any](value *T, set func(T) *ent.IssueUpdateOne, clear func() *ent.IssueUpdateOne) {
	if value == nil {
		clear()
		return
	}
	set(*value)
}

func (self *IssueRepository) SetLabels(ctx context.Context, tx repository.TxInterface, id uuid.UUID, add, remove []uuid.UUID) error {
	if len(add) == 0 && len(remove) == 0 {
		return nil
	}
	return self.base.Client(tx).Issue.UpdateOneID(id).RemoveLabelIDs(remove...).AddLabelIDs(add...).Exec(ctx)
}

func (self *IssueRepository) RecordMove(ctx context.Context, tx repository.TxInterface, issueID, oldTeamID uuid.UUID, oldNumber int) error {
	return self.base.Client(tx).IssueMove.Create().
		SetIssueID(issueID).
		SetOldTeamID(oldTeamID).
		SetOldNumber(oldNumber).
		Exec(ctx)
}

func (self *IssueRepository) Delete(ctx context.Context, tx repository.TxInterface, id uuid.UUID) error {
	client := self.base.Client(tx)
	if _, err := client.Issue.Update().Where(issue.ParentID(id)).ClearParentID().Save(ctx); err != nil {
		return err
	}
	return client.Issue.DeleteOneID(id).Exec(ctx)
}
