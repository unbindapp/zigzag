package import_service

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/attachment"
	"github.com/unbindapp/zigzag-api/ent/comment"
	"github.com/unbindapp/zigzag-api/ent/issue"
	"github.com/unbindapp/zigzag-api/ent/issuehistory"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/user"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/internal/importers/linear"
	"github.com/unbindapp/zigzag-api/internal/models"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

type ImportDatabaseSuite struct {
	repositorytest.RepositoryBaseSuite
	service   *ImportService
	admin     *ent.User
	workspace *ent.Workspace
	export    linear.LinearExport
}

func (suite *ImportDatabaseSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	repo := repositories.NewRepositories(suite.DB)
	suite.service = NewImportService(&config.Config{}, repo)
	suite.admin = suite.DB.User.Create().SetEmail("admin@example.com").SetName("Admin").SetDisplayName("admin").SaveX(suite.Ctx)
	suite.DB.User.Create().SetEmail("ngyekta@gmail.com").SetName("Existing").SetDisplayName("ngyekta").SaveX(suite.Ctx)
	var err error
	suite.workspace, err = workspace_repo.NewWorkspaceRepository(suite.DB).Create(suite.Ctx, nil, &workspace_repo.CreateWorkspaceInput{Name: "Unbind", Slug: "unbind", CreatorID: suite.admin.ID})
	suite.Require().NoError(err)

	raw, err := os.ReadFile("../../importers/linear/testdata/unbind-export.json")
	suite.Require().NoError(err)
	suite.Require().NoError(json.Unmarshal(raw, &suite.export))
}

func (suite *ImportDatabaseSuite) run(dryRun bool) *schema.ImportReport {
	started := time.Now()
	report, err := suite.service.ImportLinear(suite.Ctx, suite.admin.ID, &ImportLinearInput{WorkspaceID: suite.workspace.ID, DryRun: &dryRun, Export: suite.export})
	suite.Require().NoError(err)
	suite.Less(time.Since(started), 30*time.Second)
	suite.Equal(dryRun, report.DryRun)
	return report
}

type tableCounts struct {
	issues, archived, comments, attachments, labels, states, projects, history, relations, users int
}

func (suite *ImportDatabaseSuite) counts() tableCounts {
	ws := issue.WorkspaceID(suite.workspace.ID)
	return tableCounts{
		issues:      suite.DB.Issue.Query().Where(ws).CountX(suite.Ctx),
		archived:    suite.DB.Issue.Query().Where(ws, issue.ArchivedAtNotNil()).CountX(suite.Ctx),
		comments:    suite.DB.Comment.Query().Where(comment.HasIssueWith(ws)).CountX(suite.Ctx),
		attachments: suite.DB.Attachment.Query().Where(attachment.SourceType("import")).CountX(suite.Ctx),
		labels:      suite.DB.Label.Query().CountX(suite.Ctx),
		states:      suite.DB.WorkflowState.Query().Where(workflowstate.HasTeamWith(team.Key("UNB"))).CountX(suite.Ctx),
		projects:    suite.DB.Project.Query().CountX(suite.Ctx),
		history:     suite.DB.IssueHistory.Query().Where(issuehistory.SourceEQ(schema.HistorySourceImport)).CountX(suite.Ctx),
		relations:   suite.DB.IssueRelation.Query().CountX(suite.Ctx),
		users:       suite.DB.User.Query().CountX(suite.Ctx),
	}
}

func (suite *ImportDatabaseSuite) TestDryRunWritesNothing() {
	before := suite.counts()
	report := suite.run(true)

	suite.Equal(456, report.Counts[linear.EntityIssues].Created)
	suite.Equal(before, suite.counts())
	suite.False(suite.DB.Team.Query().Where(team.Key("UNB")).ExistX(suite.Ctx))
}

func (suite *ImportDatabaseSuite) TestImportingTwiceCreatesNothingTheSecondTime() {
	first := suite.run(false)
	suite.Equal(456, first.Counts[linear.EntityIssues].Created)
	suite.Equal(2, first.Counts[linear.EntityUsers].Created)
	suite.Equal(1, first.Counts[linear.EntityUsers].Updated)

	after := suite.counts()
	suite.Equal(456, after.issues)
	suite.Equal(192, after.archived)
	suite.Equal(21, after.comments)
	suite.Equal(28, after.attachments)
	suite.Equal(9, after.labels)
	suite.Equal(7, after.states)
	suite.Equal(1, after.projects)
	suite.Equal(first.Counts[linear.EntityHistory].Created, after.history)
	suite.Equal(first.Counts[linear.EntityRelations].Created, after.relations)
	suite.Equal(4, after.users)

	unb := suite.DB.Team.Query().Where(team.Key("UNB")).OnlyX(suite.Ctx)
	suite.Equal(458, unb.IssueCounter)
	suite.NotNil(unb.DefaultStateID)
	recent := suite.DB.Issue.Query().Where(issue.TeamID(unb.ID), issue.Number(455)).WithLabels().WithProject().OnlyX(suite.Ctx)
	suite.Len(recent.Edges.Labels, 2)
	suite.NotNil(recent.Edges.Project)
	suite.Equal(time.Date(2026, 10, 15, 18, 20, 19, 197000000, time.UTC), recent.CreatedAt.UTC())
	placeholder := suite.DB.User.Query().Where(user.Email("unbindapp@gmail.com")).OnlyX(suite.Ctx)
	suite.False(placeholder.Active)
	suite.Empty(placeholder.PasswordHash)

	second := suite.run(false)
	for _, entity := range linear.Entities {
		suite.Zero(second.Counts[entity].Created, entity)
	}
	suite.Equal(456, second.Counts[linear.EntityIssues].Updated)
	suite.Equal(after, suite.counts())

	imports, err := suite.service.List(suite.Ctx, suite.admin.ID, &ListImportsInput{WorkspaceID: suite.workspace.ID})
	suite.Require().NoError(err)
	suite.Len(imports, 2)
	for _, job := range imports {
		suite.Equal(schema.ImportStatusCompleted, job.Status)
		suite.Equal(456, job.Report.Counts[linear.EntityIssues].Created+job.Report.Counts[linear.EntityIssues].Updated)
		suite.Equal(models.TransformUserSummary(suite.admin).ID, job.StartedBy.ID)
	}
}

func (suite *ImportDatabaseSuite) TestRestoreArchivedImportsEverythingActive() {
	suite.export.Issues = suite.export.Issues[:50]
	report, err := suite.service.ImportLinear(suite.Ctx, suite.admin.ID, &ImportLinearInput{
		WorkspaceID: suite.workspace.ID, DryRun: new(bool), RestoreArchived: true, Export: suite.export,
	})
	suite.Require().NoError(err)
	suite.Equal(50, report.Counts[linear.EntityIssues].Created)
	suite.Zero(suite.counts().archived)
}

func TestImportDatabaseSuite(t *testing.T) {
	suite.Run(t, new(ImportDatabaseSuite))
}
