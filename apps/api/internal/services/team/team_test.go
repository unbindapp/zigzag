package team_service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type TeamServiceSuite struct {
	services.ServiceTestSuite
	service     *TeamService
	userID      uuid.UUID
	workspaceID uuid.UUID
	team        *ent.Team
}

func (suite *TeamServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewTeamService(suite.MockRepo)
	suite.service.now = func() time.Time { return time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC) }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.team = &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Engineering", Key: "ENG"}
	suite.ExpectTx()
}

func (suite *TeamServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *TeamServiceSuite) expectRespond() {
	suite.MockTeamRepo.EXPECT().MemberCounts(suite.Ctx, mock.Anything).Return(map[uuid.UUID]int{suite.team.ID: 1}, nil).Maybe()
	suite.MockCycleRepo.EXPECT().ActiveCycleIDs(suite.Ctx, mock.Anything, mock.Anything).Return(map[uuid.UUID]uuid.UUID{}, nil).Maybe()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, mock.Anything, suite.userID).Return(true, nil).Maybe()
}

func (suite *TeamServiceSuite) TestCreateNeedsAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTeamInput{WorkspaceID: suite.workspaceID, Name: "Eng", Key: "ENG"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *TeamServiceSuite) TestCreateValidatesAndRejectsTakenKeys() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)

	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTeamInput{WorkspaceID: suite.workspaceID, Name: "Eng", Key: "1BAD"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTeamInput{WorkspaceID: suite.workspaceID, Name: "Eng", Key: "ENG", Timezone: "Mars/Olympus"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockTeamRepo.EXPECT().KeyOrNameTaken(suite.Ctx, mock.Anything, suite.workspaceID, "ENG", "Eng", (*uuid.UUID)(nil)).Return(true, false, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTeamInput{WorkspaceID: suite.workspaceID, Name: "Eng", Key: "eng"})
	suite.ErrorIs(err, errdefs.ErrConflict)
}

func (suite *TeamServiceSuite) TestCreateAddsTheCreator() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.expectRespond()
	suite.MockTeamRepo.EXPECT().KeyOrNameTaken(suite.Ctx, mock.Anything, suite.workspaceID, "ENG", "Engineering", (*uuid.UUID)(nil)).Return(false, false, nil).Once()
	suite.MockTeamRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *team_repo.CreateTeamInput) bool {
		return input.Key == "ENG" && input.CreatorID == suite.userID && input.WorkspaceID == suite.workspaceID
	})).Return(suite.team, nil).Once()

	resp, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTeamInput{WorkspaceID: suite.workspaceID, Name: "Engineering", Key: "eng"})
	suite.Require().NoError(err)
	suite.True(resp.IsMember)
	suite.Equal(1, resp.MemberCount)
	suite.NotNil(resp.States)
}

func (suite *TeamServiceSuite) TestPrivateTeamsHideFromNonMembers() {
	private := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Private: true}
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, private.ID).Return(private, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, private.ID, suite.userID).Return(false, nil).Once()

	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetTeamInput{TeamID: private.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *TeamServiceSuite) TestListSkipsPrivateTeams() {
	private := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Private: true}
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockTeamRepo.EXPECT().ListByWorkspace(suite.Ctx, suite.workspaceID, false).Return([]*ent.Team{suite.team, private}, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, private.ID, suite.userID).Return(false, nil).Once()
	suite.expectRespond()

	teams, err := suite.service.List(suite.Ctx, suite.userID, &ListTeamsInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Require().Len(teams, 1)
	suite.Equal(suite.team.ID, teams[0].ID)
}

func (suite *TeamServiceSuite) TestMembersJoinOpenTeamsThemselves() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleMember, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil)
	suite.MockWorkspaceRepo.EXPECT().GetMember(suite.Ctx, mock.Anything, suite.workspaceID, suite.userID).Return(&ent.WorkspaceMember{}, nil).Once()
	suite.MockTeamRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, suite.team.ID, suite.userID, false).Return(nil).Once()
	suite.MockTeamRepo.EXPECT().GetMember(suite.Ctx, suite.team.ID, suite.userID).Return(&ent.TeamMember{Edges: ent.TeamMemberEdges{User: &ent.User{ID: suite.userID}}}, nil).Once()
	_, err := suite.service.AddMember(suite.Ctx, suite.userID, &AddTeamMemberInput{TeamID: suite.team.ID, UserID: suite.userID})
	suite.NoError(err)
}

