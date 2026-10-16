package project_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateProjectInput struct {
	WorkspaceID          uuid.UUID
	Name                 string
	Slug                 string
	StatusID             uuid.UUID
	Summary              string
	Description          string
	Icon                 string
	Color                string
	LeadID               *uuid.UUID
	Priority             int
	CreatorID            *uuid.UUID
	StartDate            *time.Time
	TargetDate           *time.Time
	StartDateResolution  schema.DateResolution
	TargetDateResolution schema.DateResolution
	SortOrder            float64
	StartedAt            *time.Time
	CompletedAt          *time.Time
	CanceledAt           *time.Time
	TeamIDs              []uuid.UUID
	MemberIDs            []uuid.UUID
	LabelIDs             []uuid.UUID
	InitiativeIDs        []uuid.UUID
}

func (self *ProjectRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateProjectInput) (*ent.Project, error) {
	m := self.base.Client(tx).Project.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetName(input.Name).
		SetSlug(input.Slug).
		SetStatusID(input.StatusID).
		SetSummary(input.Summary).
		SetDescription(input.Description).
		SetIcon(input.Icon).
		SetColor(input.Color).
		SetNillableLeadID(input.LeadID).
		SetPriority(input.Priority).
		SetNillableCreatorID(input.CreatorID).
		SetNillableStartDate(input.StartDate).
		SetNillableTargetDate(input.TargetDate).
		SetSortOrder(input.SortOrder).
		SetNillableStartedAt(input.StartedAt).
		SetNillableCompletedAt(input.CompletedAt).
		SetNillableCanceledAt(input.CanceledAt).
		AddTeamIDs(input.TeamIDs...).
		AddMemberIDs(input.MemberIDs...).
		AddLabelIDs(input.LabelIDs...).
		AddInitiativeIDs(input.InitiativeIDs...)
	if input.StartDateResolution != "" {
		m.SetStartDateResolution(input.StartDateResolution)
	}
	if input.TargetDateResolution != "" {
		m.SetTargetDateResolution(input.TargetDateResolution)
	}
	return m.Save(ctx)
}

type UpdateProjectInput struct {
	Name                 *string
	Summary              *string
	Description          *string
	Icon                 *string
	Color                *string
	StatusID             *uuid.UUID
	LeadID               *uuid.UUID
	ClearLead            bool
	Priority             *int
	StartDate            *time.Time
	ClearStartDate       bool
	TargetDate           *time.Time
	ClearTargetDate      bool
	StartDateResolution  *schema.DateResolution
	TargetDateResolution *schema.DateResolution
	SortOrder            *float64
	StartedAt            *time.Time
	CompletedAt          *time.Time
	ClearCompletedAt     bool
	CanceledAt           *time.Time
	ClearCanceledAt      bool
	TeamIDs              []uuid.UUID
	MemberIDs            []uuid.UUID
	LabelIDs             []uuid.UUID
	InitiativeIDs        []uuid.UUID
	ReplaceTeams         bool
	ReplaceMembers       bool
	ReplaceLabels        bool
	ReplaceInitiatives   bool
}

