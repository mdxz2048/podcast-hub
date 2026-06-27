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
	"strings"
	"time"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/sources"
)

var ErrNoQueuedJob = errors.New("no queued import job")

type JobService interface {
	GetJob(ctx context.Context, jobID string) (jobs.ImportJob, error)
	ClaimNextQueuedJob(ctx context.Context) (jobs.ImportJob, bool, error)
	AppendEvent(ctx context.Context, event jobs.ImportJobEvent) error
	AppendArtifact(ctx context.Context, artifact jobs.ImportJobArtifact) error
	TransitionJob(ctx context.Context, jobID string, status jobs.JobStatus, failureCode string, failureMessageRedacted string) (jobs.ImportJob, error)
}

type Executor interface {
	Execute(ctx context.Context, inputPath string, outputDir string) ExecutionResult
}

type SecretProvider interface {
	ResolveRunnerSecrets(ctx context.Context, sourceID string) ([]sources.RunnerSecret, error)
}

type ExecutionResult struct {
	Stdout   io.Reader
	ExitCode int
	Err      error
}

type Config struct {
	WorkspaceRoot            string
	ProtocolLimits           ProtocolLimits
	ArtifactLimits           ArtifactLimits
	ExecutionTimeout         time.Duration
	CancellationPollInterval time.Duration
	SecretProvider           SecretProvider
	ArtifactStoreRoot        string
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
	if config.ExecutionTimeout == 0 {
		config.ExecutionTimeout = 15 * time.Minute
	}
	if config.CancellationPollInterval == 0 {
		config.CancellationPollInterval = 250 * time.Millisecond
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
	secrets, err := r.resolveSecrets(ctx, job)
	if err != nil {
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, failureCode(err), redact(err.Error()))
		return err
	}
	workspace, inputPath, outputDir, err := r.prepareWorkspace(job, secrets)
	if err != nil {
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "workspace_prepare_failed", redact(err.Error()))
		return err
	}
	defer os.RemoveAll(workspace)

	execCtx, cancel := context.WithTimeout(ctx, r.executionTimeout(job))
	defer cancel()
	cancelled := make(chan struct{}, 1)
	stopPoll := make(chan struct{})
	go r.watchCancellation(execCtx, cancel, job.ID, cancelled, stopPoll)
	result := r.executor.Execute(execCtx, inputPath, outputDir)
	close(stopPoll)

	stdout := result.Stdout
	if stdout == nil {
		stdout = bytes.NewReader(nil)
	}
	protocol, parseErr := ParseJSONLines(job.ID, stdout, result.ExitCode, r.config.ProtocolLimits)
	for _, event := range protocol.Events {
		_ = r.jobs.AppendEvent(ctx, event)
	}
	if execCtx.Err() == context.DeadlineExceeded {
		_ = r.jobs.AppendEvent(ctx, jobs.ImportJobEvent{ImportJobID: job.ID, EventType: "runner.timeout", Level: "error", MessageRedacted: "Runner timed out and stopped execution."})
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "timeout", "Runner timed out.")
		return context.DeadlineExceeded
	}
	if wasCancelled(cancelled) {
		_ = r.jobs.AppendEvent(ctx, jobs.ImportJobEvent{ImportJobID: job.ID, EventType: "runner.cancelled", Level: "warning", MessageRedacted: "Runner stopped execution after cancellation request."})
		_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusCancelled, "", "")
		return context.Canceled
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
		if r.config.ArtifactStoreRoot != "" {
			storageKey, err := r.promoteArtifact(outputDir, artifact)
			if err != nil {
				_, _ = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "artifact_promote_failed", redact(err.Error()))
				return err
			}
			artifact.StorageKey = storageKey
		}
		_ = r.jobs.AppendArtifact(ctx, artifact)
	}
	if protocol.TerminalType == "completed" {
		_, err = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusCompleted, "", "")
		return err
	}
	_, err = r.jobs.TransitionJob(ctx, job.ID, jobs.JobStatusFailed, "connector_failed", "Connector reported failure.")
	return err
}

