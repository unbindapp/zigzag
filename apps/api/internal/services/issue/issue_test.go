package issue_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
	mocks_services_githubsync "github.com/zigzagapp-org/zigzag-api/mocks/services/githubsync"
	mocks_services_notification "github.com/zigzagapp-org/zigzag-api/mocks/services/notification"
	mocks_services_webhook "github.com/zigzagapp-org/zigzag-api/mocks/services/webhook"
)

type IssueServiceSuite struct {
	services.ServiceTestSuite
	service       *IssueService
	notifications *mocks_services_notification.NotificationServiceMock
	webhooks      *mocks_services_webhook.WebhookServiceMock
	githubSync    *mocks_services_githubsync.GithubSyncServiceMock
	now           time.Time
	userID        uuid.UUID
	otherID       uuid.UUID
	workspaceID   uuid.UUID
	team          *ent.Team
	states        map[string]*ent.WorkflowState
	issue         *ent.Issue
	events        []models.NotificationEvent
	emitted       []string
}

func (suite *IssueServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.notifications = mocks_services_notification.NewNotificationServiceMock(suite.T())
	suite.webhooks = mocks_services_webhook.NewWebhookServiceMock(suite.T())
	suite.githubSync = mocks_services_githubsync.NewGithubSyncServiceMock(suite.T())
	suite.githubSync.EXPECT().AfterIssueUpdate(mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.MockGithubRepo.EXPECT().IssueLinkByIssue(mock.Anything, mock.Anything).Return(nil, &ent.NotFoundError{}).Maybe()
	suite.service = NewIssueService(&config.Config{ExternalUIUrl: "https://zz.example.com"}, suite.MockRepo, suite.notifications, suite.webhooks, suite.githubSync)
	suite.now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID, suite.otherID, suite.workspaceID = uuid.New(), uuid.New(), uuid.New()

	teamID := uuid.New()
	suite.states = map[string]*ent.WorkflowState{}
	for _, state := range []struct {
		name string
		kind schema.StateType
	}{{"Todo", schema.StateTypeUnstarted}, {"In Progress", schema.StateTypeStarted}, {"Done", schema.StateTypeCompleted}, {"Canceled", schema.StateTypeCanceled}, {"Duplicate", schema.StateTypeCanceled}} {
		suite.states[state.name] = &ent.WorkflowState{ID: uuid.New(), TeamID: teamID, Name: state.name, Type: state.kind}
	}
	suite.team = &ent.Team{ID: teamID, WorkspaceID: suite.workspaceID, Key: "ENG", Name: "Engineering", DefaultStateID: &suite.states["Todo"].ID, CyclesEnabled: true, CycleAutoAddStarted: true}
	suite.issue = suite.newIssue(7, suite.states["Todo"])
	suite.events, suite.emitted = nil, nil
	suite.ExpectTx()

	suite.notifications.EXPECT().Notify(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, event models.NotificationEvent) error {
		suite.events = append(suite.events, event)
		return nil
	}).Maybe()
	suite.webhooks.EXPECT().Emit(mock.Anything, suite.workspaceID, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ uuid.UUID, event string, _ any) error {
		suite.emitted = append(suite.emitted, event)
		return nil
	}).Maybe()
	suite.MockTeamRepo.EXPECT().GetState(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, id uuid.UUID) (*ent.WorkflowState, error) {
		for _, state := range suite.states {
			if state.ID == id {
				return state, nil
			}
		}
		return nil, &ent.NotFoundError{}
	}).Maybe()
	suite.MockTeamRepo.EXPECT().ListStates(mock.Anything, teamID).RunAndReturn(func(context.Context, uuid.UUID) ([]*ent.WorkflowState, error) {
		return []*ent.WorkflowState{suite.states["Todo"], suite.states["In Progress"], suite.states["Done"], suite.states["Canceled"], suite.states["Duplicate"]}, nil
	}).Maybe()
}

func (suite *IssueServiceSuite) newIssue(number int, state *ent.WorkflowState) *ent.Issue {
	return &ent.Issue{
		ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: suite.team.ID, Number: number, Title: "Fix login", StateID: state.ID,
		Edges: ent.IssueEdges{Team: suite.team, State: state, Workspace: &ent.Workspace{ID: suite.workspaceID, Slug: "acme"}},
	}
}

func (suite *IssueServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *IssueServiceSuite) roles(held schema.WorkspaceRole) {
	for _, role := range []schema.WorkspaceRole{schema.WorkspaceRoleViewer, schema.WorkspaceRoleMember, schema.WorkspaceRoleAdmin} {
		var result error
		if !permissions_repo.RoleImplies(held, role) {
			result = errdefs.ErrUnauthorized
		}
		suite.role(role, result)
	}
}

func (suite *IssueServiceSuite) loads(issues ...*ent.Issue) {
	for _, issue := range issues {
		suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, issue.ID).Return(issue, nil).Maybe()
		suite.MockIssueRepo.EXPECT().LockWithEdges(mock.Anything, mock.Anything, issue.ID).Return(issue, nil).Maybe()
		suite.MockIssueRepo.EXPECT().GetByID(mock.Anything, issue.ID).Return(issue, nil).Maybe()
	}
}

func (suite *IssueServiceSuite) expectRespond() {
	suite.MockIssueRepo.EXPECT().Counts(mock.Anything, mock.Anything).Return(map[uuid.UUID]models.IssueCounts{}, nil).Maybe()
	suite.MockTeamRepo.EXPECT().ListByWorkspace(mock.Anything, suite.workspaceID, true).Return([]*ent.Team{suite.team}, nil).Maybe()
	suite.MockIssueRepo.EXPECT().Children(mock.Anything, mock.Anything).Return([]*ent.Issue{}, nil).Maybe()
	suite.MockIssueRepo.EXPECT().Relations(mock.Anything, mock.Anything).Return(nil, nil, nil).Maybe()
	suite.MockIssueRepo.EXPECT().Attachments(mock.Anything, mock.Anything).Return([]*ent.Attachment{}, nil).Maybe()
	suite.MockIssueRepo.EXPECT().Subscribers(mock.Anything, mock.Anything).Return([]*ent.User{}, nil).Maybe()
	suite.MockIssueRepo.EXPECT().Reactions(mock.Anything, mock.Anything).Return([]*ent.Reaction{}, nil).Maybe()
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil).Maybe()
}

