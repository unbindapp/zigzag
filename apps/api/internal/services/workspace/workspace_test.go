package workspace_service

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

type WorkspaceServiceSuite struct {
	services.ServiceTestSuite
	service     *WorkspaceService
	now         time.Time
	adminID     uuid.UUID
	workspaceID uuid.UUID
	workspace   *ent.Workspace
}

func (suite *WorkspaceServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewWorkspaceService(&config.Config{ExternalUIUrl: "https://zigzag.example.com"}, suite.MockRepo, suite.MockEmail)
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.adminID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.workspace = &ent.Workspace{ID: suite.workspaceID, Name: "Acme", Slug: "acme"}
	suite.ExpectTx()
}

func (suite *WorkspaceServiceSuite) allow(role schema.WorkspaceRole) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.adminID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(nil).Maybe()
}

func (suite *WorkspaceServiceSuite) deny(role schema.WorkspaceRole) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.adminID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(errdefs.ErrUnauthorized).Maybe()
}

func (suite *WorkspaceServiceSuite) expectRespond() {
	suite.MockPermissionsRepo.EXPECT().GetUserPermissionSet(suite.Ctx, mock.Anything).
		Return(permissions_repo.NewPermissionSet(false, map[uuid.UUID]schema.WorkspaceRole{suite.workspaceID: schema.WorkspaceRoleAdmin}), nil).Maybe()
	suite.MockWorkspaceRepo.EXPECT().Counts(suite.Ctx, mock.Anything).Return(map[uuid.UUID]workspace_repo.WorkspaceCounts{suite.workspaceID: {Members: 2, Teams: 1}}, nil).Maybe()
}

func (suite *WorkspaceServiceSuite) TestInvisibleWorkspacesAreNotFound() {
	suite.deny(schema.WorkspaceRoleViewer)
	suite.MockWorkspaceRepo.EXPECT().GetBySlug(suite.Ctx, "acme").Return(suite.workspace, nil).Once()
	_, err := suite.service.Get(suite.Ctx, suite.adminID, &GetWorkspaceInput{Slug: "acme"})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	_, err = suite.service.Update(suite.Ctx, suite.adminID, &UpdateWorkspaceInput{WorkspaceID: suite.workspaceID, Name: utils.ToPtr("x")})
	suite.ErrorIs(err, errdefs.ErrNotFound, "updates also hide workspaces the caller cannot see")
}

func (suite *WorkspaceServiceSuite) TestViewerCannotUpdate() {
	suite.allow(schema.WorkspaceRoleViewer)
	suite.deny(schema.WorkspaceRoleAdmin)
	_, err := suite.service.Update(suite.Ctx, suite.adminID, &UpdateWorkspaceInput{WorkspaceID: suite.workspaceID, Name: utils.ToPtr("x")})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *WorkspaceServiceSuite) TestCreateRejectsTakenSlug() {
	suite.MockWorkspaceRepo.EXPECT().SlugExists(suite.Ctx, mock.Anything, "acme", (*uuid.UUID)(nil)).Return(true, nil).Once()
	_, err := suite.service.Create(suite.Ctx, suite.adminID, &CreateWorkspaceInput{Name: "Acme", Slug: "ACME"})
	suite.ErrorIs(err, errdefs.ErrConflict)
}

