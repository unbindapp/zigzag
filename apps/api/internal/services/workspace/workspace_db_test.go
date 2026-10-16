package workspace_service

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/team"
	"github.com/unbindapp/zigzag-api/ent/teammember"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
)

type WorkspaceDatabaseSuite struct {
	repositorytest.RepositoryBaseSuite
	service *WorkspaceService
	creator *ent.User
}

func (suite *WorkspaceDatabaseSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.service = NewWorkspaceService(&config.Config{}, repositories.NewRepositories(suite.DB), nil)
	suite.creator = suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada").SetDisplayName("ada").SaveX(suite.Ctx)
}

func (suite *WorkspaceDatabaseSuite) TestCreateAddsOneDefaultTeamWithStatesAndTheCreator() {
	resp, err := suite.service.Create(suite.Ctx, suite.creator.ID, &CreateWorkspaceInput{Name: "Acme Robotics", Slug: "acme"})
	suite.Require().NoError(err)

	teams := suite.DB.Team.Query().Where(team.WorkspaceID(resp.ID)).AllX(suite.Ctx)
	suite.Require().Len(teams, 1)
	created := teams[0]
	suite.Equal("Acme Robotics", created.Name)
	suite.Equal("AR", created.Key)
	suite.Require().NotNil(resp.DefaultTeamID)
	suite.Equal(created.ID, *resp.DefaultTeamID)
	suite.Equal(1, resp.TeamCount)

	states := suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(created.ID)).CountX(suite.Ctx)
	suite.Equal(len(team_repo.DefaultStates), states)
	suite.NotNil(created.DefaultStateID)

	member := suite.DB.TeamMember.Query().Where(teammember.TeamID(created.ID)).OnlyX(suite.Ctx)
	suite.Equal(suite.creator.ID, member.UserID)
	suite.True(member.Owner)
}

func TestWorkspaceDatabaseSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceDatabaseSuite))
}
