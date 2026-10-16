package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zigzagapp-org/zigzag-api/ent"
)

func newManager(t *testing.T, issuer string) *TokenManager {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return NewTokenManager(key, issuer, "zigzag-api")
}

func TestMintAndVerify(t *testing.T) {
	manager := newManager(t, "https://zigzag.example.com")
	user := &ent.User{ID: uuid.New(), Email: "ada@example.com"}

	token, expiresAt, err := manager.MintAccessToken(user)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(AccessTokenTTL), expiresAt, time.Minute)

	claims, err := manager.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "ada@example.com", claims.Email)
}

func TestTokensMintedTogetherDiffer(t *testing.T) {
	manager := newManager(t, "https://zigzag.example.com")
	user := &ent.User{ID: uuid.New(), Email: "ada@example.com"}
	a, _, _ := manager.MintAccessToken(user)
	b, _, _ := manager.MintAccessToken(user)
	assert.NotEqual(t, a, b, "sessions store access tokens under a unique index")
}

func TestVerifyRejectsForeignTokens(t *testing.T) {
	manager := newManager(t, "https://zigzag.example.com")
	other := newManager(t, "https://zigzag.example.com")
	otherIssuer := newManager(t, "https://elsewhere.example.com")
	user := &ent.User{ID: uuid.New(), Email: "ada@example.com"}

	signedElsewhere, _, _ := other.MintAccessToken(user)
	_, err := manager.Verify(signedElsewhere)
	assert.Error(t, err, "a different signing key")

	otherIssuer.privateKey = manager.privateKey
	wrongIssuer, _, _ := otherIssuer.MintAccessToken(user)
	_, err = manager.Verify(wrongIssuer)
	assert.Error(t, err, "a different issuer")

	_, err = manager.Verify("not-a-jwt")
	assert.Error(t, err)
}