func (suite *IssueServiceSuite) member(userIDs ...uuid.UUID) {
	for _, id := range userIDs {
		suite.MockWorkspaceRepo.EXPECT().GetMember(mock.Anything, nil, suite.workspaceID, id).Return(&ent.WorkspaceMember{UserID: id}, nil).Maybe()
	}
}

func (suite *IssueServiceSuite) changes(fields ...string) func([]schema.FieldChange) bool {
	return func(changes []schema.FieldChange) bool {
		if len(changes) != len(fields) {
			return false
		}
		for i, change := range changes {
			if change.Field != fields[i] {
				return false
			}
		}
		return true
	}
}

func (suite *IssueServiceSuite) eventsOf(kind schema.NotificationType) []models.NotificationEvent {
	found := []models.NotificationEvent{}
	for _, event := range suite.events {
		if event.Type == kind {
			found = append(found, event)
		}
	}
	return found
}

func (suite *IssueServiceSuite) TestGetMasksIssuesTheCallerCannotSee() {
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, nil, suite.issue.ID).Return(nil, &ent.NotFoundError{}).Once()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{IssueID: suite.issue.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	suite.loads(suite.issue)
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err = suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{IssueID: suite.issue.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound, "a workspace the caller is not in reads as not found")
}

func (suite *IssueServiceSuite) TestPrivateTeamIssuesHideFromNonMembers() {
	suite.team.Private = true
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockTeamRepo.EXPECT().IsMember(mock.Anything, suite.team.ID, suite.userID).Return(false, nil).Once()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{IssueID: suite.issue.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *IssueServiceSuite) TestParentsInPrivateTeamsStayHidden() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectRespond()
	privateTeam := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "SEC", Private: true}
	parent := &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: privateTeam.ID, Number: 1, Title: "Secret", Edges: ent.IssueEdges{Team: privateTeam}}
	suite.issue.ParentID = &parent.ID
	suite.issue.Edges.Parent = parent
	suite.loads(suite.issue)
	suite.MockTeamRepo.EXPECT().IsMember(mock.Anything, privateTeam.ID, suite.userID).Return(false, nil).Once()

	issue, err := suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{IssueID: suite.issue.ID})
	suite.Require().NoError(err)
	suite.Nil(issue.Parent, "a public child must not reveal the title of a private parent")
}