func (r *Runner) promoteArtifact(outputDir string, artifact jobs.ImportJobArtifact) (string, error) {
	rel, err := cleanRelativePath(artifact.RelativePath)
	if err != nil {
		return "", err
	}
	storageKey := filepath.Join(artifact.ImportJobID, rel)
	sourcePath := filepath.Join(outputDir, rel)
	targetPath := filepath.Join(r.config.ArtifactStoreRoot, storageKey)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
		return "", fmt.Errorf("create artifact store directory: %w", err)
	}
	body, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("read artifact for promotion: %w", err)
	}
	if err := os.WriteFile(targetPath, body, 0o600); err != nil {
		return "", fmt.Errorf("write artifact store file: %w", err)
	}
	return storageKey, nil
}

func cleanRelativePath(value string) (string, error) {
	cleaned := filepath.Clean(strings.TrimSpace(value))
	if cleaned == "." || filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return "", ErrArtifactPathEscape
	}
	return cleaned, nil
}

func (r *Runner) executionTimeout(job jobs.ImportJob) time.Duration {
	if job.TimeoutAt != nil {
		until := time.Until(*job.TimeoutAt)
		if until > 0 {
			return until
		}
		return time.Nanosecond
	}
	return r.config.ExecutionTimeout
}

func (r *Runner) watchCancellation(ctx context.Context, cancel context.CancelFunc, jobID string, cancelled chan<- struct{}, stop <-chan struct{}) {
	ticker := time.NewTicker(r.config.CancellationPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case <-ticker.C:
			job, err := r.jobs.GetJob(ctx, jobID)
			if err != nil {
				continue
			}
			if job.CancellationRequestedAt != nil {
				select {
				case cancelled <- struct{}{}:
				default:
				}
				cancel()
				return
			}
		}
	}
}

func wasCancelled(cancelled <-chan struct{}) bool {
	select {
	case <-cancelled:
		return true
	default:
		return false
	}
}

func (r *Runner) resolveSecrets(ctx context.Context, job jobs.ImportJob) ([]sources.RunnerSecret, error) {
	if r.config.SecretProvider == nil {
		return nil, nil
	}
	secrets, err := r.config.SecretProvider.ResolveRunnerSecrets(ctx, job.ConnectorSourceID)
	if err != nil {
		return nil, fmt.Errorf("resolve runner secrets: %w", err)
	}
	return secrets, nil
}

func (r *Runner) prepareWorkspace(job jobs.ImportJob, secrets []sources.RunnerSecret) (string, string, string, error) {
	root := r.config.WorkspaceRoot
	if root == "" {
		root = filepath.Join(".local", "runner-workspaces")
	}
	workspace := filepath.Join(root, job.ID)
	inputDir := filepath.Join(workspace, "work", "input")
	outputDir := filepath.Join(workspace, "work", "output")
	secretsDir := filepath.Join(workspace, "work", "secrets")
	if err := os.MkdirAll(inputDir, 0o700); err != nil {
		return "", "", "", fmt.Errorf("create input directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return "", "", "", fmt.Errorf("create output directory: %w", err)
	}
	secretRefs := make([]map[string]string, 0, len(secrets))
	if len(secrets) > 0 {
		if err := os.MkdirAll(secretsDir, 0o700); err != nil {
			return "", "", "", fmt.Errorf("create secrets directory: %w", err)
		}
		for _, secret := range secrets {
			name, err := cleanSecretName(secret.Name)
			if err != nil {
				return "", "", "", err
			}
			secretPath := filepath.Join(secretsDir, name)
			if err := os.WriteFile(secretPath, secret.Value, 0o400); err != nil {
				return "", "", "", fmt.Errorf("write runner secret: %w", err)
			}
			secretRefs = append(secretRefs, map[string]string{
				"name": name,
				"type": string(secret.Type),
				"path": "/work/secrets/" + name,
			})
		}
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
			"input_file":  "/work/input/job.json",
			"output_dir":  "/work/output",
			"secrets_dir": "/work/secrets",
		},
		"secrets": secretRefs,
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

func cleanSecretName(name string) (string, error) {
	if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return "", errors.New("runner secret name is invalid")
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return "", errors.New("runner secret name is invalid")
	}
	return name, nil
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
	case errors.Is(err, sources.ErrMissingRequiredSecrets):
		return "missing_required_secrets"
	case errors.Is(err, sources.ErrSecretRevoked):
		return "secret_revoked"
	default:
		return "runner_failed"
	}
}
