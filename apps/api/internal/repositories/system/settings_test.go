package system_repo

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type SettingsSuite struct {
	repository.RepositoryBaseSuite
	repo *SystemRepository
}

func (suite *SettingsSuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewSystemRepository(suite.DB)
}

func (suite *SettingsSuite) TestGetCreatesTheSingletonOnce() {
	first, err := suite.repo.GetSystemSettings(suite.Ctx, nil)
	suite.Require().NoError(err)
	suite.Equal("Zigzag", first.InstanceName)
	suite.False(first.AllowPublicSignup)

	second, err := suite.repo.GetSystemSettings(suite.Ctx, nil)
	suite.Require().NoError(err)
	suite.Equal(first.ID, second.ID)
	suite.Equal(1, suite.DB.SystemSetting.Query().CountX(suite.Ctx))
}

func (suite *SettingsSuite) TestUpdateEmailAndClear() {
	email := &schema.EmailSettings{Host: "smtp.resend.com", Port: 465, Username: "resend", FromAddress: "a@b.c", Encryption: "tls"}
	updated, err := suite.repo.UpdateSystemSettings(suite.Ctx, &SystemSettingUpdateInput{
		InstanceName:           utils.ToPtr("Acme"),
		EmailSettings:          email,
		EmailPasswordEncrypted: []byte{1, 2, 3},
	})
	suite.Require().NoError(err)
	suite.Equal("Acme", updated.InstanceName)
	suite.Equal("smtp.resend.com", updated.EmailSettings.Host)
	suite.Equal([]byte{1, 2, 3}, updated.EmailPasswordEncrypted)

	kept, err := suite.repo.UpdateSystemSettings(suite.Ctx, &SystemSettingUpdateInput{AllowPublicSignup: utils.ToPtr(true)})
	suite.Require().NoError(err)
	suite.True(kept.AllowPublicSignup)
	suite.Equal([]byte{1, 2, 3}, kept.EmailPasswordEncrypted, "omitted fields are kept")

	cleared, err := suite.repo.UpdateSystemSettings(suite.Ctx, &SystemSettingUpdateInput{ClearEmailSettings: true})
	suite.Require().NoError(err)
	suite.Nil(cleared.EmailSettings)
	suite.Empty(cleared.EmailPasswordEncrypted)
	suite.Equal("Acme", cleared.InstanceName)
}

func TestSettingsSuite(t *testing.T) {
	suite.Run(t, new(SettingsSuite))
}
