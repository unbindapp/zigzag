package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func key(b byte) []byte { return bytes.Repeat([]byte{b}, 32) }

func TestEncryptRoundTrip(t *testing.T) {
	enc, err := NewEncryptor(key(1))
	require.NoError(t, err)

	for _, plaintext := range [][]byte{[]byte("re_123456789"), {}, bytes.Repeat([]byte("x"), 4096)} {
		ciphertext, err := enc.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotContains(t, string(ciphertext), string(plaintext)+"\x00")
		decrypted, err := enc.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, string(plaintext), string(decrypted))
	}
}

func TestEncryptUsesFreshNonces(t *testing.T) {
	enc, _ := NewEncryptor(key(1))
	a, _ := enc.Encrypt([]byte("secret"))
	b, _ := enc.Encrypt([]byte("secret"))
	assert.NotEqual(t, a, b)
}

func TestDecryptRejectsTamperingAndWrongKeys(t *testing.T) {
	enc, _ := NewEncryptor(key(1))
	other, _ := NewEncryptor(key(2))
	ciphertext, _ := enc.Encrypt([]byte("secret"))

	_, err := other.Decrypt(ciphertext)
	assert.ErrorIs(t, err, ErrInvalidCiphertext)

	tampered := append([]byte{}, ciphertext...)
	tampered[len(tampered)-1] ^= 0xff
	_, err = enc.Decrypt(tampered)
	assert.ErrorIs(t, err, ErrInvalidCiphertext)

	_, err = enc.Decrypt([]byte("short"))
	assert.ErrorIs(t, err, ErrInvalidCiphertext)
}

func TestNewEncryptorNeeds32Bytes(t *testing.T) {
	_, err := NewEncryptor([]byte("too short"))
	assert.Error(t, err)
}
