package project_repo

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/project"
	"github.com/unbindapp/zigzag-api/ent/projectmilestone"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

const importBatchSize = 500

type ImportedProject struct {
	ID          uuid.UUID
	Existing    bool
	ExternalID  string
	Name        string
	Slug        string
	Summary     string
	Description string
	StatusID    uuid.UUID
	LeadID      *uuid.UUID
	Priority    int
	StartDate   *time.Time
	TargetDate  *time.Time
	SortOrder   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	CanceledAt  *time.Time
	ArchivedAt  *time.Time
	TrashedAt   *time.Time
	TeamIDs     []uuid.UUID
	MemberIDs   []uuid.UUID
	LabelIDs    []uuid.UUID
}

type ImportedMilestone struct {
	ID          uuid.UUID
	Existing    bool
	ProjectID   uuid.UUID
	Name        string
	Description string
	TargetDate  *time.Time
	SortOrder   float64
}

func (self *ProjectRepository) ImportedProjects(ctx context.Context, workspaceID uuid.UUID, source string) ([]*ent.Project, error) {
	return self.base.DB.Project.Query().
		Where(project.WorkspaceID(workspaceID), project.ImportSource(source), project.ImportExternalIDNEQ("")).
		All(ctx)
}

func (self *ProjectRepository) MilestonesOfImportedProjects(ctx context.Context, workspaceID uuid.UUID, source string) ([]*ent.ProjectMilestone, error) {
	return self.base.DB.ProjectMilestone.Query().
		Where(projectmilestone.HasProjectWith(project.WorkspaceID(workspaceID), project.ImportSource(source))).
		All(ctx)
}

func (self *ProjectRepository) ImportProjects(ctx context.Context, tx repository.TxInterface, workspaceID uuid.UUID, source string, projects []ImportedProject) error {
	db := self.base.Client(tx)
	var builders []*ent.ProjectCreate
	for _, p := range projects {
		if p.Existing {
			if err := self.updateImportedProject(ctx, db, p); err != nil {
				return err
			}
			continue
		}
		builders = append(builders, db.Project.Create().
			SetID(p.ID).
			SetWorkspaceID(workspaceID).
			SetImportSource(source).
			SetImportExternalID(p.ExternalID).
			SetName(p.Name).
			SetSlug(p.Slug).
			SetSummary(p.Summary).
			SetDescription(p.Description).
			SetStatusID(p.StatusID).
			SetNillableLeadID(p.LeadID).
			SetPriority(p.Priority).
			SetNillableStartDate(p.StartDate).
			SetNillableTargetDate(p.TargetDate).
			SetSortOrder(p.SortOrder).
			SetCreatedAt(p.CreatedAt).
			SetUpdatedAt(p.UpdatedAt).
			SetNillableStartedAt(p.StartedAt).
			SetNillableCompletedAt(p.CompletedAt).
			SetNillableCanceledAt(p.CanceledAt).
			SetNillableArchivedAt(p.ArchivedAt).
			SetNillableTrashedAt(p.TrashedAt).
			AddTeamIDs(p.TeamIDs...).
			AddMemberIDs(p.MemberIDs...).
			AddLabelIDs(p.LabelIDs...))
	}
	for batch := range slices.Chunk(builders, importBatchSize) {
		if err := db.Project.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (self *ProjectRepository) updateImportedProject(ctx context.Context, db *ent.Client, p ImportedProject) error {
	update := db.Project.UpdateOneID(p.ID).
		SetName(p.Name).
		SetSummary(p.Summary).
		SetDescription(p.Description).
		SetStatusID(p.StatusID).
		SetPriority(p.Priority).
		SetSortOrder(p.SortOrder).
		SetUpdatedAt(p.UpdatedAt).
		ClearTeams().AddTeamIDs(p.TeamIDs...).
		ClearMembers().AddMemberIDs(p.MemberIDs...).
		ClearLabels().AddLabelIDs(p.LabelIDs...)
	setOrClear(p.LeadID, update.SetLeadID, update.ClearLeadID)
	setOrClear(p.StartDate, update.SetStartDate, update.ClearStartDate)
	setOrClear(p.TargetDate, update.SetTargetDate, update.ClearTargetDate)
	setOrClear(p.StartedAt, update.SetStartedAt, update.ClearStartedAt)
	setOrClear(p.CompletedAt, update.SetCompletedAt, update.ClearCompletedAt)
	setOrClear(p.CanceledAt, update.SetCanceledAt, update.ClearCanceledAt)
	setOrClear(p.ArchivedAt, update.SetArchivedAt, update.ClearArchivedAt)
	setOrClear(p.TrashedAt, update.SetTrashedAt, update.ClearTrashedAt)
	return update.Exec(ctx)
}

func (self *ProjectRepository) ImportMilestones(ctx context.Context, tx repository.TxInterface, milestones []ImportedMilestone) error {
	db := self.base.Client(tx)
	var builders []*ent.ProjectMilestoneCreate
	for _, ms := range milestones {
		if ms.Existing {
			update := db.ProjectMilestone.UpdateOneID(ms.ID).SetName(ms.Name).SetDescription(ms.Description).SetSortOrder(ms.SortOrder)
			setOrClear(ms.TargetDate, update.SetTargetDate, update.ClearTargetDate)
			if err := update.Exec(ctx); err != nil {
				return err
			}
			continue
		}
		builders = append(builders, db.ProjectMilestone.Create().
			SetID(ms.ID).
			SetProjectID(ms.ProjectID).
			SetName(ms.Name).
			SetDescription(ms.Description).
			SetNillableTargetDate(ms.TargetDate).
			SetSortOrder(ms.SortOrder))
	}
	for batch := range slices.Chunk(builders, importBatchSize) {
		if err := db.ProjectMilestone.CreateBulk(batch...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func setOrClear[T, B any](value *T, set func(T) B, clear func() B) {
	if value == nil {
		clear()
		return
	}
	set(*value)
}
