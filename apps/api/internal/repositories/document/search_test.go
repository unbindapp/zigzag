package document_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
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

	found, err := suite.repo.Search(suite.Ctx, []uuid.UUID{suite.workspace.ID}, nil, "release", 10)
	suite.Require().NoError(err)
	suite.Require().Len(found, 2)
	suite.Equal(byTitle.ID, found[0].ID)
	suite.Equal(byContent.ID, found[1].ID)

	none, err := suite.repo.Search(suite.Ctx, nil, nil, "release", 10)
	suite.Require().NoError(err)
	suite.Empty(none)
}

func (suite *DocumentSearchSuite) TestDocumentsOfIssuesInHiddenTeamsAreLeftOut() {
	open := suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Engineering").SetKey("ENG").SaveX(suite.Ctx)
	secret := suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Security").SetKey("SEC").SetPrivate(true).SaveX(suite.Ctx)
	issueIn := func(team *ent.Team) *ent.Issue {
		state := suite.DB.WorkflowState.Create().SetTeamID(team.ID).SetName("Todo").SetType(schema.StateTypeUnstarted).SetColor("#e2e2e2").SaveX(suite.Ctx)
		return suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(team.ID).SetNumber(1).SetTitle("x").SetStateID(state.ID).SaveX(suite.Ctx)
	}
	base := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	loose := suite.document("Release plan", "", base)
	onOpen := suite.document("Release spec", "", base)
	suite.DB.Document.UpdateOne(onOpen).SetIssueID(issueIn(open).ID).ExecX(suite.Ctx)
	onSecret := suite.document("Release incident", "", base)
	suite.DB.Document.UpdateOne(onSecret).SetIssueID(issueIn(secret).ID).ExecX(suite.Ctx)

	found, err := suite.repo.Search(suite.Ctx, []uuid.UUID{suite.workspace.ID}, []uuid.UUID{open.ID}, "release", 10)
	suite.Require().NoError(err)
	ids := []uuid.UUID{}
	for _, document := range found {
		ids = append(ids, document.ID)
	}
	suite.ElementsMatch([]uuid.UUID{loose.ID, onOpen.ID}, ids)

	listed, err := suite.repo.List(suite.Ctx, &ListDocumentsFilter{WorkspaceID: suite.workspace.ID, VisibleTeamIDs: []uuid.UUID{open.ID}})
	suite.Require().NoError(err)
	suite.Len(listed, 2)
}

func TestDocumentSearchSuite(t *testing.T) {
	suite.Run(t, new(DocumentSearchSuite))
}