func (self *ProjectRepository) Update(ctx context.Context, tx repository.TxInterface, id uuid.UUID, input *UpdateProjectInput) error {
	m := self.base.Client(tx).Project.UpdateOneID(id)
	if input.Name != nil {
		m.SetName(*input.Name)
	}
	if input.Summary != nil {
		m.SetSummary(*input.Summary)
	}
	if input.Description != nil {
		m.SetDescription(*input.Description)
	}
	if input.Icon != nil {
		m.SetIcon(*input.Icon)
	}
	if input.Color != nil {
		m.SetColor(*input.Color)
	}
	if input.StatusID != nil {
		m.SetStatusID(*input.StatusID)
	}
	if input.LeadID != nil {
		m.SetLeadID(*input.LeadID)
	}
	if input.ClearLead {
		m.ClearLeadID()
	}
	if input.Priority != nil {
		m.SetPriority(*input.Priority)
	}
	if input.StartDate != nil {
		m.SetStartDate(*input.StartDate)
	}
	if input.ClearStartDate {
		m.ClearStartDate()
	}
	if input.TargetDate != nil {
		m.SetTargetDate(*input.TargetDate)
	}
	if input.ClearTargetDate {
		m.ClearTargetDate()
	}
	if input.StartDateResolution != nil {
		m.SetStartDateResolution(*input.StartDateResolution)
	}
	if input.TargetDateResolution != nil {
		m.SetTargetDateResolution(*input.TargetDateResolution)
	}
	if input.SortOrder != nil {
		m.SetSortOrder(*input.SortOrder)
	}
	if input.StartedAt != nil {
		m.SetStartedAt(*input.StartedAt)
	}
	if input.CompletedAt != nil {
		m.SetCompletedAt(*input.CompletedAt)
	}
	if input.ClearCompletedAt {
		m.ClearCompletedAt()
	}
	if input.CanceledAt != nil {
		m.SetCanceledAt(*input.CanceledAt)
	}
	if input.ClearCanceledAt {
		m.ClearCanceledAt()
	}
	if input.ReplaceTeams {
		m.ClearTeams().AddTeamIDs(input.TeamIDs...)
	}
	if input.ReplaceMembers {
		m.ClearMembers().AddMemberIDs(input.MemberIDs...)
	}
	if input.ReplaceLabels {
		m.ClearLabels().AddLabelIDs(input.LabelIDs...)
	}
	if input.ReplaceInitiatives {
		m.ClearInitiatives().AddInitiativeIDs(input.InitiativeIDs...)
	}
	updated, err := m.Save(ctx)
	if err != nil || input.SortOrder == nil {
		return err
	}
	_, err = self.base.RespaceIfCrowded(ctx, tx, repository.ProjectSortOrder, updated.WorkspaceID, updated.ID, updated.SortOrder)
	return err
}

func (self *ProjectRepository) SetArchivedAt(ctx context.Context, id uuid.UUID, at *time.Time) error {
	m := self.base.DB.Project.UpdateOneID(id)
	if at == nil {
		return m.ClearArchivedAt().Exec(ctx)
	}
	return m.SetArchivedAt(*at).Exec(ctx)
}

func (self *ProjectRepository) SetTrashedAt(ctx context.Context, id uuid.UUID, at *time.Time) error {
	m := self.base.DB.Project.UpdateOneID(id)
	if at == nil {
		return m.ClearTrashedAt().Exec(ctx)
	}
	return m.SetTrashedAt(*at).Exec(ctx)
}

func (self *ProjectRepository) SetHealth(ctx context.Context, tx repository.TxInterface, id uuid.UUID, health *schema.Health) error {
	m := self.base.Client(tx).Project.UpdateOneID(id)
	if health == nil {
		return m.ClearHealth().Exec(ctx)
	}
	return m.SetHealth(*health).Exec(ctx)
}

func (self *ProjectRepository) Delete(ctx context.Context, tx repository.TxInterface, id uuid.UUID) error {
	return self.base.Client(tx).Project.DeleteOneID(id).Exec(ctx)
}

func (self *ProjectRepository) IssuesInProject(ctx context.Context, tx repository.TxInterface, projectID uuid.UUID) ([]*ent.Issue, error) {
	return self.base.Client(tx).Issue.Query().
		Where(issue.ProjectID(projectID)).
		Select(issue.FieldID, issue.FieldMilestoneID).
		WithMilestone().
		All(ctx)
}

func (self *ProjectRepository) IssueIDsInMilestone(ctx context.Context, tx repository.TxInterface, milestoneID uuid.UUID) ([]uuid.UUID, error) {
	return self.base.Client(tx).Issue.Query().Where(issue.MilestoneID(milestoneID)).IDs(ctx)
}
