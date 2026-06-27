package sources

import "testing"

func TestSecretCipherDecryptRoundTrip(t *testing.T) {
	cipher, err := NewSecretCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	payload, err := cipher.Encrypt([]byte("secret-value"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	plaintext, err := cipher.Decrypt(payload)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plaintext) != "secret-value" {
		t.Fatalf("unexpected plaintext")
	}
}
