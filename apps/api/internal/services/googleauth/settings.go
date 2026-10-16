package googleauth_service

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/models"
	system_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/system"
)

const secretPrefixLength = 10

var domainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

func (self *GoogleAuthService) RedirectURI() string {
	return self.cfg.ExternalUIUrl + self.cfg.APIPathPrefix() + "/auth/google/callback"
}

func (self *GoogleAuthService) GetSettings(ctx context.Context, requesterUserID uuid.UUID) (*models.GoogleAuthSettingsResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return nil, err
	}
	settings, err := self.repo.System().GetSystemSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	return self.transform(settings)
}

type UpdateGoogleAuthSettingsInput struct {
	Enabled      bool    `json:"enabled" required:"true" doc:"Offer Sign in with Google on the sign-in page"`
	ClientID     string  `json:"client_id" required:"true" minLength:"1" maxLength:"255" doc:"OAuth client id of a Google Cloud web application client"`
	ClientSecret *string `json:"client_secret,omitempty" required:"false" minLength:"1" maxLength:"255" doc:"OAuth client secret. Omit to keep the stored one."`
	HostedDomain *string `json:"hosted_domain,omitempty" required:"false" maxLength:"253" doc:"Only accept Google Workspace accounts of this domain, for example example.com. Empty or omitted accepts any Google account."`
}

func (self *GoogleAuthService) UpdateSettings(ctx context.Context, requesterUserID uuid.UUID, input *UpdateGoogleAuthSettingsInput) (*models.GoogleAuthSettingsResponse, error) {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return nil, err
	}
	clientID := strings.TrimSpace(input.ClientID)
	if clientID == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Client id is required")
	}
	domain := ""
	if input.HostedDomain != nil {
		domain = strings.ToLower(strings.TrimSpace(*input.HostedDomain))
	}
	if domain != "" && !domainPattern.MatchString(domain) {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Hosted domain must be a domain such as example.com")
	}

	update := &system_repo.SystemSettingUpdateInput{GoogleAuth: &schema.GoogleAuthSettings{Enabled: input.Enabled, ClientID: clientID, HostedDomain: domain}}
	secret := ""
	if input.ClientSecret != nil {
		secret = strings.TrimSpace(*input.ClientSecret)
	}
	if input.ClientSecret != nil && secret == "" {
		return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Client secret cannot be empty; omit it to keep the stored one")
	}
	if secret != "" {
		encrypted, err := self.encryptor.Encrypt([]byte(secret))
		if err != nil {
			return nil, errdefs.NewInternalError(err, "Failed to encrypt the client secret")
		}
		update.GoogleSecretEncrypted = encrypted
	}
	if secret == "" && input.Enabled {
		current, err := self.repo.System().GetSystemSettings(ctx, nil)
		if err != nil {
			return nil, err
		}
		if len(current.GoogleClientSecretEncrypted) == 0 {
			return nil, errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Add the client secret before enabling Google sign-in")
		}
	}

	settings, err := self.repo.System().UpdateSystemSettings(ctx, update)
	if err != nil {
		return nil, err
	}
	return self.transform(settings)
}

func (self *GoogleAuthService) ClearSettings(ctx context.Context, requesterUserID uuid.UUID) error {
	if err := self.repo.Permissions().CheckSystemAdmin(ctx, requesterUserID); err != nil {
		return err
	}
	_, err := self.repo.System().UpdateSystemSettings(ctx, &system_repo.SystemSettingUpdateInput{ClearGoogleAuth: true})
	return err
}

func (self *GoogleAuthService) Providers(ctx context.Context) (*models.AuthProvidersResponse, error) {
	settings, err := self.repo.System().GetSystemSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &models.AuthProvidersResponse{Google: models.AuthProviderStatus{Enabled: googleEnabled(settings)}}, nil
}

func (self *GoogleAuthService) transform(settings *ent.SystemSetting) (*models.GoogleAuthSettingsResponse, error) {
	resp := &models.GoogleAuthSettingsResponse{RedirectURI: self.RedirectURI()}
	if settings.GoogleAuth != nil {
		resp.Enabled = settings.GoogleAuth.Enabled
		resp.ClientID = settings.GoogleAuth.ClientID
		resp.HostedDomain = settings.GoogleAuth.HostedDomain
	}
	if len(settings.GoogleClientSecretEncrypted) == 0 {
		return resp, nil
	}
	secret, err := self.encryptor.Decrypt(settings.GoogleClientSecretEncrypted)
	if err != nil {
		return nil, errdefs.NewInternalError(err, "Failed to decrypt the Google client secret, was ENCRYPTION_KEY changed?")
	}
	resp.Configured = resp.ClientID != ""
	resp.SecretPrefix = string(secret[:min(secretPrefixLength, len(secret)/2)])
	return resp, nil
}

func googleEnabled(settings *ent.SystemSetting) bool {
	return settings.GoogleAuth != nil && settings.GoogleAuth.Enabled && settings.GoogleAuth.ClientID != "" && len(settings.GoogleClientSecretEncrypted) > 0
}
