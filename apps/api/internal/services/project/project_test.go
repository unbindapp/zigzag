package project_service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/models"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
	"github.com/unbindapp/zigzag-api/internal/services"
	mocks_services_notification "github.com/unbindapp/zigzag-api/mocks/services/notification"
	mocks_services_webhook "github.com/unbindapp/zigzag-api/mocks/services/webhook"
)

type ProjectServiceSuite struct {
	services.ServiceTestSuite
	service       *ProjectService
	notifications *mocks_services_notification.NotificationServiceMock
	webhooks      *mocks_services_webhook.WebhookServiceMock
	now           time.Time
	userID        uuid.UUID
	workspaceID   uuid.UUID
	teamID        uuid.UUID
	backlog       *ent.ProjectStatus
	project       *ent.Project
}

func (suite *ProjectServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.notifications = mocks_services_notification.NewNotificationServiceMock(suite.T())
	suite.webhooks = mocks_services_webhook.NewWebhookServiceMock(suite.T())
	suite.service = NewProjectService(nil, suite.MockRepo, suite.notifications, suite.webhooks)
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.teamID = uuid.New()
	suite.backlog = &ent.ProjectStatus{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Backlog", Type: schema.ProjectStatusBacklog}
	suite.project = &ent.Project{
		ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Launch", Slug: "launch-1e0e8889", StatusID: suite.backlog.ID,
		Edges: ent.ProjectEdges{Status: suite.backlog},
	}
	suite.ExpectTx()
}

func TestProjectServiceSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceSuite))
}

func (suite *ProjectServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *ProjectServiceSuite) roles(granted schema.WorkspaceRole) {
	order := []schema.WorkspaceRole{schema.WorkspaceRoleViewer, schema.WorkspaceRoleMember, schema.WorkspaceRoleAdmin}
	allowed := true
	for _, role := range order {
		if allowed {
			suite.role(role, nil)
		} else {
			suite.role(role, errdefs.ErrUnauthorized)
		}
		if role == granted {
			allowed = false
		}
	}
}

func (suite *ProjectServiceSuite) outsider() {
	for _, role := range []schema.WorkspaceRole{schema.WorkspaceRoleViewer, schema.WorkspaceRoleMember, schema.WorkspaceRoleAdmin} {
		suite.role(role, errdefs.ErrUnauthorized)
	}
}

func (suite *ProjectServiceSuite) expectTotals() {
	suite.MockProjectRepo.EXPECT().IssueTotals(suite.Ctx, mock.Anything).Return(map[uuid.UUID]*models.IssueTotals{}, nil).Maybe()
	suite.MockProjectRepo.EXPECT().MilestoneIssueTotals(suite.Ctx, mock.Anything).Return(map[uuid.UUID]*models.IssueTotals{}, nil).Maybe()
}

func (suite *ProjectServiceSuite) expectProject(project *ent.Project) {
	suite.MockProjectRepo.EXPECT().GetByID(suite.Ctx, project.ID).Return(project, nil).Maybe()
}

func (suite *ProjectServiceSuite) expectValidReferences() {
	suite.MockProjectRepo.EXPECT().CountTeamsInWorkspace(suite.Ctx, suite.workspaceID, mock.Anything).RunAndReturn(
		func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int, error) { return len(ids), nil }).Maybe()
	suite.MockProjectRepo.EXPECT().CountWorkspaceMembers(suite.Ctx, suite.workspaceID, mock.Anything).RunAndReturn(
		func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int, error) { return len(ids), nil }).Maybe()
}

func (suite *ProjectServiceSuite) createInput() *CreateProjectInput {
	return &CreateProjectInput{WorkspaceID: suite.workspaceID, Name: "Launch", TeamIDs: []uuid.UUID{suite.teamID}}
}

