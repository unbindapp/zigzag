package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignedValueRoundTrip(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	signed := SignValue(secret, "google-login", []byte(`{"s":"state"}`))

	payload, err := VerifySignedValue(secret, "google-login", signed)
	require.NoError(t, err)
	assert.JSONEq(t, `{"s":"state"}`, string(payload))
}

func TestSignedValueRejectsTampering(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	signed := SignValue(secret, "google-login", []byte(`{"s":"state"}`))
	forged := SignValue(secret, "google-login", []byte(`{"s":"other"}`))
	_, mac, _ := strings.Cut(forged, ".")
	payload, _, _ := strings.Cut(signed, ".")

	cases := map[string]string{
		"swapped payload": payload + "." + mac,
		"other purpose":   SignValue(secret, "csrf", []byte(`{"s":"state"}`)),
		"other secret":    SignValue([]byte("another-secret-another-secret-xx"), "google-login", []byte(`{"s":"state"}`)),
		"no signature":    payload,
		"empty":           "",
	}
	for name, value := range cases {
		_, err := VerifySignedValue(secret, "google-login", value)
		assert.ErrorIs(t, err, ErrInvalidSignature, name)
	}
}
