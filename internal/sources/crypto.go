package sources

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const EncryptionVersionAESGCMV1 = "aes-gcm-v1"

type SecretCipher struct {
	aead cipher.AEAD
}

func NewSecretCipher(masterKey string) (*SecretCipher, error) {
	key, err := decodeMasterKey(masterKey)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create secret cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create secret aead: %w", err)
	}
	return &SecretCipher{aead: aead}, nil
}

func (c *SecretCipher) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("create secret nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	payload := append(nonce, sealed...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func decodeMasterKey(masterKey string) ([]byte, error) {
	if decoded, err := base64.StdEncoding.DecodeString(masterKey); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if len([]byte(masterKey)) == 32 {
		return []byte(masterKey), nil
	}
	return nil, fmt.Errorf("SECRETS_MASTER_KEY must be 32 raw bytes or base64-encoded 32 bytes")
}
