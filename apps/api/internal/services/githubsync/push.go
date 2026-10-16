package github_sync_service

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
)

func (self *GithubSyncService) pushIssue(ctx context.Context, issueID uuid.UUID, changes []schema.FieldChange) {
	target, err := self.issueTarget(ctx, issueID)
	if err == nil && target != nil {
		err = self.applyIssueChanges(ctx, target, issueID, changes)
	}
	self.record(ctx, target, err)
}

func (self *GithubSyncService) applyIssueChanges(ctx context.Context, target *target, issueID uuid.UUID, changes []schema.FieldChange) error {
	issue, err := self.repo.Issue().GetWithEdges(ctx, nil, issueID)
	if err != nil {
		return err
	}
	patch := &github.IssuePatch{}
	patched := false
	var failures []error
	for _, change := range changes {
		switch change.Field {
		case "title":
			patch.Title, patched = &issue.Title, true
		case "description":
			patch.Body, patched = &issue.Description, true
		case "state":
			state, reason, err := self.stateChange(ctx, change)
			if err != nil {
				return err
			}
			if state != "" {
				patch.State, patch.StateReason, patched = state, reason, true
			}
		case "labels_added":
			failures = append(failures, self.addLabels(ctx, target, issue, idsOf(change.To)))
		case "labels_removed":
			failures = append(failures, self.removeLabels(ctx, target, idsOf(change.From)))
		case "assignee":
			failures = append(failures, self.reassign(ctx, target, idsOf(change.From), idsOf(change.To)))
		}
	}
	var githubUpdatedAt *time.Time
	if patched {
		updated, err := self.client.UpdateIssue(ctx, target.token, target.repoSync.FullName, target.link.Number, patch)
		if err != nil {
			failures = append(failures, err)
		} else {
			githubUpdatedAt = &updated.UpdatedAt
		}
	}
	if err := self.repo.Github().RecordOutbound(ctx, target.link.ID, githubUpdatedAt, self.now()); err != nil {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}

func (self *GithubSyncService) stateChange(ctx context.Context, change schema.FieldChange) (string, string, error) {
	from, err := self.stateOf(ctx, change.From)
	if err != nil {
		return "", "", err
	}
	to, err := self.stateOf(ctx, change.To)
	if err != nil {
		return "", "", err
	}
	state, reason := githubState(to.Type)
	if from != nil {
		previous, previousReason := githubState(from.Type)
		if previous == state && previousReason == reason {
			return "", "", nil
		}
	}
	return state, reason, nil
}

func githubState(kind schema.StateType) (string, string) {
	switch kind {
	case schema.StateTypeCompleted:
		return github.StateClosed, github.StateReasonCompleted
	case schema.StateTypeCanceled:
		return github.StateClosed, github.StateReasonNotPlanned
	default:
		return github.StateOpen, ""
	}
}

func (self *GithubSyncService) stateOf(ctx context.Context, raw any) (*ent.WorkflowState, error) {
	ids := idsOf(raw)
	if len(ids) == 0 {
		return nil, nil
	}
	return self.repo.Team().GetState(ctx, ids[0])
}

func (self *GithubSyncService) addLabels(ctx context.Context, target *target, issue *ent.Issue, ids []uuid.UUID) error {
	names := []string{}
	for _, label := range issue.Edges.Labels {
		if !slices.Contains(ids, label.ID) {
			continue
		}
		if err := self.client.EnsureLabel(ctx, target.token, target.repoSync.FullName, label.Name, github.GithubColor(label.Color)); err != nil {
			return err
		}
		names = append(names, label.Name)
	}
	if len(names) == 0 {
		return nil
	}
	return self.client.AddLabels(ctx, target.token, target.repoSync.FullName, target.link.Number, names)
}

func (self *GithubSyncService) removeLabels(ctx context.Context, target *target, ids []uuid.UUID) error {
	labels, err := self.repo.Issue().LabelsByIDs(ctx, ids)
	if err != nil {
		return err
	}
	for _, label := range labels {
		if err := self.client.RemoveLabel(ctx, target.token, target.repoSync.FullName, target.link.Number, label.Name); err != nil {
			return err
		}
	}
	return nil
}

func (self *GithubSyncService) reassign(ctx context.Context, target *target, from, to []uuid.UUID) error {
	previous, err := self.loginsOf(ctx, from)
	if err != nil {
		return err
	}
	next, err := self.loginsOf(ctx, to)
	if err != nil {
		return err
	}
	if len(previous) > 0 {
		if err := self.client.RemoveAssignees(ctx, target.token, target.repoSync.FullName, target.link.Number, previous); err != nil {
			return err
		}
	}
	if len(next) == 0 {
		return nil
	}
	return self.client.AddAssignees(ctx, target.token, target.repoSync.FullName, target.link.Number, next)
}

func (self *GithubSyncService) loginsOf(ctx context.Context, ids []uuid.UUID) ([]string, error) {
	logins := []string{}
	for _, id := range ids {
		user, err := self.repo.User().GetByID(ctx, id)
		if ent.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if user.GithubLogin != nil {
			logins = append(logins, *user.GithubLogin)
		}
	}
	return logins, nil
}

func idsOf(raw any) []uuid.UUID {
	var values []string
	switch typed := raw.(type) {
	case string:
		values = []string{typed}
	case []string:
		values = typed
	case []any:
		for _, value := range typed {
			if text, ok := value.(string); ok {
				values = append(values, text)
			}
		}
	}
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if id, err := uuid.Parse(value); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
