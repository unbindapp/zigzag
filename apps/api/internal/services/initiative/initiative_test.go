package initiative_service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	initiative_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/initiative"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type InitiativeServiceSuite struct {
	services.ServiceTestSuite
	service     *InitiativeService
	now         time.Time
	userID      uuid.UUID
	workspaceID uuid.UUID
	initiative  *ent.Initiative
}

func (suite *InitiativeServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewInitiativeService(nil, suite.MockRepo)
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.initiative = &ent.Initiative{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Growth", Slug: "growth-1a2b3c4d", Status: schema.InitiativeStatusPlanned}
	suite.ExpectTx()
}

func TestInitiativeServiceSuite(t *testing.T) {
	suite.Run(t, new(InitiativeServiceSuite))
}

func (suite *InitiativeServiceSuite) roles(granted schema.WorkspaceRole) {
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

func (suite *InitiativeServiceSuite) outsider() {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, mock.Anything).Return(errdefs.ErrUnauthorized).Maybe()
}

func (suite *InitiativeServiceSuite) expectRespond(initiative *ent.Initiative) {
	suite.MockInitiativeRepo.EXPECT().GetByID(suite.Ctx, initiative.ID).Return(initiative, nil).Maybe()
	suite.MockInitiativeRepo.EXPECT().LatestHealths(suite.Ctx, mock.Anything).Return(map[uuid.UUID]schema.Health{}, nil).Maybe()
	suite.MockProjectRepo.EXPECT().IssueTotals(suite.Ctx, mock.Anything).Return(map[uuid.UUID]*models.IssueTotals{}, nil).Maybe()
}

func (suite *InitiativeServiceSuite) TestListHidesOtherWorkspaces() {
	suite.outsider()
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListInitiativesInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *InitiativeServiceSuite) TestListSumsProgressAndHealthOverProjects() {
	suite.roles(schema.WorkspaceRoleViewer)
	first, second := uuid.New(), uuid.New()
	initiative := *suite.initiative
	initiative.Edges.Projects = []*ent.Project{{ID: first}, {ID: second}}
	suite.MockInitiativeRepo.EXPECT().List(suite.Ctx, mock.MatchedBy(func(filter *initiative_repo.ListInitiativesFilter) bool {
		return *filter.Status == schema.InitiativeStatusActive && filter.OwnerID == nil
	})).Return([]*ent.Initiative{&initiative}, nil).Once()
	suite.MockInitiativeRepo.EXPECT().LatestHealths(suite.Ctx, []uuid.UUID{initiative.ID}).Return(map[uuid.UUID]schema.Health{initiative.ID: schema.HealthAtRisk}, nil).Once()
	suite.MockProjectRepo.EXPECT().IssueTotals(suite.Ctx, []uuid.UUID{first, second}).Return(map[uuid.UUID]*models.IssueTotals{
		first:  {Counts: map[schema.StateType]int{schema.StateTypeCompleted: 1}, Scopes: map[schema.StateType]float64{schema.StateTypeCompleted: 3}},
		second: {Counts: map[schema.StateType]int{schema.StateTypeUnstarted: 1}, Scopes: map[schema.StateType]float64{schema.StateTypeUnstarted: 1}},
	}, nil).Once()

	initiatives, err := suite.service.List(suite.Ctx, suite.userID, &ListInitiativesInput{WorkspaceID: suite.workspaceID, Status: schema.InitiativeStatusActive})
	suite.Require().NoError(err)
	suite.Require().Len(initiatives, 1)
	suite.Equal(75, initiatives[0].Progress.Percent)
	suite.Equal(schema.HealthAtRisk, *initiatives[0].Health)
	suite.Len(initiatives[0].Projects, 2)
}

func (suite *InitiativeServiceSuite) TestListRejectsUnknownStatus() {
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListInitiativesInput{WorkspaceID: suite.workspaceID, Status: "someday"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *InitiativeServiceSuite) TestCreateNeedsMember() {
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateInitiativeInput{WorkspaceID: suite.workspaceID, Name: "Growth"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *InitiativeServiceSuite) TestCreateValidatesReferences() {
	suite.roles(schema.WorkspaceRoleMember)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateInitiativeInput{WorkspaceID: suite.workspaceID, Name: "Growth", Status: "someday"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	owner := uuid.New()
	suite.MockProjectRepo.EXPECT().CountWorkspaceMembers(suite.Ctx, suite.workspaceID, []uuid.UUID{owner}).Return(0, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateInitiativeInput{WorkspaceID: suite.workspaceID, Name: "Growth", OwnerID: owner})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	foreign := uuid.New()
	suite.MockInitiativeRepo.EXPECT().CountProjectsInWorkspace(suite.Ctx, suite.workspaceID, []uuid.UUID{foreign}).Return(0, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateInitiativeInput{WorkspaceID: suite.workspaceID, Name: "Growth", ProjectIDs: []uuid.UUID{foreign}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *InitiativeServiceSuite) TestCreateSlugsAndAppends() {
	suite.roles(schema.WorkspaceRoleMember)
	project := uuid.New()
	suite.MockInitiativeRepo.EXPECT().CountProjectsInWorkspace(suite.Ctx, suite.workspaceID, []uuid.UUID{project}).Return(1, nil).Once()
	suite.MockInitiativeRepo.EXPECT().MaxSortOrder(suite.Ctx, suite.workspaceID).Return(2, nil).Once()
	suite.MockInitiativeRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *initiative_repo.CreateInitiativeInput) bool {
		return input.Name == "Q4 Growth" && input.Slug[:10] == "q4-growth-" && len(input.Slug) == 18 && input.SortOrder == 3 &&
			len(input.ProjectIDs) == 1 && input.TargetDate.Format(models.DateFormat) == "2026-12-31"
	})).Return(suite.initiative, nil).Once()
	suite.expectRespond(suite.initiative)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateInitiativeInput{
		WorkspaceID: suite.workspaceID, Name: " Q4 Growth ", TargetDate: "2026-12-31", ProjectIDs: []uuid.UUID{project, project},
	})
	suite.NoError(err)
}

func (suite *InitiativeServiceSuite) TestUpdateClearsOwnerWithTheNilUUID() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond(suite.initiative)
	suite.MockInitiativeRepo.EXPECT().Update(suite.Ctx, mock.Anything, suite.initiative.ID, mock.MatchedBy(func(input *initiative_repo.UpdateInitiativeInput) bool {
		return input.ClearOwner && input.OwnerID == nil && !input.ReplaceProjects
	})).Return(nil).Once()
	nilOwner := uuid.Nil
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateInitiativeInput{InitiativeID: suite.initiative.ID, OwnerID: &nilOwner})
	suite.NoError(err)
}

