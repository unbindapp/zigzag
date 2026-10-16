package project_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type ProjectRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo      *ProjectRepository
	user      *ent.User
	workspace *ent.Workspace
	team      *ent.Team
	states    map[schema.StateType]*ent.WorkflowState
	statuses  map[schema.ProjectStatusType]*ent.ProjectStatus
	number    int
}

func TestProjectRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProjectRepositorySuite))
}

func (suite *ProjectRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewProjectRepository(suite.DB)
	suite.number = 0
	suite.user = suite.DB.User.Create().SetEmail("ada@example.com").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.team = suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Engineering").SetKey("ENG").SaveX(suite.Ctx)
	suite.states = map[schema.StateType]*ent.WorkflowState{}
	for _, stateType := range []schema.StateType{schema.StateTypeBacklog, schema.StateTypeStarted, schema.StateTypeCompleted, schema.StateTypeCanceled} {
		suite.states[stateType] = suite.DB.WorkflowState.Create().SetTeamID(suite.team.ID).SetName(string(stateType)).SetType(stateType).SetColor("#ffffff").SaveX(suite.Ctx)
	}
	suite.statuses = map[schema.ProjectStatusType]*ent.ProjectStatus{}
	for i, statusType := range []schema.ProjectStatusType{schema.ProjectStatusCanceled, schema.ProjectStatusStarted, schema.ProjectStatusBacklog} {
		suite.statuses[statusType] = suite.DB.ProjectStatus.Create().SetWorkspaceID(suite.workspace.ID).SetName(string(statusType)).SetType(statusType).SetColor("#ffffff").SetPosition(float64(i)).SaveX(suite.Ctx)
	}
}

func (suite *ProjectRepositorySuite) project(name string, status schema.ProjectStatusType) *ent.Project {
	created, err := suite.repo.Create(suite.Ctx, nil, &CreateProjectInput{
		WorkspaceID: suite.workspace.ID, Name: name, Slug: name, StatusID: suite.statuses[status].ID,
		TeamIDs: []uuid.UUID{suite.team.ID}, MemberIDs: []uuid.UUID{suite.user.ID},
	})
	suite.Require().NoError(err)
	return created
}

func (suite *ProjectRepositorySuite) issue(project *ent.Project, milestone *ent.ProjectMilestone, stateType schema.StateType, estimate *float64, trashed bool) {
	suite.number++
	create := suite.DB.Issue.Create().
		SetWorkspaceID(suite.workspace.ID).SetTeamID(suite.team.ID).SetNumber(suite.number).SetTitle("work").
		SetStateID(suite.states[stateType].ID).SetProjectID(project.ID).SetNillableEstimate(estimate)
	if milestone != nil {
		create.SetMilestoneID(milestone.ID)
	}
	if trashed {
		create.SetTrashedAt(time.Now())
	}
	create.SaveX(suite.Ctx)
}

func (suite *ProjectRepositorySuite) TestIssueTotalsCountAndWeighByEstimate() {
	project := suite.project("launch", schema.ProjectStatusStarted)
	other := suite.project("other", schema.ProjectStatusBacklog)
	three := 3.0
	zero := 0.0
	suite.issue(project, nil, schema.StateTypeCompleted, &three, false)
	suite.issue(project, nil, schema.StateTypeCompleted, nil, false)
	suite.issue(project, nil, schema.StateTypeStarted, &zero, false)
	suite.issue(project, nil, schema.StateTypeBacklog, &three, true)
	suite.issue(other, nil, schema.StateTypeCanceled, nil, false)

	totals, err := suite.repo.IssueTotals(suite.Ctx, []uuid.UUID{project.ID, other.ID})
	suite.Require().NoError(err)
	suite.Equal(2, totals[project.ID].Counts[schema.StateTypeCompleted])
	suite.Equal(4.0, totals[project.ID].Scopes[schema.StateTypeCompleted], "an unestimated issue weighs 1")
	suite.Equal(1.0, totals[project.ID].Scopes[schema.StateTypeStarted], "a zero estimate weighs 1")
	suite.Zero(totals[project.ID].Counts[schema.StateTypeBacklog], "trashed issues are left out")
	suite.Equal(1, totals[other.ID].Counts[schema.StateTypeCanceled])

	empty, err := suite.repo.IssueTotals(suite.Ctx, nil)
	suite.NoError(err)
	suite.Empty(empty)
}

