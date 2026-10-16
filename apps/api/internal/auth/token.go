package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
)

const (
	KeyID = "zigzag-oauth-key"

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 24 * time.Hour * 14

	TokenScope = "openid profile email offline_access"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenManager struct {
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	csrfSecret []byte
}

func NewTokenManager(privateKey *rsa.PrivateKey, issuer, audience string) *TokenManager {
	return &TokenManager{
		privateKey: privateKey,
		issuer:     issuer,
		audience:   audience,
		csrfSecret: CSRFSecret(privateKey),
	}
}

func (self *TokenManager) CSRFSecret() []byte {
	return self.csrfSecret
}

type VerifiedClaims struct {
	UserID uuid.UUID
	Email  string
}

func (self *TokenManager) MintAccessToken(user *ent.User) (token string, expiresAt time.Time, err error) {
	now := time.Now()
	expiresAt = now.Add(AccessTokenTTL)

	claims := jwt.MapClaims{
		"iss":            self.issuer,
		"jti":            uuid.NewString(),
		"sub":            user.ID.String(),
		"aud":            self.audience,
		"email":          user.Email,
		"email_verified": true,
		"iat":            now.Unix(),
		"exp":            expiresAt.Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	jwtToken.Header["kid"] = KeyID

	signed, err := jwtToken.SignedString(self.privateKey)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (self *TokenManager) Verify(tokenString string) (*VerifiedClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return &self.privateKey.PublicKey, nil
		},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(self.issuer),
		jwt.WithAudience(self.audience),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	subject, _ := claims["sub"].(string)
	userID, err := uuid.Parse(subject)
	if err != nil {
		return nil, fmt.Errorf("%w: subject is not a user id", ErrInvalidToken)
	}
	email, _ := claims["email"].(string)
	return &VerifiedClaims{UserID: userID, Email: email}, nil
}

func (self *TokenManager) Issuer() string {
	return self.issuer
}

func NewRefreshToken() string {
	return uuid.NewString()
}
