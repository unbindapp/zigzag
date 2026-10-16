package user_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/mail"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	team_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/team"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
	workspace_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/workspace"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
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

	outage := errors.New("connection refused")
	suite.MockUserRepo.EXPECT().Authenticate(suite.Ctx, "ada@example.com", "pw").Return(nil, outage).Once()
	_, err = suite.service.Login(suite.Ctx, "ada@example.com", "pw")
	suite.ErrorIs(err, outage, "a database outage is a server error, not a wrong password")
}

func (suite *UserServiceSuite) TestWhoamiListsWorkspacesWithRolesAndKeyLimits() {
	workspaceID := uuid.New()
	ctx := permissions_repo.WithAPIKeyAccess(suite.Ctx, permissions_repo.APIKeyAccess{Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{workspaceID}})
	suite.MockUserRepo.EXPECT().GetByID(ctx, suite.userID).Return(&ent.User{ID: suite.userID, Email: "ada@example.com", Active: true}, nil).Once()
	suite.MockPermissionsRepo.EXPECT().GetUserPermissionSet(ctx, suite.userID).Return(permissionSet(workspaceID, schema.WorkspaceRoleViewer), nil).Once()
	iconID := uuid.New()
	suite.MockWorkspaceRepo.EXPECT().GetByIDs(ctx, []uuid.UUID{workspaceID}).Return([]*ent.Workspace{{ID: workspaceID, Name: "Acme", Slug: "acme", IconUploadID: &iconID}}, nil).Once()

	resp, err := suite.service.Whoami(ctx, suite.userID)
	suite.Require().NoError(err)
	suite.Require().Len(resp.Workspaces, 1)
	suite.Equal(schema.WorkspaceRoleViewer, resp.Workspaces[0].Role)
	suite.Equal(utils.ToPtr("/api/go/uploads/get?upload_id="+iconID.String()), resp.Workspaces[0].IconURL)
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
	suite.Run("placeholders without a password or Google link get nothing", func() {
		suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "ada@example.com").Return(&ent.User{ID: suite.userID, Email: "ada@example.com", Active: true}, nil).Once()
		suite.NoError(suite.service.RequestPasswordReset(suite.Ctx, "ada@example.com"))
	})
	suite.Run("a Google account without a password can set one", func() {
		suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "ada@example.com").Return(&ent.User{ID: suite.userID, Email: "ada@example.com", Active: true, GoogleSubject: utils.ToPtr("sub-1")}, nil).Once()
		suite.MockEmail.EXPECT().Configured(suite.Ctx).Return(true, nil).Once()
		suite.MockUserRepo.EXPECT().CreatePasswordResetToken(suite.Ctx, suite.userID, mock.Anything, suite.now.Add(passwordResetTTL)).Return(&ent.PasswordResetToken{}, nil).Once()
		suite.MockEmail.EXPECT().SendAsync(suite.Ctx, "ada@example.com", mock.Anything).Return(true).Once()
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
		suite.Contains(sent.Button.URL, "https://zigzag.example.com/reset-password?token=")
		token := sent.Button.URL[len("https://zigzag.example.com/reset-password?token="):]
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
	_, err := suite.service.Bootstrap(suite.Ctx, bootstrapInput())
	suite.ErrorIs(err, errdefs.ErrAlreadyBootstrapped)
}

func (suite *UserServiceSuite) TestBootstrapCreatesSystemAdminWorkspaceAndDefaultTeam() {
	created := &ent.User{ID: suite.userID, Email: "ada@example.com"}
	suite.MockSystemRepo.EXPECT().LockSettings(suite.Ctx, mock.Anything).Return(nil).Once()
	suite.MockUserRepo.EXPECT().Count(suite.Ctx, mock.Anything).Return(0, nil).Once()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "ada").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *user_repo.CreateUserInput) bool {
		return input.SystemAdmin && input.Active && input.Email == "ada@example.com" && input.DisplayName == "ada"
	})).Return(created, nil).Once()
	workspace := &ent.Workspace{ID: uuid.New(), Name: "Acme", Slug: "acme-corp"}
	suite.MockWorkspaceRepo.EXPECT().Create(suite.Ctx, mock.Anything, &workspace_repo.CreateWorkspaceInput{Name: "Acme", Slug: "acme-corp", CreatorID: suite.userID}).
		Return(workspace, nil).Once()
	team := &ent.Team{ID: uuid.New(), WorkspaceID: workspace.ID, Name: "Acme", Key: "ACM"}
	suite.MockTeamRepo.EXPECT().Create(suite.Ctx, mock.Anything, &team_repo.CreateTeamInput{WorkspaceID: workspace.ID, Name: "Acme", Key: "ACM", CreatorID: suite.userID}).
		Return(team, nil).Once()

	result, err := suite.service.Bootstrap(suite.Ctx, bootstrapInput())
	suite.Require().NoError(err)
	suite.Equal(created, result.User)
	suite.Equal("acme-corp", result.Workspace.Slug)
	suite.Equal(team, result.Team)
}