func (suite *IssueServiceSuite) TestGetByIdentifierResolvesInVisibleWorkspaces() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.loads(suite.issue)
	suite.expectRespond()
	set := permissions_repo.NewPermissionSet(false, map[uuid.UUID]schema.WorkspaceRole{suite.workspaceID: schema.WorkspaceRoleViewer})
	suite.MockPermissionsRepo.EXPECT().GetUserPermissionSet(mock.Anything, suite.userID).Return(set, nil)
	suite.MockIssueRepo.EXPECT().FindByIdentifier(mock.Anything, []uuid.UUID{suite.workspaceID}, "ENG", 7).Return([]uuid.UUID{suite.issue.ID}, nil).Once()

	issue, err := suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{Identifier: "eng-7"})
	suite.Require().NoError(err)
	suite.Equal("ENG-7", issue.Identifier)
	suite.Equal("https://zz.example.com/acme/issue/ENG-7", issue.URL)
	suite.Equal("eng-7-fix-login", issue.GitBranchName)
	suite.NotNil(issue.Relations)

	_, err = suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{Identifier: "not an id"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockIssueRepo.EXPECT().FindByIdentifier(mock.Anything, []uuid.UUID{suite.workspaceID}, "ENG", 8).Return([]uuid.UUID{uuid.New(), uuid.New()}, nil).Once()
	_, err = suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{Identifier: "ENG-8"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "an identifier shared by two workspaces asks for workspace_id")

	suite.MockIssueRepo.EXPECT().FindByIdentifier(mock.Anything, []uuid.UUID{suite.workspaceID}, "ENG", 9).Return([]uuid.UUID{}, nil).Once()
	_, err = suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{Identifier: "ENG-9", WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *IssueServiceSuite) TestListChecksTheWorkspaceAndValidatesFilters() {
	suite.role(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err := suite.service.List(suite.Ctx, suite.userID, &ListIssuesInput{WorkspaceID: suite.workspaceID})
	suite.ErrorIs(err, errdefs.ErrNotFound)

	suite.SetupTest()
	suite.roles(schema.WorkspaceRoleViewer)
	for _, input := range []*ListIssuesInput{
		{WorkspaceID: suite.workspaceID, PaginationInput: models.PaginationInput{Cursor: "garbage!"}},
		{WorkspaceID: suite.workspaceID, AssigneeIDs: []string{"bob"}},
		{WorkspaceID: suite.workspaceID, Priorities: []int{5}},
		{WorkspaceID: suite.workspaceID, StateTypes: []schema.StateType{"review"}},
		{WorkspaceID: suite.workspaceID, DueBefore: "next week"},
	} {
		_, err := suite.service.List(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput)
	}
}

func (suite *IssueServiceSuite) TestListPagesAndKeepsToVisibleTeams() {
	suite.roles(schema.WorkspaceRoleViewer)
	suite.expectRespond()
	hidden := uuid.New()
	issues := []*ent.Issue{suite.issue, suite.newIssue(8, suite.states["Todo"]), suite.newIssue(9, suite.states["Todo"])}
	suite.MockIssueRepo.EXPECT().List(mock.Anything, mock.MatchedBy(func(filter *issue_repo.ListFilter) bool {
		return filter.After == nil && filter.Limit == 3 && len(filter.TeamIDs) == 1 && filter.TeamIDs[0] == suite.team.ID &&
			filter.Unassigned && len(filter.AssigneeIDs) == 1 && filter.OrderBy == issue_repo.OrderCreatedAt && filter.Descending
	})).Return(issues, 10, nil).Once()

	input := &ListIssuesInput{
		WorkspaceID: suite.workspaceID, TeamIDs: []uuid.UUID{suite.team.ID, hidden}, AssigneeIDs: []string{"null", suite.otherID.String()},
		OrderBy: "created_at", PaginationInput: models.PaginationInput{Limit: 2},
	}
	result, err := suite.service.List(suite.Ctx, suite.userID, input)
	suite.Require().NoError(err)
	suite.Len(result.Issues, 2)
	suite.Equal(10, result.Total)
	suite.Require().NotNil(result.NextCursor)

	suite.MockIssueRepo.EXPECT().List(mock.Anything, mock.MatchedBy(func(filter *issue_repo.ListFilter) bool {
		return filter.After != nil && len(filter.After.Values) == 2
	})).Return(issues[2:], 10, nil).Once()
	input.Cursor = *result.NextCursor
	result, err = suite.service.List(suite.Ctx, suite.userID, input)
	suite.Require().NoError(err)
	suite.Len(result.Issues, 1)
	suite.Nil(result.NextCursor)

	input.OrderBy = "title"
	_, err = suite.service.List(suite.Ctx, suite.userID, input)
	suite.ErrorIs(err, errdefs.ErrInvalidInput, "a cursor only continues the ordering it came from")
}

func (suite *IssueServiceSuite) TestCreateNeedsMember() {
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{TeamID: suite.team.ID, Title: "x"})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)

	suite.SetupTest()
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.roles("")
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{TeamID: suite.team.ID, Title: "x"})
	suite.ErrorIs(err, errdefs.ErrNotFound, "outsiders do not learn the team exists")

	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, uuid.Nil).Return(nil, &ent.NotFoundError{})
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{Title: "x"})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *IssueServiceSuite) TestCreateValidatesReferences() {
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.roles(schema.WorkspaceRoleMember)
	foreignState := &ent.WorkflowState{ID: uuid.New(), TeamID: uuid.New(), Type: schema.StateTypeUnstarted}
	suite.states["Foreign"] = foreignState
	stranger := uuid.New()
	suite.MockWorkspaceRepo.EXPECT().GetMember(mock.Anything, nil, suite.workspaceID, stranger).Return(nil, &ent.NotFoundError{})
	group := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Area", IsGroup: true}
	teamLabel := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Ops only", TeamID: new(uuid.New())}
	archived := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Old", ArchivedAt: &suite.now}
	single := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Size", IsGroup: true, GroupSingleSelect: true}
	small := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "S", ParentID: &single.ID, Edges: ent.LabelEdges{Parent: single}}
	large := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "L", ParentID: &single.ID, Edges: ent.LabelEdges{Parent: single}}
	for _, label := range []*ent.Label{group, teamLabel, archived} {
		suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{label.ID}).Return([]*ent.Label{label}, nil)
	}
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{small.ID, large.ID}).Return([]*ent.Label{small, large}, nil)
	missing := uuid.New()
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{missing}).Return([]*ent.Label{}, nil)
	otherProject := &ent.Project{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockIssueRepo.EXPECT().GetProject(mock.Anything, otherProject.ID).Return(otherProject, nil)
	project := &ent.Project{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockIssueRepo.EXPECT().GetProject(mock.Anything, project.ID).Return(project, nil)
	milestone := &ent.ProjectMilestone{ID: uuid.New(), ProjectID: uuid.New()}
	suite.MockIssueRepo.EXPECT().GetMilestone(mock.Anything, milestone.ID).Return(milestone, nil)
	foreignCycle := &ent.Cycle{ID: uuid.New(), TeamID: uuid.New()}
	suite.MockCycleRepo.EXPECT().GetByID(mock.Anything, foreignCycle.ID).Return(foreignCycle, nil)
	foreignParent := &ent.Issue{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, foreignParent.ID).Return(foreignParent, nil)
	privateTeam := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "SEC", Private: true}
	hiddenParent := &ent.Issue{ID: uuid.New(), WorkspaceID: suite.workspaceID, TeamID: privateTeam.ID, Edges: ent.IssueEdges{Team: privateTeam}}
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, hiddenParent.ID).Return(hiddenParent, nil)
	suite.MockTeamRepo.EXPECT().IsMember(mock.Anything, privateTeam.ID, suite.userID).Return(false, nil)

	for name, input := range map[string]*CreateIssueInput{
		"parent in a private team": {Title: "x", ParentID: &hiddenParent.ID},
		"blank title":              {Title: "   "},
		"state of another team":    {Title: "x", StateID: &foreignState.ID},
		"assignee outside":         {Title: "x", AssigneeID: &stranger},
		"label group":              {Title: "x", LabelIDs: []uuid.UUID{group.ID}},
		"another team's label":     {Title: "x", LabelIDs: []uuid.UUID{teamLabel.ID}},
		"archived label":           {Title: "x", LabelIDs: []uuid.UUID{archived.ID}},
		"unknown label":            {Title: "x", LabelIDs: []uuid.UUID{missing}},
		"two single-select":        {Title: "x", LabelIDs: []uuid.UUID{small.ID, large.ID}},
		"foreign project":          {Title: "x", ProjectID: &otherProject.ID},
		"milestone elsewhere":      {Title: "x", ProjectID: &project.ID, MilestoneID: &milestone.ID},
		"cycle of another team":    {Title: "x", CycleID: &foreignCycle.ID},
		"parent elsewhere":         {Title: "x", ParentID: &foreignParent.ID},
		"zero estimate":            {Title: "x", Estimate: new(0.0)},
		"bad due date":             {Title: "x", DueDate: "soon"},
	} {
		input.TeamID = suite.team.ID
		_, err := suite.service.Create(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}
}

func (suite *IssueServiceSuite) TestCreateNumbersSubscribesRecordsAndNotifies() {
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.roles(schema.WorkspaceRoleMember)
	mentioned := uuid.New()
	suite.member(suite.otherID, mentioned)
	suite.MockIssueRepo.EXPECT().TopSortOrder(mock.Anything, mock.Anything, suite.states["Todo"].ID).Return(-3, nil).Once()
	suite.MockTeamRepo.EXPECT().NextIssueNumber(mock.Anything, mock.Anything, suite.team.ID).Return(7, nil).Once()
	suite.MockIssueRepo.EXPECT().Create(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateIssueInput) bool {
		return input.Number == 7 && input.StateID == suite.states["Todo"].ID && input.SortOrder == -3 && input.Priority == 0 &&
			input.DescriptionText == "ping @Cat" && *input.CreatorID == suite.userID &&
			len(input.SubscriberIDs) == 3 && input.SubscriberIDs[0] == suite.userID && input.SubscriberIDs[1] == mentioned && input.SubscriberIDs[2] == suite.otherID
	})).Return(suite.issue, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, &suite.userID, schema.HistorySourceUser, mock.MatchedBy(suite.changes("created"))).Return(nil).Once()
	suite.loads(suite.issue)
	suite.expectRespond()

	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{
		TeamID: suite.team.ID, Title: " Fix login ", Description: "ping @[Cat](user:" + mentioned.String() + ")", AssigneeID: &suite.otherID,
	})
	suite.Require().NoError(err)
	assigned := suite.eventsOf(schema.NotificationIssueAssigned)
	suite.Require().Len(assigned, 1)
	suite.Equal([]uuid.UUID{suite.otherID}, assigned[0].RecipientIDs)
	suite.Equal("ENG-7 Fix login", assigned[0].Title)
	suite.Equal([]uuid.UUID{mentioned}, suite.eventsOf(schema.NotificationIssueMentioned)[0].RecipientIDs)
	suite.Equal([]string{"issue.created"}, suite.emitted)
}