func (suite *ProjectRepositorySuite) TestMilestoneTotalsOnlyCountTheirIssues() {
	project := suite.project("launch", schema.ProjectStatusStarted)
	milestone, err := suite.repo.CreateMilestone(suite.Ctx, &CreateMilestoneInput{ProjectID: project.ID, Name: "Beta"})
	suite.Require().NoError(err)
	suite.issue(project, milestone, schema.StateTypeCompleted, nil, false)
	suite.issue(project, nil, schema.StateTypeCompleted, nil, false)

	totals, err := suite.repo.MilestoneIssueTotals(suite.Ctx, []uuid.UUID{milestone.ID})
	suite.Require().NoError(err)
	suite.Equal(1, totals[milestone.ID].Counts[schema.StateTypeCompleted])
}

func (suite *ProjectRepositorySuite) TestStatusesOrderByTypeThenPosition() {
	statuses, err := suite.repo.ListStatuses(suite.Ctx, suite.workspace.ID)
	suite.Require().NoError(err)
	types := []schema.ProjectStatusType{}
	for _, status := range statuses {
		types = append(types, status.Type)
	}
	suite.Equal([]schema.ProjectStatusType{schema.ProjectStatusBacklog, schema.ProjectStatusStarted, schema.ProjectStatusCanceled}, types)

	fallback, err := suite.repo.DefaultStatus(suite.Ctx, suite.workspace.ID)
	suite.Require().NoError(err)
	suite.Equal(suite.statuses[schema.ProjectStatusBacklog].ID, fallback.ID)
}

func (suite *ProjectRepositorySuite) TestDeleteStatusMovesItsProjects() {
	project := suite.project("launch", schema.ProjectStatusCanceled)
	err := suite.repo.DeleteStatus(suite.Ctx, nil, suite.statuses[schema.ProjectStatusCanceled].ID, suite.statuses[schema.ProjectStatusBacklog].ID)
	suite.Require().NoError(err)
	moved, err := suite.repo.GetByID(suite.Ctx, project.ID)
	suite.Require().NoError(err)
	suite.Equal(suite.statuses[schema.ProjectStatusBacklog].ID, moved.StatusID)
}

func (suite *ProjectRepositorySuite) TestListFiltersAndHidesArchived() {
	launch := suite.project("launch", schema.ProjectStatusStarted)
	archived := suite.project("archived", schema.ProjectStatusStarted)
	suite.project("idea", schema.ProjectStatusBacklog)
	now := time.Now()
	suite.Require().NoError(suite.repo.SetArchivedAt(suite.Ctx, archived.ID, &now))

	started := schema.ProjectStatusStarted
	found, err := suite.repo.List(suite.Ctx, &ListProjectsFilter{WorkspaceID: suite.workspace.ID, StatusType: &started, TeamID: &suite.team.ID})
	suite.Require().NoError(err)
	suite.Require().Len(found, 1)
	suite.Equal(launch.ID, found[0].ID)
	suite.NotNil(found[0].Edges.Status)
	suite.Len(found[0].Edges.Teams, 1)
	suite.Len(found[0].Edges.Members, 1)

	found, err = suite.repo.List(suite.Ctx, &ListProjectsFilter{WorkspaceID: suite.workspace.ID, IncludeArchived: true, Query: "ARCH"})
	suite.Require().NoError(err)
	suite.Require().Len(found, 1)
	suite.Equal(archived.ID, found[0].ID)
}

