package runner

import (
	"errors"
	"strings"
	"testing"
)

func TestParseJSONLinesFailures(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		exitCode int
		limits   ProtocolLimits
		want     error
	}{
		{name: "invalid JSON Lines", stdout: "{not-json}\n", exitCode: 1, want: ErrProtocolInvalidJSON},
		{name: "unknown event", stdout: `{"type":"mystery"}` + "\n", exitCode: 1, want: ErrProtocolUnknownEvent},
		{name: "missing terminal", stdout: `{"type":"log","message":"started"}` + "\n", exitCode: 0, want: ErrProtocolMissingTerminal},
		{name: "duplicate terminal", stdout: `{"type":"completed"}` + "\n" + `{"type":"completed"}` + "\n", exitCode: 0, want: ErrProtocolDuplicateTerminal},
		{name: "event after completed", stdout: `{"type":"completed"}` + "\n" + `{"type":"log","message":"late"}` + "\n", exitCode: 0, want: ErrProtocolEventAfterTerminal},
		{name: "exit code mismatch", stdout: `{"type":"completed"}` + "\n", exitCode: 2, want: ErrProtocolExitMismatch},
		{name: "stdout line too long", stdout: `{"type":"log","message":"` + strings.Repeat("x", 80) + `"}` + "\n", exitCode: 1, limits: ProtocolLimits{MaxLineBytes: 40, MaxTotalBytes: 1000, MaxEventCount: 10}, want: ErrProtocolLineTooLong},
		{name: "event count exceeded", stdout: `{"type":"log"}` + "\n" + `{"type":"log"}` + "\n" + `{"type":"failed"}` + "\n", exitCode: 1, limits: ProtocolLimits{MaxLineBytes: 1000, MaxTotalBytes: 1000, MaxEventCount: 2}, want: ErrProtocolTooManyEvents},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseJSONLines("job1", strings.NewReader(tt.stdout), tt.exitCode, tt.limits)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestParseJSONLinesRedactsEvents(t *testing.T) {
	result, err := ParseJSONLines("job1", strings.NewReader(`{"type":"log","message":"token=abc123 cookie=session123","metadata":{"authorization":"Bearer abc123"}}`+"\n"+`{"type":"failed","message":"password=hunter2"}`+"\n"), 1, ProtocolLimits{})
	if err != nil {
		t.Fatalf("parse JSON Lines: %v", err)
	}
	combined := result.Events[0].MessageRedacted + result.Events[0].MetadataRedacted + result.Events[1].MessageRedacted
	for _, leaked := range []string{"abc123", "session123", "hunter2", "Bearer"} {
		if strings.Contains(combined, leaked) {
			t.Fatalf("expected redaction to remove %q from %s", leaked, combined)
		}
	}
}
