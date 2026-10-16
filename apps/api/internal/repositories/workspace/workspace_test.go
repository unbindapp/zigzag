package workspace_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type WorkspaceRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo  *WorkspaceRepository
	owner *ent.User
	other *ent.User
}

func (suite *WorkspaceRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewWorkspaceRepository(suite.DB)
	suite.owner = suite.DB.User.Create().SetEmail("owner@example.com").SetDisplayName("owner").SaveX(suite.Ctx)
	suite.other = suite.DB.User.Create().SetEmail("other@example.com").SetDisplayName("other").SaveX(suite.Ctx)
}

func (suite *WorkspaceRepositorySuite) workspace(slug string) *ent.Workspace {
	created, err := suite.repo.Create(suite.Ctx, nil, &CreateWorkspaceInput{Name: slug, Slug: slug, CreatorID: suite.owner.ID})
	suite.Require().NoError(err)
	return created
}

func (suite *WorkspaceRepositorySuite) TestCreateSeedsStatusesAndAdmin() {
	created := suite.workspace("acme")
	statuses := suite.DB.ProjectStatus.Query().AllX(suite.Ctx)
	suite.Len(statuses, len(DefaultProjectStatuses))
	types := map[schema.ProjectStatusType]bool{}
	for _, s := range statuses {
		types[s.Type] = true
	}
	suite.Len(types, 6, "one status of every type")

	member, err := suite.repo.GetMember(suite.Ctx, nil, created.ID, suite.owner.ID)
	suite.Require().NoError(err)
	suite.Equal(schema.WorkspaceRoleAdmin, member.Role)

	_, err = suite.repo.Create(suite.Ctx, nil, &CreateWorkspaceInput{Name: "dup", Slug: "acme", CreatorID: suite.owner.ID})
	suite.True(ent.IsConstraintError(err))
	taken, _ := suite.repo.SlugExists(suite.Ctx, nil, "acme", nil)
	suite.True(taken)
	taken, _ = suite.repo.SlugExists(suite.Ctx, nil, "acme", &created.ID)
	suite.False(taken)
}

func (suite *WorkspaceRepositorySuite) TestMembers() {
	w := suite.workspace("acme")
	added, err := suite.repo.AddMember(suite.Ctx, nil, w.ID, suite.other.ID, schema.WorkspaceRoleViewer)
	suite.Require().NoError(err)
	suite.Equal(schema.WorkspaceRoleViewer, added.Role)
	suite.Equal("other@example.com", added.Edges.User.Email)

	again, err := suite.repo.AddMember(suite.Ctx, nil, w.ID, suite.other.ID, schema.WorkspaceRoleAdmin)
	suite.Require().NoError(err)
	suite.Equal(schema.WorkspaceRoleViewer, again.Role, "adding twice keeps the existing role")

	updated, err := suite.repo.UpdateMemberRole(suite.Ctx, w.ID, suite.other.ID, schema.WorkspaceRoleMember)
	suite.Require().NoError(err)
	suite.Equal(schema.WorkspaceRoleMember, updated.Role)
	_, err = suite.repo.UpdateMemberRole(suite.Ctx, w.ID, uuid.New(), schema.WorkspaceRoleMember)
	suite.True(ent.IsNotFound(err))

	admins, _ := suite.repo.CountAdmins(suite.Ctx, w.ID)
	suite.Equal(1, admins)
	members, _ := suite.repo.ListMembers(suite.Ctx, w.ID)
	suite.Len(members, 2)

	counts, err := suite.repo.Counts(suite.Ctx, []uuid.UUID{w.ID})
	suite.Require().NoError(err)
	suite.Equal(2, counts[w.ID].Members)
	suite.Equal(0, counts[w.ID].Teams)
}

func (suite *WorkspaceRepositorySuite) TestRemoveMemberLeavesTheWorkspacesTeams() {
	acme := suite.workspace("acme")
	globex := suite.workspace("globex")
	suite.repo.AddMember(suite.Ctx, nil, acme.ID, suite.other.ID, schema.WorkspaceRoleMember)
	suite.repo.AddMember(suite.Ctx, nil, globex.ID, suite.other.ID, schema.WorkspaceRoleMember)
	acmeTeam := suite.DB.Team.Create().SetWorkspaceID(acme.ID).SetName("Eng").SetKey("ENG").SaveX(suite.Ctx)
	globexTeam := suite.DB.Team.Create().SetWorkspaceID(globex.ID).SetName("Eng").SetKey("ENG").SaveX(suite.Ctx)
	suite.DB.TeamMember.Create().SetTeamID(acmeTeam.ID).SetUserID(suite.other.ID).SaveX(suite.Ctx)
	suite.DB.TeamMember.Create().SetTeamID(globexTeam.ID).SetUserID(suite.other.ID).SaveX(suite.Ctx)

	teamIDs, _ := suite.repo.MemberTeamIDs(suite.Ctx, acme.ID)
	suite.Equal([]uuid.UUID{acmeTeam.ID}, teamIDs[suite.other.ID])

	suite.Require().NoError(suite.repo.RemoveMember(suite.Ctx, acme.ID, suite.other.ID))
	suite.Equal(1, suite.DB.TeamMember.Query().CountX(suite.Ctx), "only the globex team membership stays")
	suite.True(ent.IsNotFound(suite.repo.RemoveMember(suite.Ctx, acme.ID, suite.other.ID)))
}