func (suite *UserServiceSuite) TestBootstrapFailsWhenTheDefaultTeamFails() {
	suite.MockSystemRepo.EXPECT().LockSettings(suite.Ctx, mock.Anything).Return(nil).Once()
	suite.MockUserRepo.EXPECT().Count(suite.Ctx, mock.Anything).Return(0, nil).Once()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "ada").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).Return(&ent.User{ID: suite.userID}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).Return(&ent.Workspace{ID: uuid.New(), Name: "Acme"}, nil).Once()
	failure := errors.New("insert failed")
	suite.MockTeamRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).Return(nil, failure).Once()

	_, err := suite.service.Bootstrap(suite.Ctx, bootstrapInput())
	suite.ErrorIs(err, failure)
}

func (suite *UserServiceSuite) TestBootstrapValidatesBeforeWriting() {
	input := bootstrapInput()
	input.WorkspaceSlug = "!"
	_, err := suite.service.Bootstrap(suite.Ctx, input)
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
	suite.Run("blank name", func() {
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{Name: utils.ToPtr("   ")})
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	})
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
	suite.Run("avatar that is not an image", func() {
		uploadID := uuid.New()
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, UploaderID: &suite.userID, ContentType: "application/pdf"}, nil).Once()
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{AvatarUploadID: &uploadID})
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	})
	suite.Run("raster avatar without decoded dimensions", func() {
		uploadID := uuid.New()
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, UploaderID: &suite.userID, ContentType: "image/x-icon"}, nil).Once()
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{AvatarUploadID: &uploadID})
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	})
	for name, upload := range map[string]*ent.Upload{
		"png avatar":              {ContentType: "image/png", Width: utils.ToPtr(64), Height: utils.ToPtr(64)},
		"svg avatar without size": {ContentType: "image/svg+xml"},
	} {
		suite.Run(name, func() {
			upload.ID = uuid.New()
			upload.UploaderID = &suite.userID
			suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, upload.ID).Return(upload, nil).Once()
			suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, suite.userID, &user_repo.UpdateUserInput{AvatarUploadID: &upload.ID}).
				Return(&ent.User{ID: suite.userID, AvatarUploadID: &upload.ID}, nil).Once()
			resp, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{AvatarUploadID: &upload.ID})
			suite.Require().NoError(err)
			suite.Equal(utils.ToPtr("/api/go/uploads/get?upload_id="+upload.ID.String()), resp.AvatarURL)
		})
	}
	suite.Run("nil uuid clears the avatar", func() {
		suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, suite.userID, mock.MatchedBy(func(input *user_repo.UpdateUserInput) bool { return input.ClearAvatar })).
			Return(&ent.User{ID: suite.userID}, nil).Once()
		nilID := uuid.Nil
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{AvatarUploadID: &nilID})
		suite.NoError(err)
	})
}

func (suite *UserServiceSuite) TestGithubUsernames() {
	suite.Run("invalid", func() {
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{GithubLogin: utils.ToPtr("not a login")})
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	})
	suite.Run("linked to someone else", func() {
		suite.MockUserRepo.EXPECT().GetByGithubLogin(suite.Ctx, "octocat").Return(&ent.User{ID: uuid.New()}, nil).Once()
		_, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{GithubLogin: utils.ToPtr("@OctoCat")})
		suite.ErrorIs(err, errdefs.ErrConflict)
	})
	suite.Run("stored lowercased", func() {
		suite.MockUserRepo.EXPECT().GetByGithubLogin(suite.Ctx, "octocat").Return(nil, &ent.NotFoundError{}).Once()
		suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, suite.userID, mock.MatchedBy(func(input *user_repo.UpdateUserInput) bool {
			return input.GithubLogin != nil && *input.GithubLogin == "octocat"
		})).Return(&ent.User{ID: suite.userID, GithubLogin: utils.ToPtr("octocat")}, nil).Once()
		updated, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{GithubLogin: utils.ToPtr("OctoCat")})
		suite.Require().NoError(err)
		suite.Equal("octocat", *updated.GithubLogin)
	})
	suite.Run("empty clears", func() {
		suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, suite.userID, mock.MatchedBy(func(input *user_repo.UpdateUserInput) bool {
			return input.GithubLogin != nil && *input.GithubLogin == ""
		})).Return(&ent.User{ID: suite.userID}, nil).Once()
		updated, err := suite.service.UpdateMe(suite.Ctx, suite.userID, &UpdateMeInput{GithubLogin: utils.ToPtr(" ")})
		suite.Require().NoError(err)
		suite.Nil(updated.GithubLogin)
	})
	suite.Run("system admins set it for others", func() {
		target := &ent.User{ID: uuid.New()}
		suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil).Once()
		suite.MockUserRepo.EXPECT().GetByID(suite.Ctx, target.ID).Return(target, nil).Once()
		suite.MockUserRepo.EXPECT().GetByGithubLogin(suite.Ctx, "eve-gh").Return(target, nil).Once()
		suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, target.ID, mock.MatchedBy(func(input *user_repo.UpdateUserInput) bool {
			return input.GithubLogin != nil && *input.GithubLogin == "eve-gh"
		})).Return(target, nil).Once()
		_, err := suite.service.UpdateUser(suite.Ctx, suite.userID, &UpdateUserInput{UserID: target.ID, GithubLogin: utils.ToPtr("Eve-GH")})
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
