package services

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	mocks_repositories "github.com/zigzagapp-org/zigzag-api/mocks/repositories"
	mocks_repository_apikey "github.com/zigzagapp-org/zigzag-api/mocks/repository/apikey"
	mocks_repository_comment "github.com/zigzagapp-org/zigzag-api/mocks/repository/comment"
	mocks_repository_cycle "github.com/zigzagapp-org/zigzag-api/mocks/repository/cycle"
	mocks_repository_document "github.com/zigzagapp-org/zigzag-api/mocks/repository/document"
	mocks_repository_favorite "github.com/zigzagapp-org/zigzag-api/mocks/repository/favorite"
	mocks_repository_github "github.com/zigzagapp-org/zigzag-api/mocks/repository/github"
	mocks_repository_importjob "github.com/zigzagapp-org/zigzag-api/mocks/repository/importjob"
	mocks_repository_initiative "github.com/zigzagapp-org/zigzag-api/mocks/repository/initiative"
	mocks_repository_issue "github.com/zigzagapp-org/zigzag-api/mocks/repository/issue"
	mocks_repository_label "github.com/zigzagapp-org/zigzag-api/mocks/repository/label"
	mocks_repository_notification "github.com/zigzagapp-org/zigzag-api/mocks/repository/notification"
	mocks_repository_oauth "github.com/zigzagapp-org/zigzag-api/mocks/repository/oauth"
	mocks_repository_oauthserver "github.com/zigzagapp-org/zigzag-api/mocks/repository/oauthserver"
	mocks_repository_permissions "github.com/zigzagapp-org/zigzag-api/mocks/repository/permissions"
	mocks_repository_project "github.com/zigzagapp-org/zigzag-api/mocks/repository/project"
	mocks_repository_system "github.com/zigzagapp-org/zigzag-api/mocks/repository/system"
	mocks_repository_team "github.com/zigzagapp-org/zigzag-api/mocks/repository/team"
	mocks_repository_template "github.com/zigzagapp-org/zigzag-api/mocks/repository/template"
	mocks_repository_tx "github.com/zigzagapp-org/zigzag-api/mocks/repository/tx"
	mocks_repository_upload "github.com/zigzagapp-org/zigzag-api/mocks/repository/upload"
	mocks_repository_user "github.com/zigzagapp-org/zigzag-api/mocks/repository/user"
	mocks_repository_view "github.com/zigzagapp-org/zigzag-api/mocks/repository/view"
	mocks_repository_webhook "github.com/zigzagapp-org/zigzag-api/mocks/repository/webhook"
	mocks_repository_workspace "github.com/zigzagapp-org/zigzag-api/mocks/repository/workspace"
	mocks_services_email "github.com/zigzagapp-org/zigzag-api/mocks/services/email"
)

