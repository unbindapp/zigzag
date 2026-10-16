package document_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type DocumentSearchSuite struct {
	repository.RepositoryBaseSuite
	repo      *DocumentRepository
	workspace *ent.Workspace
}

func (suite *DocumentSearchSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewDocumentRepository(suite.DB)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
}

func (suite *DocumentSearchSuite) document(title, content string, updated time.Time) *ent.Document {
	return suite.DB.Document.Create().
		SetWorkspaceID(suite.workspace.ID).
		SetTitle(title).
		SetSlug(title).
		SetContent(content).
		SetUpdatedAt(updated).
		SaveX(suite.Ctx)
}

func (suite *DocumentSearchSuite) TestRanksTitleMatchesAndHidesArchived() {
	base := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	byContent := suite.document("Onboarding", "Read the release checklist first", base.Add(time.Hour))
	byTitle := suite.document("Release checklist", "", base)
	archived := suite.document("Release notes", "", base)
	suite.DB.Document.UpdateOne(archived).SetArchivedAt(base).ExecX(suite.Ctx)

	found, err := suite.repo.Search(suite.Ctx, []uuid.UUID{suite.workspace.ID}, "release", 10)
	suite.Require().NoError(err)
	suite.Require().Len(found, 2)
	suite.Equal(byTitle.ID, found[0].ID)
	suite.Equal(byContent.ID, found[1].ID)

	none, err := suite.repo.Search(suite.Ctx, nil, "release", 10)
	suite.Require().NoError(err)
	suite.Empty(none)
}

func TestDocumentSearchSuite(t *testing.T) {
	suite.Run(t, new(DocumentSearchSuite))
}
