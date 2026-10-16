package template_service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	template_repo "github.com/unbindapp/zigzag-api/internal/repositories/template"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type TemplateServiceSuite struct {
	services.ServiceTestSuite
	service     *TemplateService
	userID      uuid.UUID
	workspaceID uuid.UUID
	team        *ent.Team
}

func (suite *TemplateServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewTemplateService(nil, suite.MockRepo)
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.team = &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.ExpectTx()
}

func (suite *TemplateServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *TemplateServiceSuite) asAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Maybe()
	suite.MockLabelRepo.EXPECT().GetMany(suite.Ctx, ([]uuid.UUID)(nil)).Return([]*ent.Label{}, nil).Maybe()
}

func (suite *TemplateServiceSuite) TestCreateNeedsAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *TemplateServiceSuite) TestTemplatesOfOtherWorkspacesAreNotFound() {
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	current := &ent.Template{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockTemplateRepo.EXPECT().GetByID(suite.Ctx, current.ID).Return(current, nil).Once()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetTemplateInput{TemplateID: current.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *TemplateServiceSuite) TestCreateChecksReferences() {
	suite.asAdmin()
	priority := 7
	stateID := uuid.New()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{Priority: &priority}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{StateID: &stateID}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockTeamRepo.EXPECT().GetState(suite.Ctx, stateID).Return(&ent.WorkflowState{ID: stateID, TeamID: uuid.New()}, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, TeamID: &suite.team.ID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{StateID: &stateID}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	otherTeam := uuid.New()
	teamLabel := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: &otherTeam}
	suite.MockLabelRepo.EXPECT().GetMany(suite.Ctx, []uuid.UUID{teamLabel.ID}).Return([]*ent.Label{teamLabel}, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{LabelIDs: []uuid.UUID{teamLabel.ID}}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	missing := uuid.New()
	suite.MockLabelRepo.EXPECT().GetMany(suite.Ctx, []uuid.UUID{missing}).Return([]*ent.Label{}, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{LabelIDs: []uuid.UUID{missing}}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	stranger := uuid.New()
	suite.MockWorkspaceRepo.EXPECT().GetMember(suite.Ctx, mock.Anything, suite.workspaceID, stranger).Return(nil, &ent.NotFoundError{}).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{AssigneeID: &stranger}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	projectID := uuid.New()
	suite.MockTemplateRepo.EXPECT().ProjectInWorkspace(suite.Ctx, suite.workspaceID, projectID).Return(false, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{ProjectID: &projectID}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: "sprint", Name: "Bug"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *TemplateServiceSuite) TestCreateStoresTeamTemplates() {
	suite.asAdmin()
	stateID := uuid.New()
	suite.MockTeamRepo.EXPECT().GetState(suite.Ctx, stateID).Return(&ent.WorkflowState{ID: stateID, TeamID: suite.team.ID}, nil).Once()
	suite.MockTemplateRepo.EXPECT().NextSortOrder(suite.Ctx, mock.Anything, suite.workspaceID).Return(3, nil).Once()
	data := schema.TemplateData{Title: "Bug: ", StateID: &stateID, SubIssues: []schema.TemplateSubIssue{{Title: "Reproduce"}}}
	suite.MockTemplateRepo.EXPECT().Create(suite.Ctx, mock.Anything, &template_repo.CreateTemplateInput{
		WorkspaceID: suite.workspaceID, TeamID: &suite.team.ID, Kind: schema.TemplateKindIssue, Name: "Bug report", Data: data, SortOrder: 3, CreatorID: &suite.userID,
	}).Return(&ent.Template{ID: uuid.New(), Name: "Bug report", Data: data, SortOrder: 3}, nil).Once()

	template, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, TeamID: &suite.team.ID, Kind: schema.TemplateKindIssue, Name: "Bug report", Data: data})
	suite.Require().NoError(err)
	suite.Equal("Bug: ", template.Data.Title)
}

func (suite *TemplateServiceSuite) TestUpdateRevalidatesDataAgainstTheTemplatesTeam() {
	suite.asAdmin()
	current := &ent.Template{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockTemplateRepo.EXPECT().GetByID(suite.Ctx, current.ID).Return(current, nil)
	stateID := uuid.New()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateTemplateInput{TemplateID: current.ID, Data: &schema.TemplateData{StateID: &stateID}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockTemplateRepo.EXPECT().Delete(suite.Ctx, current.ID).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &DeleteTemplateInput{TemplateID: current.ID}))
}

func (suite *TemplateServiceSuite) TestListAddsTheTeamsTemplates() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	kind := schema.TemplateKindIssue
	suite.MockTemplateRepo.EXPECT().List(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.team.ID}, &kind).Return([]*ent.Template{{ID: uuid.New()}}, nil).Once()
	templates, err := suite.service.List(suite.Ctx, suite.userID, &ListTemplatesInput{WorkspaceID: suite.workspaceID, TeamID: suite.team.ID, Kind: kind})
	suite.Require().NoError(err)
	suite.Len(templates, 1)
}

func (suite *TemplateServiceSuite) TestIDsFromAnotherWorkspaceAreRefused() {
	suite.asAdmin()
	foreignTeam := &ent.Team{ID: uuid.New(), WorkspaceID: uuid.New()}
	otherTeamsState := &ent.WorkflowState{ID: uuid.New(), TeamID: uuid.New()}
	foreignLabel := &ent.Label{ID: uuid.New(), WorkspaceID: uuid.New()}
	teamTemplate := &ent.Template{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: &suite.team.ID}
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, foreignTeam.ID).Return(foreignTeam, nil).Maybe()
	suite.MockTeamRepo.EXPECT().GetState(suite.Ctx, otherTeamsState.ID).Return(otherTeamsState, nil).Maybe()
	suite.MockLabelRepo.EXPECT().GetMany(suite.Ctx, []uuid.UUID{foreignLabel.ID}).Return([]*ent.Label{foreignLabel}, nil).Maybe()
	suite.MockTemplateRepo.EXPECT().GetByID(suite.Ctx, teamTemplate.ID).Return(teamTemplate, nil).Maybe()

	cases := map[string]func() error{
		"create with a team of another workspace": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, TeamID: &foreignTeam.ID, Kind: schema.TemplateKindIssue, Name: "Bug"})
			return err
		},
		"create with a state of another team": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, TeamID: &suite.team.ID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{StateID: &otherTeamsState.ID}})
			return err
		},
		"create with a label of another workspace": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateTemplateInput{WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Name: "Bug", Data: schema.TemplateData{LabelIDs: []uuid.UUID{foreignLabel.ID}}})
			return err
		},
		"update with a state of another team": func() error {
			_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateTemplateInput{TemplateID: teamTemplate.ID, Data: &schema.TemplateData{StateID: &otherTeamsState.ID}})
			return err
		},
	}
	for name, run := range cases {
		suite.Run(name, func() {
			err := run()
			suite.Truef(errors.Is(err, errdefs.ErrInvalidInput) || errors.Is(err, errdefs.ErrNotFound), "got %v", err)
		})
	}
}

func TestTemplateServiceSuite(t *testing.T) {
	suite.Run(t, new(TemplateServiceSuite))
}
