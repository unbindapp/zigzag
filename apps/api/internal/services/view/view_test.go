package view_service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	view_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/view"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type ViewServiceSuite struct {
	services.ServiceTestSuite
	service     *ViewService
	userID      uuid.UUID
	otherID     uuid.UUID
	workspaceID uuid.UUID
	display     schema.ViewDisplay
}

func (suite *ViewServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewViewService(nil, suite.MockRepo)
	suite.userID = uuid.New()
	suite.otherID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.display = schema.ViewDisplay{Layout: "list", GroupBy: "state", OrderBy: "priority"}
}

func (suite *ViewServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *ViewServiceSuite) view(ownerID uuid.UUID, shared bool) *ent.View {
	found := &ent.View{ID: uuid.New(), WorkspaceID: suite.workspaceID, OwnerID: ownerID, Shared: shared, Name: "Mine", Kind: schema.ViewKindIssue, Display: suite.display}
	suite.MockViewRepo.EXPECT().GetByID(suite.Ctx, found.ID).Return(found, nil).Maybe()
	return found
}

func (suite *ViewServiceSuite) TestPrivateViewsOfOthersAreNotFound() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	private := suite.view(suite.otherID, false)
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetViewInput{ViewID: private.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	shared := suite.view(suite.otherID, true)
	view, err := suite.service.Get(suite.Ctx, suite.userID, &GetViewInput{ViewID: shared.ID})
	suite.Require().NoError(err)
	suite.False(view.IsOwner)
	suite.NotNil(view.Filters)
}

func (suite *ViewServiceSuite) TestCreateNeedsMember() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleMember, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateViewInput{WorkspaceID: suite.workspaceID, Name: "Mine", Kind: schema.ViewKindIssue, Display: suite.display})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ViewServiceSuite) TestCreateValidatesFiltersAndDisplay() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleMember, nil)
	cases := []*CreateViewInput{
		{WorkspaceID: suite.workspaceID, Name: "Mine", Kind: "board", Display: suite.display},
		{WorkspaceID: suite.workspaceID, Name: "Mine", Kind: schema.ViewKindIssue, Filters: map[string]any{"workspace_id": "x"}, Display: suite.display},
		{WorkspaceID: suite.workspaceID, Name: "Mine", Kind: schema.ViewKindIssue, Display: schema.ViewDisplay{Layout: "timeline"}},
		{WorkspaceID: suite.workspaceID, Name: "Mine", Kind: schema.ViewKindIssue, Display: schema.ViewDisplay{Layout: "list", GroupBy: "mood"}},
		{WorkspaceID: suite.workspaceID, Name: "Mine", Kind: schema.ViewKindIssue, Color: "blue", Display: suite.display},
	}
	for _, input := range cases {
		_, err := suite.service.Create(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	}
}

func (suite *ViewServiceSuite) TestCreateOwnsTheView() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleMember, nil)
	filters := map[string]any{"state_types": []any{"started"}}
	suite.MockViewRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *view_repo.CreateViewInput) bool {
		return input.OwnerID == suite.userID && input.Name == "In progress" && input.Shared && input.Filters["state_types"] != nil
	})).Return(&ent.View{ID: uuid.New(), OwnerID: suite.userID, Name: "In progress", Shared: true, Filters: filters}, nil).Once()

	view, err := suite.service.Create(suite.Ctx, suite.userID, &CreateViewInput{
		WorkspaceID: suite.workspaceID, Name: " In progress ", Kind: schema.ViewKindIssue, Shared: true, Filters: filters, Display: suite.display,
	})
	suite.Require().NoError(err)
	suite.True(view.IsOwner)
}

func (suite *ViewServiceSuite) TestOwnersEditTheirViewsAndAdminsEditSharedOnes() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleMember, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	name := "Renamed"

	mine := suite.view(suite.userID, false)
	suite.MockViewRepo.EXPECT().Update(suite.Ctx, mine.ID, &view_repo.UpdateViewInput{Name: &name}).Return(mine, nil).Once()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateViewInput{ViewID: mine.ID, Name: &name})
	suite.NoError(err)

	theirs := suite.view(suite.otherID, true)
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateViewInput{ViewID: theirs.ID, Name: &name})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.service.Delete(suite.Ctx, suite.userID, &DeleteViewInput{ViewID: theirs.ID}), errdefs.ErrUnauthorized)
}

func (suite *ViewServiceSuite) TestAdminsCannotMakeOthersViewsPrivate() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	theirs := suite.view(suite.otherID, true)
	private := false
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateViewInput{ViewID: theirs.ID, Shared: &private})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockViewRepo.EXPECT().Delete(suite.Ctx, theirs.ID).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &DeleteViewInput{ViewID: theirs.ID}))
}

func (suite *ViewServiceSuite) TestListHidesViewsOfPrivateTeams() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	private := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Private: true}
	hidden := &ent.View{ID: uuid.New(), WorkspaceID: suite.workspaceID, OwnerID: suite.otherID, Shared: true, TeamID: &private.ID}
	open := &ent.View{ID: uuid.New(), WorkspaceID: suite.workspaceID, OwnerID: suite.userID}
	suite.MockViewRepo.EXPECT().ListVisible(suite.Ctx, view_repo.ListViewsFilter{WorkspaceID: suite.workspaceID, ViewerID: suite.userID}).Return([]*ent.View{hidden, open}, nil).Once()
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, private.ID).Return(private, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, private.ID, suite.userID).Return(false, nil).Once()

	views, err := suite.service.List(suite.Ctx, suite.userID, &ListViewsInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Require().Len(views, 1)
	suite.Equal(open.ID, views[0].ID)
}

func (suite *ViewServiceSuite) TestIDsFromAnotherWorkspaceAreRefused() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleMember, nil)
	foreignTeam := &ent.Team{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, foreignTeam.ID).Return(foreignTeam, nil).Maybe()
	onForeignTeam := suite.view(suite.userID, false)
	onForeignTeam.TeamID = &foreignTeam.ID
	renamed := "Renamed"

	cases := map[string]func() error{
		"create with a team of another workspace": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateViewInput{WorkspaceID: suite.workspaceID, TeamID: &foreignTeam.ID, Name: "Mine", Kind: schema.ViewKindIssue, Display: suite.display})
			return err
		},
		"update a view tied to a team of another workspace": func() error {
			_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateViewInput{ViewID: onForeignTeam.ID, Name: &renamed})
			return err
		},
	}
	for name, run := range cases {
		suite.Run(name, func() {
			suite.ErrorIs(run(), errdefs.ErrNotFound)
		})
	}
}

func TestViewServiceSuite(t *testing.T) {
	suite.Run(t, new(ViewServiceSuite))
}
