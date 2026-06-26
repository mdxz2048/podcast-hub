package security

import "testing"

func TestArgon2idHasher_HashAndVerify(t *testing.T) {
	hasher := DefaultArgon2idHasher()
	hash, err := hasher.HashPassword("very-strong-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	ok, err := hasher.VerifyPassword("very-strong-password", hash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !ok {
		t.Fatalf("expected password to verify")
	}
	bad, err := hasher.VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if bad {
		t.Fatalf("expected wrong password to fail verification")
	}
}
