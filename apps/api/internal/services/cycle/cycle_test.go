package cycle_service

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
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	"github.com/unbindapp/zigzag-api/internal/services"
	mocks_services_webhook "github.com/unbindapp/zigzag-api/mocks/services/webhook"
)

type CycleServiceSuite struct {
	services.ServiceTestSuite
	service     *CycleService
	webhooks    *mocks_services_webhook.WebhookServiceMock
	emitted     []map[string]any
	now         time.Time
	userID      uuid.UUID
	workspaceID uuid.UUID
	team        *ent.Team
	cycle       *ent.Cycle
}

func (suite *CycleServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	suite.webhooks = mocks_services_webhook.NewWebhookServiceMock(suite.T())
	suite.emitted = nil
	suite.webhooks.EXPECT().Emit(mock.Anything, mock.Anything, "cycle.updated", mock.Anything).RunAndReturn(func(_ context.Context, _ uuid.UUID, _ string, payload any) error {
		suite.emitted = append(suite.emitted, payload.(map[string]any))
		return nil
	}).Maybe()
	suite.service = NewCycleService(nil, suite.MockRepo, suite.webhooks)
	suite.service.now = func() time.Time { return suite.now }
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
	suite.team = &ent.Team{
		ID: uuid.New(), WorkspaceID: suite.workspaceID, Key: "ENG", Timezone: "UTC", EstimateType: schema.EstimateTypeNone,
		CyclesEnabled: true, CycleDurationWeeks: 2, CycleStartDay: 1, CycleAutoMove: true,
	}
	suite.cycle = &ent.Cycle{
		ID: uuid.New(), TeamID: suite.team.ID, Number: 3,
		StartsAt: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), EndsAt: time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
	}
	suite.ExpectTx()
	suite.MockCycleRepo.EXPECT().LockTeam(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (suite *CycleServiceSuite) role(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *CycleServiceSuite) asAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Maybe()
}

func (suite *CycleServiceSuite) tally(byType map[schema.StateType]cycle_repo.StateTally) {
	suite.MockCycleRepo.EXPECT().Tally(suite.Ctx, mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, _ repository.TxInterface, ids []uuid.UUID) (map[uuid.UUID]map[schema.StateType]cycle_repo.StateTally, error) {
			out := map[uuid.UUID]map[schema.StateType]cycle_repo.StateTally{}
			for _, id := range ids {
				out[id] = byType
			}
			return out, nil
		}).Maybe()
}

func (suite *CycleServiceSuite) TestListComputesProgressWithoutCanceledIssues() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockCycleRepo.EXPECT().ListByTeam(suite.Ctx, suite.team.ID, false).Return([]*ent.Cycle{suite.cycle}, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{
		schema.StateTypeCompleted: {Count: 1, Points: 5},
		schema.StateTypeStarted:   {Count: 2, Points: 3},
		schema.StateTypeUnstarted: {Count: 1, Points: 1},
		schema.StateTypeCanceled:  {Count: 4, Points: 8},
	})

	cycles, err := suite.service.List(suite.Ctx, suite.userID, &ListCyclesInput{TeamID: suite.team.ID})
	suite.Require().NoError(err)
	suite.Require().Len(cycles, 1)
	progress := cycles[0].Progress
	suite.Equal(4.0, progress.Scope)
	suite.Equal(1.0, progress.Completed)
	suite.Equal(2.0, progress.Started)
	suite.Equal(25, progress.Percent)
	suite.Equal(4, progress.IssueCount)
	suite.Equal(1, progress.CompletedIssueCount)
	suite.True(cycles[0].IsActive)
	suite.False(cycles[0].IsPast || cycles[0].IsFuture)

	suite.team.EstimateType = schema.EstimateTypeFibonacci
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockCycleRepo.EXPECT().ListByTeam(suite.Ctx, suite.team.ID, false).Return([]*ent.Cycle{suite.cycle}, nil).Once()
	cycles, err = suite.service.List(suite.Ctx, suite.userID, &ListCyclesInput{TeamID: suite.team.ID})
	suite.Require().NoError(err)
	suite.Equal(9.0, cycles[0].Progress.Scope)
	suite.Equal(56, cycles[0].Progress.Percent, "percent is rounded like project progress")
}

