package webhook_service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/oauthserver"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
)

const deliveryTimeout = 10 * time.Second

var Events = []string{"issue.created", "issue.updated", "issue.deleted", "comment.created", "project.updated", "cycle.updated"}

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i WebhookServiceInterface -p webhook_service -s WebhookService -o webhook_service_iface.go
type WebhookService struct {
	cfg         *config.Config
	repo        repositories.RepositoriesInterface
	client      *http.Client
	policy      oauthserver.IPPolicy
	async       func(func())
	retryDelays []time.Duration
	now         func() time.Time
}

func NewWebhookService(cfg *config.Config, repo repositories.RepositoriesInterface) *WebhookService {
	return &WebhookService{
		cfg:         cfg,
		repo:        repo,
		client:      newClient(oauthserver.DefaultIPPolicy),
		policy:      oauthserver.DefaultIPPolicy,
		async:       func(fn func()) { go fn() },
		retryDelays: []time.Duration{2 * time.Second, 10 * time.Second, 30 * time.Second},
		now:         time.Now,
	}
}

func newClient(policy oauthserver.IPPolicy) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext:         oauthserver.SafeDialContext(policy),
			TLSHandshakeTimeout: deliveryTimeout,
			ForceAttemptHTTP2:   true,
		},
		Timeout: deliveryTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (self *WebhookService) checkAdmin(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return self.repo.Permissions().Check(ctx, userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleAdmin, WorkspaceID: workspaceID}})
}

func (self *WebhookService) adminWebhook(ctx context.Context, userID, webhookID uuid.UUID) (*ent.Webhook, error) {
	notFound := errdefs.NewCustomError(errdefs.ErrTypeNotFound, "Webhook not found")
	hook, err := self.repo.Webhook().GetByID(ctx, webhookID)
	if ent.IsNotFound(err) {
		return nil, notFound
	}
	if err != nil {
		return nil, err
	}
	viewer := []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleViewer, WorkspaceID: hook.WorkspaceID}}
	if err := self.repo.Permissions().Check(ctx, userID, viewer); err != nil {
		return nil, errdefs.MaskAsNotFound(err, "Webhook not found")
	}
	if err := self.checkAdmin(ctx, userID, hook.WorkspaceID); err != nil {
		return nil, err
	}
	return hook, nil
}