func (suite *WorkspaceRepositorySuite) TestInvitations() {
	w := suite.workspace("acme")
	now := time.Now()
	first, err := suite.repo.CreateInvitation(suite.Ctx, &CreateInvitationInput{WorkspaceID: w.ID, Email: "Grace@Example.com", Role: schema.WorkspaceRoleMember, TokenHash: "h1", ExpiresAt: now.Add(time.Hour), InvitedBy: suite.owner.ID})
	suite.Require().NoError(err)
	suite.Equal("grace@example.com", first.Email)
	suite.NotNil(first.TeamIds)

	second, err := suite.repo.CreateInvitation(suite.Ctx, &CreateInvitationInput{WorkspaceID: w.ID, Email: "grace@example.com", Role: schema.WorkspaceRoleAdmin, TokenHash: "h2", ExpiresAt: now.Add(time.Hour), InvitedBy: suite.owner.ID})
	suite.Require().NoError(err)
	_, err = suite.repo.GetInvitationByTokenHash(suite.Ctx, "h1")
	suite.True(ent.IsNotFound(err), "re-inviting replaces the earlier invitation")

	found, err := suite.repo.GetInvitationByTokenHash(suite.Ctx, "h2")
	suite.Require().NoError(err)
	suite.Equal(w.Name, found.Edges.Workspace.Name)

	pending, _ := suite.repo.ListPendingInvitations(suite.Ctx, w.ID, now)
	suite.Len(pending, 1)

	claimed, err := suite.repo.ClaimInvitation(suite.Ctx, nil, second.ID, now)
	suite.NoError(err)
	suite.True(claimed)
	claimed, _ = suite.repo.ClaimInvitation(suite.Ctx, nil, second.ID, now)
	suite.False(claimed, "an invitation is accepted once")

	pending, _ = suite.repo.ListPendingInvitations(suite.Ctx, w.ID, now)
	suite.Empty(pending)

	suite.True(ent.IsNotFound(suite.repo.DeleteInvitation(suite.Ctx, uuid.New(), second.ID)), "invitations are scoped to their workspace")
	suite.NoError(suite.repo.DeleteInvitation(suite.Ctx, w.ID, second.ID))
}

func (suite *WorkspaceRepositorySuite) TestDeleteCascades() {
	w := suite.workspace("acme")
	suite.DB.Team.Create().SetWorkspaceID(w.ID).SetName("Eng").SetKey("ENG").SaveX(suite.Ctx)
	suite.Require().NoError(suite.repo.Delete(suite.Ctx, w.ID))
	suite.Zero(suite.DB.Team.Query().CountX(suite.Ctx))
	suite.Zero(suite.DB.WorkspaceMember.Query().CountX(suite.Ctx))
	suite.Zero(suite.DB.ProjectStatus.Query().CountX(suite.Ctx))
}

func (suite *WorkspaceRepositorySuite) upload(workspaceID uuid.UUID, key string) *ent.Upload {
	return suite.DB.Upload.Create().SetWorkspaceID(workspaceID).SetFileName("logo.png").SetContentType("image/png").SetSizeBytes(1).SetStorageKey(key).SaveX(suite.Ctx)
}

func (suite *WorkspaceRepositorySuite) TestIconUpload() {
	w := suite.workspace("acme")
	logo := suite.upload(w.ID, "logo")
	updated, err := suite.repo.Update(suite.Ctx, w.ID, &UpdateWorkspaceInput{IconUploadID: &logo.ID})
	suite.Require().NoError(err)
	suite.Equal(&logo.ID, updated.IconUploadID)

	updated, err = suite.repo.Update(suite.Ctx, w.ID, &UpdateWorkspaceInput{ClearIconUpload: true})
	suite.Require().NoError(err)
	suite.Nil(updated.IconUploadID)

	_, err = suite.repo.Update(suite.Ctx, w.ID, &UpdateWorkspaceInput{IconUploadID: &logo.ID})
	suite.Require().NoError(err)
	suite.DB.Upload.DeleteOneID(logo.ID).ExecX(suite.Ctx)
	suite.Nil(suite.DB.Workspace.GetX(suite.Ctx, w.ID).IconUploadID, "deleting the upload clears the logo")

	other := suite.workspace("other")
	otherLogo := suite.upload(other.ID, "other-logo")
	_, err = suite.repo.Update(suite.Ctx, other.ID, &UpdateWorkspaceInput{IconUploadID: &otherLogo.ID})
	suite.Require().NoError(err)
	suite.Require().NoError(suite.repo.Delete(suite.Ctx, other.ID), "a workspace with a logo can be deleted")
	suite.Zero(suite.DB.Upload.Query().CountX(suite.Ctx))
}

func TestWorkspaceRepositorySuite(t *testing.T) {
	suite.Run(t, new(WorkspaceRepositorySuite))
}

func TestWorkspaceRepositoryOnPostgres(t *testing.T) {
	if !repository.PostgresConfigured() {
		t.Skip("set " + repository.PostgresURLEnv + " to run the workspace repository against Postgres")
	}
	run := new(WorkspaceRepositorySuite)
	run.Postgres = true
	suite.Run(t, run)
}
