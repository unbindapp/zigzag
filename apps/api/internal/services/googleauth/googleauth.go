package googleauth_service

import (
	"context"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/crypto"
	"github.com/zigzagapp-org/zigzag-api/internal/repositories/repositories"
	"golang.org/x/oauth2"
)

const (
	googleIssuer    = "https://accounts.google.com"
	googleJWKSURL   = "https://www.googleapis.com/oauth2/v3/certs"
	googleAuthURL   = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL  = "https://oauth2.googleapis.com/token"
	exchangeTimeout = 10 * time.Second
)

type GoogleIdentity struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	HostedDomain  string `json:"hd"`
	Name          string `json:"name"`
	Nonce         string `json:"nonce"`
}

type IDTokenVerifier interface {
	Verify(ctx context.Context, clientID, rawIDToken string) (*GoogleIdentity, error)
}

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i GoogleAuthServiceInterface -p googleauth_service -s GoogleAuthService -o googleauth_service_iface.go
type GoogleAuthService struct {
	cfg         *config.Config
	repo        repositories.RepositoriesInterface
	encryptor   *crypto.Encryptor
	stateSecret []byte
	verifier    IDTokenVerifier
	endpoint    oauth2.Endpoint
	httpClient  *http.Client
	now         func() time.Time
}

func NewGoogleAuthService(cfg *config.Config, repo repositories.RepositoriesInterface, encryptor *crypto.Encryptor, stateSecret []byte) *GoogleAuthService {
	return &GoogleAuthService{
		cfg:         cfg,
		repo:        repo,
		encryptor:   encryptor,
		stateSecret: stateSecret,
		verifier:    &googleVerifier{keys: oidc.NewRemoteKeySet(context.Background(), googleJWKSURL)},
		endpoint:    oauth2.Endpoint{AuthURL: googleAuthURL, TokenURL: googleTokenURL, AuthStyle: oauth2.AuthStyleInParams},
		httpClient:  &http.Client{Timeout: exchangeTimeout},
		now:         time.Now,
	}
}

type googleVerifier struct {
	keys oidc.KeySet
}

func (self *googleVerifier) Verify(ctx context.Context, clientID, rawIDToken string) (*GoogleIdentity, error) {
	token, err := oidc.NewVerifier(googleIssuer, self.keys, &oidc.Config{ClientID: clientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	identity := &GoogleIdentity{}
	if err := token.Claims(identity); err != nil {
		return nil, err
	}
	identity.Subject = token.Subject
	identity.Nonce = token.Nonce
	return identity, nil
}