func (suite *IssueServiceSuite) TestCreateInStartedStateJoinsTheActiveCycle() {
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.roles(schema.WorkspaceRoleMember)
	cycleID := uuid.New()
	suite.MockCycleRepo.EXPECT().ActiveCycleIDs(mock.Anything, []uuid.UUID{suite.team.ID}, suite.now).Return(map[uuid.UUID]uuid.UUID{suite.team.ID: cycleID}, nil).Once()
	suite.MockIssueRepo.EXPECT().TopSortOrder(mock.Anything, mock.Anything, mock.Anything).Return(0, nil)
	suite.MockTeamRepo.EXPECT().NextIssueNumber(mock.Anything, mock.Anything, suite.team.ID).Return(8, nil)
	suite.MockIssueRepo.EXPECT().Create(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateIssueInput) bool {
		return input.CycleID != nil && *input.CycleID == cycleID && input.StartedAt != nil && input.StartedAt.Equal(suite.now)
	})).Return(suite.issue, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.loads(suite.issue)
	suite.expectRespond()

	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{TeamID: suite.team.ID, Title: "Go", StateID: &suite.states["In Progress"].ID})
	suite.NoError(err)
}

func (suite *IssueServiceSuite) TestCreateFromTemplateFillsGapsAndAddsSubIssues() {
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.roles(schema.WorkspaceRoleMember)
	template := &ent.Template{ID: uuid.New(), WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindIssue, Data: schema.TemplateData{
		Description: "Steps", Priority: new(2), SubIssues: []schema.TemplateSubIssue{{Title: "Write test"}},
	}}
	suite.MockTemplateRepo.EXPECT().GetByID(mock.Anything, template.ID).Return(template, nil)
	suite.MockIssueRepo.EXPECT().TopSortOrder(mock.Anything, mock.Anything, mock.Anything).Return(0, nil)
	suite.MockTeamRepo.EXPECT().NextIssueNumber(mock.Anything, mock.Anything, suite.team.ID).Return(9, nil)
	suite.MockIssueRepo.EXPECT().Create(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateIssueInput) bool {
		return input.Title == "Bug" && input.Description == "Steps" && input.Priority == 2
	})).Return(suite.issue, nil).Once()
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, suite.issue.ID).Return(suite.issue, nil)
	child := suite.newIssue(10, suite.states["Todo"])
	suite.MockIssueRepo.EXPECT().Create(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateIssueInput) bool {
		return input.Title == "Write test" && *input.ParentID == suite.issue.ID
	})).Return(child, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.loads(suite.issue, child)
	suite.expectRespond()

	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{TeamID: suite.team.ID, Title: "Bug", TemplateID: &template.ID})
	suite.NoError(err)
	suite.Equal([]string{"issue.created", "issue.created"}, suite.emitted, "template sub-issues are announced like any new issue")

	project := &ent.Template{ID: uuid.New(), WorkspaceID: suite.workspaceID, Kind: schema.TemplateKindProject}
	suite.MockTemplateRepo.EXPECT().GetByID(mock.Anything, project.ID).Return(project, nil)
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{TeamID: suite.team.ID, Title: "Bug", TemplateID: &project.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *IssueServiceSuite) TestUpdateNeedsMember() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleViewer)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, Title: new("x")})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *IssueServiceSuite) TestUpdateStateAssigneeAndPriority() {
	previous := uuid.New()
	suite.issue.AssigneeID = &previous
	suite.issue.Edges.Assignee = &ent.User{ID: previous, Name: "Prev"}
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.member(suite.otherID)
	suite.MockUserRepo.EXPECT().GetByID(mock.Anything, suite.otherID).Return(&ent.User{ID: suite.otherID, Name: "Other"}, nil)
	cycleID := uuid.New()
	suite.MockCycleRepo.EXPECT().ActiveCycleIDs(mock.Anything, []uuid.UUID{suite.team.ID}, suite.now).Return(map[uuid.UUID]uuid.UUID{suite.team.ID: cycleID}, nil).Once()
	suite.MockCycleRepo.EXPECT().GetByID(mock.Anything, cycleID).Return(&ent.Cycle{ID: cycleID, Number: 4}, nil).Once()
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{}).Return([]*ent.Label{}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.StateID == suite.states["In Progress"].ID && fields.StartedAt != nil && *fields.CycleID == cycleID &&
			*fields.AssigneeID == suite.otherID && fields.Priority == 1
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, suite.issue.ID, []uuid.UUID{}, []uuid.UUID{}).Return(nil)
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID, suite.otherID).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, &suite.userID, schema.HistorySourceUser, mock.MatchedBy(func(changes []schema.FieldChange) bool {
		return suite.changes("state", "priority", "assignee", "cycle")(changes) &&
			changes[0].FromLabel == "Todo" && changes[0].ToLabel == "In Progress" && changes[1].ToLabel == "Urgent" && changes[2].ToLabel == "Other" && changes[3].ToLabel == "Cycle 4"
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, suite.issue.ID).Return([]uuid.UUID{suite.userID, suite.otherID}, nil)
	suite.expectRespond()

	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, StateID: &suite.states["In Progress"].ID, AssigneeID: &suite.otherID, Priority: new(1)})
	suite.Require().NoError(err)
	suite.Equal([]uuid.UUID{suite.otherID}, suite.eventsOf(schema.NotificationIssueAssigned)[0].RecipientIDs)
	suite.Equal([]uuid.UUID{previous}, suite.eventsOf(schema.NotificationIssueUnassigned)[0].RecipientIDs)
	suite.Equal([]uuid.UUID{suite.otherID}, suite.eventsOf(schema.NotificationIssueStateChanged)[0].RecipientIDs, "the actor is never notified")
	suite.Len(suite.eventsOf(schema.NotificationIssuePriorityChanged), 1)
	suite.Equal([]string{"issue.updated"}, suite.emitted)
}

