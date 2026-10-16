package upload_repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type UploadRepositorySuite struct {
	repositorytest.RepositoryBaseSuite
	repo *UploadRepository
}

func (suite *UploadRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewUploadRepository(suite.DB)
}

func (suite *UploadRepositorySuite) TestListAfterPagesInIDOrder() {
	user := suite.DB.User.Create().SetEmail("ada@example.com").SetName("Ada").SetDisplayName("ada").SaveX(suite.Ctx)
	workspace := suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	created := map[uuid.UUID]bool{}
	for range 5 {
		upload, err := suite.repo.Create(suite.Ctx, &CreateUploadInput{ID: uuid.New(), WorkspaceID: workspace.ID, UploaderID: user.ID, FileName: "f", ContentType: "text/plain", SizeBytes: 1, StorageKey: uuid.NewString()})
		suite.Require().NoError(err)
		created[upload.ID] = true
	}

	seen := map[uuid.UUID]bool{}
	after := uuid.Nil
	for {
		page, err := suite.repo.ListAfter(suite.Ctx, after, 2)
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
	suite.Equal(created, seen)
}

func TestUploadRepositorySuite(t *testing.T) {
	suite.Run(t, new(UploadRepositorySuite))
}