func (suite *TeamServiceSuite) TestMembersCannotAddOthers() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	_, err := suite.service.AddMember(suite.Ctx, suite.userID, &AddTeamMemberInput{TeamID: suite.team.ID, UserID: uuid.New()})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *TeamServiceSuite) TestAddedUserMustBeInTheWorkspace() {
	other := uuid.New()
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().GetMember(suite.Ctx, mock.Anything, suite.workspaceID, other).Return(nil, &ent.NotFoundError{}).Once()
	_, err := suite.service.AddMember(suite.Ctx, suite.userID, &AddTeamMemberInput{TeamID: suite.team.ID, UserID: other})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *TeamServiceSuite) TestCreateStateValidates() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil)

	_, err := suite.service.CreateState(suite.Ctx, suite.userID, &CreateStateInput{TeamID: suite.team.ID, Name: "QA", Type: "review", Color: "#ffffff"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockTeamRepo.EXPECT().StateNameTaken(suite.Ctx, suite.team.ID, "Done", (*uuid.UUID)(nil)).Return(true, nil).Once()
	_, err = suite.service.CreateState(suite.Ctx, suite.userID, &CreateStateInput{TeamID: suite.team.ID, Name: "Done", Type: schema.StateTypeCompleted, Color: "#ffffff"})
	suite.ErrorIs(err, errdefs.ErrConflict)

	suite.MockTeamRepo.EXPECT().StateNameTaken(suite.Ctx, suite.team.ID, "QA", (*uuid.UUID)(nil)).Return(false, nil)
	_, err = suite.service.CreateState(suite.Ctx, suite.userID, &CreateStateInput{TeamID: suite.team.ID, Name: "QA", Type: schema.StateTypeStarted, Color: "red"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockTeamRepo.EXPECT().NextStatePosition(suite.Ctx, suite.team.ID, schema.StateTypeStarted).Return(2, nil).Once()
	suite.MockTeamRepo.EXPECT().CreateState(suite.Ctx, &team_repo.CreateStateInput{TeamID: suite.team.ID, Name: "QA", Type: schema.StateTypeStarted, Color: "#ffffff", Position: 2}).
		Return(&ent.WorkflowState{ID: uuid.New(), Name: "QA"}, nil).Once()
	state, err := suite.service.CreateState(suite.Ctx, suite.userID, &CreateStateInput{TeamID: suite.team.ID, Name: "QA", Type: schema.StateTypeStarted, Color: "#FFFFFF"})
	suite.Require().NoError(err)
	suite.Equal("QA", state.Name)
}

func (suite *TeamServiceSuite) TestDeleteStateNeedsAnotherStateOfTheTeam() {
	stateID := uuid.New()
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockTeamRepo.EXPECT().GetState(suite.Ctx, stateID).Return(&ent.WorkflowState{ID: stateID, TeamID: suite.team.ID}, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil)

	suite.ErrorIs(suite.service.DeleteState(suite.Ctx, suite.userID, &DeleteStateInput{StateID: stateID, MoveIssuesToStateID: stateID}), errdefs.ErrInvalidInput)

	foreign := uuid.New()
	suite.MockTeamRepo.EXPECT().GetState(suite.Ctx, foreign).Return(&ent.WorkflowState{ID: foreign, TeamID: uuid.New()}, nil).Once()
	suite.ErrorIs(suite.service.DeleteState(suite.Ctx, suite.userID, &DeleteStateInput{StateID: stateID, MoveIssuesToStateID: foreign}), errdefs.ErrInvalidInput)

	sibling := uuid.New()
	suite.MockTeamRepo.EXPECT().GetState(suite.Ctx, sibling).Return(&ent.WorkflowState{ID: sibling, TeamID: suite.team.ID}, nil).Once()
	suite.MockTeamRepo.EXPECT().DeleteState(suite.Ctx, stateID, sibling).Return(nil).Once()
	suite.NoError(suite.service.DeleteState(suite.Ctx, suite.userID, &DeleteStateInput{StateID: stateID, MoveIssuesToStateID: sibling}))
}

func TestTeamServiceSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceSuite))
}
