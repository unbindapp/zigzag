package user_repo

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type UserRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo *UserRepository
}

func (suite *UserRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewUserRepository(suite.DB)
}

func (suite *UserRepositorySuite) create(email, handle, password string) *ent.User {
	user, err := suite.repo.Create(suite.Ctx, nil, &CreateUserInput{Email: email, Password: password, Name: handle, DisplayName: handle, Active: true})
	suite.Require().NoError(err)
	return user
}

func (suite *UserRepositorySuite) TestCreateNormalizesAndHashes() {
	user := suite.create("  Ada@Example.COM ", "Ada", "correct-horse")
	suite.Equal("ada@example.com", user.Email)
	suite.Equal("ada", user.DisplayName)
	suite.NotEqual("correct-horse", user.PasswordHash)
	suite.True(user.Active)

	_, err := suite.repo.Create(suite.Ctx, nil, &CreateUserInput{Email: "ada@example.com", DisplayName: "other"})
	suite.True(ent.IsConstraintError(err), "emails are unique")
	_, err = suite.repo.Create(suite.Ctx, nil, &CreateUserInput{Email: "x@example.com", DisplayName: "ADA"})
	suite.True(ent.IsConstraintError(err), "display names are unique ignoring case")
	_, err = suite.repo.Create(suite.Ctx, nil, &CreateUserInput{Email: "", DisplayName: "x"})
	suite.ErrorIs(err, ErrInvalidUserInput)
}

func (suite *UserRepositorySuite) TestAuthenticate() {
	suite.create("ada@example.com", "ada", "correct-horse")
	placeholder, err := suite.repo.Create(suite.Ctx, nil, &CreateUserInput{Email: "ghost@example.com", DisplayName: "ghost"})
	suite.Require().NoError(err)
	suite.Empty(placeholder.PasswordHash)

	found, err := suite.repo.Authenticate(suite.Ctx, "ADA@example.com", "correct-horse")
	suite.NoError(err)
	suite.Equal("ada@example.com", found.Email)

	_, err = suite.repo.Authenticate(suite.Ctx, "ada@example.com", "wrong")
	suite.ErrorIs(err, ErrInvalidPassword)
	_, err = suite.repo.Authenticate(suite.Ctx, "missing@example.com", "whatever")
	suite.ErrorIs(err, ErrUserNotFound)
	_, err = suite.repo.Authenticate(suite.Ctx, "ghost@example.com", "")
	suite.ErrorIs(err, ErrInvalidUserInput)
	_, err = suite.repo.Authenticate(suite.Ctx, "ghost@example.com", "anything")
	suite.ErrorIs(err, ErrInvalidPassword, "accounts without a password cannot sign in")
}

func (suite *UserRepositorySuite) TestUpdatePasswordAndVerify() {
	user := suite.create("ada@example.com", "ada", "correct-horse")
	suite.NoError(suite.repo.UpdatePassword(suite.Ctx, nil, user.ID, "battery-staple"))
	suite.NoError(suite.repo.VerifyPassword(suite.Ctx, user.ID, "battery-staple"))
	suite.ErrorIs(suite.repo.VerifyPassword(suite.Ctx, user.ID, "correct-horse"), ErrInvalidPassword)
	suite.ErrorIs(suite.repo.UpdatePassword(suite.Ctx, nil, user.ID, ""), ErrInvalidUserInput)
}

func (suite *UserRepositorySuite) TestUpdateAndAvatar() {
	user := suite.create("ada@example.com", "ada", "correct-horse")
	workspace := suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	upload := suite.DB.Upload.Create().SetWorkspaceID(workspace.ID).SetFileName("a.png").SetContentType("image/png").SetSizeBytes(1).SetStorageKey("k").SaveX(suite.Ctx)

	updated, err := suite.repo.Update(suite.Ctx, nil, user.ID, &UpdateUserInput{Name: utils.ToPtr("Ada L."), DisplayName: utils.ToPtr("AdaL"), AvatarUploadID: &upload.ID})
	suite.Require().NoError(err)
	suite.Equal("Ada L.", updated.Name)
	suite.Equal("adal", updated.DisplayName)
	suite.Equal(upload.ID, *updated.AvatarUploadID)

	suite.DB.Upload.DeleteOneID(upload.ID).ExecX(suite.Ctx)
	reloaded, _ := suite.repo.GetByID(suite.Ctx, user.ID)
	suite.Nil(reloaded.AvatarUploadID, "deleting the upload clears the avatar")
}