func (suite *CycleServiceSuite) TestPrivateTeamCyclesAreNotFound() {
	suite.team.Private = true
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockCycleRepo.EXPECT().GetByID(suite.Ctx, suite.cycle.ID).Return(suite.cycle, nil).Once()
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockTeamRepo.EXPECT().IsMember(suite.Ctx, suite.team.ID, suite.userID).Return(false, nil).Once()
	_, err := suite.service.Get(suite.Ctx, suite.userID, &GetCycleInput{CycleID: suite.cycle.ID})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *CycleServiceSuite) TestGetByNumberAddsTodaysLivePoint() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockCycleRepo.EXPECT().GetByNumber(suite.Ctx, suite.team.ID, 3).Return(suite.cycle, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{schema.StateTypeCompleted: {Count: 2, Points: 2}})
	suite.MockCycleRepo.EXPECT().ScopeHistory(suite.Ctx, suite.cycle.ID).Return([]*ent.CycleScopeHistory{
		{Date: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), Scope: 2},
	}, nil).Once()

	cycle, err := suite.service.Get(suite.Ctx, suite.userID, &GetCycleInput{TeamID: suite.team.ID, Number: 3})
	suite.Require().NoError(err)
	suite.Require().Len(cycle.ScopeHistory, 2)
	suite.Equal("2026-10-15", cycle.ScopeHistory[0].Date)
	suite.Equal("2026-10-16", cycle.ScopeHistory[1].Date)
	suite.Equal(2.0, cycle.ScopeHistory[1].Completed)

	_, err = suite.service.Get(suite.Ctx, suite.userID, &GetCycleInput{TeamID: suite.team.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *CycleServiceSuite) TestCreateNeedsAdmin() {
	suite.role(schema.WorkspaceRoleViewer, nil)
	suite.role(schema.WorkspaceRoleAdmin, errdefs.ErrUnauthorized)
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateCycleInput{TeamID: suite.team.ID, StartsAt: suite.now, EndsAt: suite.now.Add(time.Hour)})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *CycleServiceSuite) TestCreateValidatesDates() {
	suite.asAdmin()
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateCycleInput{TeamID: suite.team.ID, StartsAt: suite.now, EndsAt: suite.now})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateCycleInput{TeamID: suite.team.ID, StartsAt: suite.now, EndsAt: suite.now.AddDate(1, 0, 0)})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	ends := suite.now.AddDate(0, 0, 14)
	suite.MockCycleRepo.EXPECT().Overlaps(suite.Ctx, mock.Anything, suite.team.ID, suite.now, ends, (*uuid.UUID)(nil)).Return(true, nil).Once()
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateCycleInput{TeamID: suite.team.ID, StartsAt: suite.now, EndsAt: ends})
	suite.ErrorIs(err, errdefs.ErrConflict)

	suite.team.CyclesEnabled = false
	_, err = suite.service.Create(suite.Ctx, suite.userID, &CreateCycleInput{TeamID: suite.team.ID, StartsAt: suite.now, EndsAt: ends})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *CycleServiceSuite) TestCreateNumbersCycles() {
	suite.asAdmin()
	ends := suite.now.AddDate(0, 0, 14)
	suite.MockCycleRepo.EXPECT().Overlaps(suite.Ctx, mock.Anything, suite.team.ID, suite.now, ends, (*uuid.UUID)(nil)).Return(false, nil).Once()
	suite.MockCycleRepo.EXPECT().NextNumber(suite.Ctx, mock.Anything, suite.team.ID).Return(4, nil).Once()
	suite.MockCycleRepo.EXPECT().Create(suite.Ctx, mock.Anything, &cycle_repo.CreateCycleInput{TeamID: suite.team.ID, Number: 4, Name: "Polish", StartsAt: suite.now, EndsAt: ends}).
		Return(&ent.Cycle{ID: uuid.New(), TeamID: suite.team.ID, Number: 4, Name: "Polish", StartsAt: suite.now, EndsAt: ends}, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{})

	cycle, err := suite.service.Create(suite.Ctx, suite.userID, &CreateCycleInput{TeamID: suite.team.ID, Name: " Polish ", StartsAt: suite.now, EndsAt: ends})
	suite.Require().NoError(err)
	suite.Equal(4, cycle.Number)
	suite.True(cycle.IsActive)
	suite.Equal([]string{"created"}, suite.emittedActions())
}

