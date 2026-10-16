package user_service

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent/teammember"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
)

type BootstrapDatabaseSuite struct {
	repositorytest.RepositoryBaseSuite
	service *UserService
}

func (suite *BootstrapDatabaseSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.service = NewUserService(&config.Config{}, repositories.NewRepositories(suite.DB), nil)
}

func (suite *BootstrapDatabaseSuite) TestBootstrapAddsOneDefaultTeamWithStatesAndTheAdmin() {
	result, err := suite.service.Bootstrap(suite.Ctx, bootstrapInput())
	suite.Require().NoError(err)

	teams := suite.DB.Team.Query().AllX(suite.Ctx)
	suite.Require().Len(teams, 1)
	suite.Equal(result.Team.ID, teams[0].ID)
	suite.Equal(result.Workspace.ID, teams[0].WorkspaceID)
	suite.Equal("Acme", teams[0].Name)
	suite.Equal("ACM", teams[0].Key)

	states := suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(teams[0].ID)).CountX(suite.Ctx)
	suite.Equal(len(team_repo.DefaultStates), states)

	member := suite.DB.TeamMember.Query().Where(teammember.TeamID(teams[0].ID)).OnlyX(suite.Ctx)
	suite.Equal(result.User.ID, member.UserID)
	suite.True(member.Owner)
}

func TestBootstrapDatabaseSuite(t *testing.T) {
	suite.Run(t, new(BootstrapDatabaseSuite))
}
