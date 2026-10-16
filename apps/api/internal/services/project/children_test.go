package project_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
)

func (suite *ProjectServiceSuite) TestCreateStatusNeedsAdmin() {
	suite.roles(schema.WorkspaceRoleMember)
	_, err := suite.service.CreateStatus(suite.Ctx, suite.userID, &CreateProjectStatusInput{WorkspaceID: suite.workspaceID, Name: "Review", Type: schema.ProjectStatusStarted, Color: "#ffffff"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ProjectServiceSuite) TestCreateStatusValidates() {
	suite.roles(schema.WorkspaceRoleAdmin)
	_, err := suite.service.CreateStatus(suite.Ctx, suite.userID, &CreateProjectStatusInput{WorkspaceID: suite.workspaceID, Name: "Review", Type: "review", Color: "#ffffff"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockProjectRepo.EXPECT().StatusNameTaken(suite.Ctx, suite.workspaceID, "Backlog", (*uuid.UUID)(nil)).Return(true, nil).Once()
	_, err = suite.service.CreateStatus(suite.Ctx, suite.userID, &CreateProjectStatusInput{WorkspaceID: suite.workspaceID, Name: " Backlog ", Type: schema.ProjectStatusBacklog, Color: "#ffffff"})
	suite.ErrorIs(err, errdefs.ErrConflict)

	suite.MockProjectRepo.EXPECT().StatusNameTaken(suite.Ctx, suite.workspaceID, "Review", (*uuid.UUID)(nil)).Return(false, nil).Once()
	_, err = suite.service.CreateStatus(suite.Ctx, suite.userID, &CreateProjectStatusInput{WorkspaceID: suite.workspaceID, Name: "Review", Type: schema.ProjectStatusStarted, Color: "white"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestCreateStatusAppendsWithinItsType() {
	suite.roles(schema.WorkspaceRoleAdmin)
	suite.MockProjectRepo.EXPECT().StatusNameTaken(suite.Ctx, suite.workspaceID, "Review", (*uuid.UUID)(nil)).Return(false, nil).Once()
	suite.MockProjectRepo.EXPECT().NextStatusPosition(suite.Ctx, suite.workspaceID, schema.ProjectStatusStarted).Return(3, nil).Once()
	suite.MockProjectRepo.EXPECT().CreateStatus(suite.Ctx, &project_repo.CreateStatusInput{
		WorkspaceID: suite.workspaceID, Name: "Review", Type: schema.ProjectStatusStarted, Color: "#abcdef", Position: 3,
	}).Return(&ent.ProjectStatus{ID: uuid.New(), Name: "Review"}, nil).Once()
	status, err := suite.service.CreateStatus(suite.Ctx, suite.userID, &CreateProjectStatusInput{WorkspaceID: suite.workspaceID, Name: "Review", Type: schema.ProjectStatusStarted, Color: "#ABCDEF"})
	suite.Require().NoError(err)
	suite.Equal("Review", status.Name)
}

func (suite *ProjectServiceSuite) TestDeleteStatusMovesProjectsWithinTheWorkspace() {
	suite.roles(schema.WorkspaceRoleAdmin)
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, suite.backlog.ID).Return(suite.backlog, nil)

	err := suite.service.DeleteStatus(suite.Ctx, suite.userID, &DeleteProjectStatusInput{StatusID: suite.backlog.ID, MoveToStatusID: suite.backlog.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	foreign := &ent.ProjectStatus{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, foreign.ID).Return(foreign, nil).Once()
	err = suite.service.DeleteStatus(suite.Ctx, suite.userID, &DeleteProjectStatusInput{StatusID: suite.backlog.ID, MoveToStatusID: foreign.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	planned := &ent.ProjectStatus{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, planned.ID).Return(planned, nil).Once()
	suite.MockProjectRepo.EXPECT().DeleteStatus(suite.Ctx, mock.Anything, suite.backlog.ID, planned.ID).Return(nil).Once()
	suite.NoError(suite.service.DeleteStatus(suite.Ctx, suite.userID, &DeleteProjectStatusInput{StatusID: suite.backlog.ID, MoveToStatusID: planned.ID}))
}

func (suite *ProjectServiceSuite) TestUpdateStatusHidesOtherWorkspaces() {
	suite.outsider()
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, suite.backlog.ID).Return(suite.backlog, nil).Once()
	_, err := suite.service.UpdateStatus(suite.Ctx, suite.userID, &UpdateProjectStatusInput{StatusID: suite.backlog.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *ProjectServiceSuite) expectWorkspaceOfProject() {
	suite.MockProjectRepo.EXPECT().ResolveWorkspaceOfProject(suite.Ctx, suite.project.ID).Return(suite.workspaceID, nil).Maybe()
}

func (suite *ProjectServiceSuite) TestMilestonesNeedMemberToChange() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectProject(suite.project)
	suite.expectWorkspaceOfProject()
	_, err := suite.service.CreateMilestone(suite.Ctx, suite.userID, &CreateMilestoneInput{ProjectID: suite.project.ID, Name: "Beta"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)

	milestone := &ent.ProjectMilestone{ID: uuid.New(), ProjectID: suite.project.ID}
	suite.MockProjectRepo.EXPECT().GetMilestone(suite.Ctx, milestone.ID).Return(milestone, nil).Once()
	err = suite.service.DeleteMilestone(suite.Ctx, suite.userID, &MilestoneRefInput{MilestoneID: milestone.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ProjectServiceSuite) TestDeleteMilestoneRecordsHistoryOnItsIssues() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectWorkspaceOfProject()
	milestone := &ent.ProjectMilestone{ID: uuid.New(), ProjectID: suite.project.ID, Name: "Beta"}
	issueID := uuid.New()
	suite.MockProjectRepo.EXPECT().GetMilestone(suite.Ctx, milestone.ID).Return(milestone, nil).Once()
	suite.MockProjectRepo.EXPECT().IssueIDsInMilestone(suite.Ctx, mock.Anything, milestone.ID).Return([]uuid.UUID{issueID}, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistoryMany(suite.Ctx, mock.Anything, []issue_repo.HistoryEntry{{
		IssueID: issueID, ActorID: &suite.userID, Source: schema.HistorySourceUser,
		Changes: []schema.FieldChange{{Field: "milestone", From: milestone.ID.String(), FromLabel: "Beta"}},
	}}).Return(nil).Once()
	suite.MockProjectRepo.EXPECT().DeleteMilestone(suite.Ctx, mock.Anything, milestone.ID).Return(nil).Once()
	suite.NoError(suite.service.DeleteMilestone(suite.Ctx, suite.userID, &MilestoneRefInput{MilestoneID: milestone.ID}))
}

func (suite *ProjectServiceSuite) TestCreateMilestoneValidatesAndAppends() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	_, err := suite.service.CreateMilestone(suite.Ctx, suite.userID, &CreateMilestoneInput{ProjectID: suite.project.ID, Name: "Beta", TargetDate: "soon"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	created := &ent.ProjectMilestone{ID: uuid.New(), ProjectID: suite.project.ID, Name: "Beta"}
	suite.MockProjectRepo.EXPECT().NextMilestoneSortOrder(suite.Ctx, suite.project.ID).Return(2, nil).Once()
	suite.MockProjectRepo.EXPECT().CreateMilestone(suite.Ctx, mock.MatchedBy(func(input *project_repo.CreateMilestoneInput) bool {
		return input.Name == "Beta" && input.SortOrder == 2 && input.TargetDate.Format(models.DateFormat) == "2026-11-01"
	})).Return(created, nil).Once()
	suite.MockProjectRepo.EXPECT().MilestoneIssueTotals(suite.Ctx, []uuid.UUID{created.ID}).Return(map[uuid.UUID]*models.IssueTotals{
		created.ID: {Counts: map[schema.StateType]int{schema.StateTypeCompleted: 1, schema.StateTypeUnstarted: 1}, Scopes: map[schema.StateType]float64{schema.StateTypeCompleted: 1, schema.StateTypeUnstarted: 3}},
	}, nil).Once()
	milestone, err := suite.service.CreateMilestone(suite.Ctx, suite.userID, &CreateMilestoneInput{ProjectID: suite.project.ID, Name: " Beta ", TargetDate: "2026-11-01"})
	suite.Require().NoError(err)
	suite.Equal(25, milestone.Progress.Percent)
	suite.Equal(2, milestone.IssueCount)
}

func (suite *ProjectServiceSuite) TestUpdateMilestoneClearsTheDate() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectWorkspaceOfProject()
	milestone := &ent.ProjectMilestone{ID: uuid.New(), ProjectID: suite.project.ID}
	suite.MockProjectRepo.EXPECT().GetMilestone(suite.Ctx, milestone.ID).Return(milestone, nil).Once()
	suite.MockProjectRepo.EXPECT().UpdateMilestone(suite.Ctx, milestone.ID, mock.MatchedBy(func(input *project_repo.UpdateMilestoneInput) bool {
		return input.ClearTargetDate && input.TargetDate == nil
	})).Return(milestone, nil).Once()
	suite.MockProjectRepo.EXPECT().MilestoneIssueTotals(suite.Ctx, mock.Anything).Return(map[uuid.UUID]*models.IssueTotals{}, nil).Once()
	empty := ""
	_, err := suite.service.UpdateMilestone(suite.Ctx, suite.userID, &UpdateMilestoneInput{MilestoneID: milestone.ID, TargetDate: &empty})
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestCreateUpdateValidates() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	_, err := suite.service.CreateUpdate(suite.Ctx, suite.userID, &CreateProjectUpdateInput{ProjectID: suite.project.ID, Body: "  ", Health: schema.HealthOnTrack})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.CreateUpdate(suite.Ctx, suite.userID, &CreateProjectUpdateInput{ProjectID: suite.project.ID, Body: "Shipping", Health: "great"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestCreateUpdateSetsHealthAndNotifiesTheAudience() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	member := uuid.New()
	created := &ent.ProjectPost{ID: uuid.New(), ProjectID: suite.project.ID, Body: "Shipping **soon**", Health: schema.HealthAtRisk}
	suite.MockProjectRepo.EXPECT().CreateUpdate(suite.Ctx, mock.Anything, &project_repo.CreateUpdateInput{
		ProjectID: suite.project.ID, AuthorID: suite.userID, Body: "Shipping **soon**", Health: schema.HealthAtRisk,
	}).Return(created, nil).Once()
	atRisk := schema.HealthAtRisk
	suite.MockProjectRepo.EXPECT().SetHealth(suite.Ctx, mock.Anything, suite.project.ID, &atRisk).Return(nil).Once()
	suite.MockProjectRepo.EXPECT().GetUpdate(suite.Ctx, created.ID).Return(created, nil).Once()
	suite.MockProjectRepo.EXPECT().AudienceIDs(suite.Ctx, suite.project.ID).Return([]uuid.UUID{suite.userID, member}, nil).Once()
	suite.notifications.EXPECT().Notify(suite.Ctx, mock.MatchedBy(func(event models.NotificationEvent) bool {
		return event.Type == schema.NotificationProjectUpdate && len(event.RecipientIDs) == 1 && event.RecipientIDs[0] == member &&
			*event.ProjectUpdateID == created.ID && event.Body == "Shipping soon" && event.Title == "New update on Launch: At risk"
	})).Return(nil).Once()

	update, err := suite.service.CreateUpdate(suite.Ctx, suite.userID, &CreateProjectUpdateInput{ProjectID: suite.project.ID, Body: " Shipping **soon** ", Health: schema.HealthAtRisk})
	suite.Require().NoError(err)
	suite.Equal(created.ID, update.ID)
}

func (suite *ProjectServiceSuite) TestOnlyAuthorsOrAdminsChangeUpdates() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectWorkspaceOfProject()
	author := uuid.New()
	update := &ent.ProjectPost{ID: uuid.New(), ProjectID: suite.project.ID, AuthorID: &author}
	suite.MockProjectRepo.EXPECT().GetUpdate(suite.Ctx, update.ID).Return(update, nil)

	body := "edited"
	_, err := suite.service.EditUpdate(suite.Ctx, suite.userID, &EditProjectUpdateInput{UpdateID: update.ID, Body: &body})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
	err = suite.service.DeleteUpdate(suite.Ctx, suite.userID, &ProjectUpdateRefInput{UpdateID: update.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ProjectServiceSuite) TestAuthorDeletesAndHealthFallsBack() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectWorkspaceOfProject()
	update := &ent.ProjectPost{ID: uuid.New(), ProjectID: suite.project.ID, AuthorID: &suite.userID}
	suite.MockProjectRepo.EXPECT().GetUpdate(suite.Ctx, update.ID).Return(update, nil).Once()
	suite.MockProjectRepo.EXPECT().DeleteUpdate(suite.Ctx, mock.Anything, update.ID).Return(nil).Once()
	onTrack := schema.HealthOnTrack
	suite.MockProjectRepo.EXPECT().LatestHealth(suite.Ctx, mock.Anything, suite.project.ID).Return(&onTrack, nil).Once()
	suite.MockProjectRepo.EXPECT().SetHealth(suite.Ctx, mock.Anything, suite.project.ID, &onTrack).Return(nil).Once()
	suite.NoError(suite.service.DeleteUpdate(suite.Ctx, suite.userID, &ProjectUpdateRefInput{UpdateID: update.ID}))
}

func (suite *ProjectServiceSuite) TestAdminEditsSomeoneElsesUpdate() {
	suite.roles(schema.WorkspaceRoleAdmin)
	suite.expectWorkspaceOfProject()
	author := uuid.New()
	update := &ent.ProjectPost{ID: uuid.New(), ProjectID: suite.project.ID, AuthorID: &author}
	suite.MockProjectRepo.EXPECT().GetUpdate(suite.Ctx, update.ID).Return(update, nil).Twice()
	offTrack := schema.HealthOffTrack
	suite.MockProjectRepo.EXPECT().EditUpdate(suite.Ctx, mock.Anything, update.ID, &project_repo.EditUpdateInput{Health: &offTrack, EditedAt: suite.now}).Return(nil).Once()
	suite.MockProjectRepo.EXPECT().LatestHealth(suite.Ctx, mock.Anything, suite.project.ID).Return(&offTrack, nil).Once()
	suite.MockProjectRepo.EXPECT().SetHealth(suite.Ctx, mock.Anything, suite.project.ID, &offTrack).Return(nil).Once()
	_, err := suite.service.EditUpdate(suite.Ctx, suite.userID, &EditProjectUpdateInput{UpdateID: update.ID, Health: &offTrack})
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestLinksNeedHTTPURLs() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	for _, raw := range []string{"javascript:alert(1)", "ftp://files.example.com", "example.com", "https://"} {
		_, err := suite.service.CreateLink(suite.Ctx, suite.userID, &CreateProjectLinkInput{ProjectID: suite.project.ID, URL: raw, Label: "Spec"})
		suite.ErrorIs(err, errdefs.ErrInvalidInput, raw)
	}
	suite.MockProjectRepo.EXPECT().NextLinkSortOrder(suite.Ctx, suite.project.ID).Return(1, nil).Once()
	suite.MockProjectRepo.EXPECT().CreateLink(suite.Ctx, &project_repo.CreateLinkInput{
		ProjectID: suite.project.ID, URL: "https://example.com/spec", Label: "Spec", CreatorID: suite.userID, SortOrder: 1,
	}).Return(&ent.ProjectLink{ID: uuid.New(), CreatedAt: time.Now()}, nil).Once()
	_, err := suite.service.CreateLink(suite.Ctx, suite.userID, &CreateProjectLinkInput{ProjectID: suite.project.ID, URL: " https://example.com/spec ", Label: " Spec "})
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestLinksOfHiddenProjectsAreNotFound() {
	suite.outsider()
	suite.expectWorkspaceOfProject()
	link := &ent.ProjectLink{ID: uuid.New(), ProjectID: suite.project.ID}
	suite.MockProjectRepo.EXPECT().GetLink(suite.Ctx, link.ID).Return(link, nil).Once()
	err := suite.service.DeleteLink(suite.Ctx, suite.userID, &ProjectLinkRefInput{LinkID: link.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}
