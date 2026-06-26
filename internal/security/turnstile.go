package security

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrTurnstileRequired = errors.New("turnstile_required")
	ErrTurnstileFailed   = errors.New("turnstile_failed")
)

type TurnstileVerifier interface {
	Verify(ctx context.Context, token, remoteIP string) error
}

type MockTurnstileVerifier struct{}

func (v MockTurnstileVerifier) Verify(_ context.Context, token, _ string) error {
	if strings.TrimSpace(token) == "" {
		return ErrTurnstileRequired
	}
	if token != "mock-pass" {
		return ErrTurnstileFailed
	}
	return nil
}

type CloudflareTurnstileVerifier struct {
	SecretKey  string
	HTTPClient *http.Client
}

func (v CloudflareTurnstileVerifier) Verify(ctx context.Context, token, remoteIP string) error {
	if strings.TrimSpace(token) == "" {
		return ErrTurnstileRequired
	}
	form := url.Values{}
	form.Set("secret", v.SecretKey)
	form.Set("response", token)
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", remoteIP)
	}
	client := v.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return ErrTurnstileFailed
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return ErrTurnstileFailed
	}
	defer res.Body.Close()
	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return ErrTurnstileFailed
	}
	if !payload.Success {
		return ErrTurnstileFailed
	}
	return nil
}
