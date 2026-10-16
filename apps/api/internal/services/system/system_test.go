package system_service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	system_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/system"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type SystemServiceSuite struct {
	services.ServiceTestSuite
	service *SystemService
	userID  uuid.UUID
}

func (suite *SystemServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewSystemService(suite.MockRepo)
	suite.userID = uuid.New()
}

func (suite *SystemServiceSuite) TestOnlySystemAdmins() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(errdefs.ErrUnauthorized).Twice()
	_, err := suite.service.GetSettings(suite.Ctx, suite.userID)
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	_, err = suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateSettingsInput{})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *SystemServiceSuite) TestUpdateTrimsAndRefusesEmptyNames() {
	suite.MockPermissionsRepo.EXPECT().CheckSystemAdmin(suite.Ctx, suite.userID).Return(nil)
	_, err := suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateSettingsInput{InstanceName: utils.ToPtr("  ")})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockSystemRepo.EXPECT().UpdateSystemSettings(suite.Ctx, &system_repo.SystemSettingUpdateInput{InstanceName: utils.ToPtr("Acme"), AllowPublicSignup: utils.ToPtr(true)}).
		Return(&ent.SystemSetting{InstanceName: "Acme", AllowPublicSignup: true}, nil).Once()
	resp, err := suite.service.UpdateSettings(suite.Ctx, suite.userID, &UpdateSettingsInput{InstanceName: utils.ToPtr(" Acme "), AllowPublicSignup: utils.ToPtr(true)})
	suite.Require().NoError(err)
	suite.Equal("Acme", resp.InstanceName)
	suite.False(resp.Email.Configured)
}

func TestSystemServiceSuite(t *testing.T) {
	suite.Run(t, new(SystemServiceSuite))
}
