package apikey_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/auth"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	apikey_repo "github.com/unbindapp/zigzag-api/internal/repositories/apikey"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type APIKeyServiceSuite struct {
	services.ServiceTestSuite
	service     *APIKeyService
	requester   uuid.UUID
	workspaceID uuid.UUID
}

func (suite *APIKeyServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewAPIKeyService(suite.MockRepo)
	suite.requester = uuid.New()
	suite.workspaceID = uuid.New()
	suite.MockAPIKeyRepo.EXPECT().GetNamesByUser(mock.Anything, mock.Anything, suite.requester).Return([]string{"taken"}, nil).Maybe()
}

func (suite *APIKeyServiceSuite) expectCheck(role schema.WorkspaceRole, workspaceID uuid.UUID, result error) {
	suite.MockPermissionsRepo.EXPECT().
		Check(suite.Ctx, suite.requester, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: workspaceID}}).
		Return(result).Once()
}

func (suite *APIKeyServiceSuite) expectCreate() *apikey_repo.CreateAPIKeyInput {
	stored := &apikey_repo.CreateAPIKeyInput{}
	suite.MockAPIKeyRepo.EXPECT().
		Create(suite.Ctx, mock.AnythingOfType("*apikey_repo.CreateAPIKeyInput")).
		RunAndReturn(func(_ context.Context, input *apikey_repo.CreateAPIKeyInput) (*ent.APIKey, error) {
			*stored = *input
			return &ent.APIKey{ID: uuid.New(), UserID: input.UserID, Name: input.Name, TokenPrefix: input.TokenPrefix, TokenHash: input.TokenHash, Role: input.Role, FullAccess: input.FullAccess, WorkspaceIds: input.WorkspaceIDs, Privileges: input.Privileges}, nil
		}).Once()
	return stored
}

func (suite *APIKeyServiceSuite) TestCreateStoresHashAndReturnsTokenOnce() {
	suite.expectCheck(schema.WorkspaceRoleViewer, suite.workspaceID, nil)
	stored := suite.expectCreate()

	resp, err := suite.service.Create(suite.Ctx, suite.requester, &models.APIKeyCreateInput{Name: "ci", Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{suite.workspaceID}})
	suite.Require().NoError(err)

	suite.True(auth.IsAPIKey(resp.Token))
	suite.Equal(auth.HashAPIKey(resp.Token), stored.TokenHash, "only the hash is persisted")
	suite.Equal(suite.requester, stored.UserID, "keys are always minted for the requester")
	suite.Equal(resp.TokenPrefix, stored.TokenPrefix)
	suite.NotContains(resp.TokenPrefix, resp.Token[len(resp.TokenPrefix):])
	suite.Equal(schema.WorkspaceRoleViewer, stored.Role)
	suite.False(stored.FullAccess)
	suite.Equal([]uuid.UUID{suite.workspaceID}, stored.WorkspaceIDs)
	suite.Equal([]uuid.UUID{suite.workspaceID}, resp.WorkspaceIDs)
	suite.NotNil(resp.Privileges)
}

