package security

import (
	"context"
	"testing"
)

func TestMockTurnstileVerifier(t *testing.T) {
	verifier := MockTurnstileVerifier{}
	if err := verifier.Verify(context.Background(), "", "127.0.0.1"); err == nil {
		t.Fatalf("expected missing token to fail")
	}
	if err := verifier.Verify(context.Background(), "invalid-token", "127.0.0.1"); err == nil {
		t.Fatalf("expected invalid token to fail")
	}
	if err := verifier.Verify(context.Background(), "mock-pass", "127.0.0.1"); err != nil {
		t.Fatalf("expected mock-pass token to pass, got %v", err)
	}
}
