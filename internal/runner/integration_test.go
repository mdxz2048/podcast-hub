package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

func TestIntegrationDockerTrustedAdminFixture(t *testing.T) {
	if os.Getenv("RUNNER_INTEGRATION_TEST") != "1" {
		t.Skip("set RUNNER_INTEGRATION_TEST=1 to run real Docker fixture integration")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skipf("Docker is unavailable for integration test: %v", err)
	}
	fixtureDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("fixture path: %v", err)
	}
	root := t.TempDir()
	jobService := &fakeJobService{job: jobs.ImportJob{
		ID:                 "job-integration",
		ConnectorSourceID:  "source1",
		ConnectorVersionID: "version1",
		Status:             jobs.JobStatusQueued,
		TriggerType:        "manual",
		AuthMode:           "reusable_session",
		ExecutionMode:      "unattended",
		CreatedAt:          time.Now(),
	}}
	r := New(jobService, DockerTrustedAdminExecutor{Config: DockerTrustedAdminConfig{
		Image:                "python:3.12-alpine",
		ConnectorPackagePath: fixtureDir,
		Timeout:              30 * time.Second,
	}}, Config{
		WorkspaceRoot: root,
		SecretProvider: fakeSecretProvider{secrets: []sources.RunnerSecret{{
			Name:  "session_file",
			Type:  sources.SecretTypeFile,
			Value: []byte("integration-secret-value"),
		}}},
		ExecutionTimeout: 30 * time.Second,
	})
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("run Docker integration fixture: %v", err)
	}
	if jobService.job.Status != jobs.JobStatusCompleted {
		t.Fatalf("expected completed job, got %+v", jobService.job)
	}
	if len(jobService.artifacts) != 1 {
		t.Fatalf("expected one artifact, got %d", len(jobService.artifacts))
	}
	if jobService.artifacts[0].SizeBytes <= 0 || jobService.artifacts[0].SHA256 == "" {
		t.Fatalf("expected hash and size metadata, got %+v", jobService.artifacts[0])
	}
	if _, err := os.Stat(filepath.Join(root, "job-integration")); !os.IsNotExist(err) {
		t.Fatalf("expected workspace cleanup, stat err=%v", err)
	}
}
