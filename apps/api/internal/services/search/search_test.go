package search_service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
)

type fixedTeams struct {
	visible *issue_service.VisibleTeams
}

func (self *fixedTeams) VisibleTeams(context.Context, uuid.UUID, uuid.UUID) (*issue_service.VisibleTeams, error) {
	return self.visible, nil
}

type SearchServiceSuite struct {
	services.ServiceTestSuite
	service     *SearchService
	userID      uuid.UUID
	workspaceID uuid.UUID
	team        *ent.Team
}

func (suite *SearchServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.team = &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "ENG"}
	visible := &issue_service.VisibleTeams{WorkspaceID: suite.workspaceID, Teams: map[uuid.UUID]*ent.Team{suite.team.ID: suite.team}}
	suite.service = NewSearchService(&config.Config{}, suite.MockRepo, &fixedTeams{visible: visible})
}

func (suite *SearchServiceSuite) viewer(result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: suite.workspaceID}}).Return(result).Once()
}

func (suite *SearchServiceSuite) TestMasksWorkspacesOutOfReach() {
	suite.viewer(errdefs.ErrUnauthorized)
	_, err := suite.service.Search(suite.Ctx, suite.userID, &SearchInput{WorkspaceID: suite.workspaceID, Query: "login"})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *SearchServiceSuite) TestGroupsIssuesProjectsAndDocuments() {
	suite.viewer(nil)
	issue := &ent.Issue{ID: uuid.New(), Number: 7, Title: "Fix login", Edges: ent.IssueEdges{Team: suite.team, State: &ent.WorkflowState{Name: "Todo", Type: schema.StateTypeUnstarted}}}
	suite.MockIssueRepo.EXPECT().Search(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.team.ID}, "login", defaultLimit).Return([]uuid.UUID{issue.ID}, nil).Once()
	suite.MockIssueRepo.EXPECT().GetManyWithEdges(suite.Ctx, mock.Anything, []uuid.UUID{issue.ID}).Return([]*ent.Issue{issue}, nil).Once()
	suite.MockProjectRepo.EXPECT().Search(suite.Ctx, []uuid.UUID{suite.workspaceID}, "login", defaultLimit).Return([]*ent.Project{{ID: uuid.New(), Name: "Login revamp", Slug: "login-revamp-1a2b3c4d"}}, nil).Once()
	suite.MockDocumentRepo.EXPECT().Search(suite.Ctx, []uuid.UUID{suite.workspaceID}, "login", defaultLimit).Return([]*ent.Document{{ID: uuid.New(), Title: "Login spec", Slug: "login-spec"}}, nil).Once()

	result, err := suite.service.Search(suite.Ctx, suite.userID, &SearchInput{WorkspaceID: suite.workspaceID, Query: "login"})
	suite.Require().NoError(err)
	suite.Require().Len(result.Issues, 1)
	suite.Equal("ENG-7", result.Issues[0].Identifier)
	suite.Equal("Todo", result.Issues[0].State.Name)
	suite.Require().Len(result.Projects, 1)
	suite.Equal("login-revamp-1a2b3c4d", result.Projects[0].Slug)
	suite.Require().Len(result.Documents, 1)
	suite.Equal("login-spec", result.Documents[0].Slug)
}

func (suite *SearchServiceSuite) TestPassesTheLimit() {
	suite.viewer(nil)
	suite.MockIssueRepo.EXPECT().Search(suite.Ctx, suite.workspaceID, mock.Anything, "x", 3).Return([]uuid.UUID{}, nil).Once()
	suite.MockIssueRepo.EXPECT().GetManyWithEdges(suite.Ctx, mock.Anything, []uuid.UUID{}).Return([]*ent.Issue{}, nil).Once()
	suite.MockProjectRepo.EXPECT().Search(suite.Ctx, mock.Anything, "x", 3).Return([]*ent.Project{}, nil).Once()
	suite.MockDocumentRepo.EXPECT().Search(suite.Ctx, mock.Anything, "x", 3).Return([]*ent.Document{}, nil).Once()

	result, err := suite.service.Search(suite.Ctx, suite.userID, &SearchInput{WorkspaceID: suite.workspaceID, Query: "x", Limit: 3})
	suite.Require().NoError(err)
	suite.Empty(result.Issues)
	suite.Empty(result.Projects)
	suite.Empty(result.Documents)
}

func TestSearchServiceSuite(t *testing.T) {
	suite.Run(t, new(SearchServiceSuite))
}