func (suite *CycleServiceSuite) TestUpdateChecksTheResultingRange() {
	suite.asAdmin()
	suite.MockCycleRepo.EXPECT().GetByID(suite.Ctx, suite.cycle.ID).Return(suite.cycle, nil)
	early := suite.cycle.StartsAt.Add(-time.Hour)
	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateCycleInput{CycleID: suite.cycle.ID, EndsAt: &early})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *CycleServiceSuite) TestDeleteRecordsTheCycleLeavingIssues() {
	suite.asAdmin()
	issueID := uuid.New()
	suite.MockCycleRepo.EXPECT().GetByID(suite.Ctx, suite.cycle.ID).Return(suite.cycle, nil).Once()
	suite.MockCycleRepo.EXPECT().IssueIDsInCycle(suite.Ctx, mock.Anything, suite.cycle.ID).Return([]uuid.UUID{issueID}, nil).Once()
	suite.MockCycleRepo.EXPECT().RecordHistory(suite.Ctx, mock.Anything, []cycle_repo.HistoryEntry{{
		IssueID: issueID, ActorID: &suite.userID, Source: schema.HistorySourceUser,
		Changes: []schema.FieldChange{{Field: "cycle", From: suite.cycle.ID.String(), FromLabel: "Cycle 3"}},
	}}).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().Delete(suite.Ctx, mock.Anything, suite.cycle.ID).Return(nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{})
	suite.NoError(suite.service.Delete(suite.Ctx, suite.userID, &CycleRefInput{CycleID: suite.cycle.ID}))
	suite.Equal([]string{"deleted"}, suite.emittedActions())
	suite.Equal(suite.team.ID, suite.emitted[0]["team_id"], "team-scoped webhooks match on team_id")
}

func (suite *CycleServiceSuite) emittedActions() []string {
	actions := make([]string, len(suite.emitted))
	for i, payload := range suite.emitted {
		actions[i] = payload["action"].(string)
	}
	return actions
}

func (suite *CycleServiceSuite) TestProcessCyclesPlansCurrentAndNextCycle() {
	suite.MockCycleRepo.EXPECT().TeamsWithCyclesEnabled(suite.Ctx).Return([]*ent.Team{suite.team}, nil).Once()
	suite.MockCycleRepo.EXPECT().CountEndingAfter(suite.Ctx, mock.Anything, suite.team.ID, suite.now).Return(0, nil).Once()
	suite.MockCycleRepo.EXPECT().LatestCycle(suite.Ctx, mock.Anything, suite.team.ID).Return(nil, nil).Once()
	suite.MockCycleRepo.EXPECT().NextNumber(suite.Ctx, mock.Anything, suite.team.ID).Return(1, nil).Once()
	suite.MockCycleRepo.EXPECT().NextNumber(suite.Ctx, mock.Anything, suite.team.ID).Return(2, nil).Once()
	monday := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	var created []*cycle_repo.CreateCycleInput
	suite.MockCycleRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, _ repository.TxInterface, input *cycle_repo.CreateCycleInput) (*ent.Cycle, error) {
			created = append(created, input)
			return &ent.Cycle{ID: uuid.New(), TeamID: input.TeamID, Number: input.Number, StartsAt: input.StartsAt, EndsAt: input.EndsAt}, nil
		}).Twice()
	suite.MockCycleRepo.EXPECT().EndedOpenCycles(suite.Ctx, suite.now).Return(nil, nil).Once()
	suite.MockCycleRepo.EXPECT().ActiveCycles(suite.Ctx, suite.now).Return(nil, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{})

	suite.Require().NoError(suite.service.ProcessCycles(suite.Ctx))
	suite.Equal([]string{"created", "created"}, suite.emittedActions())
	suite.Require().Len(created, 2)
	suite.True(created[0].StartsAt.Equal(monday))
	suite.True(created[0].EndsAt.Equal(monday.AddDate(0, 0, 14)))
	suite.True(created[1].StartsAt.Equal(monday.AddDate(0, 0, 14)))
	suite.Equal(2, created[1].Number)
}

