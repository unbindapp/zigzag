package import_service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/importers/linear"
	"github.com/unbindapp/zigzag-api/internal/models"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	importjob_repo "github.com/unbindapp/zigzag-api/internal/repositories/importjob"
)

const maxExportIssues = 50000

var errDryRun = errors.New("dry run")

type ImportLinearInput struct {
	WorkspaceID        uuid.UUID           `json:"workspace_id" required:"true" format:"uuid" doc:"Workspace to import into; needs admin"`
	DryRun             *bool               `json:"dry_run,omitempty" required:"false" default:"true" doc:"Default true: runs the whole import in a transaction that is rolled back, so the report is exact and nothing is kept. Send false to import for real"`
	RestoreArchived    bool                `json:"restore_archived,omitempty" required:"false" doc:"Import archived issues and projects as active instead of archived"`
	RestoreTrashed     bool                `json:"restore_trashed,omitempty" required:"false" doc:"Import trashed issues and projects as active instead of trashed"`
	CreateMissingUsers *bool               `json:"create_missing_users,omitempty" required:"false" default:"true" doc:"Default true: users without a Zigzag account get an inactive placeholder account; false leaves their issues and comments without a person"`
	Export             linear.LinearExport `json:"export" required:"true" doc:"The Linear export: users, teams, workflow_states, labels, projects, milestones, cycles and issues with their comments, attachments, relations and stateHistory"`
}

func (self *ImportService) ImportLinear(ctx context.Context, userID uuid.UUID, input *ImportLinearInput) (*schema.ImportReport, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, err
	}
	if len(input.Export.Issues) > maxExportIssues {
		return nil, invalid("An export can hold at most 50000 issues; split it by team")
	}
	dryRun := input.DryRun == nil || *input.DryRun

	job, err := self.repo.ImportJob().Create(ctx, nil, &importjob_repo.CreateImportJobInput{
		WorkspaceID: input.WorkspaceID,
		Source:      schema.ImportSourceLinear,
		StartedByID: &userID,
		DryRun:      dryRun,
	})
	if err != nil {
		return nil, err
	}

	report, err := self.runLinear(ctx, userID, input, dryRun)
	report.ImportID = job.ID
	report.DryRun = dryRun
	status := schema.ImportStatusCompleted
	if err != nil {
		status = schema.ImportStatusFailed
		report.Warnings = append(report.Warnings, schema.ImportWarning{Entity: "import", Message: "The import stopped: " + err.Error()})
	}
	if _, finishErr := self.repo.ImportJob().Finish(ctx, job.ID, status, &report, self.now()); finishErr != nil {
		log.Error("Failed to record the import report", "import_id", job.ID, "err", finishErr)
	}
	if err != nil {
		return nil, errdefs.NewInternalError(err, "The import failed; the steps that finished were kept and running it again continues from there")
	}
	return &report, nil
}

func (self *ImportService) runLinear(ctx context.Context, userID uuid.UUID, input *ImportLinearInput, dryRun bool) (schema.ImportReport, error) {
	snapshot, err := self.snapshot(ctx, input.WorkspaceID, &input.Export)
	if err != nil {
		return models.NormalizeImportReport(nil), err
	}
	plan := linear.Build(&input.Export, snapshot, linear.Options{
		ActorID:            userID,
		RestoreArchived:    input.RestoreArchived,
		RestoreTrashed:     input.RestoreTrashed,
		CreateMissingUsers: input.CreateMissingUsers == nil || *input.CreateMissingUsers,
		Now:                self.now().UTC(),
	})
	report := models.NormalizeImportReport(&schema.ImportReport{Counts: plan.Counts, Warnings: plan.Warnings, UserMapping: plan.UserMapping})

	groups := self.writeGroups(ctx, input.WorkspaceID, plan)
	if dryRun {
		err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
			for _, write := range groups {
				if err := write(tx); err != nil {
					return err
				}
			}
			return errDryRun
		})
		if errors.Is(err, errDryRun) {
			return report, nil
		}
		return report, err
	}
	for _, write := range groups {
		if err := self.repo.WithTx(ctx, write); err != nil {
			return report, err
		}
	}
	return report, nil
}