func (suite *APIKeyServiceSuite) TestCreateChecksEveryWorkspaceAtTheKeyRole() {
	other := uuid.New()
	suite.expectCheck(schema.WorkspaceRoleMember, suite.workspaceID, nil)
	suite.expectCheck(schema.WorkspaceRoleMember, other, errdefs.ErrUnauthorized)

	_, err := suite.service.Create(suite.Ctx, suite.requester, &models.APIKeyCreateInput{Name: "ci", Role: schema.WorkspaceRoleMember, WorkspaceIDs: []uuid.UUID{suite.workspaceID, other}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	suite.Contains(err.Error(), "you do not have member access to workspace "+other.String())
}

func (suite *APIKeyServiceSuite) TestCreateFullAccessNeedsNoCheck() {
	stored := suite.expectCreate()
	resp, err := suite.service.Create(suite.Ctx, suite.requester, &models.APIKeyCreateInput{Name: "me", Role: schema.WorkspaceRoleAdmin, FullAccess: true})
	suite.Require().NoError(err)
	suite.True(stored.FullAccess)
	suite.NotNil(stored.WorkspaceIDs, "workspace ids are stored as an empty list, never null")
	suite.Empty(resp.WorkspaceIDs)
}

func (suite *APIKeyServiceSuite) TestCreateRejectsTakenName() {
	_, err := suite.service.Create(suite.Ctx, suite.requester, &models.APIKeyCreateInput{Name: " taken ", Role: schema.WorkspaceRoleAdmin, FullAccess: true})
	suite.ErrorIs(err, errdefs.ErrConflict)
	suite.ErrorContains(err, `An API key named "taken" already exists`)
}

func (suite *APIKeyServiceSuite) TestCreateRejectsMalformedInput() {
	past := time.Now().Add(-time.Minute)
	ws := []uuid.UUID{suite.workspaceID}
	tests := map[string]*models.APIKeyCreateInput{
		"expiry in the past":          {Name: "ci", ExpiresAt: &past, Role: schema.WorkspaceRoleViewer, WorkspaceIDs: ws},
		"unknown role":                {Name: "ci", Role: "owner", WorkspaceIDs: ws},
		"full access with workspaces": {Name: "ci", Role: schema.WorkspaceRoleViewer, FullAccess: true, WorkspaceIDs: ws},
		"scoped without workspaces":   {Name: "ci", Role: schema.WorkspaceRoleViewer},
		"duplicate workspace":         {Name: "ci", Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{suite.workspaceID, suite.workspaceID}},
		"unknown privilege":           {Name: "ci", Role: schema.WorkspaceRoleViewer, FullAccess: true, Privileges: []schema.KeyPrivilege{"read_everything"}},
	}
	for name, input := range tests {
		suite.Run(name, func() {
			_, err := suite.service.Create(suite.Ctx, suite.requester, input)
			suite.ErrorIs(err, errdefs.ErrInvalidInput)
		})
	}
}

func (suite *APIKeyServiceSuite) TestListOwnKeys() {
	suite.MockAPIKeyRepo.EXPECT().ListByUser(suite.Ctx, suite.requester).Return([]*ent.APIKey{{ID: uuid.New(), UserID: suite.requester}}, nil).Once()
	keys, err := suite.service.List(suite.Ctx, suite.requester, &models.APIKeyListInput{})
	suite.NoError(err)
	suite.Len(keys, 1)
	suite.NotNil(keys[0].WorkspaceIDs)
}

func (suite *APIKeyServiceSuite) TestListOtherUserRequiresSystemAdmin() {
	other := uuid.New()
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.requester).Return(errdefs.ErrUnauthorized).Once()
	_, err := suite.service.List(suite.Ctx, suite.requester, &models.APIKeyListInput{UserID: other})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)

	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.requester).Return(nil).Once()
	suite.MockAPIKeyRepo.EXPECT().ListByUser(suite.Ctx, other).Return(nil, nil).Once()
	keys, err := suite.service.List(suite.Ctx, suite.requester, &models.APIKeyListInput{UserID: other})
	suite.NoError(err)
	suite.Empty(keys)
}

func (suite *APIKeyServiceSuite) TestDeleteOwnKey() {
	id := uuid.New()
	suite.MockAPIKeyRepo.EXPECT().GetByID(suite.Ctx, id).Return(&ent.APIKey{ID: id, UserID: suite.requester}, nil).Once()
	suite.MockAPIKeyRepo.EXPECT().Delete(suite.Ctx, id).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.requester, id))
}

func (suite *APIKeyServiceSuite) TestDeleteOthersKeyMaskedUnlessSystemAdmin() {
	id := uuid.New()
	key := &ent.APIKey{ID: id, UserID: uuid.New()}
	suite.MockAPIKeyRepo.EXPECT().GetByID(suite.Ctx, id).Return(key, nil).Once()
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.requester).Return(errdefs.ErrUnauthorized).Once()
	err := suite.service.Delete(suite.Ctx, suite.requester, id)
	suite.ErrorIs(err, errdefs.ErrNotFound, "non-admins cannot tell someone else's key from a missing one")
	suite.False(errors.Is(err, errdefs.ErrUnauthorized))

	suite.MockAPIKeyRepo.EXPECT().GetByID(suite.Ctx, id).Return(key, nil).Once()
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.requester).Return(nil).Once()
	suite.MockAPIKeyRepo.EXPECT().Delete(suite.Ctx, id).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.requester, id))
}

func (suite *APIKeyServiceSuite) TestDeleteMissingKey() {
	id := uuid.New()
	suite.MockAPIKeyRepo.EXPECT().GetByID(suite.Ctx, id).Return(nil, &ent.NotFoundError{}).Once()
	suite.ErrorIs(suite.service.Delete(suite.Ctx, suite.requester, id), errdefs.ErrNotFound)
}

func TestAPIKeyServiceSuite(t *testing.T) {
	suite.Run(t, new(APIKeyServiceSuite))
}