func (suite *UserRepositorySuite) TestListFiltersByWorkspaceAndQuery() {
	ada := suite.create("ada@example.com", "ada", "x-password")
	suite.create("bob@example.com", "bob", "x-password")
	workspace := suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.DB.WorkspaceMember.Create().SetWorkspaceID(workspace.ID).SetUserID(ada.ID).SetRole("member").SaveX(suite.Ctx)

	all, err := suite.repo.List(suite.Ctx, &ListUsersInput{})
	suite.NoError(err)
	suite.Len(all, 2)

	members, err := suite.repo.List(suite.Ctx, &ListUsersInput{WorkspaceIDs: []uuid.UUID{workspace.ID}})
	suite.NoError(err)
	suite.Len(members, 1)
	suite.Equal(ada.ID, members[0].ID)

	none, err := suite.repo.List(suite.Ctx, &ListUsersInput{WorkspaceIDs: []uuid.UUID{}})
	suite.NoError(err)
	suite.Empty(none, "an empty workspace list matches nobody")

	byQuery, err := suite.repo.List(suite.Ctx, &ListUsersInput{Query: "BOB"})
	suite.NoError(err)
	suite.Len(byQuery, 1)
}

func (suite *UserRepositorySuite) TestNotificationSettingsDefaultOn() {
	user := suite.create("ada@example.com", "ada", "x-password")
	settings, err := suite.repo.GetNotificationSettings(suite.Ctx, user.ID)
	suite.Require().NoError(err)
	suite.True(settings.EmailEnabled && settings.Assigned && settings.Mentioned && settings.Commented && settings.StateChanged && settings.DueSoon && settings.ProjectUpdates)

	updated, err := suite.repo.UpdateNotificationSettings(suite.Ctx, user.ID, &UpdateNotificationSettingsInput{Commented: utils.ToPtr(false)})
	suite.Require().NoError(err)
	suite.False(updated.Commented)
	suite.True(updated.Mentioned)
	suite.Equal(1, suite.DB.NotificationSetting.Query().CountX(suite.Ctx))
}

func (suite *UserRepositorySuite) TestPasswordResetTokenIsClaimedOnce() {
	user := suite.create("ada@example.com", "ada", "x-password")
	now := time.Now()
	token, err := suite.repo.CreatePasswordResetToken(suite.Ctx, user.ID, "hash", now.Add(time.Hour))
	suite.Require().NoError(err)

	found, err := suite.repo.GetPasswordResetToken(suite.Ctx, "hash")
	suite.Require().NoError(err)
	suite.Equal(user.ID, found.Edges.User.ID)

	claimed, err := suite.repo.ClaimPasswordResetToken(suite.Ctx, nil, token.ID, now)
	suite.NoError(err)
	suite.True(claimed)
	claimed, err = suite.repo.ClaimPasswordResetToken(suite.Ctx, nil, token.ID, now)
	suite.NoError(err)
	suite.False(claimed, "a reset token works once")

	expired, _ := suite.repo.CreatePasswordResetToken(suite.Ctx, user.ID, "old", now.Add(-time.Minute))
	claimed, _ = suite.repo.ClaimPasswordResetToken(suite.Ctx, nil, expired.ID, now)
	suite.False(claimed, "an expired token cannot be claimed")

	deleted, err := suite.repo.DeletePasswordResetTokensBefore(suite.Ctx, now.Add(time.Second))
	suite.NoError(err)
	suite.Equal(2, deleted)
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

func (suite *UserRepositorySuite) TestPasswordsOverBcryptsLimitAreInvalidInput() {
	_, err := HashPassword(strings.Repeat("é", 37))
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "37 accented letters are 74 bytes")
	_, err = HashPassword(strings.Repeat("é", 36))
	suite.NoError(err)
}
