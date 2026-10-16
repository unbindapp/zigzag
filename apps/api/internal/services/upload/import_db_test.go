package upload_service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/uploadblob"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type ImportDirSuite struct {
	repositorytest.RepositoryBaseSuite
	service     *UploadService
	dir         string
	workspaceID uuid.UUID
}

func (suite *ImportDirSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.service = NewUploadService(&config.Config{}, repositories.NewRepositories(suite.DB))
	suite.dir = suite.T().TempDir()
	suite.workspaceID = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx).ID
}

func (suite *ImportDirSuite) upload(ext string) *ent.Upload {
	id := uuid.New()
	return suite.DB.Upload.Create().
		SetID(id).
		SetWorkspaceID(suite.workspaceID).
		SetFileName("file" + ext).
		SetContentType("application/octet-stream").
		SetSizeBytes(1).
		SetStorageKey(suite.workspaceID.String() + "/" + id.String() + ext).
		SaveX(suite.Ctx)
}

func (suite *ImportDirSuite) writeFile(upload *ent.Upload, content string) {
	path := filepath.Join(suite.dir, filepath.FromSlash(upload.StorageKey))
	suite.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o750))
	suite.Require().NoError(os.WriteFile(path, []byte(content), 0o600))
}

func (suite *ImportDirSuite) data(upload *ent.Upload) string {
	blob, err := suite.DB.UploadBlob.Query().Where(uploadblob.UploadID(upload.ID)).Only(suite.Ctx)
	if ent.IsNotFound(err) {
		return ""
	}
	suite.Require().NoError(err)
	return string(blob.Data)
}

func (suite *ImportDirSuite) TestImportsFilesOfUploadsWithoutBytes() {
	png := suite.upload(".png")
	noExtension := suite.upload("")
	gone := suite.upload(".pdf")
	suite.writeFile(png, "png bytes")
	suite.writeFile(noExtension, "plain bytes")

	result, err := suite.service.ImportDir(suite.Ctx, suite.dir)
	suite.Require().NoError(err)
	suite.Equal(&ImportResult{Imported: 2, Missing: 1}, result)
	suite.Equal("png bytes", suite.data(png))
	suite.Equal("plain bytes", suite.data(noExtension))
	suite.Empty(suite.data(gone))
}

func (suite *ImportDirSuite) TestRunningAgainKeepsImportedBytes() {
	stored := suite.upload(".txt")
	suite.writeFile(stored, "first")
	_, err := suite.service.ImportDir(suite.Ctx, suite.dir)
	suite.Require().NoError(err)
	suite.writeFile(stored, "changed on disk")

	result, err := suite.service.ImportDir(suite.Ctx, suite.dir)
	suite.Require().NoError(err)
	suite.Equal(&ImportResult{}, result)
	suite.Equal("first", suite.data(stored))
}

func (suite *ImportDirSuite) TestKeysCannotLeaveTheDirectory() {
	outside := suite.DB.Upload.Create().
		SetWorkspaceID(suite.workspaceID).
		SetFileName("secret").
		SetContentType("text/plain").
		SetSizeBytes(1).
		SetStorageKey("../secret").
		SaveX(suite.Ctx)
	suite.Require().NoError(os.WriteFile(filepath.Join(filepath.Dir(suite.dir), "secret"), []byte("secret"), 0o600))

	result, err := suite.service.ImportDir(suite.Ctx, suite.dir)
	suite.Require().NoError(err)
	suite.Equal(&ImportResult{Failed: 1}, result)
	suite.Empty(suite.data(outside))
}

func (suite *ImportDirSuite) TestAMissingDirectoryIsAnError() {
	_, err := suite.service.ImportDir(suite.Ctx, filepath.Join(suite.dir, "nope"))
	suite.Error(err)
}

func TestImportDirSuite(t *testing.T) {
	suite.Run(t, new(ImportDirSuite))
}

func TestImportDirOnPostgres(t *testing.T) {
	if !repositorytest.PostgresConfigured() {
		t.Skip("set " + repositorytest.PostgresURLEnv + " to run the upload import against Postgres")
	}
	run := new(ImportDirSuite)
	run.Postgres = true
	suite.Run(t, run)
}