func (self *ImportService) writeGroups(ctx context.Context, workspaceID uuid.UUID, plan *linear.Plan) []func(repository.TxInterface) error {
	return []func(repository.TxInterface) error{
		func(tx repository.TxInterface) error {
			if err := self.repo.User().CreatePlaceholders(ctx, tx, plan.Users); err != nil {
				return err
			}
			return self.repo.Workspace().AddMembers(ctx, tx, workspaceID, plan.Members)
		},
		func(tx repository.TxInterface) error {
			if err := self.repo.Team().ImportTeams(ctx, tx, workspaceID, plan.Teams, plan.States); err != nil {
				return err
			}
			return self.repo.Team().AddMembers(ctx, tx, plan.TeamMembers)
		},
		func(tx repository.TxInterface) error {
			return self.repo.Label().ImportLabels(ctx, tx, workspaceID, plan.Labels)
		},
		func(tx repository.TxInterface) error {
			if err := self.repo.Project().ImportProjects(ctx, tx, workspaceID, linear.Source, plan.Projects); err != nil {
				return err
			}
			return self.repo.Project().ImportMilestones(ctx, tx, plan.Milestones)
		},
		func(tx repository.TxInterface) error {
			return self.repo.Cycle().ImportCycles(ctx, tx, plan.Cycles)
		},
		func(tx repository.TxInterface) error {
			if err := self.repo.Issue().ImportIssues(ctx, tx, workspaceID, linear.Source, plan.Issues); err != nil {
				return err
			}
			if err := self.repo.Issue().SetParents(ctx, tx, plan.Parents); err != nil {
				return err
			}
			if err := self.repo.Team().RaiseIssueCounters(ctx, tx, plan.Counters); err != nil {
				return err
			}
			if err := self.repo.Issue().ImportAttachments(ctx, tx, plan.Attachments); err != nil {
				return err
			}
			return self.repo.Issue().ImportHistory(ctx, tx, plan.History)
		},
		func(tx repository.TxInterface) error {
			if err := self.repo.Comment().ImportComments(ctx, tx, linear.Source, plan.Comments); err != nil {
				return err
			}
			return self.repo.Issue().ImportRelations(ctx, tx, plan.Relations)
		},
	}
}

func (self *ImportService) snapshot(ctx context.Context, workspaceID uuid.UUID, export *linear.LinearExport) (*linear.Snapshot, error) {
	snapshot := &linear.Snapshot{}
	emails := make([]string, len(export.Users))
	for i, u := range export.Users {
		emails[i] = u.Email
	}
	var err error
	if snapshot.Users, err = self.repo.User().GetByEmails(ctx, emails); err != nil {
		return nil, err
	}
	if snapshot.TakenDisplayNames, err = self.repo.User().DisplayNamesWithPrefixes(ctx, linear.DisplayNameBases(export)); err != nil {
		return nil, err
	}
	if snapshot.Members, err = self.repo.Workspace().ListMembers(ctx, workspaceID); err != nil {
		return nil, err
	}
	if snapshot.Teams, err = self.repo.Team().ListByWorkspace(ctx, workspaceID, true); err != nil {
		return nil, err
	}
	teamIDs := make([]uuid.UUID, len(snapshot.Teams))
	for i, team := range snapshot.Teams {
		teamIDs[i] = team.ID
	}
	if snapshot.States, err = self.repo.Team().StatesOfWorkspace(ctx, workspaceID); err != nil {
		return nil, err
	}
	if snapshot.TeamMembers, err = self.repo.Team().MembersOfWorkspace(ctx, workspaceID); err != nil {
		return nil, err
	}
	if snapshot.Labels, err = self.repo.Label().List(ctx, workspaceID, teamIDs, true); err != nil {
		return nil, err
	}
	if snapshot.ProjectStatuses, err = self.repo.Project().ListStatuses(ctx, workspaceID); err != nil {
		return nil, err
	}
	if snapshot.Projects, err = self.repo.Project().ImportedProjects(ctx, workspaceID, linear.Source); err != nil {
		return nil, err
	}
	if snapshot.Milestones, err = self.repo.Project().MilestonesOfImportedProjects(ctx, workspaceID, linear.Source); err != nil {
		return nil, err
	}
	if snapshot.Cycles, err = self.repo.Cycle().CyclesOfWorkspace(ctx, workspaceID); err != nil {
		return nil, err
	}
	if snapshot.Issues, err = self.repo.Issue().IssueKeys(ctx, workspaceID); err != nil {
		return nil, err
	}
	if snapshot.Comments, err = self.repo.Comment().ImportedComments(ctx, workspaceID, linear.Source); err != nil {
		return nil, err
	}
	if snapshot.Attachments, err = self.repo.Issue().ImportedAttachments(ctx, workspaceID, linear.Source); err != nil {
		return nil, err
	}
	if snapshot.Relations, err = self.repo.Issue().ImportedRelations(ctx, workspaceID, linear.Source); err != nil {
		return nil, err
	}
	if snapshot.HistoryIssueIDs, err = self.repo.Issue().IssueIDsWithImportHistory(ctx, workspaceID, linear.Source); err != nil {
		return nil, err
	}
	return snapshot, nil
}

type ListImportsInput struct {
	WorkspaceID uuid.UUID `query:"workspace_id" required:"true" format:"uuid" doc:"Workspace whose imports to list; needs admin"`
}

func (self *ImportService) List(ctx context.Context, userID uuid.UUID, input *ListImportsInput) ([]*models.ImportResponse, error) {
	if err := self.checkAdmin(ctx, userID, input.WorkspaceID); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Workspace not found")
	}
	jobs, err := self.repo.ImportJob().ListByWorkspace(ctx, input.WorkspaceID, 50)
	if err != nil {
		return nil, err
	}
	return models.TransformImportJobs(jobs), nil
}
