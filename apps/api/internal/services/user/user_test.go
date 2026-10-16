package user_service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/mail"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	user_repo "github.com/unbindapp/zigzag-api/internal/repositories/user"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type UserServiceSuite struct {
	services.ServiceTestSuite
	service *UserService
	now     time.Time
	userID  uuid.UUID
}

func (suite *UserServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewUserService(&config.Config{ExternalUIUrl: "https://zigzag.example.com"}, suite.MockRepo, suite.MockEmail)
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID = uuid.New()
	suite.ExpectTx()
}

func (suite *UserServiceSuite) TestLoginRefusesInactiveAccounts() {
	suite.MockUserRepo.EXPECT().Authenticate(suite.Ctx, "ada@example.com", "pw").Return(&ent.User{ID: suite.userID, Active: false}, nil).Once()
	_, err := suite.service.Login(suite.Ctx, "ada@example.com", "pw")
	suite.ErrorIs(err, ErrInvalidCredentials)

	suite.MockUserRepo.EXPECT().Authenticate(suite.Ctx, "ada@example.com", "bad").Return(nil, user_repo.ErrInvalidPassword).Once()
	_, err = suite.service.Login(suite.Ctx, "ada@example.com", "bad")
	suite.ErrorIs(err, ErrInvalidCredentials, "wrong password and unknown account answer alike")
}

func (suite *UserServiceSuite) TestWhoamiListsWorkspacesWithRolesAndKeyLimits() {
	workspaceID := uuid.New()
	ctx := permissions_repo.WithAPIKeyAccess(suite.Ctx, permissions_repo.APIKeyAccess{Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{workspaceID}})
	suite.MockUserRepo.EXPECT().GetByID(ctx, suite.userID).Return(&ent.User{ID: suite.userID, Email: "ada@example.com", Active: true}, nil).Once()
	suite.MockPermissionsRepo.EXPECT().GetUserPermissionSet(ctx, suite.userID).Return(permissionSet(workspaceID, schema.WorkspaceRoleViewer), nil).Once()
	suite.MockWorkspaceRepo.EXPECT().GetByIDs(ctx, []uuid.UUID{workspaceID}).Return([]*ent.Workspace{{ID: workspaceID, Name: "Acme", Slug: "acme"}}, nil).Once()

	resp, err := suite.service.Whoami(ctx, suite.userID)
	suite.Require().NoError(err)
	suite.Require().Len(resp.Workspaces, 1)
	suite.Equal(schema.WorkspaceRoleViewer, resp.Workspaces[0].Role)
	suite.Require().NotNil(resp.APIKey)
	suite.Equal([]uuid.UUID{workspaceID}, resp.APIKey.WorkspaceIDs)
	suite.NotNil(resp.APIKey.Privileges)
}

func (suite *UserServiceSuite) TestRequestPasswordReset() {
	suite.Run("unknown address answers success without mail", func() {
		suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "nobody@example.com").Return(nil, &ent.NotFoundError{}).Once()
		suite.NoError(suite.service.RequestPasswordReset(suite.Ctx, "Nobody@Example.com"))
	})
	suite.Run("mail not configured sends nothing", func() {
		suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "ada@example.com").Return(&ent.User{ID: suite.userID, Email: "ada@example.com", Active: true, PasswordHash: "h"}, nil).Once()
		suite.MockEmail.EXPECT().Configured(suite.Ctx).Return(false, nil).Once()
		suite.NoError(suite.service.RequestPasswordReset(suite.Ctx, "ada@example.com"))
	})
	suite.Run("configured mail stores a hashed token and links to it", func() {
		var storedHash string
		var sent mail.Content
		suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "ada@example.com").Return(&ent.User{ID: suite.userID, Email: "ada@example.com", Active: true, PasswordHash: "h"}, nil).Once()
		suite.MockEmail.EXPECT().Configured(suite.Ctx).Return(true, nil).Once()
		suite.MockUserRepo.EXPECT().CreatePasswordResetToken(suite.Ctx, suite.userID, mock.Anything, suite.now.Add(passwordResetTTL)).
			RunAndReturn(func(_ context.Context, _ uuid.UUID, hash string, _ time.Time) (*ent.PasswordResetToken, error) {
				storedHash = hash
				return &ent.PasswordResetToken{}, nil
			}).Once()
		suite.MockEmail.EXPECT().SendAsync(suite.Ctx, "ada@example.com", mock.Anything).
			RunAndReturn(func(_ context.Context, _ string, content mail.Content) bool {
				sent = content
				return true
			}).Once()

		suite.NoError(suite.service.RequestPasswordReset(suite.Ctx, "ada@example.com"))
		suite.Contains(sent.ButtonURL, "https://zigzag.example.com/reset-password?token=")
		token := sent.ButtonURL[len("https://zigzag.example.com/reset-password?token="):]
		suite.Equal(utils.HashToken(token), storedHash, "only the hash of the emailed token is stored")
	})
}

