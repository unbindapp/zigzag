package webhook_service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
)

const (
	SignatureHeader = "X-Zigzag-Signature"
	EventHeader     = "X-Zigzag-Event"
	DeliveryHeader  = "X-Zigzag-Delivery"
)

type Envelope struct {
	ID          uuid.UUID `json:"id"`
	Event       string    `json:"event"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	Data        any       `json:"data"`
}

type delivery struct {
	id    uuid.UUID
	event string
	body  []byte
}

// Emit returns before any endpoint is called.
func (self *WebhookService) Emit(ctx context.Context, workspaceID uuid.UUID, event string, payload any) error {
	hooks, err := self.repo.Webhook().ListEnabled(ctx, workspaceID)
	if err != nil {
		return err
	}
	hooks = slices.DeleteFunc(hooks, func(hook *ent.Webhook) bool { return !slices.Contains(hook.Events, event) })
	if len(hooks) == 0 {
		return nil
	}
	sent, err := self.newDelivery(workspaceID, event, payload)
	if err != nil {
		return err
	}
	teams := self.teamsOf(ctx, sent.body)
	for _, hook := range hooks {
		if hook.TeamID != nil && !slices.Contains(teams, *hook.TeamID) {
			continue
		}
		self.async(func() { self.deliver(hook, sent) })
	}
	return nil
}

func (self *WebhookService) newDelivery(workspaceID uuid.UUID, event string, payload any) (*delivery, error) {
	id := uuid.New()
	body, err := json.Marshal(Envelope{ID: id, Event: event, WorkspaceID: workspaceID, CreatedAt: self.now().UTC(), Data: payload})
	if err != nil {
		return nil, fmt.Errorf("encoding webhook payload: %w", err)
	}
	return &delivery{id: id, event: event, body: body}, nil
}

func (self *WebhookService) deliver(hook *ent.Webhook, sent *delivery) {
	for attempt := 0; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
		_, err := self.post(ctx, hook, sent)
		cancel()
		if err == nil {
			return
		}
		if attempt >= len(self.retryDelays) {
			log.Warn("Webhook delivery failed", "webhook", hook.ID, "event", sent.event, "delivery", sent.id, "attempts", attempt+1, "err", err)
			return
		}
		time.Sleep(self.retryDelays[attempt])
	}
}

func (self *WebhookService) post(ctx context.Context, hook *ent.Webhook, sent *delivery) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(sent.body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Zigzag-Webhooks/1")
	req.Header.Set(EventHeader, sent.event)
	req.Header.Set(DeliveryHeader, sent.id.String())
	req.Header.Set(SignatureHeader, Sign(hook.Secret, sent.body))

	resp, err := self.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, fmt.Errorf("endpoint answered %s", resp.Status)
	}
	return resp.StatusCode, nil
}

func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

type teamRefs struct {
	Data struct {
		TeamID  *uuid.UUID `json:"team_id"`
		IssueID *uuid.UUID `json:"issue_id"`
		Issue   *struct {
			TeamID uuid.UUID `json:"team_id"`
		} `json:"issue"`
		Teams []struct {
			ID uuid.UUID `json:"id"`
		} `json:"teams"`
	} `json:"data"`
}

func (self *WebhookService) teamsOf(ctx context.Context, body []byte) []uuid.UUID {
	var refs teamRefs
	if err := json.Unmarshal(body, &refs); err != nil {
		return nil
	}
	if refs.Data.TeamID != nil {
		return []uuid.UUID{*refs.Data.TeamID}
	}
	if refs.Data.Issue != nil {
		return []uuid.UUID{refs.Data.Issue.TeamID}
	}
	if len(refs.Data.Teams) > 0 {
		teams := make([]uuid.UUID, len(refs.Data.Teams))
		for i, team := range refs.Data.Teams {
			teams[i] = team.ID
		}
		return teams
	}
	if refs.Data.IssueID == nil {
		return nil
	}
	teamID, err := self.repo.Webhook().TeamOfIssue(ctx, *refs.Data.IssueID)
	if err != nil {
		return nil
	}
	return []uuid.UUID{teamID}
}
