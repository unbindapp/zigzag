package cycle_repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

type CycleRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo   *CycleRepository
	team   *ent.Team
	states map[schema.StateType]*ent.WorkflowState
	start  time.Time
}

func (suite *CycleRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewCycleRepository(suite.DB)
	workspace := suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.team = suite.DB.Team.Create().SetWorkspaceID(workspace.ID).SetName("Eng").SetKey("ENG").SetCyclesEnabled(true).SaveX(suite.Ctx)
	suite.states = map[schema.StateType]*ent.WorkflowState{}
	for _, stateType := range []schema.StateType{schema.StateTypeUnstarted, schema.StateTypeStarted, schema.StateTypeCompleted, schema.StateTypeCanceled} {
		suite.states[stateType] = suite.DB.WorkflowState.Create().SetTeamID(suite.team.ID).SetName(string(stateType)).SetType(stateType).SetColor("#ffffff").SaveX(suite.Ctx)
	}
	suite.start = time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
}

func (suite *CycleRepositorySuite) cycle(number int, startsAt time.Time) *ent.Cycle {
	created, err := suite.repo.Create(suite.Ctx, nil, &CreateCycleInput{TeamID: suite.team.ID, Number: number, StartsAt: startsAt, EndsAt: startsAt.AddDate(0, 0, 14)})
	suite.Require().NoError(err)
	return created
}

func (suite *CycleRepositorySuite) issue(number int, cycleID uuid.UUID, stateType schema.StateType, estimate *float64) *ent.Issue {
	return suite.DB.Issue.Create().
		SetWorkspaceID(suite.team.WorkspaceID).SetTeamID(suite.team.ID).SetNumber(number).SetTitle("Issue").
		SetStateID(suite.states[stateType].ID).SetCycleID(cycleID).SetNillableEstimate(estimate).
		SaveX(suite.Ctx)
}

func (suite *CycleRepositorySuite) TestTallySumsByStateType() {
	current := suite.cycle(1, suite.start)
	three := 3.0
	suite.issue(1, current.ID, schema.StateTypeCompleted, &three)
	suite.issue(2, current.ID, schema.StateTypeStarted, nil)
	suite.issue(3, current.ID, schema.StateTypeStarted, &three)
	trashed := suite.issue(4, current.ID, schema.StateTypeUnstarted, nil)
	suite.DB.Issue.UpdateOne(trashed).SetTrashedAt(suite.start).ExecX(suite.Ctx)

	tallies, err := suite.repo.Tally(suite.Ctx, nil, []uuid.UUID{current.ID})
	suite.Require().NoError(err)
	byType := tallies[current.ID]
	suite.Equal(StateTally{Count: 1, Points: 3}, byType[schema.StateTypeCompleted])
	suite.Equal(StateTally{Count: 2, Points: 4}, byType[schema.StateTypeStarted])
	suite.Zero(byType[schema.StateTypeUnstarted].Count, "trashed issues do not count")
}

func (suite *CycleRepositorySuite) TestOverlapsIgnoresTouchingCycles() {
	current := suite.cycle(1, suite.start)
	touching, err := suite.repo.Overlaps(suite.Ctx, nil, suite.team.ID, current.EndsAt, current.EndsAt.AddDate(0, 0, 14), nil)
	suite.Require().NoError(err)
	suite.False(touching)
	inside, err := suite.repo.Overlaps(suite.Ctx, nil, suite.team.ID, suite.start.AddDate(0, 0, 3), suite.start.AddDate(0, 0, 20), nil)
	suite.Require().NoError(err)
	suite.True(inside)
	self, err := suite.repo.Overlaps(suite.Ctx, nil, suite.team.ID, suite.start, current.EndsAt, &current.ID)
	suite.Require().NoError(err)
	suite.False(self)
}