func (suite *UserServiceSuite) TestResetPassword() {
	token := "reset-token"
	valid := &ent.PasswordResetToken{ID: uuid.New(), UserID: suite.userID, ExpiresAt: suite.now.Add(time.Minute), Edges: ent.PasswordResetTokenEdges{User: &ent.User{ID: suite.userID, Active: true}}}

	suite.Run("short password", func() {
		suite.ErrorIs(suite.service.ResetPassword(suite.Ctx, &ResetPasswordInput{Token: token, Password: "short"}), errdefs.ErrInvalidInput)
	})
	suite.Run("expired token", func() {
		expired := *valid
		expired.ExpiresAt = suite.now.Add(-time.Second)
		suite.MockUserRepo.EXPECT().GetPasswordResetToken(suite.Ctx, utils.HashToken(token)).Return(&expired, nil).Once()
		suite.ErrorIs(suite.service.ResetPassword(suite.Ctx, &ResetPasswordInput{Token: token, Password: "long-enough"}), errdefs.ErrInvalidInput)
	})
	suite.Run("token claimed concurrently", func() {
		suite.MockUserRepo.EXPECT().GetPasswordResetToken(suite.Ctx, utils.HashToken(token)).Return(valid, nil).Once()
		suite.MockUserRepo.EXPECT().ClaimPasswordResetToken(suite.Ctx, mock.Anything, valid.ID, suite.now).Return(false, nil).Once()
		suite.ErrorIs(suite.service.ResetPassword(suite.Ctx, &ResetPasswordInput{Token: token, Password: "long-enough"}), errdefs.ErrInvalidInput)
	})
	suite.Run("success sets the password and signs out sessions", func() {
		suite.MockUserRepo.EXPECT().GetPasswordResetToken(suite.Ctx, utils.HashToken(token)).Return(valid, nil).Once()
		suite.MockUserRepo.EXPECT().ClaimPasswordResetToken(suite.Ctx, mock.Anything, valid.ID, suite.now).Return(true, nil).Once()
		suite.MockUserRepo.EXPECT().UpdatePassword(suite.Ctx, mock.Anything, suite.userID, "long-enough").Return(nil).Once()
		suite.MockOauthRepo.EXPECT().RevokeAllForUser(suite.Ctx, suite.userID).Return(nil).Once()
		suite.NoError(suite.service.ResetPassword(suite.Ctx, &ResetPasswordInput{Token: token, Password: "long-enough"}))
	})
}

func (suite *UserServiceSuite) TestBootstrapOnlyOnce() {
	suite.MockSystemRepo.EXPECT().LockSettings(suite.Ctx, mock.Anything).Return(nil).Once()
	suite.MockUserRepo.EXPECT().Count(suite.Ctx, mock.Anything).Return(1, nil).Once()
	_, _, err := suite.service.Bootstrap(suite.Ctx, bootstrapInput())
	suite.ErrorIs(err, errdefs.ErrAlreadyBootstrapped)
}

func (suite *UserServiceSuite) TestBootstrapCreatesSystemAdminAndWorkspace() {
	created := &ent.User{ID: suite.userID, Email: "ada@example.com"}
	suite.MockSystemRepo.EXPECT().LockSettings(suite.Ctx, mock.Anything).Return(nil).Once()
	suite.MockUserRepo.EXPECT().Count(suite.Ctx, mock.Anything).Return(0, nil).Once()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "ada").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *user_repo.CreateUserInput) bool {
		return input.SystemAdmin && input.Active && input.Email == "ada@example.com" && input.DisplayName == "ada"
	})).Return(created, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().Create(suite.Ctx, mock.Anything, &workspace_repo.CreateWorkspaceInput{Name: "Acme", Slug: "acme-corp", CreatorID: suite.userID}).
		Return(&ent.Workspace{ID: uuid.New(), Slug: "acme-corp"}, nil).Once()

	user, workspace, err := suite.service.Bootstrap(suite.Ctx, bootstrapInput())
	suite.Require().NoError(err)
	suite.Equal(created, user)
	suite.Equal("acme-corp", workspace.Slug)
}

func (suite *UserServiceSuite) TestBootstrapValidatesBeforeWriting() {
	input := bootstrapInput()
	input.WorkspaceSlug = "!"
	_, _, err := suite.service.Bootstrap(suite.Ctx, input)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *UserServiceSuite) TestAdminOperationsNeedSystemAdmin() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(errdefs.ErrUnauthorized).Times(3)
	_, err := suite.service.CreateUser(suite.Ctx, suite.userID, &CreateUserInput{})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	_, err = suite.service.UpdateUser(suite.Ctx, suite.userID, &UpdateUserInput{UserID: uuid.New()})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.service.DeleteUser(suite.Ctx, suite.userID, &DeleteUserInput{UserID: uuid.New()}), errdefs.ErrUnauthorized)
}

