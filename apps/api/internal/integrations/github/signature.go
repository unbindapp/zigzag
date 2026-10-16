package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const SignatureHeader = "X-Hub-Signature-256"

func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func ValidSignature(secret, header string, body []byte) bool {
	given, ok := strings.CutPrefix(header, "sha256=")
	if !ok {
		return false
	}
	decoded, err := hex.DecodeString(given)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(decoded, mac.Sum(nil))
}
