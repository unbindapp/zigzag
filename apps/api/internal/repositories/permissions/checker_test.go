package permissions_repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type CheckerSuite struct {
	repository.RepositoryBaseSuite
	repo *PermissionsRepository

	admin    *ent.User
	member   *ent.User
	viewer   *ent.User
	outsider *ent.User
	sysadmin *ent.User
	inactive *ent.User

	acme   *ent.Workspace
	globex *ent.Workspace
}

func (suite *CheckerSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewPermissionsRepository(suite.DB)

	user := func(name string) *ent.User {
		return suite.DB.User.Create().SetEmail(name + "@example.com").SetDisplayName(name).SaveX(suite.Ctx)
	}
	suite.admin = user("admin")
	suite.member = user("member")
	suite.viewer = user("viewer")
	suite.outsider = user("outsider")
	suite.sysadmin = suite.DB.User.Create().SetEmail("root@example.com").SetDisplayName("root").SetSystemAdmin(true).SaveX(suite.Ctx)
	suite.inactive = suite.DB.User.Create().SetEmail("gone@example.com").SetDisplayName("gone").SetActive(false).SaveX(suite.Ctx)

	suite.acme = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.globex = suite.DB.Workspace.Create().SetName("Globex").SetSlug("globex").SaveX(suite.Ctx)

	join := func(u *ent.User, w *ent.Workspace, role schema.WorkspaceRole) {
		suite.DB.WorkspaceMember.Create().SetUserID(u.ID).SetWorkspaceID(w.ID).SetRole(role).SaveX(suite.Ctx)
	}
	join(suite.admin, suite.acme, schema.WorkspaceRoleAdmin)
	join(suite.member, suite.acme, schema.WorkspaceRoleMember)
	join(suite.viewer, suite.acme, schema.WorkspaceRoleViewer)
	join(suite.member, suite.globex, schema.WorkspaceRoleAdmin)
	join(suite.inactive, suite.acme, schema.WorkspaceRoleAdmin)
}

func (suite *CheckerSuite) check(ctx context.Context, user *ent.User, role schema.WorkspaceRole, workspace uuid.UUID) error {
	return suite.repo.Check(ctx, user.ID, []PermissionCheck{{Role: role, WorkspaceID: workspace}})
}

func (suite *CheckerSuite) TestRoleImplication() {
	tests := []struct {
		name    string
		user    func() *ent.User
		role    schema.WorkspaceRole
		allowed bool
	}{
		{"admin acts as admin", func() *ent.User { return suite.admin }, schema.WorkspaceRoleAdmin, true},
		{"admin acts as member", func() *ent.User { return suite.admin }, schema.WorkspaceRoleMember, true},
		{"admin acts as viewer", func() *ent.User { return suite.admin }, schema.WorkspaceRoleViewer, true},
		{"member is not admin", func() *ent.User { return suite.member }, schema.WorkspaceRoleAdmin, false},
		{"member acts as member", func() *ent.User { return suite.member }, schema.WorkspaceRoleMember, true},
		{"member acts as viewer", func() *ent.User { return suite.member }, schema.WorkspaceRoleViewer, true},
		{"viewer is not member", func() *ent.User { return suite.viewer }, schema.WorkspaceRoleMember, false},
		{"viewer reads", func() *ent.User { return suite.viewer }, schema.WorkspaceRoleViewer, true},
		{"outsider cannot read", func() *ent.User { return suite.outsider }, schema.WorkspaceRoleViewer, false},
		{"system admin is admin everywhere", func() *ent.User { return suite.sysadmin }, schema.WorkspaceRoleAdmin, true},
		{"inactive user is refused", func() *ent.User { return suite.inactive }, schema.WorkspaceRoleViewer, false},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.check(suite.Ctx, tt.user(), tt.role, suite.acme.ID)
			if tt.allowed {
				suite.NoError(err)
				return
			}
			suite.ErrorIs(err, errdefs.ErrUnauthorized)
		})
	}
}

func (suite *CheckerSuite) TestRolesAreScopedPerWorkspace() {
	suite.NoError(suite.check(suite.Ctx, suite.member, schema.WorkspaceRoleAdmin, suite.globex.ID))
	suite.ErrorIs(suite.check(suite.Ctx, suite.admin, schema.WorkspaceRoleViewer, suite.globex.ID), errdefs.ErrUnauthorized)
}

func (suite *CheckerSuite) TestAnyPassingCheckGrants() {
	err := suite.repo.Check(suite.Ctx, suite.admin.ID, []PermissionCheck{
		{Role: schema.WorkspaceRoleViewer, WorkspaceID: suite.globex.ID},
		{Role: schema.WorkspaceRoleAdmin, WorkspaceID: suite.acme.ID},
	})
	suite.NoError(err)
}