func (suite *UserServiceSuite) TestTheLastSystemAdminStays() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil)
	admin := &ent.User{ID: uuid.New(), SystemAdmin: true, Active: true}
	suite.MockUserRepo.EXPECT().GetByID(suite.Ctx, admin.ID).Return(admin, nil)
	suite.MockUserRepo.EXPECT().CountSystemAdmins(suite.Ctx, mock.Anything).Return(0, nil)

	_, err := suite.service.UpdateUser(suite.Ctx, suite.userID, &UpdateUserInput{UserID: admin.ID, SystemAdmin: utils.ToPtr(false)})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.UpdateUser(suite.Ctx, suite.userID, &UpdateUserInput{UserID: admin.ID, Active: utils.ToPtr(false)})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	suite.ErrorIs(suite.service.DeleteUser(suite.Ctx, suite.userID, &DeleteUserInput{UserID: admin.ID}), errdefs.ErrInvalidInput)
	suite.ErrorIs(suite.service.DeleteUser(suite.Ctx, suite.userID, &DeleteUserInput{UserID: suite.userID}), errdefs.ErrInvalidInput, "nobody deletes themselves")
}

func (suite *UserServiceSuite) TestDeactivatingSignsOutSessions() {
	target := &ent.User{ID: uuid.New(), Active: true}
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil).Once()
	suite.MockUserRepo.EXPECT().GetByID(suite.Ctx, target.ID).Return(target, nil).Once()
	suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, target.ID, mock.Anything).Return(&ent.User{ID: target.ID}, nil).Once()
	suite.MockOauthRepo.EXPECT().RevokeAllForUser(suite.Ctx, target.ID).Return(nil).Once()
	_, err := suite.service.UpdateUser(suite.Ctx, suite.userID, &UpdateUserInput{UserID: target.ID, Active: utils.ToPtr(false)})
	suite.NoError(err)
}

func (suite *UserServiceSuite) TestUpdateMeValidates() {
	suite.Run("taken display name", func() {
		suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "bob").Return(&ent.User{ID: uuid.New()}, nil).Once()
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{DisplayName: utils.ToPtr("Bob")})
		suite.ErrorIs(err, errdefs.ErrConflict)
	})
	suite.Run("bad time zone", func() {
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{Timezone: utils.ToPtr("Mars/Olympus")})
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	})
	suite.Run("avatar uploaded by someone else", func() {
		uploadID := uuid.New()
		other := uuid.New()
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, UploaderID: &other, ContentType: "image/png"}, nil).Once()
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{AvatarUploadID: &uploadID})
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	})
	suite.Run("nil uuid clears the avatar", func() {
		suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, suite.userID, mock.MatchedBy(func(input *user_repo.UpdateUserInput) bool { return input.ClearAvatar })).
			Return(&ent.User{ID: suite.userID}, nil).Once()
		nilID := uuid.Nil
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{AvatarUploadID: &nilID})
		suite.NoError(err)
	})
}

func (suite *UserServiceSuite) TestChangePasswordNeedsTheCurrentOne() {
	suite.MockUserRepo.EXPECT().VerifyPassword(suite.Ctx, suite.userID, "wrong").Return(user_repo.ErrInvalidPassword).Once()
	err := suite.service.ChangePassword(suite.Ctx, suite.userID, &ChangePasswordInput{CurrentPassword: "wrong", NewPassword: "new-password"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *UserServiceSuite) TestListUsersScopesNonAdminsToTheirWorkspaces() {
	workspaceID := uuid.New()
	suite.MockPermissionsRepo.EXPECT().GetUserPermissionSet(suite.Ctx, suite.userID).Return(permissionSet(workspaceID, schema.WorkspaceRoleMember), nil).Once()
	suite.MockUserRepo.EXPECT().List(suite.Ctx, &user_repo.ListUsersInput{Query: "a", Limit: 10, WorkspaceIDs: []uuid.UUID{workspaceID}}).Return(nil, nil).Once()
	_, err := suite.service.ListUsers(suite.Ctx, suite.userID, &ListUsersInput{Query: "a", Limit: 10})
	suite.NoError(err)
}

func bootstrapInput() *BootstrapInput {
	return &BootstrapInput{Email: "Ada@Example.com", Password: "correct-horse", Name: "Ada", DisplayName: "ada", WorkspaceName: "Acme", WorkspaceSlug: "Acme Corp"}
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func permissionSet(workspaceID uuid.UUID, role schema.WorkspaceRole) *permissions_repo.PermissionSet {
	return permissions_repo.NewPermissionSet(false, map[uuid.UUID]schema.WorkspaceRole{workspaceID: role})
}