func (suite *WorkspaceServiceSuite) TestCreateMakesTheCreatorAdminOfTheWorkspaceAndItsDefaultTeam() {
	suite.expectRespond()
	suite.MockWorkspaceRepo.EXPECT().SlugExists(suite.Ctx, mock.Anything, "acme", (*uuid.UUID)(nil)).Return(false, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().Create(suite.Ctx, mock.Anything, &workspace_repo.CreateWorkspaceInput{Name: "Acme", Slug: "acme", CreatorID: suite.adminID}).Return(suite.workspace, nil).Once()
	teamID := uuid.New()
	suite.MockTeamRepo.EXPECT().Create(suite.Ctx, mock.Anything, &team_repo.CreateTeamInput{WorkspaceID: suite.workspaceID, Name: "Acme", Key: "ACM", CreatorID: suite.adminID}).
		Return(&ent.Team{ID: teamID, WorkspaceID: suite.workspaceID}, nil).Once()
	resp, err := suite.service.Create(suite.Ctx, suite.adminID, &CreateWorkspaceInput{Name: " Acme ", Slug: "acme"})
	suite.Require().NoError(err)
	suite.Equal(schema.WorkspaceRoleAdmin, resp.Role)
	suite.Equal(2, resp.MemberCount)
	suite.Equal(&teamID, resp.DefaultTeamID)
}

func (suite *WorkspaceServiceSuite) TestCreateFailsWhenTheDefaultTeamFails() {
	suite.MockWorkspaceRepo.EXPECT().SlugExists(suite.Ctx, mock.Anything, "acme", (*uuid.UUID)(nil)).Return(false, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).Return(suite.workspace, nil).Once()
	failure := errors.New("insert failed")
	suite.MockTeamRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).Return(nil, failure).Once()
	_, err := suite.service.Create(suite.Ctx, suite.adminID, &CreateWorkspaceInput{Name: "Acme", Slug: "acme"})
	suite.ErrorIs(err, failure)
}

func (suite *WorkspaceServiceSuite) TestUpdateIconUpload() {
	uploadID := uuid.New()
	update := func(id uuid.UUID) error {
		_, err := suite.service.Update(suite.Ctx, suite.adminID, &UpdateWorkspaceInput{WorkspaceID: suite.workspaceID, IconUploadID: &id})
		return err
	}
	suite.allow(schema.WorkspaceRoleViewer)
	suite.allow(schema.WorkspaceRoleAdmin)
	suite.expectRespond()

	suite.Run("unknown upload", func() {
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(nil, &ent.NotFoundError{}).Once()
		suite.ErrorIs(update(uploadID), errdefs.ErrInvalidInput)
	})
	suite.Run("upload of another workspace", func() {
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, WorkspaceID: uuid.New(), ContentType: "image/png", Width: utils.ToPtr(64)}, nil).Once()
		suite.ErrorIs(update(uploadID), errdefs.ErrInvalidInput)
	})
	suite.Run("not an image", func() {
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, WorkspaceID: suite.workspaceID, ContentType: "application/pdf"}, nil).Once()
		suite.ErrorIs(update(uploadID), errdefs.ErrInvalidInput)
	})
	suite.Run("raster image without decoded dimensions", func() {
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, WorkspaceID: suite.workspaceID, ContentType: "image/x-icon"}, nil).Once()
		suite.ErrorIs(update(uploadID), errdefs.ErrInvalidInput)
	})
	suite.Run("svg without dimensions", func() {
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, WorkspaceID: suite.workspaceID, ContentType: "image/svg+xml"}, nil).Once()
		suite.MockWorkspaceRepo.EXPECT().Update(suite.Ctx, suite.workspaceID, &workspace_repo.UpdateWorkspaceInput{IconUploadID: &uploadID}).
			Return(&ent.Workspace{ID: suite.workspaceID, IconUploadID: &uploadID}, nil).Once()
		suite.NoError(update(uploadID))
	})
	suite.Run("image of this workspace", func() {
		suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, uploadID).Return(&ent.Upload{ID: uploadID, WorkspaceID: suite.workspaceID, ContentType: "image/png", Width: utils.ToPtr(64)}, nil).Once()
		suite.MockWorkspaceRepo.EXPECT().Update(suite.Ctx, suite.workspaceID, &workspace_repo.UpdateWorkspaceInput{IconUploadID: &uploadID}).
			Return(&ent.Workspace{ID: suite.workspaceID, IconUploadID: &uploadID}, nil).Once()
		resp, err := suite.service.Update(suite.Ctx, suite.adminID, &UpdateWorkspaceInput{WorkspaceID: suite.workspaceID, IconUploadID: &uploadID})
		suite.Require().NoError(err)
		suite.Equal(utils.ToPtr("/api/go/uploads/get?upload_id="+uploadID.String()), resp.IconURL)
	})
	suite.Run("nil uuid clears the logo", func() {
		suite.MockWorkspaceRepo.EXPECT().Update(suite.Ctx, suite.workspaceID, &workspace_repo.UpdateWorkspaceInput{ClearIconUpload: true}).
			Return(&ent.Workspace{ID: suite.workspaceID}, nil).Once()
		resp, err := suite.service.Update(suite.Ctx, suite.adminID, &UpdateWorkspaceInput{WorkspaceID: suite.workspaceID, IconUploadID: &uuid.Nil})
		suite.Require().NoError(err)
		suite.Nil(resp.IconURL)
	})
}