func (suite *CheckerSuite) TestMalformedChecksDeny() {
	suite.ErrorIs(suite.repo.Check(suite.Ctx, suite.admin.ID, nil), errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.check(suite.Ctx, suite.admin, "", suite.acme.ID), errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.check(suite.Ctx, suite.admin, "owner", suite.acme.ID), errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.check(suite.Ctx, suite.sysadmin, schema.WorkspaceRoleViewer, uuid.Nil), errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.check(suite.Ctx, suite.sysadmin, schema.WorkspaceRoleViewer, uuid.New()), errdefs.ErrUnauthorized, "system admins only pass on workspaces that exist")
	suite.ErrorIs(suite.repo.Check(suite.Ctx, uuid.New(), []PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: suite.acme.ID}}), errdefs.ErrUnauthorized)
}

func (suite *CheckerSuite) TestSystemCallerBypasses() {
	suite.NoError(suite.check(WithSystemCaller(suite.Ctx), suite.outsider, schema.WorkspaceRoleAdmin, suite.acme.ID))
}

func (suite *CheckerSuite) TestAPIKeyNarrowing() {
	tests := []struct {
		name    string
		user    func() *ent.User
		access  APIKeyAccess
		role    schema.WorkspaceRole
		ws      func() uuid.UUID
		allowed bool
	}{
		{"full access viewer key reads", func() *ent.User { return suite.admin }, APIKeyAccess{Role: schema.WorkspaceRoleViewer, FullAccess: true}, schema.WorkspaceRoleViewer, func() uuid.UUID { return suite.acme.ID }, true},
		{"viewer key cannot write even for an admin", func() *ent.User { return suite.admin }, APIKeyAccess{Role: schema.WorkspaceRoleViewer, FullAccess: true}, schema.WorkspaceRoleMember, func() uuid.UUID { return suite.acme.ID }, false},
		{"member key writes", func() *ent.User { return suite.admin }, APIKeyAccess{Role: schema.WorkspaceRoleMember, FullAccess: true}, schema.WorkspaceRoleMember, func() uuid.UUID { return suite.acme.ID }, true},
		{"member key is not admin", func() *ent.User { return suite.admin }, APIKeyAccess{Role: schema.WorkspaceRoleMember, FullAccess: true}, schema.WorkspaceRoleAdmin, func() uuid.UUID { return suite.acme.ID }, false},
		{"admin key never exceeds the owner", func() *ent.User { return suite.viewer }, APIKeyAccess{Role: schema.WorkspaceRoleAdmin, FullAccess: true}, schema.WorkspaceRoleMember, func() uuid.UUID { return suite.acme.ID }, false},
		{"listed workspace passes", func() *ent.User { return suite.member }, APIKeyAccess{Role: schema.WorkspaceRoleAdmin, WorkspaceIDs: []uuid.UUID{suite.globex.ID}}, schema.WorkspaceRoleAdmin, func() uuid.UUID { return suite.globex.ID }, true},
		{"unlisted workspace is refused", func() *ent.User { return suite.member }, APIKeyAccess{Role: schema.WorkspaceRoleAdmin, WorkspaceIDs: []uuid.UUID{suite.globex.ID}}, schema.WorkspaceRoleViewer, func() uuid.UUID { return suite.acme.ID }, false},
		{"system admin key is narrowed too", func() *ent.User { return suite.sysadmin }, APIKeyAccess{Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{suite.globex.ID}}, schema.WorkspaceRoleViewer, func() uuid.UUID { return suite.acme.ID }, false},
		{"system admin key reaches listed workspaces", func() *ent.User { return suite.sysadmin }, APIKeyAccess{Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{suite.acme.ID}}, schema.WorkspaceRoleViewer, func() uuid.UUID { return suite.acme.ID }, true},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := WithAPIKeyAccess(suite.Ctx, tt.access)
			err := suite.check(ctx, tt.user(), tt.role, tt.ws())
			if tt.allowed {
				suite.NoError(err)
				return
			}
			suite.ErrorIs(err, errdefs.ErrUnauthorized)
		})
	}
}

func (suite *CheckerSuite) TestCheckSystemAdmin() {
	suite.NoError(suite.repo.CheckSystemAdmin(suite.Ctx, suite.sysadmin.ID))
	suite.ErrorIs(suite.repo.CheckSystemAdmin(suite.Ctx, suite.admin.ID), errdefs.ErrUnauthorized)
	suite.ErrorIs(suite.repo.CheckSystemAdmin(WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleMember, FullAccess: true}), suite.sysadmin.ID), errdefs.ErrUnauthorized)
	suite.NoError(suite.repo.CheckSystemAdmin(WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleAdmin, FullAccess: true}), suite.sysadmin.ID))
	suite.ErrorIs(suite.repo.CheckSystemAdmin(WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleAdmin, WorkspaceIDs: []uuid.UUID{suite.acme.ID}}), suite.sysadmin.ID), errdefs.ErrUnauthorized)
}

