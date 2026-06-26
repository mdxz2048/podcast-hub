package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

var ErrNoQueuedJob = errors.New("no queued import job")

type JobService interface {
	ClaimNextQueuedJob(ctx context.Context) (jobs.ImportJob, bool, error)
	AppendEvent(ctx context.Context, event jobs.ImportJobEvent) error
	AppendArtifact(ctx context.Context, artifact jobs.ImportJobArtifact) error
	TransitionJob(ctx context.Context, jobID string, status jobs.JobStatus, failureCode string, failureMessageRedacted string) (jobs.ImportJob, error)
}

type Executor interface {
	Execute(ctx context.Context, inputPath string, outputDir string) ExecutionResult
}

type ExecutionResult struct {
	Stdout   io.Reader
	ExitCode int
	Err      error
}

type Config struct {
	WorkspaceRoot  string
	ProtocolLimits ProtocolLimits
	ArtifactLimits ArtifactLimits
}

type Runner struct {
	jobs     JobService
	executor Executor
	config   Config
}

func New(jobService JobService, executor Executor, config Config) *Runner {
	if config.ProtocolLimits.MaxLineBytes == 0 {
		config.ProtocolLimits = DefaultProtocolLimits()
	}
	if config.ArtifactLimits.MaxArtifacts == 0 {
		config.ArtifactLimits = DefaultArtifactLimits()
	}
	return &Runner{jobs: jobService, executor: executor, config: config}
}

func (r *Runner) RunOnce(ctx context.Context) error {
	job, found, err := r.jobs.ClaimNextQueuedJob(ctx)
	if err != nil {
		return err
	}
	if !found {
		return ErrNoQueuedJob
	}
	workspace, inputPath, outputDir, err := r.prepareWorkspace(job)
	if err != nil {
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "workspace_prepare_failed", redact(err.Error()))
		return err
	}
	defer os.RemoveAll(workspace)

	result := r.executor.Execute(ctx, inputPath, outputDir)
	stdout := result.Stdout
	if stdout == nil {
		stdout = bytes.NewReader(nil)
	}
	protocol, parseErr := ParseJSONLines(job.ID, stdout, result.ExitCode, r.config.ProtocolLimits)
	for _, event := range protocol.Events {
		_ = r.jobs.AppendEvent(ctx, event)
	}
	if parseErr != nil {
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, failureCode(parseErr), redact(parseErr.Error()))
		return parseErr
	}
	if result.Err != nil {
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "executor_failed", redact(result.Err.Error()))
		return result.Err
	}
	artifacts, artifactErr := ValidateArtifacts(job.ID, outputDir, protocol.DeclaredArtifacts, r.config.ArtifactLimits)
	if artifactErr != nil {
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, failureCode(artifactErr), redact(artifactErr.Error()))
		return artifactErr
	}
	for _, artifact := range artifacts {
		_ = r.jobs.AppendArtifact(ctx, artifact)
	}
	if protocol.TerminalType == "completed" {
		_, err = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusCompleted, "", "")
		return err
	}
	_, err = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "connector_failed", "Connector reported failure.")
	return err
}

func (r *Runner) prepareWorkspace(job jobs.ImportJob) (string, string, string, error) {
	root := r.config.WorkspaceRoot
	if root == "" {
		root = filepath.Join(".local", "runner-workspaces")
	}
	workspace := filepath.Join(root, job.ID)
	inputDir := filepath.Join(workspace, "work", "input")
	outputDir := filepath.Join(workspace, "work", "output")
	if err := os.MkdirAll(inputDir, 0o700); err != nil {
		return "", "", "", fmt.Errorf("create input directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return "", "", "", fmt.Errorf("create output directory: %w", err)
	}
	input := map[string]any{
		"schema_version": "1.0",
		"job": map[string]any{
			"id":             job.ID,
			"ingestion_type": "connector",
			"trigger_type":   job.TriggerType,
			"auth_mode":      job.AuthMode,
			"execution_mode": job.ExecutionMode,
			"created_at":     job.CreatedAt.Format(time.RFC3339),
		},
		"source": map[string]any{
			"id": job.ConnectorSourceID,
		},
		"connector": map[string]any{
			"version_id": job.ConnectorVersionID,
		},
		"paths": map[string]any{
			"input_file": "/work/input/job.json",
			"output_dir": "/work/output",
		},
	}
	body, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", "", "", fmt.Errorf("encode job input: %w", err)
	}
	inputPath := filepath.Join(inputDir, "job.json")
	if err := os.WriteFile(inputPath, body, 0o600); err != nil {
		return "", "", "", fmt.Errorf("write job input: %w", err)
	}
	return workspace, inputPath, outputDir, nil
}

func failureCode(err error) string {
	switch {
	case errors.Is(err, ErrProtocolInvalidJSON):
		return "invalid_json_lines"
	case errors.Is(err, ErrProtocolUnknownEvent):
		return "unknown_event"
	case errors.Is(err, ErrProtocolMissingTerminal):
		return "missing_terminal_event"
	case errors.Is(err, ErrProtocolDuplicateTerminal):
		return "duplicate_terminal_event"
	case errors.Is(err, ErrProtocolEventAfterTerminal):
		return "event_after_terminal"
	case errors.Is(err, ErrProtocolExitMismatch):
		return "exit_code_mismatch"
	case errors.Is(err, ErrProtocolLineTooLong), errors.Is(err, ErrProtocolTooMuchOutput):
		return "stdout_limit_exceeded"
	case errors.Is(err, ErrProtocolTooManyEvents):
		return "event_limit_exceeded"
	case errors.Is(err, ErrArtifactPathEscape):
		return "artifact_path_escape"
	case errors.Is(err, ErrArtifactInvalidType):
		return "artifact_invalid_type"
	case errors.Is(err, ErrArtifactTooMany), errors.Is(err, ErrArtifactTooLarge):
		return "artifact_limit_exceeded"
	case errors.Is(err, ErrArtifactDuplicate):
		return "artifact_duplicate"
	case errors.Is(err, ErrArtifactUndeclared):
		return "artifact_undeclared"
	default:
		return "runner_failed"
	}
}
