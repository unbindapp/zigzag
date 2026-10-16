package oauthserver

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

const (
	pkceMinLen = 43
	pkceMaxLen = 128
)

func ValidPKCEValue(value string) bool {
	if len(value) < pkceMinLen || len(value) > pkceMaxLen {
		return false
	}
	for _, c := range value {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '.', c == '_', c == '~':
		default:
			return false
		}
	}
	return true
}

func VerifyS256(verifier, challenge string) bool {
	if !ValidPKCEValue(verifier) {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
