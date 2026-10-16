package notification_service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/models"
)

const (
	dueSoonWindow = 24 * time.Hour
	dueSoonRepeat = 48 * time.Hour
)

func (self *NotificationService) NotifyDueSoon(ctx context.Context) error {
	now := self.now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	issues, err := self.repo.Notification().DueBetween(ctx, today, now.Add(dueSoonWindow))
	if err != nil {
		return err
	}

	var errs []error
	for _, issue := range issues {
		if issue.AssigneeID == nil {
			continue
		}
		recent, err := self.repo.Notification().HasRecent(ctx, *issue.AssigneeID, issue.ID, schema.NotificationIssueDueSoon, now.Add(-dueSoonRepeat))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if recent {
			continue
		}
		if err := self.Notify(ctx, dueSoonEvent(issue, today)); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func dueSoonEvent(issue *ent.Issue, today time.Time) models.NotificationEvent {
	body := "Due tomorrow"
	if !issue.DueDate.UTC().After(today) {
		body = "Due today"
	}
	issueID := issue.ID
	return models.NotificationEvent{
		Type:         schema.NotificationIssueDueSoon,
		WorkspaceID:  issue.WorkspaceID,
		RecipientIDs: []uuid.UUID{*issue.AssigneeID},
		IssueID:      &issueID,
		Title:        issue.Title,
		Body:         body,
	}
}
