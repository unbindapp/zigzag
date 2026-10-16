package auth

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
)

func CSRFSecret(privateKey *rsa.PrivateKey) []byte {
	der := x509.MarshalPKCS1PrivateKey(privateKey)
	sum := sha256.Sum256(der)
	return sum[:]
}

func MintCSRFToken(secret []byte, sessionID string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sessionID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func VerifyCSRFToken(secret []byte, sessionID, presented string) bool {
	expected := MintCSRFToken(secret, sessionID)
	return hmac.Equal([]byte(expected), []byte(presented))
}
