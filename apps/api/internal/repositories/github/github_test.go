package github_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type GithubRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo      *GithubRepository
	workspace *ent.Workspace
	user      *ent.User
}

func (suite *GithubRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewGithubRepository(suite.DB)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.user = suite.DB.User.Create().SetEmail("admin@example.com").SetDisplayName("admin").SetPasswordHash("x").SaveX(suite.Ctx)
}

func (suite *GithubRepositorySuite) create() *ent.GithubIntegration {
	created, err := suite.repo.Create(suite.Ctx, &CreateIntegrationInput{WorkspaceID: suite.workspace.ID, CreatedByID: suite.user.ID, SecretEncrypted: []byte("sealed"), SecretPrefix: "zzgh_abcd"})
	suite.Require().NoError(err)
	return created
}

func (suite *GithubRepositorySuite) TestCreateLoadsTheCreator() {
	created := suite.create()
	suite.Equal(suite.workspace.ID, created.WorkspaceID)
	suite.Require().NotNil(created.Edges.CreatedBy)
	suite.Equal(suite.user.ID, created.Edges.CreatedBy.ID)
	suite.Empty(created.Edges.Repositories)
}

func (suite *GithubRepositorySuite) TestOneIntegrationPerWorkspace() {
	suite.create()
	_, err := suite.repo.Create(suite.Ctx, &CreateIntegrationInput{WorkspaceID: suite.workspace.ID, CreatedByID: suite.user.ID, SecretEncrypted: []byte("x"), SecretPrefix: "zzgh_x"})
	suite.True(ent.IsConstraintError(err))
}

func (suite *GithubRepositorySuite) TestUpdateSecret() {
	created := suite.create()
	updated, err := suite.repo.UpdateSecret(suite.Ctx, created.ID, []byte("resealed"), "zzgh_efgh")
	suite.Require().NoError(err)
	suite.Equal([]byte("resealed"), updated.SecretEncrypted)
	suite.Equal("zzgh_efgh", updated.SecretPrefix)
}

func (suite *GithubRepositorySuite) TestRecordDeliveryUpsertsRepositories() {
	created := suite.create()
	first := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	second := first.Add(time.Minute)

	suite.Require().NoError(suite.repo.RecordDelivery(suite.Ctx, created.ID, "unbindapp/zigzag", "ping", first))
	suite.Require().NoError(suite.repo.RecordDelivery(suite.Ctx, created.ID, "unbindapp/unbind", "push", first))
	suite.Require().NoError(suite.repo.RecordDelivery(suite.Ctx, created.ID, "unbindapp/zigzag", "pull_request.opened", second))
	suite.Require().NoError(suite.repo.RecordDelivery(suite.Ctx, created.ID, "", "ping", second))

	found, err := suite.repo.GetByWorkspace(suite.Ctx, suite.workspace.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(found.LastDeliveryAt)
	suite.True(found.LastDeliveryAt.Equal(second))
	suite.Require().Len(found.Edges.Repositories, 2)
	suite.Equal("unbindapp/zigzag", found.Edges.Repositories[0].FullName)
	suite.Equal("pull_request.opened", found.Edges.Repositories[0].LastEvent)
	suite.True(found.Edges.Repositories[0].LastEventAt.Equal(second))
	suite.Equal("unbindapp/unbind", found.Edges.Repositories[1].FullName)
}

func (suite *GithubRepositorySuite) TestDeleteRemovesRepositories() {
	created := suite.create()
	suite.Require().NoError(suite.repo.RecordDelivery(suite.Ctx, created.ID, "unbindapp/zigzag", "ping", time.Now()))
	suite.Require().NoError(suite.repo.Delete(suite.Ctx, created.ID))

	_, err := suite.repo.GetByWorkspace(suite.Ctx, suite.workspace.ID)
	suite.True(ent.IsNotFound(err))
	suite.Zero(suite.DB.GithubRepository.Query().CountX(suite.Ctx))
}

func (suite *GithubRepositorySuite) TestGetByWorkspaceIsScoped() {
	suite.create()
	_, err := suite.repo.GetByWorkspace(suite.Ctx, uuid.New())
	suite.True(ent.IsNotFound(err))
}

func TestGithubRepositorySuite(t *testing.T) {
	suite.Run(t, new(GithubRepositorySuite))
}
