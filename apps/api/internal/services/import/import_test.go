package import_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/importers/linear"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	importjob_repo "github.com/unbindapp/zigzag-api/internal/repositories/importjob"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type ImportServiceSuite struct {
	services.ServiceTestSuite
	service     *ImportService
	userID      uuid.UUID
	workspaceID uuid.UUID
	jobID       uuid.UUID
	finished    []schema.ImportStatus
	reports     []*schema.ImportReport
	txs         int
}

func (suite *ImportServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewImportService(&config.Config{}, suite.MockRepo)
	suite.service.now = func() time.Time { return time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC) }
	suite.userID, suite.workspaceID, suite.jobID = uuid.New(), uuid.New(), uuid.New()
	suite.finished, suite.reports, suite.txs = nil, nil, 0
}

func (suite *ImportServiceSuite) allowAdmin(allowed bool) {
	var result error
	if !allowed {
		result = errdefs.ErrUnauthorized
	}
	suite.MockPermissionsRepo.EXPECT().
		Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleAdmin, WorkspaceID: suite.workspaceID}}).
		Return(result).Once()
}

func (suite *ImportServiceSuite) expectJob(dryRun bool) {
	suite.MockImportJobRepo.EXPECT().Create(mock.Anything, nil, mock.Anything).
		RunAndReturn(func(_ context.Context, _ repository.TxInterface, input *importjob_repo.CreateImportJobInput) (*ent.ImportJob, error) {
			suite.Equal(dryRun, input.DryRun)
			suite.Equal(suite.workspaceID, input.WorkspaceID)
			suite.Equal(&suite.userID, input.StartedByID)
			return &ent.ImportJob{ID: suite.jobID, WorkspaceID: suite.workspaceID, DryRun: dryRun}, nil
		}).Once()
	suite.MockImportJobRepo.EXPECT().Finish(mock.Anything, suite.jobID, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ uuid.UUID, status schema.ImportStatus, report *schema.ImportReport, _ time.Time) (*ent.ImportJob, error) {
			suite.finished = append(suite.finished, status)
			suite.reports = append(suite.reports, report)
			return &ent.ImportJob{ID: suite.jobID}, nil
		}).Once()
}

