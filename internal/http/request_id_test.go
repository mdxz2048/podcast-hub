package http

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/mdxz2048/podcast-hub/config"
)

func TestRequestIDIsServerGeneratedRandomAndNonSemantic(t *testing.T) {
	server := NewServer(config.Config{
		AppEnv:            "development",
		FrontendOrigin:    "http://127.0.0.1:5173",
		SessionCookieName: "podcast_hub_session",
		CSRFHeaderName:    "X-CSRF-Token",
	}, nil, nil, HealthDependencies{}, nil, nil, nil)
	handler := server.Router()

	first := requestErrorID(t, handler, "attacker.local/leaked-host-000001")
	second := requestErrorID(t, handler, "user@host.local/path/token")
	if first == second {
		t.Fatalf("expected unique random request ids, got %q twice", first)
	}

	requestIDPattern := regexp.MustCompile(`^req_[a-f0-9]{32}$`)
	for _, requestID := range []string{first, second} {
		if !requestIDPattern.MatchString(requestID) {
			t.Fatalf("request id %q is not the expected opaque format", requestID)
		}
		for _, forbidden := range []string{".local", "host", "attacker", "/Users/", "token", "cookie", "http://", "https://", "@", "/"} {
			if strings.Contains(strings.ToLower(requestID), strings.ToLower(forbidden)) {
				t.Fatalf("request id leaked semantic data %q: %q", forbidden, requestID)
			}
		}
	}
}

func requestErrorID(t *testing.T, handler stdhttp.Handler, suppliedRequestID string) string {
	t.Helper()
	req := httptest.NewRequest(stdhttp.MethodGet, "/admin/me", nil)
	req.Header.Set("X-Request-Id", suppliedRequestID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var response errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v; body=%s", err, rec.Body.String())
	}
	if response.Error.RequestID == "" {
		t.Fatalf("expected request id in error response")
	}
	if response.Error.RequestID == suppliedRequestID {
		t.Fatalf("request id accepted caller-supplied value %q", suppliedRequestID)
	}
	if header := rec.Header().Get("X-Request-Id"); header != response.Error.RequestID {
		t.Fatalf("expected response header request id %q, got %q", response.Error.RequestID, header)
	}
	return response.Error.RequestID
}