func (suite *WorkspaceServiceSuite) TestViewerCannotSetTheIconUpload() {
	uploadID := uuid.New()
	suite.allow(schema.WorkspaceRoleViewer)
	suite.deny(schema.WorkspaceRoleAdmin)
	_, err := suite.service.Update(suite.Ctx, suite.adminID, &UpdateWorkspaceInput{WorkspaceID: suite.workspaceID, IconUploadID: &uploadID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *WorkspaceServiceSuite) TestNarrowedCredentialsCreateWorkspacesOnlyWithFullWriteAccess() {
	for name, access := range map[string]permissions_repo.APIKeyAccess{
		"scoped to a workspace": {Role: schema.WorkspaceRoleAdmin, WorkspaceIDs: []uuid.UUID{suite.workspaceID}},
		"read only":             {Role: schema.WorkspaceRoleViewer, FullAccess: true},
	} {
		ctx := permissions_repo.WithAPIKeyAccess(suite.Ctx, access)
		_, err := suite.service.Create(ctx, suite.adminID, &CreateWorkspaceInput{Name: "Acme", Slug: "acme"})
		suite.ErrorIs(err, errdefs.ErrUnauthorized, name)
	}
}

func (suite *WorkspaceServiceSuite) TestDeleteNeedsTheSlugRepeated() {
	suite.allow(schema.WorkspaceRoleViewer)
	suite.allow(schema.WorkspaceRoleAdmin)
	suite.MockWorkspaceRepo.EXPECT().GetByID(suite.Ctx, suite.workspaceID).Return(suite.workspace, nil)
	suite.ErrorIs(suite.service.Delete(suite.Ctx, suite.adminID, &DeleteWorkspaceInput{WorkspaceID: suite.workspaceID, ConfirmSlug: "nope"}), errdefs.ErrInvalidInput)

	suite.MockWorkspaceRepo.EXPECT().Delete(suite.Ctx, suite.workspaceID).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.adminID, &DeleteWorkspaceInput{WorkspaceID: suite.workspaceID, ConfirmSlug: "acme"}))
}

func (suite *WorkspaceServiceSuite) TestAddMemberAddsExistingAccountsAtOnce() {
	suite.allow(schema.WorkspaceRoleViewer)
	suite.allow(schema.WorkspaceRoleAdmin)
	teamID := uuid.New()
	existing := &ent.User{ID: uuid.New(), Email: "grace@example.com", Active: true, PasswordHash: "h"}
	suite.MockTeamRepo.EXPECT().TeamsInWorkspace(suite.Ctx, suite.workspaceID, []uuid.UUID{teamID}).Return([]*ent.Team{{ID: teamID}}, nil).Once()
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "grace@example.com").Return(existing, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().GetMember(suite.Ctx, mock.Anything, suite.workspaceID, existing.ID).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, suite.workspaceID, existing.ID, schema.WorkspaceRoleMember).
		Return(&ent.WorkspaceMember{UserID: existing.ID, Role: schema.WorkspaceRoleMember, Edges: ent.WorkspaceMemberEdges{User: existing}}, nil).Once()
	suite.MockTeamRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, teamID, existing.ID, false).Return(nil).Once()
	suite.MockUserRepo.EXPECT().GetByID(suite.Ctx, suite.adminID).Return(&ent.User{ID: suite.adminID, Name: "Ada"}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().GetByID(suite.Ctx, suite.workspaceID).Return(suite.workspace, nil).Once()
	var sent mail.Content
	suite.MockEmail.EXPECT().SendAsync(suite.Ctx, "grace@example.com", mock.Anything).RunAndReturn(func(_ context.Context, _ string, content mail.Content) bool {
		sent = content
		return true
	}).Once()

	resp, err := suite.service.AddMember(suite.Ctx, suite.adminID, &AddMemberInput{WorkspaceID: suite.workspaceID, Email: "Grace@Example.com", Role: schema.WorkspaceRoleMember, TeamIDs: []uuid.UUID{teamID}})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Member)
	suite.Nil(resp.Invitation)
	suite.Equal([]uuid.UUID{teamID}, resp.Member.TeamIDs)
	suite.Equal("Ada added you to the "+suite.workspace.Name+" workspace as a member.", sent.Intro[0])
	suite.Equal(mail.Button{Label: "Open workspace", URL: "https://zigzag.example.com/" + suite.workspace.Slug}, sent.Button)
}

