package importjob_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/importjob"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
)

type CreateImportJobInput struct {
	WorkspaceID uuid.UUID
	Source      schema.ImportSource
	StartedByID *uuid.UUID
	DryRun      bool
}

func (self *ImportJobRepository) Create(ctx context.Context, tx repository.TxInterface, input *CreateImportJobInput) (*ent.ImportJob, error) {
	return self.base.Client(tx).ImportJob.Create().
		SetWorkspaceID(input.WorkspaceID).
		SetSource(input.Source).
		SetNillableStartedByID(input.StartedByID).
		SetDryRun(input.DryRun).
		Save(ctx)
}

func (self *ImportJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.ImportJob, error) {
	return self.base.DB.ImportJob.Get(ctx, id)
}

func (self *ImportJobRepository) ResolveWorkspaceOfImportJob(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	found, err := self.base.DB.ImportJob.Query().Where(importjob.ID(id)).Select(importjob.FieldWorkspaceID).Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return found.WorkspaceID, nil
}

func (self *ImportJobRepository) Finish(ctx context.Context, id uuid.UUID, status schema.ImportStatus, report *schema.ImportReport, finishedAt time.Time) (*ent.ImportJob, error) {
	return self.base.DB.ImportJob.UpdateOneID(id).
		SetStatus(status).
		SetReport(report).
		SetFinishedAt(finishedAt).
		Save(ctx)
}

func (self *ImportJobRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*ent.ImportJob, error) {
	return self.base.DB.ImportJob.Query().
		Where(importjob.WorkspaceID(workspaceID)).
		WithStartedBy().
		Order(ent.Desc(importjob.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}