func (suite *IssueServiceSuite) TestUpdateRejectsBadValues() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	child := suite.newIssue(8, suite.states["Todo"])
	suite.loads(child)
	suite.MockIssueRepo.EXPECT().AncestorIDs(mock.Anything, mock.Anything, child.ID).Return([]uuid.UUID{suite.issue.ID}, nil)
	foreign := &ent.WorkflowState{ID: uuid.New(), TeamID: uuid.New()}
	suite.states["Foreign"] = foreign

	for name, input := range map[string]*UpdateIssueInput{
		"own parent":    {ParentID: &suite.issue.ID},
		"descendant":    {ParentID: &child.ID},
		"foreign state": {StateID: &foreign.ID},
		"blank title":   {Title: new(" ")},
		"due date":      {DueDate: new("tomorrow")},
		"snooze":        {SnoozedUntil: new("later")},
		"negative":      {Estimate: new(-2.0)},
		"zero estimate": {Estimate: new(0.0)},
	} {
		input.IssueID = suite.issue.ID
		_, err := suite.service.Update(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}
}

func (suite *IssueServiceSuite) TestSingleSelectGroupReplacesTheSibling() {
	group := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Size", IsGroup: true, GroupSingleSelect: true}
	small := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "S", ParentID: &group.ID, Edges: ent.LabelEdges{Parent: group}}
	large := &ent.Label{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "L", ParentID: &group.ID, Edges: ent.LabelEdges{Parent: group}}
	suite.issue.Edges.Labels = []*ent.Label{small}
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{small.ID, large.ID}).Return([]*ent.Label{small, large}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, suite.issue.ID, []uuid.UUID{large.ID}, []uuid.UUID{small.ID}).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil)
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything, mock.Anything, mock.MatchedBy(suite.changes("labels_added", "labels_removed"))).Return(nil).Once()
	suite.expectRespond()

	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, AddLabelIDs: []uuid.UUID{large.ID}})
	suite.NoError(err)
}

func (suite *IssueServiceSuite) TestMoveToAnotherTeamRenumbersAndKeepsARedirect() {
	ops := &ent.Team{ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "OPS"}
	opsTodo := &ent.WorkflowState{ID: uuid.New(), TeamID: ops.ID, Name: "Todo", Type: schema.StateTypeUnstarted}
	suite.states["OpsTodo"] = opsTodo
	cycleID := uuid.New()
	suite.issue.CycleID = &cycleID
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, ops.ID).Return(ops, nil)
	suite.MockTeamRepo.EXPECT().ListStates(mock.Anything, ops.ID).Return([]*ent.WorkflowState{opsTodo}, nil)
	suite.MockIssueRepo.EXPECT().RecordMove(mock.Anything, mock.Anything, suite.issue.ID, suite.team.ID, 7).Return(nil).Once()
	suite.MockTeamRepo.EXPECT().NextIssueNumber(mock.Anything, mock.Anything, ops.ID).Return(3, nil).Once()
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{}).Return([]*ent.Label{}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.TeamID == ops.ID && fields.Number == 3 && fields.StateID == opsTodo.ID && fields.CycleID == nil
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil)
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything, mock.Anything, mock.MatchedBy(func(changes []schema.FieldChange) bool {
		return suite.changes("team", "state", "cycle")(changes) && changes[0].FromLabel == "ENG-7" && changes[0].ToLabel == "OPS-3"
	})).Return(nil).Once()
	suite.expectRespond()

	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, TeamID: &ops.ID})
	suite.NoError(err)

	elsewhere := &ent.Team{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, elsewhere.ID).Return(elsewhere, nil)
	_, err = suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, TeamID: &elsewhere.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *IssueServiceSuite) TestBulkUpdateNeedsMemberOnEveryIssue() {
	other := suite.newIssue(8, suite.states["Todo"])
	other.WorkspaceID = uuid.New()
	suite.loads(suite.issue, other)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: other.WorkspaceID}}).Return(errdefs.ErrUnauthorized)
	_, err := suite.service.BulkUpdate(suite.Ctx, suite.userID, &BulkUpdateIssuesInput{IssueIDs: []uuid.UUID{suite.issue.ID, other.ID}, Priority: new(1)})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *IssueServiceSuite) TestBulkUpdateWritesOneHistoryEntryPerIssue() {
	second := suite.newIssue(8, suite.states["Todo"])
	suite.loads(suite.issue, second)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{}).Return([]*ent.Label{}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything, mock.Anything, mock.MatchedBy(suite.changes("priority"))).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, second.ID, mock.Anything, mock.Anything, mock.MatchedBy(suite.changes("priority"))).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().GetManyWithEdges(mock.Anything, nil, []uuid.UUID{suite.issue.ID, second.ID}).Return([]*ent.Issue{suite.issue, second}, nil)
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil)
	suite.expectRespond()

	issues, err := suite.service.BulkUpdate(suite.Ctx, suite.userID, &BulkUpdateIssuesInput{IssueIDs: []uuid.UUID{suite.issue.ID, second.ID, suite.issue.ID}, Priority: new(1)})
	suite.Require().NoError(err)
	suite.Len(issues, 2)
	suite.Equal([]string{"issue.updated", "issue.updated"}, suite.emitted)
}

func (suite *IssueServiceSuite) TestDeleteNeedsAdmin() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.ErrorIs(suite.service.Delete(suite.Ctx, suite.userID, &IssueRefInput{IssueID: suite.issue.ID}), errdefs.ErrUnauthorized)

	suite.SetupTest()
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleAdmin)
	child := uuid.New()
	suite.MockIssueRepo.EXPECT().ChildIDs(mock.Anything, mock.Anything, suite.issue.ID).Return([]uuid.UUID{child}, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistoryMany(mock.Anything, mock.Anything, []issue_repo.HistoryEntry{{
		IssueID: child, ActorID: &suite.userID, Source: schema.HistorySourceUser,
		Changes: []schema.FieldChange{{Field: "parent", From: suite.issue.ID.String(), FromLabel: "ENG-7"}},
	}}).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().Delete(mock.Anything, mock.Anything, suite.issue.ID).Return(nil).Once()
	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &IssueRefInput{IssueID: suite.issue.ID}))
	suite.Equal([]string{"issue.deleted"}, suite.emitted)
}

