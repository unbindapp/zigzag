package email_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/crypto"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/mail"
	system_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/system"
	mocks_repositories "github.com/zigzagapp-org/zigzag-api/mocks/repositories"
	mocks_repository_permissions "github.com/zigzagapp-org/zigzag-api/mocks/repository/permissions"
	mocks_repository_system "github.com/zigzagapp-org/zigzag-api/mocks/repository/system"
	mocks_repository_user "github.com/zigzagapp-org/zigzag-api/mocks/repository/user"
)

type recordingSender struct {
	sent     []mail.Message
	settings []mail.Settings
	err      error
}

func (self *recordingSender) Send(_ context.Context, settings mail.Settings, msg mail.Message) error {
	self.settings = append(self.settings, settings)
	self.sent = append(self.sent, msg)
	return self.err
}

// the shared ServiceTestSuite imports the email mock, which would import this package
type EmailServiceSuite struct {
	suite.Suite
	Ctx                 context.Context
	MockRepo            *mocks_repositories.RepositoriesMock
	MockPermissionsRepo *mocks_repository_permissions.PermissionsRepositoryMock
	MockSystemRepo      *mocks_repository_system.SystemRepositoryMock
	MockUserRepo        *mocks_repository_user.UserRepositoryMock
	service             *EmailService
	sender              *recordingSender
	encryptor           *crypto.Encryptor
	adminID             uuid.UUID
}

func (suite *EmailServiceSuite) SetupTest() {
	suite.Ctx = context.Background()
	suite.MockRepo = mocks_repositories.NewRepositoriesMock(suite.T())
	suite.MockPermissionsRepo = mocks_repository_permissions.NewPermissionsRepositoryMock(suite.T())
	suite.MockSystemRepo = mocks_repository_system.NewSystemRepositoryMock(suite.T())
	suite.MockUserRepo = mocks_repository_user.NewUserRepositoryMock(suite.T())
	suite.MockRepo.EXPECT().Permissions().Return(suite.MockPermissionsRepo).Maybe()
	suite.MockRepo.EXPECT().System().Return(suite.MockSystemRepo).Maybe()
	suite.MockRepo.EXPECT().User().Return(suite.MockUserRepo).Maybe()
	key := make([]byte, 32)
	suite.encryptor, _ = crypto.NewEncryptor(key)
	suite.sender = &recordingSender{}
	suite.service = NewEmailService(&config.Config{ExternalUIUrl: "https://zigzag.example.com"}, suite.MockRepo, suite.encryptor, suite.sender)
	suite.service.async = func(fn func()) { fn() }
	suite.adminID = uuid.New()
}

func relay() schema.EmailSettings {
	return schema.EmailSettings{Host: "smtp.resend.com", Port: 465, Username: "resend", FromAddress: "zigzag@example.com", Encryption: "tls"}
}

func (suite *EmailServiceSuite) TestUpdateEncryptsThePassword() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.adminID).Return(nil).Once()
	var stored *system_repo.SystemSettingUpdateInput
	suite.MockSystemRepo.EXPECT().UpdateSystemSettings(suite.Ctx, mock.Anything).RunAndReturn(func(_ context.Context, input *system_repo.SystemSettingUpdateInput) (*ent.SystemSetting, error) {
		stored = input
		return &ent.SystemSetting{EmailSettings: input.EmailSettings, EmailPasswordEncrypted: input.EmailPasswordEncrypted}, nil
	}).Once()

	password := "re_secret"
	resp, err := suite.service.UpdateSettings(suite.Ctx, suite.adminID, &EmailSettingsUpdateInput{EmailSettings: relay(), Password: &password})
	suite.Require().NoError(err)
	suite.True(resp.Configured)
	suite.True(resp.PasswordSet)
	suite.NotContains(string(stored.EmailPasswordEncrypted), password)
	decrypted, err := suite.encryptor.Decrypt(stored.EmailPasswordEncrypted)
	suite.NoError(err)
	suite.Equal(password, string(decrypted))
}

