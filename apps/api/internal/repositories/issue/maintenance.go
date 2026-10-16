package issue_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/issue"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/ent/team"
	"github.com/zigzagapp-org/zigzag-api/ent/workflowstate"
)

func (self *IssueRepository) TeamsForAutomation(ctx context.Context) ([]*ent.Team, error) {
	return self.base.DB.Team.Query().
		Where(team.ArchivedAtIsNil()).
		WithStates(func(q *ent.WorkflowStateQuery) {
			q.Where(workflowstate.ArchivedAtIsNil()).Order(workflowstate.ByPosition(), workflowstate.ByName())
		}).
		All(ctx)
}

func (self *IssueRepository) StaleOpenIssues(ctx context.Context, teamID uuid.UUID, updatedBefore time.Time) ([]*ent.Issue, error) {
	return self.base.DB.Issue.Query().
		Where(
			issue.TeamID(teamID),
			issue.UpdatedAtLT(updatedBefore),
			issue.ArchivedAtIsNil(),
			issue.TrashedAtIsNil(),
			issue.HasStateWith(workflowstate.TypeIn(schema.StateTypeTriage, schema.StateTypeBacklog, schema.StateTypeUnstarted)),
		).
		WithState().
		All(ctx)
}

func (self *IssueRepository) ClosedIssuesBefore(ctx context.Context, teamID uuid.UUID, closedBefore time.Time) ([]*ent.Issue, error) {
	return self.base.DB.Issue.Query().
		Where(
			issue.TeamID(teamID),
			issue.ArchivedAtIsNil(),
			issue.TrashedAtIsNil(),
			issue.Or(
				issue.And(issue.HasStateWith(workflowstate.TypeEQ(schema.StateTypeCompleted)), issue.CompletedAtLT(closedBefore)),
				issue.And(issue.HasStateWith(workflowstate.TypeEQ(schema.StateTypeCanceled)), issue.CanceledAtLT(closedBefore)),
			),
		).
		All(ctx)
}

func (self *IssueRepository) TrashedBefore(ctx context.Context, before time.Time) ([]uuid.UUID, error) {
	return self.base.DB.Issue.Query().Where(issue.TrashedAtLT(before)).IDs(ctx)
}
