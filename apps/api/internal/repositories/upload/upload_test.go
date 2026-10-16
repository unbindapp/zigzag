package upload_repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type UploadRepositorySuite struct {
	repositorytest.RepositoryBaseSuite
	repo        *UploadRepository
	userID      uuid.UUID
	workspaceID uuid.UUID
}

func (suite *UploadRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewUploadRepository(suite.DB)
	suite.userID = suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada").SetDisplayName("ada").SaveX(suite.Ctx).ID
	suite.workspaceID = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx).ID
}

func (suite *UploadRepositorySuite) create(data []byte) *ent.Upload {
	upload, err := suite.repo.Create(suite.Ctx, &CreateUploadInput{ID: uuid.New(), WorkspaceID: suite.workspaceID, UploaderID: suite.userID, FileName: "f.txt", ContentType: "text/plain", StorageKey: uuid.NewString(), Data: data})
	suite.Require().NoError(err)
	return upload
}

func (suite *UploadRepositorySuite) metadataOnly() *ent.Upload {
	return suite.DB.Upload.Create().SetWorkspaceID(suite.workspaceID).SetFileName("old.png").SetContentType("image/png").SetSizeBytes(3).SetStorageKey(uuid.NewString()).SaveX(suite.Ctx)
}

func (suite *UploadRepositorySuite) TestCreateStoresTheBytesAndTheirSize() {
	upload := suite.create([]byte("hello"))
	suite.Equal(int64(5), upload.SizeBytes)

	data, err := suite.repo.GetData(suite.Ctx, upload.ID)
	suite.Require().NoError(err)
	suite.Equal([]byte("hello"), data)
}

func (suite *UploadRepositorySuite) TestGetDataOfAnUploadWithoutBytesIsNotFound() {
	_, err := suite.repo.GetData(suite.Ctx, suite.metadataOnly().ID)
	suite.True(ent.IsNotFound(err))
}

func (suite *UploadRepositorySuite) TestDeletingAnUploadDeletesItsBytes() {
	upload := suite.create([]byte("hello"))
	suite.Require().NoError(suite.repo.Delete(suite.Ctx, upload.ID))
	suite.Zero(suite.DB.UploadBlob.Query().CountX(suite.Ctx))
}

func (suite *UploadRepositorySuite) TestListMissingDataPagesInIDOrderAndSkipsStoredUploads() {
	suite.create([]byte("stored"))
	missing := map[uuid.UUID]bool{}
	for range 5 {
		missing[suite.metadataOnly().ID] = true
	}

	seen := map[uuid.UUID]bool{}
	after := uuid.Nil
	for {
		page, err := suite.repo.ListMissingData(suite.Ctx, after, 2)
		suite.Require().NoError(err)
		for _, upload := range page {
			suite.Greater(upload.ID.String(), after.String())
			seen[upload.ID] = true
			after = upload.ID
		}
		if len(page) < 2 {
			break
		}
	}
	suite.Equal(missing, seen)
}

func (suite *UploadRepositorySuite) TestCreateDataKeepsExistingBytes() {
	upload := suite.metadataOnly()
	suite.Require().NoError(suite.repo.CreateData(suite.Ctx, upload.ID, []byte("first")))
	suite.Require().NoError(suite.repo.CreateData(suite.Ctx, upload.ID, []byte("second")))

	data, err := suite.repo.GetData(suite.Ctx, upload.ID)
	suite.Require().NoError(err)
	suite.Equal([]byte("first"), data)
}

func TestUploadRepositorySuite(t *testing.T) {
	suite.Run(t, new(UploadRepositorySuite))
}

func TestUploadRepositoryOnPostgres(t *testing.T) {
	if !repositorytest.PostgresConfigured() {
		t.Skip("set " + repositorytest.PostgresURLEnv + " to run the upload repository against Postgres")
	}
	run := new(UploadRepositorySuite)
	run.Postgres = true
	suite.Run(t, run)
}
