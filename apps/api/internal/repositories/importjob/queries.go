package importjob_repo

import (
	"context"

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
