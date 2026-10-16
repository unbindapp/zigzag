package favorite_repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type FavoriteRepositorySuite struct {
	repositorytest.RepositoryBaseSuite
	repo  *FavoriteRepository
	user  *ent.User
	acme  *ent.Workspace
	other *ent.Workspace
}

func (suite *FavoriteRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewFavoriteRepository(suite.DB)
	suite.user = suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.acme = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.other = suite.DB.Workspace.Create().SetName("Other").SetSlug("other").SaveX(suite.Ctx)
}

func (suite *FavoriteRepositorySuite) TestRemoveAndReorderStayInTheNamedWorkspace() {
	entityID := uuid.New()
	favorite, err := suite.repo.Create(suite.Ctx, nil, &CreateFavoriteInput{UserID: suite.user.ID, WorkspaceID: suite.acme.ID, Kind: schema.FavoriteKindIssue, EntityID: entityID, SortOrder: 5})
	suite.Require().NoError(err)

	removed, err := suite.repo.Delete(suite.Ctx, suite.user.ID, suite.other.ID, schema.FavoriteKindIssue, entityID)
	suite.Require().NoError(err)
	suite.Zero(removed, "a call scoped to another workspace leaves the favorite alone")

	suite.Require().NoError(suite.repo.Reorder(suite.Ctx, suite.user.ID, suite.other.ID, []uuid.UUID{favorite.ID}))
	suite.Equal(5.0, suite.DB.Favorite.GetX(suite.Ctx, favorite.ID).SortOrder)

	removed, err = suite.repo.Delete(suite.Ctx, suite.user.ID, suite.acme.ID, schema.FavoriteKindIssue, entityID)
	suite.Require().NoError(err)
	suite.Equal(1, removed)
}

func TestFavoriteRepositorySuite(t *testing.T) {
	suite.Run(t, new(FavoriteRepositorySuite))
}