func (suite *CheckerSuite) TestPermissionSet() {
	set, err := suite.repo.GetUserPermissionSet(suite.Ctx, suite.member.ID)
	suite.Require().NoError(err)
	suite.False(set.IsSystemAdmin)
	role, ok := set.WorkspaceRole(suite.acme.ID)
	suite.True(ok)
	suite.Equal(schema.WorkspaceRoleMember, role)
	role, _ = set.WorkspaceRole(suite.globex.ID)
	suite.Equal(schema.WorkspaceRoleAdmin, role)
	suite.ElementsMatch([]uuid.UUID{suite.acme.ID, suite.globex.ID}, set.VisibleWorkspaceIDs())
	suite.True(set.Can(schema.WorkspaceRoleViewer, suite.acme.ID))
	suite.False(set.Can(schema.WorkspaceRoleAdmin, suite.acme.ID))

	set, err = suite.repo.GetUserPermissionSet(suite.Ctx, suite.outsider.ID)
	suite.Require().NoError(err)
	suite.Empty(set.VisibleWorkspaceIDs())

	set, err = suite.repo.GetUserPermissionSet(suite.Ctx, suite.inactive.ID)
	suite.Require().NoError(err)
	suite.Empty(set.VisibleWorkspaceIDs())
}

func (suite *CheckerSuite) TestPermissionSetForSystemAdmin() {
	set, err := suite.repo.GetUserPermissionSet(suite.Ctx, suite.sysadmin.ID)
	suite.Require().NoError(err)
	suite.True(set.IsSystemAdmin)
	suite.ElementsMatch([]uuid.UUID{suite.acme.ID, suite.globex.ID}, set.VisibleWorkspaceIDs())
	role, _ := set.WorkspaceRole(suite.globex.ID)
	suite.Equal(schema.WorkspaceRoleAdmin, role)
}

func (suite *CheckerSuite) TestPermissionSetIsNarrowedByKey() {
	ctx := WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleViewer, WorkspaceIDs: []uuid.UUID{suite.globex.ID}})
	set, err := suite.repo.GetUserPermissionSet(ctx, suite.member.ID)
	suite.Require().NoError(err)
	suite.Equal([]uuid.UUID{suite.globex.ID}, set.VisibleWorkspaceIDs())
	role, _ := set.WorkspaceRole(suite.globex.ID)
	suite.Equal(schema.WorkspaceRoleViewer, role, "the weaker of the member role and the key role")

	ctx = WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleMember, FullAccess: true})
	set, err = suite.repo.GetUserPermissionSet(ctx, suite.sysadmin.ID)
	suite.Require().NoError(err)
	suite.False(set.IsSystemAdmin, "only a full access admin key keeps system admin")
	role, _ = set.WorkspaceRole(suite.acme.ID)
	suite.Equal(schema.WorkspaceRoleMember, role)
}

func (suite *CheckerSuite) TestHasPrivilege() {
	privilege := schema.KeyPrivilege("test_privilege")
	suite.True(HasPrivilege(suite.Ctx, privilege), "sessions hold every privilege")
	suite.False(HasPrivilege(WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleAdmin, FullAccess: true}), privilege))
	suite.True(HasPrivilege(WithAPIKeyAccess(suite.Ctx, APIKeyAccess{Role: schema.WorkspaceRoleViewer, Privileges: []schema.KeyPrivilege{privilege}}), privilege))
}

func TestCheckerSuite(t *testing.T) {
	suite.Run(t, new(CheckerSuite))
}

func TestRoleHelpers(t *testing.T) {
	cases := []struct {
		held, required schema.WorkspaceRole
		implies        bool
	}{
		{schema.WorkspaceRoleAdmin, schema.WorkspaceRoleViewer, true},
		{schema.WorkspaceRoleMember, schema.WorkspaceRoleAdmin, false},
		{schema.WorkspaceRoleViewer, schema.WorkspaceRoleViewer, true},
		{"", schema.WorkspaceRoleViewer, false},
		{"owner", "", false},
	}
	for _, c := range cases {
		if got := RoleImplies(c.held, c.required); got != c.implies {
			t.Errorf("RoleImplies(%q, %q) = %v", c.held, c.required, got)
		}
	}
	if WeakerRole(schema.WorkspaceRoleAdmin, schema.WorkspaceRoleViewer) != schema.WorkspaceRoleViewer {
		t.Error("WeakerRole should pick viewer")
	}
}