func (suite *ProjectRepositorySuite) TestLatestHealthFollowsTheNewestUpdate() {
	project := suite.project("launch", schema.ProjectStatusStarted)
	health, err := suite.repo.LatestHealth(suite.Ctx, nil, project.ID)
	suite.Require().NoError(err)
	suite.Nil(health)

	suite.DB.ProjectPost.Create().SetProjectID(project.ID).SetBody("a").SetHealth(schema.HealthOnTrack).SetCreatedAt(time.Now().Add(-time.Hour)).SaveX(suite.Ctx)
	_, err = suite.repo.CreateUpdate(suite.Ctx, nil, &CreateUpdateInput{ProjectID: project.ID, AuthorID: suite.user.ID, Body: "b", Health: schema.HealthOffTrack})
	suite.Require().NoError(err)

	health, err = suite.repo.LatestHealth(suite.Ctx, nil, project.ID)
	suite.Require().NoError(err)
	suite.Equal(schema.HealthOffTrack, *health)
	audience, err := suite.repo.AudienceIDs(suite.Ctx, project.ID)
	suite.Require().NoError(err)
	suite.Equal([]uuid.UUID{suite.user.ID}, audience)
}

func (suite *ProjectRepositorySuite) TestValidationCountsStayInTheWorkspace() {
	otherWorkspace := suite.DB.Workspace.Create().SetName("Other").SetSlug("other").SaveX(suite.Ctx)
	otherTeam := suite.DB.Team.Create().SetWorkspaceID(otherWorkspace.ID).SetName("Ops").SetKey("OPS").SaveX(suite.Ctx)
	count, err := suite.repo.CountTeamsInWorkspace(suite.Ctx, suite.workspace.ID, []uuid.UUID{suite.team.ID, otherTeam.ID})
	suite.Require().NoError(err)
	suite.Equal(1, count)

	workspaceLabel := suite.DB.Label.Create().SetWorkspaceID(suite.workspace.ID).SetName("Bug").SetColor("#ffffff").SaveX(suite.Ctx)
	teamLabel := suite.DB.Label.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(suite.team.ID).SetName("Team").SetColor("#ffffff").SaveX(suite.Ctx)
	group := suite.DB.Label.Create().SetWorkspaceID(suite.workspace.ID).SetName("Area").SetColor("#ffffff").SetIsGroup(true).SaveX(suite.Ctx)
	count, err = suite.repo.CountApplicableLabels(suite.Ctx, suite.workspace.ID, []uuid.UUID{workspaceLabel.ID, teamLabel.ID, group.ID})
	suite.Require().NoError(err)
	suite.Equal(1, count)
}

func (suite *ProjectRepositorySuite) TestIssuesLosingTheProjectOrMilestone() {
	project := suite.project("launch", schema.ProjectStatusStarted)
	beta, err := suite.repo.CreateMilestone(suite.Ctx, &CreateMilestoneInput{ProjectID: project.ID, Name: "Beta"})
	suite.Require().NoError(err)
	suite.issue(project, beta, schema.StateTypeStarted, nil, false)
	suite.issue(project, nil, schema.StateTypeStarted, nil, false)

	issues, err := suite.repo.IssuesInProject(suite.Ctx, nil, project.ID)
	suite.Require().NoError(err)
	suite.Require().Len(issues, 2)
	names := []string{}
	for _, issue := range issues {
		if issue.Edges.Milestone != nil {
			names = append(names, issue.Edges.Milestone.Name)
		}
	}
	suite.Equal([]string{"Beta"}, names, "the milestone comes along for the history label")

	inMilestone, err := suite.repo.IssueIDsInMilestone(suite.Ctx, nil, beta.ID)
	suite.Require().NoError(err)
	suite.Len(inMilestone, 1)

	suite.Require().NoError(suite.repo.Delete(suite.Ctx, nil, project.ID))
	suite.Equal(2, suite.DB.Issue.Query().CountX(suite.Ctx), "issues survive their project")
}
