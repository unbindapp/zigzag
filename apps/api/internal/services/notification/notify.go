package notification_service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	notification_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/notification"
)

const (
	titleMaxRunes = 255
	bodyMaxRunes  = 280
)

func (self *NotificationService) Notify(ctx context.Context, event models.NotificationEvent) error {
	recipients := distinctRecipients(event.RecipientIDs, event.ActorID)
	if len(recipients) == 0 {
		return nil
	}
	members, err := self.repo.Notification().ActiveMembers(ctx, event.WorkspaceID, recipients, event.IssueID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	settings, err := self.repo.Notification().Settings(ctx, members)
	if err != nil {
		return err
	}

	inputs := make([]*notification_repo.CreateNotificationInput, 0, len(members))
	for _, userID := range members {
		if !wants(settings[userID], event.Type) {
			continue
		}
		inputs = append(inputs, &notification_repo.CreateNotificationInput{
			UserID:          userID,
			ActorID:         event.ActorID,
			WorkspaceID:     event.WorkspaceID,
			Type:            event.Type,
			IssueID:         event.IssueID,
			CommentID:       event.CommentID,
			ProjectID:       event.ProjectID,
			ProjectUpdateID: event.ProjectUpdateID,
			Title:           truncate(event.Title, titleMaxRunes),
			Body:            truncate(event.Body, bodyMaxRunes),
		})
	}
	if len(inputs) == 0 {
		return nil
	}
	_, err = self.repo.Notification().CreateMany(ctx, nil, inputs)
	return err
}

func distinctRecipients(ids []uuid.UUID, actorID *uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(ids))
	recipients := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil || seen[id] || (actorID != nil && *actorID == id) {
			continue
		}
		seen[id] = true
		recipients = append(recipients, id)
	}
	return recipients
}

func wants(settings *ent.NotificationSetting, kind schema.NotificationType) bool {
	if settings == nil {
		return true
	}
	switch kind {
	case schema.NotificationIssueAssigned, schema.NotificationIssueUnassigned:
		return settings.Assigned
	case schema.NotificationIssueMentioned, schema.NotificationCommentMentioned:
		return settings.Mentioned
	case schema.NotificationIssueCommented:
		return settings.Commented
	case schema.NotificationIssueStateChanged, schema.NotificationIssuePriorityChanged, schema.NotificationIssueBlocking:
		return settings.StateChanged
	case schema.NotificationIssueDueSoon:
		return settings.DueSoon
	case schema.NotificationProjectUpdate:
		return settings.ProjectUpdates
	default:
		return true
	}
}

func truncate(value string, maxRunes int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
}
