package label_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type LabelRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo      *LabelRepository
	workspace *ent.Workspace
	team      *ent.Team
}

func (suite *LabelRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewLabelRepository(suite.DB)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.team = suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Eng").SetKey("ENG").SaveX(suite.Ctx)
}

func (suite *LabelRepositorySuite) label(name string, teamID, parentID *uuid.UUID, isGroup bool, sortOrder float64) *ent.Label {
	created, err := suite.repo.Create(suite.Ctx, nil, &CreateLabelInput{WorkspaceID: suite.workspace.ID, TeamID: teamID, Name: name, Color: "#eb5757", ParentID: parentID, IsGroup: isGroup, SortOrder: sortOrder})
	suite.Require().NoError(err)
	return created
}

func (suite *LabelRepositorySuite) TestNamesAreUniqueAmongSiblingsOnly() {
	group := suite.label("Type", nil, nil, true, 1)
	suite.label("Bug", nil, &group.ID, false, 1)

	taken, err := suite.repo.NameTaken(suite.Ctx, nil, suite.workspace.ID, nil, &group.ID, "bug", nil)
	suite.Require().NoError(err)
	suite.True(taken, "names compare without case")
	taken, err = suite.repo.NameTaken(suite.Ctx, nil, suite.workspace.ID, nil, nil, "Bug", nil)
	suite.Require().NoError(err)
	suite.False(taken, "a top level label may share a name with a group's label")
	taken, err = suite.repo.NameTaken(suite.Ctx, nil, suite.workspace.ID, &suite.team.ID, &group.ID, "Bug", nil)
	suite.Require().NoError(err)
	suite.False(taken)

	next, err := suite.repo.NextSortOrder(suite.Ctx, nil, suite.workspace.ID, nil, &group.ID)
	suite.Require().NoError(err)
	suite.Equal(2.0, next)
}

func (suite *LabelRepositorySuite) TestListScopesToWorkspaceAndTeams() {
	suite.label("Shared", nil, nil, false, 2)
	suite.label("Team", &suite.team.ID, nil, false, 1)
	archived := suite.label("Old", nil, nil, false, 3)
	_, err := suite.repo.SetArchived(suite.Ctx, archived.ID, new(time.Now()))
	suite.Require().NoError(err)

	workspaceOnly, err := suite.repo.List(suite.Ctx, suite.workspace.ID, nil, false)
	suite.Require().NoError(err)
	suite.Len(workspaceOnly, 1)

	withTeam, err := suite.repo.List(suite.Ctx, suite.workspace.ID, []uuid.UUID{suite.team.ID}, true)
	suite.Require().NoError(err)
	suite.Equal([]string{"Team", "Shared", "Old"}, []string{withTeam[0].Name, withTeam[1].Name, withTeam[2].Name})
}

func (suite *LabelRepositorySuite) TestDeleteTakesGroupsLabelsAndIssueLinks() {
	group := suite.label("Type", nil, nil, true, 1)
	child := suite.label("Bug", nil, &group.ID, false, 1)
	state := suite.DB.WorkflowState.Create().SetTeamID(suite.team.ID).SetName("Todo").SetType(schema.StateTypeUnstarted).SetColor("#ffffff").SaveX(suite.Ctx)
	tagged := suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(suite.team.ID).SetNumber(1).SetTitle("Crash").SetStateID(state.ID).AddLabelIDs(child.ID).SaveX(suite.Ctx)

	byIssue, err := suite.repo.IssuesWithLabels(suite.Ctx, nil, []uuid.UUID{group.ID, child.ID})
	suite.Require().NoError(err)
	suite.Require().Len(byIssue[tagged.ID], 1)
	suite.Equal(child.ID, byIssue[tagged.ID][0].ID)

	suite.Require().NoError(suite.repo.Delete(suite.Ctx, nil, []uuid.UUID{group.ID, child.ID}))
	suite.Zero(suite.DB.Label.Query().CountX(suite.Ctx))
	suite.Zero(suite.DB.Issue.QueryLabels(suite.DB.Issue.GetX(suite.Ctx, tagged.ID)).CountX(suite.Ctx))
}

func TestLabelRepositorySuite(t *testing.T) {
	suite.Run(t, new(LabelRepositorySuite))
}

func (suite *LabelRepositorySuite) TestDroppingALabelOntoANeighbourRespacesTheList() {
	bug := suite.label("Bug", nil, nil, false, 1)
	chore := suite.label("Chore", nil, nil, false, 2)
	docs := suite.label("Docs", nil, nil, false, 3)
	moved, err := suite.repo.Update(suite.Ctx, nil, docs.ID, &UpdateLabelInput{SortOrder: new(1.0)})
	suite.Require().NoError(err)
	suite.Equal(2048.0, moved.SortOrder, "the response carries the respaced value")
	suite.Equal(1024.0, suite.DB.Label.GetX(suite.Ctx, bug.ID).SortOrder)
	suite.Equal(3072.0, suite.DB.Label.GetX(suite.Ctx, chore.ID).SortOrder)

	untouched, err := suite.repo.Update(suite.Ctx, nil, chore.ID, &UpdateLabelInput{SortOrder: new(4000.0)})
	suite.Require().NoError(err)
	suite.Equal(4000.0, untouched.SortOrder)
}
