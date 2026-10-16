package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

type Encryptor struct {
	aead cipher.AEAD
}

func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Encryptor{aead: aead}, nil
}

func (self *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, self.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return self.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (self *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := self.aead.NonceSize()
	if len(ciphertext) < nonceSize+self.aead.Overhead() {
		return nil, ErrInvalidCiphertext
	}
	plaintext, err := self.aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}
	return plaintext, nil
}