type ServiceTestSuite struct {
	suite.Suite
	Ctx context.Context

	MockRepo             *mocks_repositories.RepositoriesMock
	MockEmail            *mocks_services_email.EmailServiceMock
	MockUserRepo         *mocks_repository_user.UserRepositoryMock
	MockOauthRepo        *mocks_repository_oauth.OauthRepositoryMock
	MockOAuthServerRepo  *mocks_repository_oauthserver.OAuthServerRepositoryMock
	MockAPIKeyRepo       *mocks_repository_apikey.APIKeyRepositoryMock
	MockSystemRepo       *mocks_repository_system.SystemRepositoryMock
	MockWorkspaceRepo    *mocks_repository_workspace.WorkspaceRepositoryMock
	MockTeamRepo         *mocks_repository_team.TeamRepositoryMock
	MockLabelRepo        *mocks_repository_label.LabelRepositoryMock
	MockProjectRepo      *mocks_repository_project.ProjectRepositoryMock
	MockInitiativeRepo   *mocks_repository_initiative.InitiativeRepositoryMock
	MockCycleRepo        *mocks_repository_cycle.CycleRepositoryMock
	MockIssueRepo        *mocks_repository_issue.IssueRepositoryMock
	MockCommentRepo      *mocks_repository_comment.CommentRepositoryMock
	MockNotificationRepo *mocks_repository_notification.NotificationRepositoryMock
	MockDocumentRepo     *mocks_repository_document.DocumentRepositoryMock
	MockViewRepo         *mocks_repository_view.ViewRepositoryMock
	MockTemplateRepo     *mocks_repository_template.TemplateRepositoryMock
	MockFavoriteRepo     *mocks_repository_favorite.FavoriteRepositoryMock
	MockWebhookRepo      *mocks_repository_webhook.WebhookRepositoryMock
	MockUploadRepo       *mocks_repository_upload.UploadRepositoryMock
	MockImportJobRepo    *mocks_repository_importjob.ImportJobRepositoryMock
	MockGithubRepo       *mocks_repository_github.GithubRepositoryMock
	MockPermissionsRepo  *mocks_repository_permissions.PermissionsRepositoryMock
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.Ctx = context.Background()
	suite.MockRepo = mocks_repositories.NewRepositoriesMock(suite.T())
	suite.MockEmail = mocks_services_email.NewEmailServiceMock(suite.T())
	suite.MockUserRepo = mocks_repository_user.NewUserRepositoryMock(suite.T())
	suite.MockOauthRepo = mocks_repository_oauth.NewOauthRepositoryMock(suite.T())
	suite.MockOAuthServerRepo = mocks_repository_oauthserver.NewOAuthServerRepositoryMock(suite.T())
	suite.MockAPIKeyRepo = mocks_repository_apikey.NewAPIKeyRepositoryMock(suite.T())
	suite.MockSystemRepo = mocks_repository_system.NewSystemRepositoryMock(suite.T())
	suite.MockWorkspaceRepo = mocks_repository_workspace.NewWorkspaceRepositoryMock(suite.T())
	suite.MockTeamRepo = mocks_repository_team.NewTeamRepositoryMock(suite.T())
	suite.MockLabelRepo = mocks_repository_label.NewLabelRepositoryMock(suite.T())
	suite.MockProjectRepo = mocks_repository_project.NewProjectRepositoryMock(suite.T())
	suite.MockInitiativeRepo = mocks_repository_initiative.NewInitiativeRepositoryMock(suite.T())
	suite.MockCycleRepo = mocks_repository_cycle.NewCycleRepositoryMock(suite.T())
	suite.MockIssueRepo = mocks_repository_issue.NewIssueRepositoryMock(suite.T())
	suite.MockCommentRepo = mocks_repository_comment.NewCommentRepositoryMock(suite.T())
	suite.MockNotificationRepo = mocks_repository_notification.NewNotificationRepositoryMock(suite.T())
	suite.MockDocumentRepo = mocks_repository_document.NewDocumentRepositoryMock(suite.T())
	suite.MockViewRepo = mocks_repository_view.NewViewRepositoryMock(suite.T())
	suite.MockTemplateRepo = mocks_repository_template.NewTemplateRepositoryMock(suite.T())
	suite.MockFavoriteRepo = mocks_repository_favorite.NewFavoriteRepositoryMock(suite.T())
	suite.MockWebhookRepo = mocks_repository_webhook.NewWebhookRepositoryMock(suite.T())
	suite.MockUploadRepo = mocks_repository_upload.NewUploadRepositoryMock(suite.T())
	suite.MockImportJobRepo = mocks_repository_importjob.NewImportJobRepositoryMock(suite.T())
	suite.MockGithubRepo = mocks_repository_github.NewGithubRepositoryMock(suite.T())
	suite.MockPermissionsRepo = mocks_repository_permissions.NewPermissionsRepositoryMock(suite.T())

	suite.MockRepo.EXPECT().User().Return(suite.MockUserRepo).Maybe()
	suite.MockRepo.EXPECT().Oauth().Return(suite.MockOauthRepo).Maybe()
	suite.MockRepo.EXPECT().OAuthServer().Return(suite.MockOAuthServerRepo).Maybe()
	suite.MockRepo.EXPECT().APIKey().Return(suite.MockAPIKeyRepo).Maybe()
	suite.MockRepo.EXPECT().System().Return(suite.MockSystemRepo).Maybe()
	suite.MockRepo.EXPECT().Workspace().Return(suite.MockWorkspaceRepo).Maybe()
	suite.MockRepo.EXPECT().Team().Return(suite.MockTeamRepo).Maybe()
	suite.MockRepo.EXPECT().Label().Return(suite.MockLabelRepo).Maybe()
	suite.MockRepo.EXPECT().Project().Return(suite.MockProjectRepo).Maybe()
	suite.MockRepo.EXPECT().Initiative().Return(suite.MockInitiativeRepo).Maybe()
	suite.MockRepo.EXPECT().Cycle().Return(suite.MockCycleRepo).Maybe()
	suite.MockRepo.EXPECT().Issue().Return(suite.MockIssueRepo).Maybe()
	suite.MockRepo.EXPECT().Comment().Return(suite.MockCommentRepo).Maybe()
	suite.MockRepo.EXPECT().Notification().Return(suite.MockNotificationRepo).Maybe()
	suite.MockRepo.EXPECT().Document().Return(suite.MockDocumentRepo).Maybe()
	suite.MockRepo.EXPECT().View().Return(suite.MockViewRepo).Maybe()
	suite.MockRepo.EXPECT().Template().Return(suite.MockTemplateRepo).Maybe()
	suite.MockRepo.EXPECT().Favorite().Return(suite.MockFavoriteRepo).Maybe()
	suite.MockRepo.EXPECT().Webhook().Return(suite.MockWebhookRepo).Maybe()
	suite.MockRepo.EXPECT().Upload().Return(suite.MockUploadRepo).Maybe()
	suite.MockRepo.EXPECT().ImportJob().Return(suite.MockImportJobRepo).Maybe()
	suite.MockRepo.EXPECT().Github().Return(suite.MockGithubRepo).Maybe()
	suite.MockRepo.EXPECT().Permissions().Return(suite.MockPermissionsRepo).Maybe()
}

// ExpectTx runs WithTx bodies directly; repository mocks match the tx argument with mock.Anything.
func (suite *ServiceTestSuite) ExpectTx() {
	suite.MockRepo.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(repository.TxInterface) error) error {
		return fn(suite.NewTxMock())
	}).Maybe()
}

func (suite *ServiceTestSuite) NewTxMock() repository.TxInterface {
	return mocks_repository_tx.NewTxMock(suite.T())
}
