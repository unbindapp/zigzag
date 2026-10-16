package team_repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/ent/workflowstate"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type TeamRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo      *TeamRepository
	user      *ent.User
	workspace *ent.Workspace
}

func (suite *TeamRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewTeamRepository(suite.DB)
	suite.user = suite.DB.User.Create().SetEmail("ada@example.com").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
}

func (suite *TeamRepositorySuite) team(key string, triage bool) *ent.Team {
	created, err := suite.repo.Create(suite.Ctx, nil, &CreateTeamInput{WorkspaceID: suite.workspace.ID, Name: key, Key: key, TriageEnabled: triage, CreatorID: suite.user.ID})
	suite.Require().NoError(err)
	return created
}

func (suite *TeamRepositorySuite) TestCreateSeedsDefaultStates() {
	created := suite.team("ENG", false)
	names := []string{}
	for _, state := range created.Edges.States {
		names = append(names, state.Name)
	}
	suite.Equal([]string{"Backlog", "Todo", "In Progress", "In Review", "Done", "Canceled", "Duplicate"}, names, "ordered by type then position")
	suite.Require().NotNil(created.DefaultStateID)
	todo := suite.DB.WorkflowState.GetX(suite.Ctx, *created.DefaultStateID)
	suite.Equal("Todo", todo.Name)
	suite.Nil(created.TriageStateID)
	suite.Equal("UTC", created.Timezone)

	owner, err := suite.repo.GetMember(suite.Ctx, created.ID, suite.user.ID)
	suite.Require().NoError(err)
	suite.True(owner.Owner, "the creator joins as owner")
}

func (suite *TeamRepositorySuite) TestCreateWithTriage() {
	created := suite.team("OPS", true)
	suite.Equal("Triage", created.Edges.States[0].Name)
	suite.Require().NotNil(created.TriageStateID)
	suite.Equal(created.Edges.States[0].ID, *created.TriageStateID)
}

func (suite *TeamRepositorySuite) TestKeyAndNameUniqueness() {
	created := suite.team("ENG", false)
	keyTaken, nameTaken, err := suite.repo.KeyOrNameTaken(suite.Ctx, nil, suite.workspace.ID, "ENG", "Other", nil)
	suite.NoError(err)
	suite.True(keyTaken)
	suite.False(nameTaken)
	keyTaken, _, _ = suite.repo.KeyOrNameTaken(suite.Ctx, nil, suite.workspace.ID, "ENG", "", &created.ID)
	suite.False(keyTaken, "a team does not collide with itself")
	_, err = suite.repo.Create(suite.Ctx, nil, &CreateTeamInput{WorkspaceID: suite.workspace.ID, Name: "x", Key: "ENG"})
	suite.True(ent.IsConstraintError(err))
}

func (suite *TeamRepositorySuite) TestNextIssueNumberCountsUp() {
	created := suite.team("ENG", false)
	for want := 1; want <= 3; want++ {
		got, err := suite.repo.NextIssueNumber(suite.Ctx, nil, created.ID)
		suite.Require().NoError(err)
		suite.Equal(want, got)
	}
	_, err := suite.repo.NextIssueNumber(suite.Ctx, nil, uuid.New())
	suite.True(ent.IsNotFound(err))
}

func (suite *TeamRepositorySuite) TestDeleteStateMovesIssuesAndDefaults() {
	created := suite.team("ENG", false)
	todo := *created.DefaultStateID
	backlog := suite.DB.WorkflowState.Query().Where(workflowstate.TeamID(created.ID), workflowstate.Name("Backlog")).OnlyX(suite.Ctx)
	issue := suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(created.ID).SetNumber(1).SetTitle("t").SetStateID(todo).SaveX(suite.Ctx)

	suite.Require().NoError(suite.repo.DeleteState(suite.Ctx, todo, backlog.ID))
	suite.Equal(backlog.ID, suite.DB.Issue.GetX(suite.Ctx, issue.ID).StateID)
	suite.Equal(backlog.ID, *suite.DB.Team.GetX(suite.Ctx, created.ID).DefaultStateID)
	_, err := suite.repo.GetState(suite.Ctx, todo)
	suite.True(ent.IsNotFound(err))
}

func (suite *TeamRepositorySuite) TestStatesHelpers() {
	created := suite.team("ENG", false)
	next, err := suite.repo.NextStatePosition(suite.Ctx, created.ID, schema.StateTypeStarted)
	suite.NoError(err)
	suite.Equal(2.0, next)
	next, _ = suite.repo.NextStatePosition(suite.Ctx, created.ID, schema.StateTypeTriage)
	suite.Equal(0.0, next)

	taken, _ := suite.repo.StateNameTaken(suite.Ctx, created.ID, "done", nil)
	suite.True(taken, "state names compare without case")

	state, err := suite.repo.CreateState(suite.Ctx, &CreateStateInput{TeamID: created.ID, Name: "QA", Type: schema.StateTypeStarted, Color: "#ffffff", Position: next})
	suite.Require().NoError(err)
	updated, err := suite.repo.UpdateState(suite.Ctx, state.ID, &UpdateStateInput{Name: utils.ToPtr("Testing")})
	suite.Require().NoError(err)
	suite.Equal("Testing", updated.Name)
	count, _ := suite.repo.CountStates(suite.Ctx, created.ID)
	suite.Equal(8, count)
}

func (suite *TeamRepositorySuite) TestMembersAndCounts() {
	created := suite.team("ENG", false)
	other := suite.DB.User.Create().SetEmail("bob@example.com").SetDisplayName("bob").SaveX(suite.Ctx)
	suite.NoError(suite.repo.AddMember(suite.Ctx, nil, created.ID, other.ID, false))
	suite.NoError(suite.repo.AddMember(suite.Ctx, nil, created.ID, other.ID, true), "adding twice is a no-op")

	counts, err := suite.repo.MemberCounts(suite.Ctx, []uuid.UUID{created.ID})
	suite.NoError(err)
	suite.Equal(2, counts[created.ID])
	isMember, _ := suite.repo.IsMember(suite.Ctx, created.ID, other.ID)
	suite.True(isMember)

	suite.NoError(suite.repo.RemoveMember(suite.Ctx, created.ID, other.ID))
	suite.True(ent.IsNotFound(suite.repo.RemoveMember(suite.Ctx, created.ID, other.ID)))
	workspaceID, err := suite.repo.ResolveWorkspaceOfTeam(suite.Ctx, created.ID)
	suite.NoError(err)
	suite.Equal(suite.workspace.ID, workspaceID)
}

func TestTeamRepositorySuite(t *testing.T) {
	suite.Run(t, new(TeamRepositorySuite))
}
