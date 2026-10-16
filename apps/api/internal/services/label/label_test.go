package label_service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	label_repo "github.com/unbindapp/zigzag-api/internal/repositories/label"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
)

type LabelServiceSuite struct {
	services.ServiceTestSuite
	service     *LabelService
	userID      uuid.UUID
	workspaceID uuid.UUID
	team        *ent.Team
	group       *ent.Label
}

func (suite *LabelServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewLabelService(nil, suite.MockRepo)
	suite.service.now = func() time.Time { return time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC) }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.team = &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "ENG"}
	suite.group = &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Type", IsGroup: true}
	suite.ExpectTx()
}

func (suite *LabelServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *LabelServiceSuite) asAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
}

func (suite *LabelServiceSuite) TestListMasksForeignWorkspaces() {
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListLabelsInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *LabelServiceSuite) TestListNestsChildrenUnderGroups() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	child := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Bug", ParentID: &suite.group.ID}
	plain := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Docs", TeamID: &suite.team.ID}
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockLabelRepo.EXPECT().List(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.team.ID}, false).Return([]*ent.Label{suite.group, child, plain}, nil).Once()

	labels, err := suite.service.List(suite.Ctx, suite.userID, &ListLabelsInput{WorkspaceID: suite.workspaceID, TeamID: suite.team.ID})
	suite.Require().NoError(err)
	suite.Require().Len(labels, 2)
	suite.Equal(suite.group.ID, labels[0].ID)
	suite.Require().Len(labels[0].Children, 1)
	suite.Equal("Bug", labels[0].Children[0].Name)
	suite.Equal(&suite.team.ID, labels[1].TeamID)
}

func (suite *LabelServiceSuite) TestListHidesPrivateTeamsFromNonMembers() {
	private := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Private: true}
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, private.ID).Return(private, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, private.ID, suite.userID).Return(false, nil).Once()
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListLabelsInput{WorkspaceID: suite.workspaceID, TeamID: private.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	suite.MockTeamRepo.EXPECT().ListByWorkspace(suite.Ctx, suite.workspaceID, true).Return([]*ent.Team{suite.team, private}, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, private.ID, suite.userID).Return(false, nil).Once()
	suite.MockLabelRepo.EXPECT().List(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.team.ID}, false).Return([]*ent.Label{}, nil).Once()
	_, err = suite.service.List(suite.Ctx, suite.userID, &ListLabelsInput{WorkspaceID: suite.workspaceID, AllTeams: true})
	suite.NoError(err)
}

func (suite *LabelServiceSuite) TestCreateNeedsAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#eb5757"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *LabelServiceSuite) TestCreateValidatesGroups() {
	suite.asAdmin()
	cases := []*CreateLabelInput{
		{WorkspaceID: suite.workspaceID, Name: "Bug", Color: "red"},
		{WorkspaceID: suite.workspaceID, Name: " ", Color: "#eb5757"},
		{WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#eb5757", GroupSingleSelect: true},
		{WorkspaceID: suite.workspaceID, Name: "Sub", Color: "#eb5757", IsGroup: true, ParentID: &suite.group.ID},
	}
	for _, input := range cases {
		_, err := suite.service.Create(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, input.Name)
	}

	plain := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, plain.ID).Return(plain, nil).Once()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#eb5757", ParentID: &plain.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, TeamID: &suite.team.ID, Name: "Bug", Color: "#eb5757", ParentID: &suite.group.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *LabelServiceSuite) TestCreateRejectsTakenNamesAmongSiblings() {
	suite.asAdmin()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	suite.MockLabelRepo.EXPECT().NameTaken(suite.Ctx, mock.Anything, suite.workspaceID, (*uuid.UUID)(nil), &suite.group.ID, "Bug", (*uuid.UUID)(nil)).Return(true, nil).Once()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#eb5757", ParentID: &suite.group.ID})
	suite.ErrorIs(err, errdefs.ErrConflict)
}

func (suite *LabelServiceSuite) TestCreateAppendsToItsSiblings() {
	suite.asAdmin()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	suite.MockLabelRepo.EXPECT().NameTaken(suite.Ctx, mock.Anything, suite.workspaceID, (*uuid.UUID)(nil), &suite.group.ID, "Bug", (*uuid.UUID)(nil)).Return(false, nil).Once()
	suite.MockLabelRepo.EXPECT().NextSortOrder(suite.Ctx, mock.Anything, suite.workspaceID, (*uuid.UUID)(nil), &suite.group.ID).Return(4, nil).Once()
	suite.MockLabelRepo.EXPECT().Create(suite.Ctx, mock.Anything, &label_repo.CreateLabelInput{
		WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#eb5757", ParentID: &suite.group.ID, SortOrder: 4,
	}).Return(&ent.Label{ID: uuid.New(), Name: "Bug", SortOrder: 4, ParentID: &suite.group.ID}, nil).Once()

	label, err := suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, Name: " Bug ", Color: "#EB5757", ParentID: &suite.group.ID})
	suite.Require().NoError(err)
	suite.Equal(4.0, label.SortOrder)
}

