package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/zigzagapp-org/zigzag-api/ent"
)

const (
	APIKeyPrefix            = "zz_"
	OAuthAccessTokenPrefix  = "zzat_"
	OAuthRefreshTokenPrefix = "zzrt_"
	apiKeySecretBytes       = 32
	apiKeyDisplayPrefixLen  = 4
)

type GeneratedAPIKey struct {
	Token  string
	Prefix string
	Hash   string
}

func NewAPIKey() (*GeneratedAPIKey, error) {
	return NewOpaqueToken(APIKeyPrefix)
}

func NewOpaqueToken(prefix string) (*GeneratedAPIKey, error) {
	secret := make([]byte, apiKeySecretBytes)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	token := prefix + base64.RawURLEncoding.EncodeToString(secret)
	return &GeneratedAPIKey{
		Token:  token,
		Prefix: token[:len(prefix)+apiKeyDisplayPrefixLen],
		Hash:   HashAPIKey(token),
	}, nil
}

func HashAPIKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func IsAPIKey(token string) bool {
	return strings.HasPrefix(token, APIKeyPrefix)
}

func IsOAuthToken(token string) bool {
	return strings.HasPrefix(token, OAuthAccessTokenPrefix) || strings.HasPrefix(token, OAuthRefreshTokenPrefix)
}

func APIKeyExpired(expiresAt *time.Time, now time.Time) bool {
	return expiresAt != nil && !expiresAt.After(now)
}

type APIKeyLookup interface {
	GetByTokenHash(ctx context.Context, tokenHash string) (*ent.APIKey, error)
}

func VerifyAPIKey(ctx context.Context, keys APIKeyLookup, token string, now time.Time) (*ent.APIKey, bool) {
	hash := HashAPIKey(token)
	key, err := keys.GetByTokenHash(ctx, hash)
	if err != nil || key.Edges.User == nil {
		return nil, false
	}
	if subtle.ConstantTimeCompare([]byte(key.TokenHash), []byte(hash)) != 1 || APIKeyExpired(key.ExpiresAt, now) {
		return nil, false
	}
	return key, true
}
