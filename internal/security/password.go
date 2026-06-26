package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

type Argon2idHasher struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

func DefaultArgon2idHasher() Argon2idHasher {
	return Argon2idHasher{
		Time:    2,
		Memory:  64 * 1024,
		Threads: 1,
		KeyLen:  32,
	}
}

func (h Argon2idHasher) HashPassword(password string) (string, error) {
	if utf8.RuneCountInString(password) < 12 {
		return "", fmt.Errorf("password must be at least 12 characters")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Threads, h.KeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", h.Memory, h.Time, h.Threads, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func (h Argon2idHasher) VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid password hash format")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Threads, uint32(len(expected)))
	return subtleEqual(expected, key), nil
}

func HashWithPepper(raw, pepper string) string {
	sum := sha256.Sum256([]byte(raw + ":" + pepper))
	return hex.EncodeToString(sum[:])
}

func NewOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func subtleEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var out byte
	for i := range a {
		out |= a[i] ^ b[i]
	}
	return out == 0
}
