package project_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type ProjectSearchSuite struct {
	repository.RepositoryBaseSuite
	repo      *ProjectRepository
	workspace *ent.Workspace
	status    *ent.ProjectStatus
}

func (suite *ProjectSearchSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewProjectRepository(suite.DB)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.status = suite.DB.ProjectStatus.Create().SetWorkspaceID(suite.workspace.ID).SetName("Planned").SetType(schema.ProjectStatusPlanned).SetColor("#e2e2e2").SaveX(suite.Ctx)
}

func (suite *ProjectSearchSuite) project(workspaceID uuid.UUID, name, summary string, updated time.Time) *ent.Project {
	return suite.DB.Project.Create().
		SetWorkspaceID(workspaceID).
		SetName(name).
		SetSlug(name).
		SetSummary(summary).
		SetStatusID(suite.status.ID).
		SetUpdatedAt(updated).
		SaveX(suite.Ctx)
}

func (suite *ProjectSearchSuite) TestRanksNameMatchesAndHidesTrash() {
	base := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	bySummary := suite.project(suite.workspace.ID, "Website", "New mobile layout", base.Add(3*time.Hour))
	contains := suite.project(suite.workspace.ID, "Native mobile", "", base.Add(2*time.Hour))
	prefix := suite.project(suite.workspace.ID, "Mobile app", "", base)
	exact := suite.project(suite.workspace.ID, "MOBILE", "", base.Add(-time.Hour))
	trashed := suite.project(suite.workspace.ID, "Mobile old", "", base)
	suite.DB.Project.UpdateOne(trashed).SetTrashedAt(base).ExecX(suite.Ctx)
	other := suite.DB.Workspace.Create().SetName("Other").SetSlug("other").SaveX(suite.Ctx)
	suite.project(other.ID, "Mobile elsewhere", "", base)

	found, err := suite.repo.Search(suite.Ctx, []uuid.UUID{suite.workspace.ID}, " mobile ", 10)
	suite.Require().NoError(err)
	names := make([]uuid.UUID, len(found))
	for i, project := range found {
		names[i] = project.ID
	}
	suite.Equal([]uuid.UUID{exact.ID, prefix.ID, contains.ID, bySummary.ID}, names)

	limited, err := suite.repo.Search(suite.Ctx, []uuid.UUID{suite.workspace.ID}, "mobile", 2)
	suite.Require().NoError(err)
	suite.Len(limited, 2)

	empty, err := suite.repo.Search(suite.Ctx, []uuid.UUID{suite.workspace.ID}, "  ", 10)
	suite.Require().NoError(err)
	suite.Empty(empty)
}

func TestProjectSearchSuite(t *testing.T) {
	suite.Run(t, new(ProjectSearchSuite))
}
