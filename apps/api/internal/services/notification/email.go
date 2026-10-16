package notification_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/mail"
	"github.com/unbindapp/zigzag-api/internal/models"
)

const (
	emailDelay        = 2 * time.Minute
	emailMaxAge       = 24 * time.Hour
	emailMaxListItems = 20
)

func (self *NotificationService) SendEmailBatch(ctx context.Context) error {
	configured, err := self.email.Configured(ctx)
	if err != nil || !configured {
		return err
	}
	now := self.now()
	pending, err := self.repo.Notification().PendingEmail(ctx, now.Add(-emailDelay), now.Add(-emailMaxAge))
	if err != nil {
		return err
	}
	byUser, order := groupByUser(pending)
	if len(order) == 0 {
		return nil
	}
	settings, err := self.repo.Notification().Settings(ctx, order)
	if err != nil {
		return err
	}

	var errs []error
	for _, userID := range order {
		batch := byUser[userID]
		recipient := batch[0].Edges.User
		if recipient != nil && recipient.Active && emailEnabled(settings[userID]) {
			if _, err := self.email.Send(ctx, recipient.Email, self.digest(batch)); err != nil {
				log.Warn("Failed to send notification email", "user", userID, "err", err)
				errs = append(errs, err)
				continue
			}
		}
		if err := self.repo.Notification().MarkEmailed(ctx, notificationIDs(batch), now); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func groupByUser(notifications []*ent.Notification) (map[uuid.UUID][]*ent.Notification, []uuid.UUID) {
	byUser := map[uuid.UUID][]*ent.Notification{}
	order := []uuid.UUID{}
	for _, row := range notifications {
		if _, seen := byUser[row.UserID]; !seen {
			order = append(order, row.UserID)
		}
		byUser[row.UserID] = append(byUser[row.UserID], row)
	}
	return byUser, order
}

func emailEnabled(settings *ent.NotificationSetting) bool {
	return settings == nil || settings.EmailEnabled
}

func notificationIDs(notifications []*ent.Notification) []uuid.UUID {
	ids := make([]uuid.UUID, len(notifications))
	for i, row := range notifications {
		ids[i] = row.ID
	}
	return ids
}

func (self *NotificationService) digest(batch []*ent.Notification) mail.Content {
	count := len(batch)
	subject := fmt.Sprintf("%d new notifications", count)
	heading := fmt.Sprintf("You have %d unread notifications", count)
	if count == 1 {
		subject = digestLine(batch[0])
		heading = "You have an unread notification"
	}

	lines := make([]string, 0, min(count, emailMaxListItems)+1)
	for _, row := range batch[:min(count, emailMaxListItems)] {
		lines = append(lines, digestLine(row))
	}
	if count > emailMaxListItems {
		lines = append(lines, fmt.Sprintf("And %d more.", count-emailMaxListItems))
	}

	content := mail.Content{
		Subject:    subject,
		Heading:    heading,
		Paragraphs: lines,
		Note:       "You get these emails for unread notifications. Turn them off under Settings, Account, Notifications.",
	}
	if workspace := batch[0].Edges.Workspace; workspace != nil {
		content.ButtonLabel = "Open inbox"
		content.ButtonURL = strings.TrimRight(self.cfg.ExternalUIUrl, "/") + "/" + workspace.Slug + "/inbox"
	}
	return content
}

func digestLine(row *ent.Notification) string {
	line := row.Title
	if summary := models.TransformIssueSummary(row.Edges.Issue); summary != nil && summary.Identifier != "" && !strings.HasPrefix(line, summary.Identifier) {
		line = summary.Identifier + " " + line
	}
	if row.Body != "" {
		line += ": " + row.Body
	}
	if row.Edges.Actor != nil {
		line += " (" + row.Edges.Actor.Name + ")"
	}
	return line
}