func (suite *WorkspaceServiceSuite) TestAddMemberRejectsForeignTeams() {
	suite.allow(schema.WorkspaceRoleViewer)
	suite.allow(schema.WorkspaceRoleAdmin)
	teamID := uuid.New()
	suite.MockTeamRepo.EXPECT().TeamsInWorkspace(suite.Ctx, suite.workspaceID, []uuid.UUID{teamID}).Return(nil, nil).Once()
	_, err := suite.service.AddMember(suite.Ctx, suite.adminID, &AddMemberInput{WorkspaceID: suite.workspaceID, Email: "a@example.com", Role: schema.WorkspaceRoleMember, TeamIDs: []uuid.UUID{teamID}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *WorkspaceServiceSuite) TestCreateInvitationReturnsLinkAndHashesToken() {
	suite.allow(schema.WorkspaceRoleViewer)
	suite.allow(schema.WorkspaceRoleAdmin)
	var stored *workspace_repo.CreateInvitationInput
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, "new@example.com").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().CreateInvitation(suite.Ctx, mock.Anything).RunAndReturn(func(_ context.Context, input *workspace_repo.CreateInvitationInput) (*ent.Invitation, error) {
		stored = input
		return &ent.Invitation{ID: uuid.New(), WorkspaceID: input.WorkspaceID, Email: input.Email, Role: input.Role, ExpiresAt: input.ExpiresAt}, nil
	}).Maybe()
	suite.MockUserRepo.EXPECT().GetByID(suite.Ctx, suite.adminID).Return(&ent.User{ID: suite.adminID, Email: "ada@example.com"}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().GetByID(suite.Ctx, suite.workspaceID).Return(suite.workspace, nil).Once()
	var sent mail.Content
	suite.MockEmail.EXPECT().SendAsync(suite.Ctx, "new@example.com", mock.Anything).RunAndReturn(func(_ context.Context, _ string, content mail.Content) bool {
		sent = content
		return false
	}).Once()

	resp, err := suite.service.CreateInvitation(suite.Ctx, suite.adminID, &CreateInvitationInput{WorkspaceID: suite.workspaceID, Email: "new@example.com", Role: schema.WorkspaceRoleViewer})
	suite.Require().NoError(err)
	suite.False(resp.Emailed, "emailed is false when mail is not configured")
	suite.Equal("ada@example.com invited you to "+suite.workspace.Name+" on Zigzag", sent.Subject)
	suite.Equal([]string{"ada@example.com invited you to the " + suite.workspace.Name + " workspace as a viewer.", "Viewers can read issues, projects and documents."}, sent.Intro)
	suite.Equal(mail.Button{Label: "Accept invitation", URL: resp.Link}, sent.Button)
	suite.Contains(resp.Link, "https://zigzag.example.com/invite/")
	token := resp.Link[len("https://zigzag.example.com/invite/"):]
	suite.Equal(utils.HashToken(token), stored.TokenHash)
	suite.Equal(suite.now.Add(invitationTTL), stored.ExpiresAt)
}

func (suite *WorkspaceServiceSuite) TestTheLastAdminCannotLeaveOrBeDemoted() {
	suite.allow(schema.WorkspaceRoleViewer)
	suite.allow(schema.WorkspaceRoleAdmin)
	suite.MockWorkspaceRepo.EXPECT().GetMember(suite.Ctx, mock.Anything, suite.workspaceID, suite.adminID).Return(&ent.WorkspaceMember{Role: schema.WorkspaceRoleAdmin}, nil)
	suite.MockWorkspaceRepo.EXPECT().CountAdmins(suite.Ctx, suite.workspaceID).Return(1, nil)

	suite.ErrorIs(suite.service.RemoveMember(suite.Ctx, suite.adminID, &RemoveMemberInput{WorkspaceID: suite.workspaceID, UserID: suite.adminID}), errdefs.ErrInvalidInput)
	_, err := suite.service.UpdateMember(suite.Ctx, suite.adminID, &UpdateMemberInput{WorkspaceID: suite.workspaceID, UserID: suite.adminID, Role: schema.WorkspaceRoleMember})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *WorkspaceServiceSuite) TestMembersMayLeave() {
	memberID := uuid.New()
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, memberID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: suite.workspaceID}}).Return(nil)
	suite.MockWorkspaceRepo.EXPECT().GetMember(suite.Ctx, mock.Anything, suite.workspaceID, memberID).Return(&ent.WorkspaceMember{Role: schema.WorkspaceRoleMember}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().RemoveMember(suite.Ctx, suite.workspaceID, memberID).Return(nil).Once()
	suite.NoError(suite.service.RemoveMember(suite.Ctx, memberID, &RemoveMemberInput{WorkspaceID: suite.workspaceID, UserID: memberID}))
}