func (suite *CycleServiceSuite) TestProcessCyclesClosesEndedCyclesAndMovesUnfinishedIssues() {
	ended := &ent.Cycle{ID: uuid.New(), TeamID: suite.team.ID, Number: 2, StartsAt: suite.now.AddDate(0, 0, -14), EndsAt: suite.now.Add(-time.Hour), Edges: ent.CycleEdges{Team: suite.team}}
	next := &ent.Cycle{ID: uuid.New(), TeamID: suite.team.ID, Number: 3, StartsAt: ended.EndsAt, EndsAt: ended.EndsAt.AddDate(0, 0, 14)}
	issueID := uuid.New()
	suite.MockCycleRepo.EXPECT().TeamsWithCyclesEnabled(suite.Ctx).Return(nil, nil).Once()
	suite.MockCycleRepo.EXPECT().EndedOpenCycles(suite.Ctx, suite.now).Return([]*ent.Cycle{ended}, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{schema.StateTypeCompleted: {Count: 3}, schema.StateTypeStarted: {Count: 1}})
	suite.MockCycleRepo.EXPECT().UpsertScopeSnapshot(suite.Ctx, mock.Anything, cycle_repo.ScopeSnapshot{
		CycleID: ended.ID, Date: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), Scope: 4, Completed: 3, Started: 1,
	}).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().Complete(suite.Ctx, mock.Anything, ended.ID, suite.now).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().NextOpenCycle(suite.Ctx, mock.Anything, suite.team.ID, ended.EndsAt).Return(next, nil).Once()
	suite.MockCycleRepo.EXPECT().UnfinishedIssueIDs(suite.Ctx, mock.Anything, ended.ID).Return([]uuid.UUID{issueID}, nil).Once()
	suite.MockCycleRepo.EXPECT().MoveIssues(suite.Ctx, mock.Anything, []uuid.UUID{issueID}, next.ID).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().RecordHistory(suite.Ctx, mock.Anything, []cycle_repo.HistoryEntry{{
		IssueID: issueID, Source: schema.HistorySourceSystem,
		Changes: []schema.FieldChange{{Field: "cycle", From: ended.ID.String(), To: next.ID.String(), FromLabel: "Cycle 2", ToLabel: "Cycle 3"}},
	}}).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().ActiveCycles(suite.Ctx, suite.now).Return(nil, nil).Once()

	suite.NoError(suite.service.ProcessCycles(suite.Ctx))
	suite.Equal([]string{"completed"}, suite.emittedActions())
}

