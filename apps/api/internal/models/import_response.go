package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type ImportResponse struct {
	ID          uuid.UUID           `json:"id" format:"uuid"`
	WorkspaceID uuid.UUID           `json:"workspace_id" format:"uuid"`
	Source      schema.ImportSource `json:"source"`
	DryRun      bool                `json:"dry_run"`
	Status      schema.ImportStatus `json:"status"`
	StartedBy   *UserSummary        `json:"started_by,omitempty" required:"false" doc:"Who ran the import; absent when the account is gone"`
	Report      schema.ImportReport `json:"report" doc:"Counts, warnings and user mapping; empty while the import runs"`
	CreatedAt   time.Time           `json:"created_at"`
	FinishedAt  *time.Time          `json:"finished_at" nullable:"true"`
}

func NormalizeImportReport(report *schema.ImportReport) schema.ImportReport {
	if report == nil {
		report = &schema.ImportReport{}
	}
	normalized := *report
	if normalized.Counts == nil {
		normalized.Counts = map[string]schema.ImportEntityCount{}
	}
	if normalized.Warnings == nil {
		normalized.Warnings = []schema.ImportWarning{}
	}
	if normalized.UserMapping == nil {
		normalized.UserMapping = []schema.ImportUserMapping{}
	}
	return normalized
}

func TransformImportJob(entity *ent.ImportJob) *ImportResponse {
	if entity == nil {
		return nil
	}
	report := NormalizeImportReport(entity.Report)
	report.ImportID = entity.ID
	report.DryRun = entity.DryRun
	return &ImportResponse{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Source:      entity.Source,
		DryRun:      entity.DryRun,
		Status:      entity.Status,
		StartedBy:   TransformUserSummary(entity.Edges.StartedBy),
		Report:      report,
		CreatedAt:   entity.CreatedAt,
		FinishedAt:  entity.FinishedAt,
	}
}

func TransformImportJobs(entities []*ent.ImportJob) []*ImportResponse {
	responses := make([]*ImportResponse, len(entities))
	for i, entity := range entities {
		responses[i] = TransformImportJob(entity)
	}
	return responses
}