func (suite *WorkspaceServiceSuite) invitation() *ent.Invitation {
	return &ent.Invitation{ID: uuid.New(), WorkspaceID: suite.workspaceID, Email: "grace@example.com", Role: schema.WorkspaceRoleMember, ExpiresAt: suite.now.Add(time.Hour), TeamIds: []uuid.UUID{}, Edges: ent.InvitationEdges{Workspace: suite.workspace}}
}

func (suite *WorkspaceServiceSuite) TestAcceptRefusesExpiredOrUsedInvitations() {
	expired := suite.invitation()
	expired.ExpiresAt = suite.now
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(expired, nil).Once()
	_, _, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *WorkspaceServiceSuite) TestAcceptByAnExistingAccountNeedsItsPassword() {
	inv := suite.invitation()
	existing := &ent.User{ID: uuid.New(), Email: inv.Email, Active: true, PasswordHash: "h"}
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(inv, nil)
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(existing, nil)
	suite.MockUserRepo.EXPECT().Authenticate(suite.Ctx, inv.Email, "").Return(nil, user_repo.ErrInvalidUserInput).Once()

	_, _, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "the link alone must not sign anyone into an existing account")
}

func (suite *WorkspaceServiceSuite) TestAcceptCreatesTheAccountAndJoinsTeams() {
	inv := suite.invitation()
	teamID := uuid.New()
	inv.TeamIds = []uuid.UUID{teamID}
	created := &ent.User{ID: uuid.New(), Email: inv.Email, Active: true}
	suite.expectRespond()
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(inv, nil).Once()
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().ClaimInvitation(suite.Ctx, mock.Anything, inv.ID, suite.now).Return(true, nil).Once()
	suite.MockUserRepo.EXPECT().GetByDisplayName(suite.Ctx, mock.Anything, "grace").Return(nil, &ent.NotFoundError{}).Once()
	suite.MockUserRepo.EXPECT().Create(suite.Ctx, mock.Anything, &user_repo.CreateUserInput{Email: inv.Email, Password: "long-enough", Name: "Grace Hopper", DisplayName: "grace", Active: true}).Return(created, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, suite.workspaceID, created.ID, schema.WorkspaceRoleMember).Return(&ent.WorkspaceMember{}, nil).Once()
	suite.MockTeamRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, teamID, created.ID, false).Return(nil).Once()

	user, workspace, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough"), Name: utils.ToPtr("Grace Hopper")})
	suite.Require().NoError(err)
	suite.Equal(created.ID, user.ID)
	suite.Equal(suite.workspaceID, workspace.ID)
}