func (suite *IssueServiceSuite) TestTrashRecordsHistoryOnce() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond()
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.TrashedAt != nil && fields.TrashedAt.Equal(suite.now)
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything, mock.Anything, mock.MatchedBy(suite.changes("trashed"))).Return(nil).Once()
	_, err := suite.service.Trash(suite.Ctx, suite.userID, &IssueRefInput{IssueID: suite.issue.ID})
	suite.NoError(err)

	_, err = suite.service.Restore(suite.Ctx, suite.userID, &IssueRefInput{IssueID: suite.issue.ID})
	suite.NoError(err, "restoring an issue that is not trashed changes nothing")

	suite.SetupTest()
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleViewer)
	_, err = suite.service.Archive(suite.Ctx, suite.userID, &IssueRefInput{IssueID: suite.issue.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *IssueServiceSuite) TestRelationsValidateAndNotify() {
	blocked := suite.newIssue(8, suite.states["Todo"])
	blocked.AssigneeID = &suite.otherID
	suite.loads(suite.issue, blocked)
	suite.roles(schema.WorkspaceRoleMember)

	_, err := suite.service.CreateRelation(suite.Ctx, suite.userID, &CreateRelationInput{IssueID: suite.issue.ID, RelatedIssueID: suite.issue.ID, Type: schema.IssueRelationBlocks})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.CreateRelation(suite.Ctx, suite.userID, &CreateRelationInput{IssueID: suite.issue.ID, RelatedIssueID: blocked.ID, Type: "parent"})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.MockIssueRepo.EXPECT().RelationExists(mock.Anything, suite.issue.ID, blocked.ID, schema.IssueRelationRelated).Return(true, nil).Once()
	_, err = suite.service.CreateRelation(suite.Ctx, suite.userID, &CreateRelationInput{IssueID: suite.issue.ID, RelatedIssueID: blocked.ID, Type: schema.IssueRelationRelated})
	suite.ErrorIs(err, errdefs.ErrConflict)

	suite.MockIssueRepo.EXPECT().RelationExists(mock.Anything, suite.issue.ID, blocked.ID, schema.IssueRelationBlocks).Return(false, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateRelation(mock.Anything, mock.Anything, suite.issue.ID, blocked.ID, schema.IssueRelationBlocks).Return(&ent.IssueRelation{ID: uuid.New(), Type: schema.IssueRelationBlocks}, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(suite.changes("relation_added"))).Return(nil).Twice()
	relation, err := suite.service.CreateRelation(suite.Ctx, suite.userID, &CreateRelationInput{IssueID: suite.issue.ID, RelatedIssueID: blocked.ID, Type: schema.IssueRelationBlocks})
	suite.Require().NoError(err)
	suite.Equal("ENG-8", relation.Issue.Identifier)
	blocking := suite.eventsOf(schema.NotificationIssueBlocking)
	suite.Require().Len(blocking, 1)
	suite.Equal([]uuid.UUID{suite.otherID}, blocking[0].RecipientIDs)
}

func (suite *IssueServiceSuite) TestDuplicateMovesToTheDuplicateState() {
	canonical := suite.newIssue(8, suite.states["Todo"])
	suite.loads(suite.issue, canonical)
	suite.roles(schema.WorkspaceRoleMember)
	suite.expectRespond()
	suite.MockIssueRepo.EXPECT().RelationExists(mock.Anything, suite.issue.ID, canonical.ID, schema.IssueRelationDuplicate).Return(false, nil)
	suite.MockIssueRepo.EXPECT().CreateRelation(mock.Anything, mock.Anything, suite.issue.ID, canonical.ID, schema.IssueRelationDuplicate).Return(&ent.IssueRelation{ID: uuid.New()}, nil)
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{}).Return([]*ent.Label{}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.StateID == suite.states["Duplicate"].ID && fields.CanceledAt != nil
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockIssueRepo.EXPECT().SubscriberIDs(mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil)

	_, err := suite.service.CreateRelation(suite.Ctx, suite.userID, &CreateRelationInput{IssueID: suite.issue.ID, RelatedIssueID: canonical.ID, Type: schema.IssueRelationDuplicate})
	suite.NoError(err)
}

func (suite *IssueServiceSuite) TestAttachmentsValidateTheirSource() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	foreign := &ent.Upload{ID: uuid.New(), WorkspaceID: uuid.New()}
	suite.MockUploadRepo.EXPECT().GetByID(mock.Anything, foreign.ID).Return(foreign, nil)
	for name, input := range map[string]*CreateAttachmentInput{
		"neither":      {Title: "x"},
		"both":         {Title: "x", URL: "https://a.b", UploadID: &foreign.ID},
		"bad scheme":   {Title: "x", URL: "javascript:alert(1)"},
		"other upload": {Title: "x", UploadID: &foreign.ID},
		"blank title":  {Title: " ", URL: "https://a.b"},
	} {
		input.IssueID = suite.issue.ID
		_, err := suite.service.CreateAttachment(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}

	upload := &ent.Upload{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockUploadRepo.EXPECT().GetByID(mock.Anything, upload.ID).Return(upload, nil)
	suite.MockIssueRepo.EXPECT().CreateAttachment(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateAttachmentInput) bool {
		return input.SourceType == "upload" && input.URL == models.UploadURL(upload.ID)
	})).Return(&ent.Attachment{ID: uuid.New(), IssueID: suite.issue.ID, Title: "Logs"}, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything, mock.Anything, mock.MatchedBy(suite.changes("attachment_added"))).Return(nil).Once()
	_, err := suite.service.CreateAttachment(suite.Ctx, suite.userID, &CreateAttachmentInput{IssueID: suite.issue.ID, Title: "Logs", UploadID: &upload.ID})
	suite.NoError(err)
}

func (suite *IssueServiceSuite) TestSubscribeOthersMustBeMembers() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	stranger := uuid.New()
	suite.MockWorkspaceRepo.EXPECT().GetMember(mock.Anything, nil, suite.workspaceID, stranger).Return(nil, &ent.NotFoundError{})
	_, err := suite.service.Subscribe(suite.Ctx, suite.userID, &SubscriptionInput{IssueID: suite.issue.ID, UserID: &stranger})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	suite.expectRespond()
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, nil, suite.issue.ID, suite.userID).Return(nil).Once()
	_, err = suite.service.Subscribe(suite.Ctx, suite.userID, &SubscriptionInput{IssueID: suite.issue.ID})
	suite.NoError(err)
}

