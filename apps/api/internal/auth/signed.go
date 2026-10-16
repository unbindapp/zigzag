package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

const googleLoginBaseName = "google_login"

var ErrInvalidSignature = errors.New("invalid signature")

func GoogleLoginCookieName(secure bool) string { return hostScopedName(googleLoginBaseName, secure) }

func GoogleLoginCookie(value string, expires time.Time, secure bool) http.Cookie {
	return newCookie(GoogleLoginCookieName(secure), value, expires, secure)
}

func ClearedGoogleLoginCookie(secure bool) http.Cookie {
	return newCookie(GoogleLoginCookieName(secure), "", time.Unix(0, 0), secure)
}

func SignValue(secret []byte, purpose string, payload []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + MintCSRFToken(secret, purpose+":"+encoded)
}

func VerifySignedValue(secret []byte, purpose, signed string) ([]byte, error) {
	encoded, mac, found := strings.Cut(signed, ".")
	if !found || !VerifyCSRFToken(secret, purpose+":"+encoded, mac) {
		return nil, ErrInvalidSignature
	}
	return base64.RawURLEncoding.DecodeString(encoded)
}