func (suite *LabelServiceSuite) TestUpdateMovesLabelsOutOfGroupsWithTheNilID() {
	suite.asAdmin()
	child := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Bug", ParentID: &suite.group.ID}
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, child.ID).Return(child, nil)
	suite.MockLabelRepo.EXPECT().NameTaken(suite.Ctx, mock.Anything, suite.workspaceID, (*uuid.UUID)(nil), (*uuid.UUID)(nil), "Bug", &child.ID).Return(false, nil).Once()
	suite.MockLabelRepo.EXPECT().Update(suite.Ctx, mock.Anything, child.ID, &label_repo.UpdateLabelInput{ClearParent: true}).Return(&ent.Label{ID: child.ID, Name: "Bug"}, nil).Once()

	nilID := uuid.Nil
	label, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateLabelInput{LabelID: child.ID, ParentID: &nilID})
	suite.Require().NoError(err)
	suite.Nil(label.ParentID)

	singleSelect := true
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateLabelInput{LabelID: child.ID, GroupSingleSelect: &singleSelect})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *LabelServiceSuite) TestUpdateKeepsGroupsAtTheTop() {
	suite.asAdmin()
	other := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, IsGroup: true}
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateLabelInput{LabelID: suite.group.ID, ParentID: &other.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *LabelServiceSuite) TestUpdateNeedsAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	name := "Kind"
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateLabelInput{LabelID: suite.group.ID, Name: &name})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *LabelServiceSuite) TestReadsOfForeignLabelsAreNotFound() {
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	_, err := suite.service.Archive(suite.Ctx, suite.userID, &LabelRefInput{LabelID: suite.group.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *LabelServiceSuite) TestArchiveStampsTheTime() {
	suite.asAdmin()
	now := suite.service.now()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	suite.MockLabelRepo.EXPECT().SetArchived(suite.Ctx, suite.group.ID, &now).Return(&ent.Label{ID: suite.group.ID, ArchivedAt: &now}, nil).Once()
	label, err := suite.service.Archive(suite.Ctx, suite.userID, &LabelRefInput{LabelID: suite.group.ID})
	suite.Require().NoError(err)
	suite.Equal(&now, label.ArchivedAt)
}

func (suite *LabelServiceSuite) TestDeleteRecordsRemovalOnIssues() {
	suite.asAdmin()
	child := &ent.Label{ID: uuid.New(), Name: "Bug", ParentID: &suite.group.ID}
	issueID := uuid.New()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, suite.group.ID).Return(suite.group, nil).Once()
	suite.MockLabelRepo.EXPECT().Children(suite.Ctx, suite.group.ID).Return([]*ent.Label{child}, nil).Once()
	ids := []uuid.UUID{suite.group.ID, child.ID}
	suite.MockLabelRepo.EXPECT().IssuesWithLabels(suite.Ctx, mock.Anything, ids).Return(map[uuid.UUID][]*ent.Label{issueID: {child}}, nil).Once()
	suite.MockLabelRepo.EXPECT().RecordHistory(suite.Ctx, mock.Anything, []label_repo.HistoryEntry{{
		IssueID: issueID,
		ActorID: &suite.userID,
		Source:  schema.HistorySourceUser,
		Changes: []schema.FieldChange{{Field: "labels_removed", From: []string{child.ID.String()}, FromLabel: "Bug"}},
	}}).Return(nil).Once()
	suite.MockLabelRepo.EXPECT().Delete(suite.Ctx, mock.Anything, ids).Return(nil).Once()

	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &LabelRefInput{LabelID: suite.group.ID}))
}

func (suite *LabelServiceSuite) TestIDsFromAnotherWorkspaceAreRefused() {
	suite.asAdmin()
	foreignTeam := &ent.Team{ID: uuid.New(), WorkspaceID: uuid.New(), Key: "OPS"}
	foreignGroup := &ent.Label{ID: uuid.New(), WorkspaceID: uuid.New(), Name: "Kind", IsGroup: true}
	plain := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Bug"}
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, foreignTeam.ID).Return(foreignTeam, nil).Maybe()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, foreignGroup.ID).Return(foreignGroup, nil).Maybe()
	suite.MockLabelRepo.EXPECT().GetByID(suite.Ctx, plain.ID).Return(plain, nil).Maybe()

	cases := map[string]func() error{
		"create with a team of another workspace": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, TeamID: &foreignTeam.ID, Name: "Bug", Color: "#eb5757"})
			return err
		},
		"create in a group of another workspace": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateLabelInput{WorkspaceID: suite.workspaceID, Name: "Bug", Color: "#eb5757", ParentID: &foreignGroup.ID})
			return err
		},
		"move into a group of another workspace": func() error {
			_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateLabelInput{LabelID: plain.ID, ParentID: &foreignGroup.ID})
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

func TestLabelServiceSuite(t *testing.T) {
	suite.Run(t, new(LabelServiceSuite))
}
