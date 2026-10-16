package favorite_service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	favorite_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/favorite"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type FavoriteServiceSuite struct {
	services.ServiceTestSuite
	service     *FavoriteService
	userID      uuid.UUID
	workspaceID uuid.UUID
}

func (suite *FavoriteServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewFavoriteService(nil, suite.MockRepo)
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.ExpectTx()
}

func (suite *FavoriteServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *FavoriteServiceSuite) TestOutsidersSeeNoWorkspace() {
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListFavoritesInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *FavoriteServiceSuite) TestListLeavesOutGoneAndHiddenEntities() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	private := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Private: true}
	project := &ent.Favorite{ID: uuid.New(), Kind: schema.FavoriteKindProject, EntityID: uuid.New()}
	gone := &ent.Favorite{ID: uuid.New(), Kind: schema.FavoriteKindProject, EntityID: uuid.New()}
	secretCycle := &ent.Favorite{ID: uuid.New(), Kind: schema.FavoriteKindCycle, EntityID: uuid.New()}
	othersView := &ent.Favorite{ID: uuid.New(), Kind: schema.FavoriteKindView, EntityID: uuid.New()}
	otherOwner := uuid.New()
	suite.MockFavoriteRepo.EXPECT().ListByUser(suite.Ctx, suite.userID, suite.workspaceID).Return([]*ent.Favorite{project, gone, secretCycle, othersView}, nil).Once()
	suite.MockFavoriteRepo.EXPECT().Resolve(suite.Ctx, schema.FavoriteKindProject, []uuid.UUID{project.EntityID, gone.EntityID}).Return(map[uuid.UUID]favorite_repo.ResolvedEntity{
		project.EntityID: {WorkspaceID: suite.workspaceID, Name: "Launch", URLParams: map[string]string{"slug": "launch-1a2b3c4d"}},
	}, nil).Once()
	suite.MockFavoriteRepo.EXPECT().Resolve(suite.Ctx, schema.FavoriteKindCycle, []uuid.UUID{secretCycle.EntityID}).Return(map[uuid.UUID]favorite_repo.ResolvedEntity{
		secretCycle.EntityID: {WorkspaceID: suite.workspaceID, TeamID: &private.ID, Name: "Cycle 1"},
	}, nil).Once()
	suite.MockFavoriteRepo.EXPECT().Resolve(suite.Ctx, schema.FavoriteKindView, []uuid.UUID{othersView.EntityID}).Return(map[uuid.UUID]favorite_repo.ResolvedEntity{
		othersView.EntityID: {WorkspaceID: suite.workspaceID, Name: "Theirs", ViewOwnerID: &otherOwner},
	}, nil).Once()
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, private.ID).Return(private, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, private.ID, suite.userID).Return(false, nil).Once()

	favorites, err := suite.service.List(suite.Ctx, suite.userID, &ListFavoritesInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Require().Len(favorites, 1)
	suite.Equal("Launch", favorites[0].Name)
	suite.Equal("launch-1a2b3c4d", favorites[0].URLParams["slug"])
}

func (suite *FavoriteServiceSuite) TestAddRefusesEntitiesOfOtherWorkspaces() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	entityID := uuid.New()
	suite.MockFavoriteRepo.EXPECT().Resolve(suite.Ctx, schema.FavoriteKindDocument, []uuid.UUID{entityID}).Return(map[uuid.UUID]favorite_repo.ResolvedEntity{
		entityID: {WorkspaceID: uuid.New(), Name: "Elsewhere"},
	}, nil).Once()
	_, err := suite.service.Add(suite.Ctx, suite.userID, &FavoriteTargetInput{WorkspaceID: suite.workspaceID, Kind: schema.FavoriteKindDocument, EntityID: entityID})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	_, err = suite.service.Add(suite.Ctx, suite.userID, &FavoriteTargetInput{WorkspaceID: suite.workspaceID, Kind: "comment", EntityID: entityID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *FavoriteServiceSuite) TestAddIsIdempotentAndAppends() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	entityID := uuid.New()
	suite.MockFavoriteRepo.EXPECT().Resolve(suite.Ctx, schema.FavoriteKindProject, []uuid.UUID{entityID}).Return(map[uuid.UUID]favorite_repo.ResolvedEntity{
		entityID: {WorkspaceID: suite.workspaceID, Name: "Launch"},
	}, nil)
	existing := &ent.Favorite{ID: uuid.New(), Kind: schema.FavoriteKindProject, EntityID: entityID, SortOrder: 1}
	suite.MockFavoriteRepo.EXPECT().Find(suite.Ctx, mock.Anything, suite.userID, schema.FavoriteKindProject, entityID).Return(existing, nil).Once()
	favorite, err := suite.service.Add(suite.Ctx, suite.userID, &FavoriteTargetInput{WorkspaceID: suite.workspaceID, Kind: schema.FavoriteKindProject, EntityID: entityID})
	suite.Require().NoError(err)
	suite.Equal(existing.ID, favorite.ID)

	suite.MockFavoriteRepo.EXPECT().Find(suite.Ctx, mock.Anything, suite.userID, schema.FavoriteKindProject, entityID).Return(nil, nil).Once()
	suite.MockFavoriteRepo.EXPECT().NextSortOrder(suite.Ctx, mock.Anything, suite.userID, suite.workspaceID).Return(5, nil).Once()
	suite.MockFavoriteRepo.EXPECT().Create(suite.Ctx, mock.Anything, &favorite_repo.CreateFavoriteInput{
		UserID: suite.userID, WorkspaceID: suite.workspaceID, Kind: schema.FavoriteKindProject, EntityID: entityID, SortOrder: 5,
	}).Return(&ent.Favorite{ID: uuid.New(), Kind: schema.FavoriteKindProject, EntityID: entityID, SortOrder: 5}, nil).Once()
	favorite, err = suite.service.Add(suite.Ctx, suite.userID, &FavoriteTargetInput{WorkspaceID: suite.workspaceID, Kind: schema.FavoriteKindProject, EntityID: entityID})
	suite.Require().NoError(err)
	suite.Equal(5.0, favorite.SortOrder)
	suite.Equal("Launch", favorite.Name)
}

func (suite *FavoriteServiceSuite) TestRemoveOfAMissingFavoriteIsNotFound() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	entityID := uuid.New()
	suite.MockFavoriteRepo.EXPECT().Delete(suite.Ctx, suite.userID, suite.workspaceID, schema.FavoriteKindIssue, entityID).Return(0, nil).Once()
	err := suite.service.Remove(suite.Ctx, suite.userID, &FavoriteTargetInput{WorkspaceID: suite.workspaceID, Kind: schema.FavoriteKindIssue, EntityID: entityID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *FavoriteServiceSuite) TestReorderReturnsTheNewOrder() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	suite.MockFavoriteRepo.EXPECT().Reorder(suite.Ctx, suite.userID, suite.workspaceID, ids).Return(nil).Once()
	suite.MockFavoriteRepo.EXPECT().ListByUser(suite.Ctx, suite.userID, suite.workspaceID).Return([]*ent.Favorite{}, nil).Once()
	favorites, err := suite.service.Reorder(suite.Ctx, suite.userID, &ReorderFavoritesInput{WorkspaceID: suite.workspaceID, FavoriteIDs: ids})
	suite.Require().NoError(err)
	suite.Empty(favorites)
}

func TestFavoriteServiceSuite(t *testing.T) {
	suite.Run(t, new(FavoriteServiceSuite))
}
