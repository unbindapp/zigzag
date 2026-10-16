package email_service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/config"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/common/log"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/crypto"
	"github.com/unbindapp/zigzag-api/internal/infrastructure/mail"
	"github.com/unbindapp/zigzag-api/internal/models"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositories"
	system_repo "github.com/unbindapp/zigzag-api/internal/repositories/system"
)

const asyncSendTimeout = 20 * time.Second

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i EmailServiceInterface -p email_service -s EmailService -o email_service_iface.go
type EmailService struct {
	cfg       *config.Config
	repo      repositories.RepositoriesInterface
	encryptor *crypto.Encryptor
	sender    mail.Sender
	async     func(func())
}

func NewEmailService(cfg *config.Config, repo repositories.RepositoriesInterface, encryptor *crypto.Encryptor, sender mail.Sender) *EmailService {
	return &EmailService{cfg: cfg, repo: repo, encryptor: encryptor, sender: sender, async: func(fn func()) { go fn() }}
}

type EmailSettingsUpdateInput struct {
	schema.EmailSettings
	Password *string `json:"password,omitempty" required:"false" doc:"SMTP password, for Resend the API key. Omit to keep the stored one; send an empty string to clear it."`
}

func (self *EmailService) GetSettings(ctx context.Context, requesterUserID uuid.UUID) (*models.EmailSettingsResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return nil, err
	}
	settings, err := self.repo.System().GetSystemSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	return models.TransformEmailSettings(settings), nil
}

func (self *EmailService) UpdateSettings(ctx context.Context, requesterUserID uuid.UUID, input *EmailSettingsUpdateInput) (*models.EmailSettingsResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return nil, err
	}
	switch mail.Encryption(input.Encryption) {
	case mail.EncryptionStartTLS, mail.EncryptionTLS, mail.EncryptionNone:
	default:
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Encryption must be starttls, tls or none")
	}

	update := &system_repo.SystemSettingUpdateInput{EmailSettings: &input.EmailSettings}
	if input.Password != nil && *input.Password == "" {
		update.ClearEmailPassword = true
	}
	if input.Password != nil && *input.Password != "" {
		encrypted, err := self.encryptor.Encrypt([]byte(*input.Password))
		if err != nil {
			return nil, errdefs.NewInternalError(err, "Failed to encrypt the SMTP password")
		}
		update.EmailPasswordEncrypted = encrypted
	}

	settings, err := self.repo.System().UpdateSystemSettings(ctx, update)
	if err != nil {
		return nil, err
	}
	return models.TransformEmailSettings(settings), nil
}

func (self *EmailService) ClearSettings(ctx context.Context, requesterUserID uuid.UUID) error {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return err
	}
	_, err := self.repo.System().UpdateSystemSettings(ctx, &system_repo.SystemSettingUpdateInput{ClearEmailSettings: true})
	return err
}

func (self *EmailService) SendTest(ctx context.Context, requesterUserID uuid.UUID) error {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return err
	}
	user, err := self.repo.User().GetByID(ctx, requesterUserID)
	if err != nil {
		return err
	}
	sent, err := self.Send(ctx, user.Email, mail.Content{
		Subject:    "Zigzag can send email",
		Heading:    "Email is working",
		Paragraphs: []string{"This is a test message from your Zigzag instance. Invitations, password resets and notifications are delivered the same way."},
	})
	if err != nil {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, fmt.Sprintf("Sending failed: %v", err))
	}
	if !sent {
		return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Email is not configured")
	}
	return nil
}

func (self *EmailService) Configured(ctx context.Context) (bool, error) {
	_, configured, _, err := self.settings(ctx)
	return configured, err
}

// Send reports sent=false without an error when no relay is configured.
func (self *EmailService) Send(ctx context.Context, to string, content mail.Content) (sent bool, err error) {
	settings, configured, instanceName, err := self.settings(ctx)
	if err != nil || !configured {
		return false, err
	}
	if content.InstanceURL == "" {
		content.InstanceURL = self.cfg.ExternalUIUrl
	}
	if content.InstanceName == "" {
		content.InstanceName = instanceName
	}
	msg, err := mail.Render(to, content)
	if err != nil {
		return false, err
	}
	if err := self.sender.Send(ctx, settings, msg); err != nil {
		return false, err
	}
	return true, nil
}

// SendAsync never blocks the caller on SMTP; queued is false when mail is not configured.
func (self *EmailService) SendAsync(ctx context.Context, to string, content mail.Content) (queued bool) {
	configured, err := self.Configured(ctx)
	if err != nil {
		log.Warn("Failed to read mail settings", "err", err)
		return false
	}
	if !configured {
		return false
	}
	self.async(func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), asyncSendTimeout)
		defer cancel()
		if _, err := self.Send(sendCtx, to, content); err != nil {
			log.Warn("Failed to send email", "to", to, "subject", content.Subject, "err", err)
		}
	})
	return true
}

func (self *EmailService) settings(ctx context.Context) (mail.Settings, bool, string, error) {
	stored, err := self.repo.System().GetSystemSettings(ctx, nil)
	if ent.IsNotFound(err) {
		return mail.Settings{}, false, "", nil
	}
	if err != nil {
		return mail.Settings{}, false, "", err
	}
	if stored.EmailSettings == nil {
		return mail.Settings{}, false, stored.InstanceName, nil
	}

	password := ""
	if len(stored.EmailPasswordEncrypted) > 0 {
		decrypted, err := self.encryptor.Decrypt(stored.EmailPasswordEncrypted)
		if err != nil {
			return mail.Settings{}, false, "", errdefs.NewInternalError(err, "Failed to decrypt the SMTP password, was ENCRYPTION_KEY changed?")
		}
		password = string(decrypted)
	}

	cfg := stored.EmailSettings
	return mail.Settings{
		Host:        cfg.Host,
		Port:        cfg.Port,
		Username:    cfg.Username,
		Password:    password,
		FromAddress: cfg.FromAddress,
		FromName:    cfg.FromName,
		Encryption:  mail.Encryption(cfg.Encryption),
	}, true, stored.InstanceName, nil
}