func (suite *ImportServiceSuite) expectEmptyWorkspace() {
	suite.MockUserRepo.EXPECT().GetByEmails(mock.Anything, mock.Anything).Return([]*ent.User{}, nil).Once()
	suite.MockUserRepo.EXPECT().DisplayNamesWithPrefixes(mock.Anything, mock.Anything).Return([]string{}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ListMembers(mock.Anything, suite.workspaceID).Return([]*ent.WorkspaceMember{}, nil).Once()
	suite.MockTeamRepo.EXPECT().ListByWorkspace(mock.Anything, suite.workspaceID, true).Return([]*ent.Team{}, nil).Once()
	suite.MockTeamRepo.EXPECT().StatesOfWorkspace(mock.Anything, suite.workspaceID).Return([]*ent.WorkflowState{}, nil).Once()
	suite.MockTeamRepo.EXPECT().MembersOfWorkspace(mock.Anything, suite.workspaceID).Return([]*ent.TeamMember{}, nil).Once()
	suite.MockLabelRepo.EXPECT().List(mock.Anything, suite.workspaceID, []uuid.UUID{}, true).Return([]*ent.Label{}, nil).Once()
	suite.MockProjectRepo.EXPECT().ListStatuses(mock.Anything, suite.workspaceID).Return([]*ent.ProjectStatus{{ID: uuid.New(), Name: "Backlog", Type: schema.ProjectStatusBacklog}}, nil).Once()
	suite.MockProjectRepo.EXPECT().ImportedProjects(mock.Anything, suite.workspaceID, linear.Source).Return([]*ent.Project{}, nil).Once()
	suite.MockProjectRepo.EXPECT().MilestonesOfImportedProjects(mock.Anything, suite.workspaceID, linear.Source).Return([]*ent.ProjectMilestone{}, nil).Once()
	suite.MockCycleRepo.EXPECT().CyclesOfWorkspace(mock.Anything, suite.workspaceID).Return([]*ent.Cycle{}, nil).Once()
	suite.MockIssueRepo.EXPECT().IssueKeys(mock.Anything, suite.workspaceID).Return([]*ent.Issue{}, nil).Once()
	suite.MockCommentRepo.EXPECT().ImportedComments(mock.Anything, suite.workspaceID, linear.Source).Return([]*ent.Comment{}, nil).Once()
	suite.MockIssueRepo.EXPECT().ImportedAttachments(mock.Anything, suite.workspaceID, linear.Source).Return([]*ent.Attachment{}, nil).Once()
	suite.MockIssueRepo.EXPECT().ImportedRelations(mock.Anything, suite.workspaceID, linear.Source).Return([]*ent.IssueRelation{}, nil).Once()
	suite.MockIssueRepo.EXPECT().IssueIDsWithImportHistory(mock.Anything, suite.workspaceID, linear.Source).Return([]uuid.UUID{}, nil).Once()
}

func (suite *ImportServiceSuite) expectWrites(failIssues error) {
	suite.MockRepo.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(repository.TxInterface) error) error {
		suite.txs++
		return fn(suite.NewTxMock())
	})
	suite.MockUserRepo.EXPECT().CreatePlaceholders(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockWorkspaceRepo.EXPECT().AddMembers(mock.Anything, mock.Anything, suite.workspaceID, mock.Anything).Return(nil).Maybe()
	suite.MockTeamRepo.EXPECT().ImportTeams(mock.Anything, mock.Anything, suite.workspaceID, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockTeamRepo.EXPECT().AddMembers(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockLabelRepo.EXPECT().ImportLabels(mock.Anything, mock.Anything, suite.workspaceID, mock.Anything).Return(nil).Maybe()
	suite.MockProjectRepo.EXPECT().ImportProjects(mock.Anything, mock.Anything, suite.workspaceID, linear.Source, mock.Anything).Return(nil).Maybe()
	suite.MockProjectRepo.EXPECT().ImportMilestones(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockCycleRepo.EXPECT().ImportCycles(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockIssueRepo.EXPECT().ImportIssues(mock.Anything, mock.Anything, suite.workspaceID, linear.Source, mock.Anything).
		RunAndReturn(func(_ context.Context, _ repository.TxInterface, _ uuid.UUID, _ string, issues []issue_repo.ImportedIssue) error {
			suite.Len(issues, 1)
			return failIssues
		}).Maybe()
	suite.MockIssueRepo.EXPECT().SetParents(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockTeamRepo.EXPECT().RaiseIssueCounters(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockIssueRepo.EXPECT().ImportAttachments(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockIssueRepo.EXPECT().ImportHistory(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	suite.MockCommentRepo.EXPECT().ImportComments(mock.Anything, mock.Anything, linear.Source, mock.Anything).Return(nil).Maybe()
	suite.MockIssueRepo.EXPECT().ImportRelations(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (suite *ImportServiceSuite) input(dryRun *bool) *ImportLinearInput {
	return &ImportLinearInput{
		WorkspaceID: suite.workspaceID,
		DryRun:      dryRun,
		Export: linear.LinearExport{
			Users:  []linear.LinearUser{{ID: "u-1", Email: "ada@example.com"}},
			Teams:  []linear.LinearTeam{{ID: "t-1", Key: "ENG", Name: "Engineering"}},
			Issues: []linear.LinearIssue{{ID: "i-1", Identifier: "ENG-1", Number: 1, TeamID: "t-1", Title: "First", CreatorID: strPtr("u-1")}},
		},
	}
}

func strPtr(value string) *string { return &value }

func (suite *ImportServiceSuite) TestImportNeedsAdmin() {
	suite.allowAdmin(false)
	_, err := suite.service.ImportLinear(suite.Ctx, suite.userID, suite.input(nil))
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ImportServiceSuite) TestImportRefusesHugeExports() {
	suite.allowAdmin(true)
	input := suite.input(nil)
	input.Export.Issues = make([]linear.LinearIssue, maxExportIssues+1)
	_, err := suite.service.ImportLinear(suite.Ctx, suite.userID, input)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ImportServiceSuite) TestDryRunIsTheDefaultAndRollsBack() {
	suite.allowAdmin(true)
	suite.expectJob(true)
	suite.expectEmptyWorkspace()
	suite.expectWrites(nil)

	report, err := suite.service.ImportLinear(suite.Ctx, suite.userID, suite.input(nil))
	suite.Require().NoError(err)
	suite.True(report.DryRun)
	suite.Equal(suite.jobID, report.ImportID)
	suite.Equal(1, suite.txs, "a dry run writes every group in one transaction")
	suite.Equal(1, report.Counts[linear.EntityIssues].Created)
	suite.Equal(1, report.Counts[linear.EntityUsers].Created)
	suite.Equal([]schema.ImportStatus{schema.ImportStatusCompleted}, suite.finished)
}

func (suite *ImportServiceSuite) TestRealImportCommitsEachGroup() {
	suite.allowAdmin(true)
	suite.expectJob(false)
	suite.expectEmptyWorkspace()
	suite.expectWrites(nil)

	dryRun := false
	report, err := suite.service.ImportLinear(suite.Ctx, suite.userID, suite.input(&dryRun))
	suite.Require().NoError(err)
	suite.False(report.DryRun)
	suite.Equal(7, suite.txs)
	suite.Equal([]schema.ImportStatus{schema.ImportStatusCompleted}, suite.finished)
}

func (suite *ImportServiceSuite) TestFailedImportIsRecorded() {
	suite.allowAdmin(true)
	suite.expectJob(false)
	suite.expectEmptyWorkspace()
	suite.expectWrites(errors.New("boom"))

	dryRun := false
	_, err := suite.service.ImportLinear(suite.Ctx, suite.userID, suite.input(&dryRun))
	var internal *errdefs.InternalError
	suite.ErrorAs(err, &internal)
	suite.Equal([]schema.ImportStatus{schema.ImportStatusFailed}, suite.finished)
	warnings := suite.reports[0].Warnings
	suite.Equal("import", warnings[len(warnings)-1].Entity)
	suite.Equal(6, suite.txs, "groups after the failing one never run")
}

func (suite *ImportServiceSuite) TestListMasksOutsidersAsNotFound() {
	suite.allowAdmin(false)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListImportsInput{WorkspaceID: suite.workspaceID})
	suite.True(errdefs.IsNotFound(err))
}

func (suite *ImportServiceSuite) TestListReturnsReports() {
	suite.allowAdmin(true)
	finished := time.Now()
	suite.MockImportJobRepo.EXPECT().ListByWorkspace(mock.Anything, suite.workspaceID, 50).Return([]*ent.ImportJob{
		{ID: suite.jobID, WorkspaceID: suite.workspaceID, Status: schema.ImportStatusCompleted, DryRun: true, FinishedAt: &finished,
			Report: &schema.ImportReport{Counts: map[string]schema.ImportEntityCount{linear.EntityIssues: {Created: 3}}}},
		{ID: uuid.New(), WorkspaceID: suite.workspaceID, Status: schema.ImportStatusRunning},
	}, nil).Once()

	imports, err := suite.service.List(suite.Ctx, suite.userID, &ListImportsInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Len(imports, 2)
	suite.Equal(3, imports[0].Report.Counts[linear.EntityIssues].Created)
	suite.Equal(suite.jobID, imports[0].Report.ImportID)
	suite.NotNil(imports[1].Report.Counts)
	suite.NotNil(imports[1].Report.Warnings)
}

func TestImportServiceSuite(t *testing.T) {
	suite.Run(t, new(ImportServiceSuite))
}