func (suite *IssueServiceSuite) TestReactionsNeedExactlyOneTarget() {
	commentID := uuid.New()
	for name, input := range map[string]*ReactionInput{
		"none":     {Emoji: "👍"},
		"two":      {Emoji: "👍", IssueID: &suite.issue.ID, CommentID: &commentID},
		"no emoji": {Emoji: "  ", IssueID: &suite.issue.ID},
	} {
		_, err := suite.service.AddReaction(suite.Ctx, suite.userID, input)
		suite.ErrorIs(err, errdefs.ErrInvalidInput, name)
	}

	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleViewer)
	suite.MockIssueRepo.EXPECT().IssueOfComment(mock.Anything, commentID).Return(suite.issue.ID, nil)
	_, err := suite.service.AddReaction(suite.Ctx, suite.userID, &ReactionInput{Emoji: "👍", CommentID: &commentID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized, "viewers cannot react")

	updateID := uuid.New()
	suite.MockIssueRepo.EXPECT().WorkspaceOfProjectUpdate(mock.Anything, updateID).Return(uuid.New(), nil)
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, mock.Anything).Return(errdefs.ErrUnauthorized).Maybe()
	_, err = suite.service.AddReaction(suite.Ctx, suite.userID, &ReactionInput{Emoji: "👍", ProjectUpdateID: &updateID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *IssueServiceSuite) TestAddReactionReturnsTheTargetsReactions() {
	suite.issue.CreatorID = &suite.otherID
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	target := issue_repo.ReactionTarget{IssueID: &suite.issue.ID}
	suite.MockIssueRepo.EXPECT().AddReaction(mock.Anything, suite.userID, "🎉", target).Return(true, nil).Once()
	suite.MockIssueRepo.EXPECT().Reactions(mock.Anything, target).Return([]*ent.Reaction{{Emoji: "🎉", UserID: suite.userID}, {Emoji: "🎉", UserID: suite.otherID}}, nil)
	groups, err := suite.service.AddReaction(suite.Ctx, suite.userID, &ReactionInput{Emoji: "🎉", IssueID: &suite.issue.ID})
	suite.Require().NoError(err)
	suite.Require().Len(groups, 1)
	suite.Equal(2, groups[0].Count)
	suite.True(groups[0].Reacted)
	notified := suite.eventsOf(schema.NotificationReaction)
	suite.Require().Len(notified, 1, "the issue author hears about a new reaction")
	suite.Equal([]uuid.UUID{suite.otherID}, notified[0].RecipientIDs)
	suite.Equal(suite.issue.ID, *notified[0].IssueID)

	suite.MockIssueRepo.EXPECT().AddReaction(mock.Anything, suite.userID, "🎉", target).Return(false, nil).Once()
	_, err = suite.service.AddReaction(suite.Ctx, suite.userID, &ReactionInput{Emoji: "🎉", IssueID: &suite.issue.ID})
	suite.Require().NoError(err)
	suite.Len(suite.eventsOf(schema.NotificationReaction), 1, "reacting again with the same emoji notifies nobody")
}

func (suite *IssueServiceSuite) TestCommentAndProjectUpdateReactionsNotifyTheirAuthors() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockIssueRepo.EXPECT().Reactions(mock.Anything, mock.Anything).Return([]*ent.Reaction{}, nil)

	comment := &ent.Comment{ID: uuid.New(), IssueID: suite.issue.ID, AuthorID: &suite.otherID}
	suite.MockIssueRepo.EXPECT().IssueOfComment(mock.Anything, comment.ID).Return(suite.issue.ID, nil)
	suite.MockCommentRepo.EXPECT().GetByID(mock.Anything, comment.ID).Return(comment, nil)
	suite.MockIssueRepo.EXPECT().AddReaction(mock.Anything, suite.userID, "👍", issue_repo.ReactionTarget{CommentID: &comment.ID}).Return(true, nil).Once()
	_, err := suite.service.AddReaction(suite.Ctx, suite.userID, &ReactionInput{Emoji: "👍", CommentID: &comment.ID})
	suite.Require().NoError(err)

	project := &ent.Project{ID: uuid.New(), WorkspaceID: suite.workspaceID, Name: "Launch"}
	update := &ent.ProjectPost{ID: uuid.New(), ProjectID: project.ID, AuthorID: &suite.userID}
	suite.MockIssueRepo.EXPECT().WorkspaceOfProjectUpdate(mock.Anything, update.ID).Return(suite.workspaceID, nil)
	suite.MockProjectRepo.EXPECT().GetUpdate(mock.Anything, update.ID).Return(update, nil)
	suite.MockProjectRepo.EXPECT().GetByID(mock.Anything, project.ID).Return(project, nil)
	suite.MockIssueRepo.EXPECT().AddReaction(mock.Anything, suite.userID, "👍", issue_repo.ReactionTarget{ProjectUpdateID: &update.ID}).Return(true, nil).Once()
	_, err = suite.service.AddReaction(suite.Ctx, suite.userID, &ReactionInput{Emoji: "👍", ProjectUpdateID: &update.ID})
	suite.Require().NoError(err)

	notified := suite.eventsOf(schema.NotificationReaction)
	suite.Require().Len(notified, 2)
	suite.Equal([]uuid.UUID{suite.otherID}, notified[0].RecipientIDs)
	suite.Equal(comment.ID, *notified[0].CommentID)
	suite.Equal(update.ID, *notified[1].ProjectUpdateID)
	suite.Equal(suite.userID, *notified[1].ActorID, "Notify drops the actor, so reacting to your own update notifies nobody")
}

