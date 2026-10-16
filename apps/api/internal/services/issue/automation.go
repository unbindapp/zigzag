package issue_service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	repository "github.com/unbindapp/zigzag-api/internal/repositories"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

const trashRetention = 30 * 24 * time.Hour

func (self *IssueService) PurgeTrashed(ctx context.Context) error {
	ids, err := self.repo.Issue().TrashedBefore(ctx, self.now().Add(-trashRetention))
	if err != nil {
		return err
	}
	var failures []error
	for _, id := range ids {
		err := self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
			return self.repo.Issue().Delete(ctx, tx, id)
		})
		if err != nil && !ent.IsNotFound(err) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (self *IssueService) AutoCloseStaleIssues(ctx context.Context) error {
	teams, err := self.repo.Issue().TeamsForAutomation(ctx)
	if err != nil {
		return err
	}
	now := self.now()
	var failures []error
	for _, team := range teams {
		if team.AutoCloseMonths == nil || *team.AutoCloseMonths <= 0 {
			continue
		}
		closed := closingState(team.Edges.States)
		if closed == nil {
			continue
		}
		stale, err := self.repo.Issue().StaleOpenIssues(ctx, team.ID, now.AddDate(0, -*team.AutoCloseMonths, 0))
		if err != nil {
			failures = append(failures, err)
			continue
		}
		for _, issue := range stale {
			fields := issue_repo.FieldsOf(issue)
			fields.StateID = closed.ID
			fields.AutoClosedAt = &now
			applyStateTimestamps(&fields, issue.Edges.State, closed, now)
			change := schema.FieldChange{Field: "state", From: issue.StateID.String(), To: closed.ID.String(), FromLabel: issue.Edges.State.Name, ToLabel: closed.Name}
			failures = append(failures, self.systemUpdate(ctx, issue.ID, &fields, change))
		}
	}
	return errors.Join(failures...)
}

func (self *IssueService) AutoArchiveDoneIssues(ctx context.Context) error {
	teams, err := self.repo.Issue().TeamsForAutomation(ctx)
	if err != nil {
		return err
	}
	now := self.now()
	var failures []error
	for _, team := range teams {
		if team.AutoArchiveMonths <= 0 {
			continue
		}
		done, err := self.repo.Issue().ClosedIssuesBefore(ctx, team.ID, now.AddDate(0, -team.AutoArchiveMonths, 0))
		if err != nil {
			failures = append(failures, err)
			continue
		}
		for _, issue := range done {
			fields := issue_repo.FieldsOf(issue)
			fields.ArchivedAt, fields.AutoArchivedAt = &now, &now
			failures = append(failures, self.systemUpdate(ctx, issue.ID, &fields, schema.FieldChange{Field: "archived"}))
		}
	}
	return errors.Join(failures...)
}

func (self *IssueService) systemUpdate(ctx context.Context, issueID uuid.UUID, fields *issue_repo.IssueFields, change schema.FieldChange) error {
	return self.repo.WithTx(ctx, func(tx repository.TxInterface) error {
		if err := self.repo.Issue().Update(ctx, tx, issueID, fields); err != nil {
			return err
		}
		return self.repo.Issue().CreateHistory(ctx, tx, issueID, nil, schema.HistorySourceSystem, []schema.FieldChange{change})
	})
}

func closingState(states []*ent.WorkflowState) *ent.WorkflowState {
	var fallback *ent.WorkflowState
	for _, state := range states {
		if state.Type != schema.StateTypeCanceled {
			continue
		}
		if state.Name == "Canceled" {
			return state
		}
		if fallback == nil {
			fallback = state
		}
	}
	return fallback
}
