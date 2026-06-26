package runner

import (
	"os"
	"strings"
	"testing"
)

func TestAPIServerDoesNotDependOnDockerRunner(t *testing.T) {
	for _, path := range []string{"../../cmd/api/main.go", "../http/server.go"} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lower := strings.ToLower(string(body))
		for _, forbidden := range []string{"docker", "docker.sock", "docker_trusted_admin", "internal/runner"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("API file %s must not depend on %s", path, forbidden)
			}
		}
	}
}