func (suite *EmailServiceSuite) TestUpdateRejectsUnknownEncryptionAndNonAdmins() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.adminID).Return(nil).Once()
	settings := relay()
	settings.Encryption = "ssl"
	_, err := suite.service.UpdateSettings(suite.Ctx, suite.adminID, &EmailSettingsUpdateInput{EmailSettings: settings})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.adminID).Return(errdefs.ErrUnauthorized).Once()
	_, err = suite.service.UpdateSettings(suite.Ctx, suite.adminID, &EmailSettingsUpdateInput{EmailSettings: relay()})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *EmailServiceSuite) TestEmptyPasswordClearsIt() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.adminID).Return(nil).Once()
	suite.MockSystemRepo.EXPECT().UpdateSystemSettings(suite.Ctx, mock.MatchedBy(func(input *system_repo.SystemSettingUpdateInput) bool {
		return input.ClearEmailPassword && input.EmailPasswordEncrypted == nil
	})).Return(&ent.SystemSetting{}, nil).Once()
	empty := ""
	_, err := suite.service.UpdateSettings(suite.Ctx, suite.adminID, &EmailSettingsUpdateInput{EmailSettings: relay(), Password: &empty})
	suite.NoError(err)
}

func (suite *EmailServiceSuite) TestSendDecryptsAndRenders() {
	encrypted, _ := suite.encryptor.Encrypt([]byte("re_secret"))
	settings := relay()
	suite.MockSystemRepo.EXPECT().GetSystemSettings(suite.Ctx, nil).Return(&ent.SystemSetting{InstanceName: "Acme Tracker", EmailSettings: &settings, EmailPasswordEncrypted: encrypted}, nil).Once()

	sent, err := suite.service.Send(suite.Ctx, "ada@example.com", mail.Content{Subject: "Hi", Heading: "Hello"})
	suite.Require().NoError(err)
	suite.True(sent)
	suite.Require().Len(suite.sender.sent, 1)
	suite.Equal("re_secret", suite.sender.settings[0].Password)
	suite.Equal(mail.EncryptionTLS, suite.sender.settings[0].Encryption)
	suite.Contains(suite.sender.sent[0].Text, "Sent by Acme Tracker at https://zigzag.example.com")
}

func (suite *EmailServiceSuite) TestSendWithoutRelaySendsNothing() {
	suite.MockSystemRepo.EXPECT().GetSystemSettings(suite.Ctx, nil).Return(&ent.SystemSetting{}, nil).Once()
	sent, err := suite.service.Send(suite.Ctx, "ada@example.com", mail.Content{})
	suite.NoError(err)
	suite.False(sent)
	suite.Empty(suite.sender.sent)
}

func (suite *EmailServiceSuite) TestSendAsyncReportsWhetherMailIsQueued() {
	settings := relay()
	suite.MockSystemRepo.EXPECT().GetSystemSettings(mock.Anything, nil).Return(&ent.SystemSetting{}, nil).Once()
	suite.False(suite.service.SendAsync(suite.Ctx, "ada@example.com", mail.Content{}))

	suite.MockSystemRepo.EXPECT().GetSystemSettings(mock.Anything, nil).Return(&ent.SystemSetting{EmailSettings: &settings}, nil).Twice()
	suite.sender.err = errors.New("relay down")
	suite.True(suite.service.SendAsync(suite.Ctx, "ada@example.com", mail.Content{Subject: "x"}), "failures after queueing are logged, not returned")
	suite.Len(suite.sender.sent, 1)
}

func (suite *EmailServiceSuite) TestSendTestReportsFailures() {
	settings := relay()
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.adminID).Return(nil)
	suite.MockUserRepo.EXPECT().GetByID(suite.Ctx, suite.adminID).Return(&ent.User{Email: "ada@example.com"}, nil)

	suite.MockSystemRepo.EXPECT().GetSystemSettings(suite.Ctx, nil).Return(&ent.SystemSetting{}, nil).Once()
	suite.ErrorIs(suite.service.SendTest(suite.Ctx, suite.adminID), errdefs.ErrInvalidInput)

	suite.MockSystemRepo.EXPECT().GetSystemSettings(suite.Ctx, nil).Return(&ent.SystemSetting{EmailSettings: &settings}, nil).Once()
	suite.sender.err = errors.New("auth failed")
	err := suite.service.SendTest(suite.Ctx, suite.adminID)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	suite.ErrorContains(err, "auth failed")
}

func TestEmailServiceSuite(t *testing.T) {
	suite.Run(t, new(EmailServiceSuite))
}