func (suite *CycleRepositorySuite) TestUnfinishedIssuesMoveToTheNextCycle() {
	current := suite.cycle(1, suite.start)
	next := suite.cycle(2, current.EndsAt)
	open := suite.issue(1, current.ID, schema.StateTypeStarted, nil)
	suite.issue(2, current.ID, schema.StateTypeCompleted, nil)
	suite.issue(3, current.ID, schema.StateTypeCanceled, nil)

	unfinished, err := suite.repo.UnfinishedIssueIDs(suite.Ctx, nil, current.ID)
	suite.Require().NoError(err)
	suite.Equal([]uuid.UUID{open.ID}, unfinished)

	found, err := suite.repo.NextOpenCycle(suite.Ctx, nil, suite.team.ID, current.EndsAt)
	suite.Require().NoError(err)
	suite.Equal(next.ID, found.ID)

	suite.Require().NoError(suite.repo.MoveIssues(suite.Ctx, nil, unfinished, next.ID))
	suite.Equal(next.ID, *suite.DB.Issue.GetX(suite.Ctx, open.ID).CycleID)

	suite.Require().NoError(suite.repo.RecordHistory(suite.Ctx, nil, []HistoryEntry{{IssueID: open.ID, Source: schema.HistorySourceSystem, Changes: []schema.FieldChange{{Field: "cycle"}}}}))
	suite.Equal(1, suite.DB.IssueHistory.Query().CountX(suite.Ctx))
}

func (suite *CycleRepositorySuite) TestScopeSnapshotsUpsertPerDay() {
	current := suite.cycle(1, suite.start)
	day := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	suite.Require().NoError(suite.repo.UpsertScopeSnapshot(suite.Ctx, nil, ScopeSnapshot{CycleID: current.ID, Date: day, Scope: 3}))
	suite.Require().NoError(suite.repo.UpsertScopeSnapshot(suite.Ctx, nil, ScopeSnapshot{CycleID: current.ID, Date: day, Scope: 5, Completed: 1}))
	suite.Require().NoError(suite.repo.UpsertScopeSnapshot(suite.Ctx, nil, ScopeSnapshot{CycleID: current.ID, Date: day.AddDate(0, 0, 1), Scope: 6}))

	history, err := suite.repo.ScopeHistory(suite.Ctx, current.ID)
	suite.Require().NoError(err)
	suite.Require().Len(history, 2)
	suite.Equal(5.0, history[0].Scope)
	suite.Equal(1.0, history[0].Completed)
}

func (suite *CycleRepositorySuite) TestDeleteTakesIssuesOutOfTheCycle() {
	current := suite.cycle(1, suite.start)
	member := suite.issue(1, current.ID, schema.StateTypeStarted, nil)
	user := suite.DB.User.Create().SetEmail("ada@example.com").SetDisplayName("ada").SaveX(suite.Ctx)
	suite.DB.Favorite.Create().SetUserID(user.ID).SetWorkspaceID(suite.team.WorkspaceID).SetKind(schema.FavoriteKindCycle).SetEntityID(current.ID).ExecX(suite.Ctx)

	suite.Require().NoError(suite.repo.Delete(suite.Ctx, nil, current.ID))
	suite.Nil(suite.DB.Issue.GetX(suite.Ctx, member.ID).CycleID)
	suite.Zero(suite.DB.Favorite.Query().CountX(suite.Ctx))
}

func (suite *CycleRepositorySuite) TestCountEndingAfterSkipsFinishedCycles() {
	past := suite.cycle(1, suite.start.AddDate(0, 0, -28))
	suite.cycle(2, suite.start)
	suite.cycle(3, suite.start.AddDate(0, 0, 14))
	suite.Require().NoError(suite.repo.Complete(suite.Ctx, nil, past.ID, suite.start))

	count, err := suite.repo.CountEndingAfter(suite.Ctx, suite.team.ID, suite.start.AddDate(0, 0, 1))
	suite.Require().NoError(err)
	suite.Equal(2, count)
	latest, err := suite.repo.LatestCycle(suite.Ctx, suite.team.ID)
	suite.Require().NoError(err)
	suite.Equal(3, latest.Number)
}

func TestCycleRepositorySuite(t *testing.T) {
	suite.Run(t, new(CycleRepositorySuite))
}