func (suite *IssueServiceSuite) TestHistoryAndSearchCheckAccess() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleViewer)
	suite.MockIssueRepo.EXPECT().ListHistory(mock.Anything, suite.issue.ID, (*repository.TimeCursor)(nil), 3).Return([]*ent.IssueHistory{{ID: uuid.New()}, {ID: uuid.New()}, {ID: uuid.New()}}, nil)
	entries, next, err := suite.service.ListHistory(suite.Ctx, suite.userID, &ListIssueHistoryInput{IssueID: suite.issue.ID, PaginationInput: models.PaginationInput{Limit: 2}})
	suite.Require().NoError(err)
	suite.Len(entries, 2)
	suite.NotNil(next)

	suite.expectRespond()
	suite.MockIssueRepo.EXPECT().FindByIdentifier(mock.Anything, []uuid.UUID{suite.workspaceID}, "ENG", 7).Return([]uuid.UUID{suite.issue.ID}, nil)
	suite.MockIssueRepo.EXPECT().Search(mock.Anything, suite.workspaceID, []uuid.UUID{suite.team.ID}, "ENG-7", 20).Return([]uuid.UUID{suite.issue.ID}, nil)
	suite.MockIssueRepo.EXPECT().GetManyWithEdges(mock.Anything, nil, []uuid.UUID{suite.issue.ID}).Return([]*ent.Issue{suite.issue}, nil)
	found, err := suite.service.Search(suite.Ctx, suite.userID, &SearchIssuesInput{WorkspaceID: suite.workspaceID, Query: "ENG-7", Limit: 20})
	suite.Require().NoError(err)
	suite.Len(found, 1, "an identifier hit and a text hit for the same issue show once")

	other := uuid.New()
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: other}}).Return(errdefs.ErrUnauthorized)
	_, err = suite.service.Search(suite.Ctx, suite.userID, &SearchIssuesInput{WorkspaceID: other, Query: "x"})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *IssueServiceSuite) TestAutomationClosesStaleAndArchivesDone() {
	months := 3
	suite.team.AutoCloseMonths = &months
	suite.team.AutoArchiveMonths = 6
	suite.team.Edges.States = []*ent.WorkflowState{suite.states["Todo"], suite.states["Duplicate"], suite.states["Canceled"]}
	suite.MockIssueRepo.EXPECT().TeamsForAutomation(mock.Anything).Return([]*ent.Team{suite.team}, nil)
	suite.MockIssueRepo.EXPECT().StaleOpenIssues(mock.Anything, suite.team.ID, suite.now.AddDate(0, -3, 0)).Return([]*ent.Issue{suite.issue}, nil).Once()
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.StateID == suite.states["Canceled"].ID && fields.AutoClosedAt != nil && fields.CanceledAt != nil
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, (*uuid.UUID)(nil), schema.HistorySourceSystem, mock.MatchedBy(suite.changes("state"))).Return(nil).Once()
	suite.NoError(suite.service.AutoCloseStaleIssues(suite.Ctx))

	done := suite.newIssue(9, suite.states["Done"])
	suite.MockIssueRepo.EXPECT().ClosedIssuesBefore(mock.Anything, suite.team.ID, suite.now.AddDate(0, -6, 0)).Return([]*ent.Issue{done}, nil).Once()
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, done.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.ArchivedAt != nil && fields.AutoArchivedAt != nil
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, done.ID, (*uuid.UUID)(nil), schema.HistorySourceSystem, mock.MatchedBy(suite.changes("archived"))).Return(nil).Once()
	suite.NoError(suite.service.AutoArchiveDoneIssues(suite.Ctx))
}

func (suite *IssueServiceSuite) TestPurgeDeletesIssuesTrashedAMonthAgo() {
	old := suite.newIssue(9, suite.states["Todo"])
	suite.MockIssueRepo.EXPECT().TrashedBefore(mock.Anything, suite.now.Add(-30*24*time.Hour)).Return([]uuid.UUID{old.ID}, nil).Once()
	suite.MockIssueRepo.EXPECT().GetWithEdges(mock.Anything, mock.Anything, old.ID).Return(old, nil).Once()
	suite.MockIssueRepo.EXPECT().ChildIDs(mock.Anything, mock.Anything, old.ID).Return([]uuid.UUID{}, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistoryMany(mock.Anything, mock.Anything, []issue_repo.HistoryEntry{}).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().Delete(mock.Anything, mock.Anything, old.ID).Return(nil).Once()
	suite.NoError(suite.service.PurgeTrashed(suite.Ctx))
	suite.Equal([]string{"issue.deleted"}, suite.emitted, "purging announces the deletion like delete-issue")
}

func TestIssueServiceSuite(t *testing.T) {
	suite.Run(t, new(IssueServiceSuite))
}

func (suite *IssueServiceSuite) TestIDsFromAnotherWorkspaceAreRefused() {
	otherWorkspace := uuid.New()
	suite.roles(schema.WorkspaceRoleMember)
	for _, role := range []schema.WorkspaceRole{schema.WorkspaceRoleViewer, schema.WorkspaceRoleMember} {
		suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: otherWorkspace}}).Return(errdefs.ErrUnauthorized).Maybe()
	}
	foreignTeam := &ent.Team{ID: uuid.New(), WorkspaceID: otherWorkspace, Key: "OPS"}
	foreign := &ent.Issue{ID: uuid.New(), WorkspaceID: otherWorkspace, TeamID: foreignTeam.ID, Number: 1, Edges: ent.IssueEdges{Team: foreignTeam}}
	suite.loads(suite.issue, foreign)
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil).Maybe()
	foreignTemplate := &ent.Template{ID: uuid.New(), WorkspaceID: otherWorkspace, Kind: schema.TemplateKindIssue}
	suite.MockTemplateRepo.EXPECT().GetByID(mock.Anything, foreignTemplate.ID).Return(foreignTemplate, nil)
	foreignLabel := &ent.Label{ID: uuid.New(), WorkspaceID: otherWorkspace, Name: "Bug"}
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{foreignLabel.ID}).Return([]*ent.Label{foreignLabel}, nil)

	cases := map[string]func() error{
		"relation to an issue of another workspace": func() error {
			_, err := suite.service.CreateRelation(suite.Ctx, suite.userID, &CreateRelationInput{IssueID: suite.issue.ID, RelatedIssueID: foreign.ID, Type: schema.IssueRelationRelated})
			return err
		},
		"template of another workspace": func() error {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateIssueInput{TeamID: suite.team.ID, Title: "x", TemplateID: &foreignTemplate.ID})
			return err
		},
		"label of another workspace": func() error {
			_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, AddLabelIDs: []uuid.UUID{foreignLabel.ID}})
			return err
		},
		"reading an issue of another workspace": func() error {
			_, err := suite.service.Get(suite.Ctx, suite.userID, &GetIssueInput{IssueID: foreign.ID})
			return err
		},
		"updating an issue of another workspace": func() error {
			_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: foreign.ID, Title: new("x")})
			return err
		},
	}
	for name, run := range cases {
		err := run()
		suite.True(errors.Is(err, errdefs.ErrInvalidInput) || errors.Is(err, errdefs.ErrNotFound), "%s: %v", name, err)
	}
}
