package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/sources"
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
		if strings.Contains(string(body), "/Users/") || strings.Contains(string(body), ".local/") {
			t.Fatalf("job input leaked host path: %s", string(body))
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

func TestRunnerSkipsQueuedCancelledJob(t *testing.T) {
	jobService := &fakeJobService{job: jobs.ImportJob{ID: "job1", Status: jobs.JobStatusCancelled}}
	r := New(jobService, fixtureExecutor{run: func(string, string) ExecutionResult {
		t.Fatal("executor should not run cancelled queued job")
		return ExecutionResult{}
	}}, Config{WorkspaceRoot: t.TempDir()})
	if err := r.RunOnce(context.Background()); !errors.Is(err, ErrNoQueuedJob) {
		t.Fatalf("expected no queued job, got %v", err)
	}
}

func TestRunnerRunningCancelStopsExecution(t *testing.T) {
	jobService := &fakeJobService{
		cancelOnGet: true,
		job: jobs.ImportJob{
			ID:                 "job1",
			ConnectorSourceID:  "source1",
			ConnectorVersionID: "version1",
			Status:             jobs.JobStatusQueued,
			TriggerType:        "manual",
			AuthMode:           "none",
			ExecutionMode:      "unattended",
			CreatedAt:          time.Now(),
		},
	}
	executor := fixtureExecutor{run: func(_ string, _ string) ExecutionResult {
		<-time.After(50 * time.Millisecond)
		return ExecutionResult{Stdout: bytes.NewReader(nil), ExitCode: 137, Err: context.Canceled}
	}}
	r := New(jobService, executor, Config{WorkspaceRoot: t.TempDir(), CancellationPollInterval: time.Millisecond})
	if err := r.RunOnce(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if jobService.job.Status != jobs.JobStatusCancelled {
		t.Fatalf("expected cancelled job, got %s", jobService.job.Status)
	}
}

func TestRunnerTimeoutFailsJob(t *testing.T) {
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
	executor := fixtureExecutor{run: func(_ string, _ string) ExecutionResult {
		time.Sleep(20 * time.Millisecond)
		return ExecutionResult{Stdout: bytes.NewReader(nil), ExitCode: 137, Err: context.DeadlineExceeded}
	}}
	r := New(jobService, executor, Config{WorkspaceRoot: t.TempDir(), ExecutionTimeout: time.Millisecond, CancellationPollInterval: time.Millisecond})
	if err := r.RunOnce(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if jobService.job.Status != jobs.JobStatusFailed || jobService.job.FailureCode != "timeout" {
		t.Fatalf("expected timeout failure, got %+v", jobService.job)
	}
}

func TestRunnerInjectsSecretsWithoutLeaking(t *testing.T) {
	root := t.TempDir()
	const secretValue = "super-secret-token"
	jobService := &fakeJobService{job: jobs.ImportJob{
		ID:                 "job1",
		ConnectorSourceID:  "source1",
		ConnectorVersionID: "version1",
		Status:             jobs.JobStatusQueued,
		TriggerType:        "manual",
		AuthMode:           "reusable_session",
		ExecutionMode:      "unattended",
		CreatedAt:          time.Now(),
	}}
	executor := fixtureExecutor{run: func(inputPath string, outputDir string) ExecutionResult {
		body, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("read input: %v", err)
		}
		if strings.Contains(string(body), secretValue) {
			t.Fatalf("job input leaked secret value: %s", string(body))
		}
		if !strings.Contains(string(body), "/work/secrets/session_file") {
			t.Fatalf("job input should contain logical secret path, got %s", string(body))
		}
		secretBody, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(inputPath)), "secrets", "session_file"))
		if err != nil {
			t.Fatalf("read injected secret: %v", err)
		}
		if string(secretBody) != secretValue {
			t.Fatalf("unexpected secret content")
		}
		writeArtifact(t, outputDir, "episodes/episode-001.json", []byte(`{"title":"Fixture"}`))
		return ExecutionResult{Stdout: bytes.NewReader([]byte(
			`{"type":"artifact_ready","artifact_type":"episode_metadata","path":"episodes/episode-001.json"}` + "\n" +
				`{"type":"completed","message":"done"}` + "\n",
		)), ExitCode: 0}
	}}
	r := New(jobService, executor, Config{
		WorkspaceRoot: root,
		SecretProvider: fakeSecretProvider{secrets: []sources.RunnerSecret{{
			Name:  "session_file",
			Type:  sources.SecretTypeFile,
			Value: []byte(secretValue),
		}}},
	})
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if jobService.job.Status != jobs.JobStatusCompleted {
		t.Fatalf("expected completed, got %s", jobService.job.Status)
	}
	for _, event := range jobService.events {
		if strings.Contains(event.MessageRedacted+event.MetadataRedacted, secretValue) {
			t.Fatalf("event leaked secret: %+v", event)
		}
	}
	for _, artifact := range jobService.artifacts {
		if strings.Contains(artifact.RelativePath+artifact.ArtifactType+artifact.SHA256, secretValue) {
			t.Fatalf("artifact leaked secret: %+v", artifact)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "job1", "work", "secrets", "session_file")); !os.IsNotExist(err) {
		t.Fatalf("expected secret workspace cleanup, stat err=%v", err)
	}
}

func TestRunnerSecretPreflightFailuresDoNotExecute(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "missing required secret", err: sources.ErrMissingRequiredSecrets, code: "missing_required_secrets"},
		{name: "revoked secret", err: sources.ErrSecretRevoked, code: "secret_revoked"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobService := &fakeJobService{job: jobs.ImportJob{
				ID:                 "job1",
				ConnectorSourceID:  "source1",
				ConnectorVersionID: "version1",
				Status:             jobs.JobStatusQueued,
				TriggerType:        "manual",
				AuthMode:           "reusable_session",
				ExecutionMode:      "unattended",
				CreatedAt:          time.Now(),
			}}
			r := New(jobService, fixtureExecutor{run: func(string, string) ExecutionResult {
				t.Fatal("executor should not run when secret preflight fails")
				return ExecutionResult{}
			}}, Config{WorkspaceRoot: t.TempDir(), SecretProvider: fakeSecretProvider{err: tt.err}})
			if err := r.RunOnce(context.Background()); err == nil {
				t.Fatal("expected secret preflight error")
			}
			if jobService.job.Status != jobs.JobStatusFailed || jobService.job.FailureCode != tt.code {
				t.Fatalf("expected %s failure, got %+v", tt.code, jobService.job)
			}
		})
	}
}

type fixtureExecutor struct {
	run func(inputPath string, outputDir string) ExecutionResult
}

func (e fixtureExecutor) Execute(_ context.Context, inputPath string, outputDir string) ExecutionResult {
	return e.run(inputPath, outputDir)
}

type fakeJobService struct {
	job         jobs.ImportJob
	claimed     bool
	cancelOnGet bool
	events      []jobs.ImportJobEvent
	artifacts   []jobs.ImportJobArtifact
}

type fakeSecretProvider struct {
	secrets []sources.RunnerSecret
	err     error
}

func (p fakeSecretProvider) ResolveRunnerSecrets(context.Context, string) ([]sources.RunnerSecret, error) {
	return p.secrets, p.err
}

func (s *fakeJobService) GetJob(context.Context, string) (jobs.ImportJob, error) {
	if s.cancelOnGet && s.job.CancellationRequestedAt == nil {
		now := time.Now()
		s.job.CancellationRequestedAt = &now
	}
	return s.job, nil
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
