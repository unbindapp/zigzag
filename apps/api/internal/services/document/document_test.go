package document_service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	document_repo "github.com/unbindapp/zigzag-api/internal/repositories/document"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type DocumentServiceSuite struct {
	services.ServiceTestSuite
	service     *DocumentService
	now         time.Time
	userID      uuid.UUID
	workspaceID uuid.UUID
	document    *ent.Document
}

func (suite *DocumentServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewDocumentService(nil, suite.MockRepo)
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.document = &ent.Document{ID: uuid.New(), WorkspaceID: suite.workspaceID, Title: "Spec", Slug: "spec-1a2b3c4d"}
}

func TestDocumentServiceSuite(t *testing.T) {
	suite.Run(t, new(DocumentServiceSuite))
}

func (suite *DocumentServiceSuite) roles(granted schema.WorkspaceRole) {
	allowed := true
	for _, role := range []schema.WorkspaceRole{schema.WorkspaceRoleViewer, schema.WorkspaceRoleMember, schema.WorkspaceRoleAdmin} {
		var result error
		if !allowed {
			result = errdefs.ErrUnauthorized
		}
		suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
		if role == granted {
			allowed = false
		}
	}
}

func (suite *DocumentServiceSuite) expectDocument() {
	suite.MockDocumentRepo.EXPECT().GetByID(suite.Ctx, suite.document.ID).Return(suite.document, nil).Maybe()
}

func (suite *DocumentServiceSuite) TestGetBySlugHidesOtherWorkspaces() {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, mock.Anything).Return(errdefs.ErrUnauthorized).Once()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetDocumentInput{WorkspaceID: suite.workspaceID, Slug: suite.document.Slug})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *DocumentServiceSuite) TestGetNeedsAnIdOrASlug() {
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetDocumentInput{Slug: "spec"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *DocumentServiceSuite) TestGetMissingSlugIsNotFound() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.MockDocumentRepo.EXPECT().GetBySlug(suite.Ctx, suite.workspaceID, "gone").Return(nil, &ent.NotFoundError{}).Once()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetDocumentInput{WorkspaceID: suite.workspaceID, Slug: "gone"})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *DocumentServiceSuite) TestCreateNeedsMember() {
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateDocumentInput{WorkspaceID: suite.workspaceID, Title: "Spec"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *DocumentServiceSuite) TestCreateRejectsParentsOfOtherWorkspaces() {
	suite.roles(schema.WorkspaceRoleMember)
	project := uuid.New()
	suite.MockProjectRepo.EXPECT().ResolveWorkspaceOfProject(suite.Ctx, project).Return(uuid.New(), nil).Once()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateDocumentInput{WorkspaceID: suite.workspaceID, Title: "Spec", ProjectID: project})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	issue := uuid.New()
	suite.MockIssueRepo.EXPECT().ResolveWorkspaceOfIssue(suite.Ctx, issue).Return(uuid.Nil, &ent.NotFoundError{}).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateDocumentInput{WorkspaceID: suite.workspaceID, Title: "Spec", IssueID: issue})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *DocumentServiceSuite) TestCreateDefaultsTheTitle() {
	suite.roles(schema.WorkspaceRoleMember)
	initiative := uuid.New()
	suite.MockInitiativeRepo.EXPECT().ResolveWorkspaceOfInitiative(suite.Ctx, initiative).Return(suite.workspaceID, nil).Once()
	suite.MockDocumentRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *document_repo.CreateDocumentInput) bool {
		return input.Title == "Untitled" && input.Slug[:9] == "untitled-" && *input.InitiativeID == initiative && *input.CreatorID == suite.userID
	})).Return(suite.document, nil).Once()
	suite.expectDocument()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateDocumentInput{WorkspaceID: suite.workspaceID, Title: "  ", InitiativeID: initiative})
	suite.NoError(err)
}

func (suite *DocumentServiceSuite) TestUpdateDetachesWithTheNilUUID() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectDocument()
	content := "# Spec"
	suite.MockDocumentRepo.EXPECT().Update(suite.Ctx, suite.document.ID, mock.MatchedBy(func(input *document_repo.UpdateDocumentInput) bool {
		return input.ClearProject && input.ProjectID == nil && *input.Content == content && input.UpdatedByID == suite.userID && input.InitiativeID == nil
	})).Return(nil).Once()
	detach := uuid.Nil
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateDocumentInput{DocumentID: suite.document.ID, ProjectID: &detach, Content: &content})
	suite.NoError(err)
}

func (suite *DocumentServiceSuite) TestUpdateNeedsMember() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectDocument()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateDocumentInput{DocumentID: suite.document.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *DocumentServiceSuite) TestDeleteNeedsAdmin() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectDocument()
	err := suite.service.Delete(suite.Ctx, suite.userID, &DocumentRefInput{DocumentID: suite.document.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *DocumentServiceSuite) TestTrashAndRestore() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectDocument()
	suite.MockDocumentRepo.EXPECT().SetTrashedAt(suite.Ctx, suite.document.ID, &suite.now).Return(nil).Once()
	suite.MockDocumentRepo.EXPECT().SetTrashedAt(suite.Ctx, suite.document.ID, (*time.Time)(nil)).Return(nil).Once()
	_, err := suite.service.Trash(suite.Ctx, suite.userID, &DocumentRefInput{DocumentID: suite.document.ID})
	suite.NoError(err)
	_, err = suite.service.Restore(suite.Ctx, suite.userID, &DocumentRefInput{DocumentID: suite.document.ID})
	suite.NoError(err)
}