func (suite *InitiativeServiceSuite) TestDeleteNeedsAdmin() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond(suite.initiative)
	err := suite.service.Delete(suite.Ctx, suite.userID, &InitiativeRefInput{InitiativeID: suite.initiative.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *InitiativeServiceSuite) TestGetHidesOtherWorkspaces() {
	suite.outsider()
	suite.expectRespond(suite.initiative)
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetInitiativeInput{InitiativeID: suite.initiative.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *InitiativeServiceSuite) TestArchiveStampsTheTime() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond(suite.initiative)
	suite.MockInitiativeRepo.EXPECT().SetArchivedAt(suite.Ctx, suite.initiative.ID, &suite.now).Return(nil).Once()
	_, err := suite.service.Archive(suite.Ctx, suite.userID, &InitiativeRefInput{InitiativeID: suite.initiative.ID})
	suite.NoError(err)
}

func (suite *InitiativeServiceSuite) TestAddProjectChecksTheWorkspace() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond(suite.initiative)
	project := uuid.New()
	suite.MockInitiativeRepo.EXPECT().CountProjectsInWorkspace(suite.Ctx, suite.workspaceID, []uuid.UUID{project}).Return(0, nil).Once()
	_, err := suite.service.AddProject(suite.Ctx, suite.userID, &InitiativeProjectInput{InitiativeID: suite.initiative.ID, ProjectID: project})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *InitiativeServiceSuite) TestCreateUpdateValidates() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond(suite.initiative)
	_, err := suite.service.CreateUpdate(suite.Ctx, suite.userID, &CreateInitiativeUpdateInput{InitiativeID: suite.initiative.ID, Body: " ", Health: schema.HealthOnTrack})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.CreateUpdate(suite.Ctx, suite.userID, &CreateInitiativeUpdateInput{InitiativeID: suite.initiative.ID, Body: "Fine", Health: "fine"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *InitiativeServiceSuite) TestOnlyAuthorsOrAdminsDeleteUpdates() {
	suite.roles(schema.WorkspaceRoleMember)
	author := uuid.New()
	update := &ent.InitiativePost{ID: uuid.New(), InitiativeID: suite.initiative.ID, AuthorID: &author}
	suite.MockInitiativeRepo.EXPECT().GetUpdate(suite.Ctx, update.ID).Return(update, nil).Once()
	suite.MockInitiativeRepo.EXPECT().ResolveWorkspaceOfInitiative(suite.Ctx, suite.initiative.ID).Return(suite.workspaceID, nil).Once()
	err := suite.service.DeleteUpdate(suite.Ctx, suite.userID, &InitiativeUpdateRefInput{UpdateID: update.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}