func (suite *ProjectServiceSuite) TestListHidesOtherWorkspaces() {
	suite.outsider()
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListProjectsInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *ProjectServiceSuite) TestListRejectsUnknownStatusType() {
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListProjectsInput{WorkspaceID: suite.workspaceID, StatusType: "someday"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestListPassesFiltersAndComputesProgress() {
	suite.roles(schema.WorkspaceRoleViewer)
	started := schema.ProjectStatusStarted
	suite.MockProjectRepo.EXPECT().List(suite.Ctx, mock.MatchedBy(func(filter *project_repo.ListProjectsFilter) bool {
		return filter.WorkspaceID == suite.workspaceID && *filter.TeamID == suite.teamID && *filter.StatusType == started &&
			filter.LeadID == nil && filter.Query == "launch" && filter.OrderBy == project_repo.OrderByTargetDate
	})).Return([]*ent.Project{suite.project}, nil).Once()
	suite.MockProjectRepo.EXPECT().IssueTotals(suite.Ctx, []uuid.UUID{suite.project.ID}).Return(map[uuid.UUID]*models.IssueTotals{
		suite.project.ID: {
			Counts: map[schema.StateType]int{schema.StateTypeCompleted: 1, schema.StateTypeStarted: 1, schema.StateTypeBacklog: 1, schema.StateTypeCanceled: 1},
			Scopes: map[schema.StateType]float64{schema.StateTypeCompleted: 3, schema.StateTypeStarted: 2, schema.StateTypeBacklog: 1, schema.StateTypeCanceled: 5},
		},
	}, nil).Once()
	suite.MockProjectRepo.EXPECT().MilestoneIssueTotals(suite.Ctx, []uuid.UUID{}).Return(map[uuid.UUID]*models.IssueTotals{}, nil).Once()

	projects, err := suite.service.List(suite.Ctx, suite.userID, &ListProjectsInput{
		WorkspaceID: suite.workspaceID, TeamID: suite.teamID, StatusType: started, Query: "  launch ", OrderBy: "target_date",
	})
	suite.Require().NoError(err)
	suite.Require().Len(projects, 1)
	suite.Equal(models.ProgressResponse{Scope: 6, Completed: 3, Started: 2, Percent: 50}, projects[0].Progress, "canceled issues are out of scope")
	suite.Equal(4, projects[0].IssueCounts.Total)
	suite.Equal(1, projects[0].IssueCounts.ByStateType.Canceled)
}

func (suite *ProjectServiceSuite) TestGetNeedsAnIdOrASlug() {
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetProjectInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestGetBySlugMasksOutsiders() {
	suite.outsider()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetProjectInput{WorkspaceID: suite.workspaceID, Slug: suite.project.Slug})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *ProjectServiceSuite) TestGetIncludesTheLatestUpdate() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectTotals()
	suite.MockProjectRepo.EXPECT().GetBySlug(suite.Ctx, suite.workspaceID, suite.project.Slug).Return(suite.project, nil).Once()
	latest := &ent.ProjectPost{ID: uuid.New(), ProjectID: suite.project.ID, Body: "All good", Health: schema.HealthOnTrack}
	suite.MockProjectRepo.EXPECT().LatestUpdate(suite.Ctx, suite.project.ID).Return(latest, nil).Once()

	project, err := suite.service.Get(suite.Ctx, suite.userID, &GetProjectInput{WorkspaceID: suite.workspaceID, Slug: suite.project.Slug})
	suite.Require().NoError(err)
	suite.Require().NotNil(project.LastUpdate)
	suite.Equal(latest.ID, project.LastUpdate.ID)
}

func (suite *ProjectServiceSuite) TestCreateNeedsMember() {
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Create(suite.Ctx, suite.userID, suite.createInput())
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ProjectServiceSuite) TestCreateHidesOtherWorkspaces() {
	suite.outsider()
	_, err := suite.service.Create(suite.Ctx, suite.userID, suite.createInput())
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *ProjectServiceSuite) TestCreateValidatesInput() {
	suite.roles(schema.WorkspaceRoleMember)
	cases := map[string]func(*CreateProjectInput){
		"blank name":            func(input *CreateProjectInput) { input.Name = "   " },
		"bad color":             func(input *CreateProjectInput) { input.Color = "blue" },
		"bad date":              func(input *CreateProjectInput) { input.StartDate = "10/16/2026" },
		"target before start":   func(input *CreateProjectInput) { input.StartDate, input.TargetDate = "2026-10-16", "2026-10-01" },
		"unknown resolution":    func(input *CreateProjectInput) { input.TargetDateResolution = "week" },
		"team of another space": func(input *CreateProjectInput) { input.TeamIDs = []uuid.UUID{uuid.New()} },
	}
	suite.MockProjectRepo.EXPECT().CountTeamsInWorkspace(suite.Ctx, suite.workspaceID, mock.Anything).Return(0, nil).Maybe()
	for name, mutate := range cases {
		input := suite.createInput()
		mutate(input)
		_, err := suite.service.Create(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}
}

func (suite *ProjectServiceSuite) TestCreateRejectsLeadsOutsideTheWorkspace() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockProjectRepo.EXPECT().CountTeamsInWorkspace(suite.Ctx, suite.workspaceID, []uuid.UUID{suite.teamID}).Return(1, nil).Once()
	stranger := uuid.New()
	suite.MockProjectRepo.EXPECT().CountWorkspaceMembers(suite.Ctx, suite.workspaceID, []uuid.UUID{stranger}).Return(0, nil).Once()
	input := suite.createInput()
	input.LeadID = stranger
	_, err := suite.service.Create(suite.Ctx, suite.userID, input)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestCreateRejectsStatusesOfOtherWorkspaces() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectValidReferences()
	foreign := &ent.ProjectStatus{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, foreign.ID).Return(foreign, nil).Once()
	input := suite.createInput()
	input.StatusID = foreign.ID
	_, err := suite.service.Create(suite.Ctx, suite.userID, input)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestCreateDefaultsAndNotifiesLeadAndMembers() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectValidReferences()
	suite.expectTotals()
	lead, member := uuid.New(), uuid.New()
	suite.MockProjectRepo.EXPECT().DefaultStatus(suite.Ctx, suite.workspaceID).Return(suite.backlog, nil).Once()
	suite.MockProjectRepo.EXPECT().MaxSortOrder(suite.Ctx, suite.workspaceID).Return(4, nil).Once()
	suite.MockProjectRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *project_repo.CreateProjectInput) bool {
		return input.StatusID == suite.backlog.ID && input.SortOrder == 5 && *input.CreatorID == suite.userID &&
			input.StartedAt == nil && len(input.MemberIDs) == 2 && len(input.Slug) == len("launch-")+8 && input.Slug[:7] == "launch-"
	})).Return(suite.project, nil).Once()
	created := *suite.project
	created.Edges.Lead = &ent.User{ID: lead, Name: "Lena"}
	suite.expectProject(&created)
	suite.notifications.EXPECT().Notify(suite.Ctx, mock.MatchedBy(func(event models.NotificationEvent) bool {
		return event.Type == schema.NotificationProjectLead && event.RecipientIDs[0] == lead && *event.ActorID == suite.userID
	})).Return(nil).Once()
	suite.notifications.EXPECT().Notify(suite.Ctx, mock.MatchedBy(func(event models.NotificationEvent) bool {
		return event.Type == schema.NotificationProjectMember && len(event.RecipientIDs) == 1 && event.RecipientIDs[0] == member
	})).Return(nil).Once()

	input := suite.createInput()
	input.LeadID = lead
	input.MemberIDs = []uuid.UUID{member, suite.userID, member}
	project, err := suite.service.Create(suite.Ctx, suite.userID, input)
	suite.Require().NoError(err)
	suite.Equal(suite.project.ID, project.ID)
}