func (suite *WorkspaceServiceSuite) TestAcceptCannotTakeOverDeactivatedOrForeignAccounts() {
	inv := suite.invitation()
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(inv, nil)

	deactivated := &ent.User{ID: uuid.New(), Email: inv.Email, Active: false, PasswordHash: "h"}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(deactivated, nil).Once()
	_, _, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "an invitation never reactivates an account a system admin turned off")

	placeholder := &ent.User{ID: uuid.New(), Email: inv.Email, Active: false}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(placeholder, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().MembershipsOfUser(suite.Ctx, placeholder.ID).Return([]*ent.WorkspaceMember{
		{WorkspaceID: suite.workspaceID}, {WorkspaceID: uuid.New()},
	}, nil).Once()
	_, _, err = suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "an admin of one workspace cannot claim a placeholder that also belongs to another")
}

func (suite *WorkspaceServiceSuite) TestAcceptNeverSetsAPasswordOnAGoogleAccount() {
	inv := suite.invitation()
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(inv, nil)

	linked := &ent.User{ID: uuid.New(), Email: inv.Email, Active: true, GoogleSubject: utils.ToPtr("sub-1")}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(linked, nil).Once()
	_, _, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "a Google account joins by signing in with Google")

	deactivated := &ent.User{ID: uuid.New(), Email: inv.Email, Active: false, GoogleSubject: utils.ToPtr("sub-1")}
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(deactivated, nil).Once()
	_, _, err = suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "a deactivated Google account is not a placeholder")
}

func (suite *WorkspaceServiceSuite) TestAcceptClaimsAPlaceholderOfTheInvitingWorkspace() {
	inv := suite.invitation()
	placeholder := &ent.User{ID: uuid.New(), Email: inv.Email, Active: false}
	claimed := &ent.User{ID: placeholder.ID, Email: inv.Email, Active: true}
	suite.expectRespond()
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(inv, nil).Once()
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(placeholder, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().MembershipsOfUser(suite.Ctx, placeholder.ID).Return([]*ent.WorkspaceMember{{WorkspaceID: suite.workspaceID}}, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().ClaimInvitation(suite.Ctx, mock.Anything, inv.ID, suite.now).Return(true, nil).Once()
	suite.MockUserRepo.EXPECT().UpdatePassword(suite.Ctx, mock.Anything, placeholder.ID, "long-enough").Return(nil).Once()
	suite.MockUserRepo.EXPECT().Update(suite.Ctx, mock.Anything, placeholder.ID, mock.Anything).Return(claimed, nil).Once()
	suite.MockWorkspaceRepo.EXPECT().AddMember(suite.Ctx, mock.Anything, suite.workspaceID, placeholder.ID, schema.WorkspaceRoleMember).Return(&ent.WorkspaceMember{}, nil).Once()

	user, _, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.Require().NoError(err)
	suite.True(user.Active)
}

func (suite *WorkspaceServiceSuite) TestAcceptLosesTheRace() {
	inv := suite.invitation()
	suite.MockWorkspaceRepo.EXPECT().GetInvitationByTokenHash(suite.Ctx, utils.HashToken("t")).Return(inv, nil).Once()
	suite.MockUserRepo.EXPECT().GetByEmail(suite.Ctx, inv.Email).Return(nil, &ent.NotFoundError{}).Once()
	suite.MockWorkspaceRepo.EXPECT().ClaimInvitation(suite.Ctx, mock.Anything, inv.ID, suite.now).Return(false, nil).Once()
	_, _, err := suite.service.AcceptInvitation(suite.Ctx, &AcceptInvitationInput{Token: "t", Password: utils.ToPtr("long-enough")})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func TestWorkspaceServiceSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceServiceSuite))
}
