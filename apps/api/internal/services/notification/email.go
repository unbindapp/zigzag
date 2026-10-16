package notification_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/mail"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
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
	items := self.digestItems(batch[:min(count, emailMaxListItems)])
	content := mail.Content{
		Subject:   digestSubject(count, items),
		Preheader: lineSentence(items[0].Lines[0]),
		Heading:   digestHeading(count, sharedWorkspace(batch)),
		Items:     items,
	}
	if count > emailMaxListItems {
		content.Note = fmt.Sprintf("And %d more in your inbox.", count-emailMaxListItems)
	}
	workspace := batch[0].Edges.Workspace
	if workspace == nil {
		return content
	}
	base := self.workspaceURL(workspace)
	content.Button = mail.Button{Label: "Open your inbox", URL: base + "/inbox"}
	content.Footer.ManageURL = base + "/settings/account/notifications"
	content.Footer.ManageLabel = "Notification settings"
	return content
}

func (self *NotificationService) workspaceURL(workspace *ent.Workspace) string {
	return strings.TrimRight(self.cfg.ExternalUIUrl, "/") + "/" + workspace.Slug
}

func (self *NotificationService) digestItems(rows []*ent.Notification) []mail.Item {
	items := []mail.Item{}
	index := map[uuid.UUID]int{}
	for _, row := range rows {
		key := digestKey(row)
		position, seen := index[key]
		if !seen {
			position = len(items)
			index[key] = position
			items = append(items, self.digestItem(row))
		}
		items[position].Lines = append(items[position].Lines, digestLine(row))
	}
	return items
}

func digestKey(row *ent.Notification) uuid.UUID {
	switch {
	case row.IssueID != nil:
		return *row.IssueID
	case row.ProjectID != nil:
		return *row.ProjectID
	default:
		return row.ID
	}
}

func (self *NotificationService) digestItem(row *ent.Notification) mail.Item {
	workspace := row.Edges.Workspace
	if workspace == nil {
		return mail.Item{Title: row.Title}
	}
	if summary := models.TransformIssueSummary(row.Edges.Issue); summary != nil {
		return mail.Item{Identifier: summary.Identifier, Title: summary.Title, URL: models.IssueURL(self.cfg.ExternalUIUrl, workspace.Slug, summary.Identifier)}
	}
	if project := row.Edges.Project; project != nil {
		return mail.Item{Title: project.Name, URL: self.workspaceURL(workspace) + "/project/" + project.Slug}
	}
	return mail.Item{Title: row.Title, URL: self.workspaceURL(workspace) + "/inbox"}
}

func digestLine(row *ent.Notification) mail.Line {
	sentence := digestSentence(row)
	actor := row.Edges.Actor
	if actor == nil {
		return mail.Line{Text: upperFirst(sentence)}
	}
	return mail.ActorLine(actor.ID.String(), actorName(actor), sentence)
}

func actorName(actor *ent.User) string {
	if actor.Name != "" {
		return actor.Name
	}
	return actor.DisplayName
}

func digestSentence(row *ent.Notification) string {
	switch row.Type {
	case schema.NotificationIssueAssigned:
		return "assigned the issue to you"
	case schema.NotificationIssueUnassigned:
		return "unassigned you from the issue"
	case schema.NotificationIssueMentioned:
		return "mentioned you"
	case schema.NotificationCommentMentioned:
		return "mentioned you in a comment"
	case schema.NotificationIssueCommented:
		return "commented"
	case schema.NotificationIssueStateChanged:
		return changedTo("status", row.Body, models.StateChangedBodyPrefix)
	case schema.NotificationIssuePriorityChanged:
		return changedTo("priority", row.Body, models.PriorityChangedBodyPrefix)
	case schema.NotificationIssueBlocking:
		if blocker, ok := strings.CutPrefix(row.Body, models.BlockedByBodyPrefix); ok {
			return "marked the issue as blocked by " + blocker
		}
		return "changed what blocks the issue"
	case schema.NotificationProjectUpdate:
		return "posted a project update"
	case schema.NotificationProjectMember:
		return "added you to the project"
	case schema.NotificationProjectLead:
		return "made you the project lead"
	case schema.NotificationReaction, schema.NotificationIssueDueSoon:
		return lowerFirst(row.Body)
	default:
		return lowerFirst(row.Title)
	}
}

func changedTo(field, body, prefix string) string {
	value, ok := strings.CutPrefix(body, prefix)
	if !ok || value == "" {
		return "changed the " + field
	}
	return "changed the " + field + " to " + value
}

func digestSubject(count int, items []mail.Item) string {
	if count > 1 {
		return fmt.Sprintf("%d unread notifications", count)
	}
	item := items[0]
	return strings.TrimSpace(item.Identifier+" "+item.Title) + ": " + lineSentence(item.Lines[0])
}

func digestHeading(count int, workspace *ent.Workspace) string {
	heading := fmt.Sprintf("You have %d unread notifications", count)
	if count == 1 {
		heading = "You have an unread notification"
	}
	if workspace != nil {
		heading += " in " + workspace.Name
	}
	return heading + "."
}

func sharedWorkspace(batch []*ent.Notification) *ent.Workspace {
	first := batch[0].Edges.Workspace
	for _, row := range batch {
		if first == nil || row.WorkspaceID != first.ID {
			return nil
		}
	}
	return first
}

func lineSentence(line mail.Line) string {
	return strings.TrimSpace(line.ActorName + " " + line.Text)
}

func upperFirst(text string) string {
	if text == "" {
		return text
	}
	first, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToUpper(first)) + text[size:]
}

func lowerFirst(text string) string {
	if text == "" {
		return text
	}
	first, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToLower(first)) + text[size:]
}
