package runner

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

func TestRunnerFixtureCompletesAndCleansWorkspace(t *testing.T) {
	root := t.TempDir()
	jobService := &fakeJobService{job: jobs.ImportJob{
		ID:                 "job1",
		ConnectorSourceID:  "source1",
		ConnectorVersionID: "version1",
		Status:             jobs.JobStatusQueued,
		TriggerType:        "manual",
		AuthMode:           "none",
		ExecutionMode:      "unattended",
		CreatedAt:          time.Now(),
	}}
	executor := fixtureExecutor{run: func(inputPath string, outputDir string) ExecutionResult {
		body, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("read input: %v", err)
		}
		if strings.Contains(strings.ToLower(string(body)), "secret") || strings.Contains(string(body), "/Users/") {
			t.Fatalf("job input leaked secret or host path: %s", string(body))
		}
		writeArtifact(t, outputDir, "episodes/episode-001.json", []byte(`{"title":"Fixture"}`))
		return ExecutionResult{Stdout: bytes.NewReader([]byte(
			`{"type":"log","message":"fixture started"}` + "\n" +
				`{"type":"artifact_ready","artifact_type":"episode_metadata","path":"episodes/episode-001.json"}` + "\n" +
				`{"type":"completed","message":"done"}` + "\n",
		)), ExitCode: 0}
	}}
	r := New(jobService, executor, Config{WorkspaceRoot: root})
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if jobService.job.Status != jobs.JobStatusCompleted {
		t.Fatalf("expected completed, got %s", jobService.job.Status)
	}
	if len(jobService.artifacts) != 1 {
		t.Fatalf("expected artifact metadata, got %d", len(jobService.artifacts))
	}
	if strings.Contains(jobService.artifacts[0].RelativePath, root) {
		t.Fatalf("artifact leaked workspace path: %+v", jobService.artifacts[0])
	}
	if _, err := os.Stat(filepath.Join(root, "job1")); !os.IsNotExist(err) {
		t.Fatalf("expected workspace cleanup, stat err=%v", err)
	}
}

type fixtureExecutor struct {
	run func(inputPath string, outputDir string) ExecutionResult
}

func (e fixtureExecutor) Execute(_ context.Context, inputPath string, outputDir string) ExecutionResult {
	return e.run(inputPath, outputDir)
}

type fakeJobService struct {
	job       jobs.ImportJob
	claimed   bool
	events    []jobs.ImportJobEvent
	artifacts []jobs.ImportJobArtifact
}

func (s *fakeJobService) ClaimNextQueuedJob(context.Context) (jobs.ImportJob, bool, error) {
	if s.claimed || s.job.Status != jobs.JobStatusQueued {
		return jobs.ImportJob{}, false, nil
	}
	s.claimed = true
	s.job.Status = jobs.JobStatusRunning
	now := time.Now()
	s.job.StartedAt = &now
	return s.job, true, nil
}

func (s *fakeJobService) AppendEvent(_ context.Context, event jobs.ImportJobEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (s *fakeJobService) AppendArtifact(_ context.Context, artifact jobs.ImportJobArtifact) error {
	s.artifacts = append(s.artifacts, artifact)
	return nil
}

func (s *fakeJobService) TransitionJob(_ context.Context, _ string, status jobs.JobStatus, failureCode string, failureMessageRedacted string) (jobs.ImportJob, error) {
	s.job.Status = status
	s.job.FailureCode = failureCode
	s.job.FailureMessageRedacted = failureMessageRedacted
	now := time.Now()
	if status == jobs.JobStatusCompleted || status == jobs.JobStatusFailed || status == jobs.JobStatusCancelled {
		s.job.FinishedAt = &now
	}
	return s.job, nil
}