func (suite *CycleServiceSuite) TestPlanTeamCyclesOnlyForTeamsWithCycles() {
	off := *suite.team
	off.CyclesEnabled = false
	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, off.ID).Return(&off, nil).Once()
	suite.NoError(suite.service.PlanTeamCycles(suite.Ctx, off.ID))

	suite.MockTeamRepo.EXPECT().GetByID(suite.Ctx, suite.team.ID).Return(suite.team, nil).Once()
	suite.MockCycleRepo.EXPECT().CountEndingAfter(suite.Ctx, mock.Anything, suite.team.ID, suite.now).Return(1, nil).Once()
	suite.MockCycleRepo.EXPECT().LatestCycle(suite.Ctx, mock.Anything, suite.team.ID).Return(suite.cycle, nil).Once()
	suite.MockCycleRepo.EXPECT().NextNumber(suite.Ctx, mock.Anything, suite.team.ID).Return(4, nil).Once()
	suite.MockCycleRepo.EXPECT().Create(suite.Ctx, mock.Anything, mock.MatchedBy(func(input *cycle_repo.CreateCycleInput) bool {
		return input.Number == 4 && input.StartsAt.Equal(suite.cycle.EndsAt)
	})).Return(&ent.Cycle{ID: uuid.New(), TeamID: suite.team.ID, Number: 4, StartsAt: suite.cycle.EndsAt, EndsAt: suite.cycle.EndsAt.AddDate(0, 0, 14)}, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{})
	suite.NoError(suite.service.PlanTeamCycles(suite.Ctx, suite.team.ID))
	suite.Equal([]string{"created"}, suite.emittedActions())
}

func (suite *CycleServiceSuite) TestProcessCyclesLeavesIssuesWithoutAutoMove() {
	suite.team.CycleAutoMove = false
	ended := &ent.Cycle{ID: uuid.New(), TeamID: suite.team.ID, Number: 2, EndsAt: suite.now.Add(-time.Hour), Edges: ent.CycleEdges{Team: suite.team}}
	suite.MockCycleRepo.EXPECT().TeamsWithCyclesEnabled(suite.Ctx).Return(nil, nil).Once()
	suite.MockCycleRepo.EXPECT().EndedOpenCycles(suite.Ctx, suite.now).Return([]*ent.Cycle{ended}, nil).Once()
	suite.tally(map[schema.StateType]cycle_repo.StateTally{})
	suite.MockCycleRepo.EXPECT().UpsertScopeSnapshot(suite.Ctx, mock.Anything, mock.Anything).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().Complete(suite.Ctx, mock.Anything, ended.ID, suite.now).Return(nil).Once()
	suite.MockCycleRepo.EXPECT().ActiveCycles(suite.Ctx, suite.now).Return([]*ent.Cycle{{ID: suite.cycle.ID, EndsAt: suite.cycle.EndsAt, Edges: ent.CycleEdges{Team: suite.team}}}, nil).Once()
	suite.MockCycleRepo.EXPECT().UpsertScopeSnapshot(suite.Ctx, mock.Anything, mock.MatchedBy(func(snapshot cycle_repo.ScopeSnapshot) bool {
		return snapshot.CycleID == suite.cycle.ID
	})).Return(nil).Once()

	suite.NoError(suite.service.ProcessCycles(suite.Ctx))
}

func (suite *CycleServiceSuite) TestNextStartChainsAfterCooldownAndRealignsStaleChains() {
	berlin, err := time.LoadLocation("Europe/Berlin")
	suite.Require().NoError(err)
	suite.team.CycleCooldownWeeks = 1
	latest := &ent.Cycle{EndsAt: time.Date(2026, 10, 19, 0, 0, 0, 0, berlin)}
	suite.True(nextStart(suite.team, latest, suite.now, berlin).Equal(time.Date(2026, 10, 26, 0, 0, 0, 0, berlin)))

	stale := &ent.Cycle{EndsAt: time.Date(2026, 6, 1, 0, 0, 0, 0, berlin)}
	suite.True(nextStart(suite.team, stale, suite.now, berlin).Equal(time.Date(2026, 10, 12, 0, 0, 0, 0, berlin)))

	suite.True(weekStart(time.Date(2026, 10, 18, 23, 0, 0, 0, time.UTC), berlin, 1).Equal(time.Date(2026, 10, 19, 0, 0, 0, 0, berlin)))
	suite.True(weekStart(suite.now, time.UTC, 7).Equal(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)))
}

func TestCycleServiceSuite(t *testing.T) {
	suite.Run(t, new(CycleServiceSuite))
}