func (suite *ProjectServiceSuite) TestCreateInAStartedStatusStampsStartedAt() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectValidReferences()
	suite.expectTotals()
	started := &ent.ProjectStatus{ID: uuid.New(), WorkspaceID: suite.workspaceID, Type: schema.ProjectStatusStarted}
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, started.ID).Return(started, nil).Once()
	suite.MockProjectRepo.EXPECT().MaxSortOrder(suite.Ctx, suite.workspaceID).Return(0, nil).Once()
	suite.MockProjectRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *project_repo.CreateProjectInput) bool {
		return input.StartedAt != nil && input.StartedAt.Equal(suite.now) && input.CompletedAt == nil
	})).Return(suite.project, nil).Once()
	suite.expectProject(suite.project)
	input := suite.createInput()
	input.StatusID = started.ID
	_, err := suite.service.Create(suite.Ctx, suite.userID, input)
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestUpdateNeedsMember() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectProject(suite.project)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateProjectInput{ProjectID: suite.project.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ProjectServiceSuite) TestUpdateOfAMissingProjectIsNotFound() {
	missing := uuid.New()
	suite.MockProjectRepo.EXPECT().GetByID(suite.Ctx, missing).Return(nil, &ent.NotFoundError{}).Once()
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateProjectInput{ProjectID: missing})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *ProjectServiceSuite) TestUpdateRejectsRemovingEveryTeam() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateProjectInput{ProjectID: suite.project.ID, TeamIDs: []uuid.UUID{}})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *ProjectServiceSuite) TestUpdateCompletingStampsAndClearsLead() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectTotals()
	canceledAt := suite.now.Add(-time.Hour)
	project := *suite.project
	project.CanceledAt = &canceledAt
	suite.expectProject(&project)
	completed := &ent.ProjectStatus{ID: uuid.New(), WorkspaceID: suite.workspaceID, Type: schema.ProjectStatusCompleted}
	suite.MockProjectRepo.EXPECT().GetStatus(suite.Ctx, completed.ID).Return(completed, nil).Once()
	suite.MockProjectRepo.EXPECT().Update(suite.Ctx, mock.Anything, project.ID, mock.MatchedBy(func(input *project_repo.UpdateProjectInput) bool {
		return *input.StatusID == completed.ID && input.CompletedAt.Equal(suite.now) && input.StartedAt.Equal(suite.now) &&
			input.ClearCanceledAt && !input.ClearCompletedAt && input.ClearLead && input.LeadID == nil
	})).Return(nil).Once()
	suite.webhooks.EXPECT().Emit(suite.Ctx, suite.workspaceID, "project.updated", mock.Anything).Return(nil).Once()

	nilLead := uuid.Nil
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateProjectInput{ProjectID: project.ID, StatusID: &completed.ID, LeadID: &nilLead})
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestUpdateNotifiesOnlyNewMembersAndNewLead() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectValidReferences()
	suite.expectTotals()
	existing, added, lead := uuid.New(), uuid.New(), uuid.New()
	project := *suite.project
	project.LeadID = &lead
	project.Edges.Members = []*ent.User{{ID: existing}}
	project.Edges.Lead = &ent.User{ID: lead}
	suite.expectProject(&project)
	suite.MockProjectRepo.EXPECT().Update(suite.Ctx, mock.Anything, project.ID, mock.MatchedBy(func(input *project_repo.UpdateProjectInput) bool {
		return input.ReplaceMembers && len(input.MemberIDs) == 2
	})).Return(nil).Once()
	suite.notifications.EXPECT().Notify(suite.Ctx, mock.MatchedBy(func(event models.NotificationEvent) bool {
		return event.Type == schema.NotificationProjectMember && len(event.RecipientIDs) == 1 && event.RecipientIDs[0] == added
	})).Return(nil).Once()
	suite.webhooks.EXPECT().Emit(suite.Ctx, suite.workspaceID, "project.updated", mock.Anything).Return(nil).Once()

	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateProjectInput{
		ProjectID: project.ID, MemberIDs: []uuid.UUID{existing, added}, LeadID: &lead,
	})
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestDeleteNeedsAdmin() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	err := suite.service.Delete(suite.Ctx, suite.userID, &ProjectRefInput{ProjectID: suite.project.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *ProjectServiceSuite) TestDeleteAsAdmin() {
	suite.roles(schema.WorkspaceRoleAdmin)
	suite.expectProject(suite.project)
	suite.MockProjectRepo.EXPECT().Delete(suite.Ctx, suite.project.ID).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &ProjectRefInput{ProjectID: suite.project.ID}))
}

func (suite *ProjectServiceSuite) TestArchiveAndTrashStampTimeAndEmit() {
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectProject(suite.project)
	suite.expectTotals()
	suite.MockProjectRepo.EXPECT().SetArchivedAt(suite.Ctx, suite.project.ID, &suite.now).Return(nil).Once()
	suite.MockProjectRepo.EXPECT().SetTrashedAt(suite.Ctx, suite.project.ID, (*time.Time)(nil)).Return(nil).Once()
	suite.webhooks.EXPECT().Emit(suite.Ctx, suite.workspaceID, "project.updated", mock.Anything).Return(nil).Twice()

	_, err := suite.service.Archive(suite.Ctx, suite.userID, &ProjectRefInput{ProjectID: suite.project.ID})
	suite.NoError(err)
	_, err = suite.service.Restore(suite.Ctx, suite.userID, &ProjectRefInput{ProjectID: suite.project.ID})
	suite.NoError(err)
}

func (suite *ProjectServiceSuite) TestArchiveNeedsMember() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectProject(suite.project)
	_, err := suite.service.Trash(suite.Ctx, suite.userID, &ProjectRefInput{ProjectID: suite.project.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}
