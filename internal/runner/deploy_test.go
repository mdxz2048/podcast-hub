package runner

import (
	"os"
	"strings"
	"testing"
)

func TestAlphaComposeDockerSocketBoundary(t *testing.T) {
	apiCompose, err := os.ReadFile("../../docker-compose.alpha.yml")
	if err != nil {
		t.Fatalf("read api compose: %v", err)
	}
	if strings.Contains(string(apiCompose), "docker.sock") {
		t.Fatal("API alpha compose must not mount Docker socket")
	}
	runnerCompose, err := os.ReadFile("../../deploy/docker-compose.runner-alpha.yml")
	if err != nil {
		t.Fatalf("read runner compose: %v", err)
	}
	body := string(runnerCompose)
	if strings.Count(body, "docker.sock") != 2 {
		t.Fatalf("runner compose should have one Docker socket bind, got:\n%s", body)
	}
	if strings.Contains(body, "ports:") {
		t.Fatal("runner compose must not expose public ports")
	}
}
