package intake

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mdxz2048/podcast-hub/internal/content"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

const maxText = 4000
const maxTitle = 240
const maxMetadataBytes = 8192

type JobService interface {
	GetJob(ctx context.Context, jobID string) (jobs.ImportJob, error)
	ListArtifacts(ctx context.Context, jobID string) ([]jobs.ImportJobArtifact, error)
}

type ArtifactReader interface {
	ReadArtifact(ctx context.Context, artifact jobs.ImportJobArtifact) ([]byte, error)
}

type Service struct {
	jobs      JobService
	content   content.Store
	artifacts ArtifactReader
}

func NewService(jobService JobService, contentStore content.Store, artifactReader ArtifactReader) *Service {
	return &Service{jobs: jobService, content: contentStore, artifacts: artifactReader}
}

type Result struct {
	IntakeRun content.IntakeRun `json:"intake_run"`
	Program   *content.Program  `json:"program,omitempty"`
	Issues    []string          `json:"validation_issues,omitempty"`
}

func (s *Service) Status(ctx context.Context, jobID string) (content.IntakeRun, bool, error) {
	return s.content.GetIntakeRun(ctx, jobID)
}

func (s *Service) Run(ctx context.Context, jobID string) (Result, error) {
	if existing, found, err := s.content.GetIntakeRun(ctx, jobID); err != nil {
		return Result{}, err
	} else if found && existing.Status == "succeeded" {
		return Result{IntakeRun: existing}, nil
	}
	job, err := s.jobs.GetJob(ctx, jobID)
	if err != nil {
		return Result{}, err
	}
	if job.Status != jobs.JobStatusCompleted {
		return Result{}, ErrJobNotCompleted
	}
	artifacts, err := s.jobs.ListArtifacts(ctx, jobID)
	if err != nil {
		return Result{}, err
	}
	bundleArtifact, found := findBundleArtifact(artifacts)
	if !found {
		run, _ := s.recordFailure(ctx, jobID, []string{"metadata_bundle artifact is missing"})
		return Result{IntakeRun: run, Issues: []string{"metadata_bundle artifact is missing"}}, ErrBundleMissing
	}
	body, err := s.artifacts.ReadArtifact(ctx, bundleArtifact)
	if err != nil {
		run, _ := s.recordFailure(ctx, jobID, []string{"metadata_bundle could not be read"})
		return Result{IntakeRun: run, Issues: []string{"metadata_bundle could not be read"}}, ErrBundleInvalid
	}
	bundle, issues := parseBundle(body, artifacts)
	if len(issues) > 0 {
		run, _ := s.recordFailure(ctx, jobID, issues)
		return Result{IntakeRun: run, Issues: issues}, ErrBundleInvalid
	}
	program, err := s.content.UpsertProgramFromSource(ctx, content.UpsertProgramInput{
		SourceID:          job.ConnectorSourceID,
		JobID:             job.ID,
		ExternalProgramID: bundle.Program.ExternalID,
		Title:             bundle.Program.Title,
		Description:       bundle.Program.Description,
		Author:            bundle.Program.Author,
		Language:          bundle.Program.Language,
	})
	if err != nil {
		return Result{}, err
	}
	if _, err := s.content.CreateOrKeepPendingReview(ctx, "program", program.ID, "metadata", job.ID); err != nil {
		return Result{}, err
	}
	_ = s.content.InsertPublicationEvent(ctx, content.PublicationEvent{ID: uuid.NewString(), TargetType: "program", TargetID: program.ID, EventType: "submitted_for_review", MetadataRedacted: `{}`, CreatedAt: time.Now()})
	artifactByPath := artifactMap(artifacts)
	for _, ep := range bundle.Episodes {
		episode, err := s.content.UpsertEpisode(ctx, content.UpsertEpisodeInput{
			ProgramID:         program.ID,
			ExternalEpisodeID: ep.ExternalID,
			Title:             ep.Title,
			Description:       ep.Description,
			PublishedAt:       ep.PublishedAt,
			DurationSeconds:   ep.DurationSeconds,
			SourceJobID:       job.ID,
		})
		if err != nil {
			return Result{}, err
		}
		if _, err := s.content.CreateOrKeepPendingReview(ctx, "episode", episode.ID, "metadata", job.ID); err != nil {
			return Result{}, err
		}
		_ = s.content.InsertPublicationEvent(ctx, content.PublicationEvent{ID: uuid.NewString(), TargetType: "episode", TargetID: episode.ID, EventType: "submitted_for_review", MetadataRedacted: `{}`, CreatedAt: time.Now()})
		for _, ref := range []struct {
			path string
			kind string
		}{{ep.AudioArtifact, "audio"}, {ep.CoverArtifact, "cover"}} {
			if ref.path == "" {
				continue
			}
			artifact := artifactByPath[ref.path]
			if _, err := s.content.CreateMediaAsset(ctx, content.CreateMediaAssetInput{OwnerType: "episode", OwnerID: episode.ID, ImportJobID: job.ID, ArtifactID: artifact.ID, MediaKind: ref.kind, StagedStorageKey: artifact.StorageKey, ContentType: "application/octet-stream", SizeBytes: artifact.SizeBytes, SHA256: artifact.SHA256}); err != nil {
				return Result{}, err
			}
		}
	}
	run := content.IntakeRun{ID: uuid.NewString(), ImportJobID: jobID, Status: "succeeded", ValidationIssuesRedacted: `[]`, ProgramID: &program.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	run, err = s.content.UpsertIntakeRun(ctx, run)
	if err != nil {
		return Result{}, err
	}
	return Result{IntakeRun: run, Program: &program}, nil
}

func (s *Service) recordFailure(ctx context.Context, jobID string, issues []string) (content.IntakeRun, error) {
	body, _ := json.Marshal(issues)
	now := time.Now()
	return s.content.UpsertIntakeRun(ctx, content.IntakeRun{ID: uuid.NewString(), ImportJobID: jobID, Status: "failed", ValidationIssuesRedacted: string(body), CreatedAt: now, UpdatedAt: now})
}

type bundle struct {
	SchemaVersion int             `json:"schema_version"`
	Program       bundleProgram   `json:"program"`
	Episodes      []bundleEpisode `json:"episodes"`
}

type bundleProgram struct {
	ExternalID    string `json:"external_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Language      string `json:"language"`
	Author        string `json:"author"`
	CoverArtifact string `json:"cover_artifact"`
}

type bundleEpisode struct {
	ExternalID      string          `json:"external_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	PublishedAt     time.Time       `json:"published_at"`
	DurationSeconds int             `json:"duration_seconds"`
	AudioArtifact   string          `json:"audio_artifact"`
	CoverArtifact   string          `json:"cover_artifact"`
	Metadata        json.RawMessage `json:"metadata"`
}

func parseBundle(body []byte, artifacts []jobs.ImportJobArtifact) (bundle, []string) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var parsed bundle
	if err := decoder.Decode(&parsed); err != nil {
		return bundle{}, []string{"metadata_bundle schema is invalid"}
	}
	var issues []string
	if parsed.SchemaVersion != 1 {
		issues = append(issues, "schema_version must be 1")
	}
	if invalidText(parsed.Program.ExternalID, maxTitle) || invalidText(parsed.Program.Title, maxTitle) || invalidText(parsed.Program.Description, maxText) || invalidText(parsed.Program.Author, maxTitle) || invalidText(parsed.Program.Language, 32) {
		issues = append(issues, "program fields are invalid")
	}
	artifactByPath := artifactMap(artifacts)
	for _, ref := range []string{parsed.Program.CoverArtifact} {
		if ref != "" {
			issues = append(issues, validateArtifactRef(ref, artifactByPath)...)
		}
	}
	if len(parsed.Episodes) == 0 {
		issues = append(issues, "episodes must not be empty")
	}
	for _, ep := range parsed.Episodes {
		if invalidText(ep.ExternalID, maxTitle) || invalidText(ep.Title, maxTitle) || invalidText(ep.Description, maxText) || ep.DurationSeconds < 0 {
			issues = append(issues, "episode fields are invalid")
		}
		issues = append(issues, validateArtifactRef(ep.AudioArtifact, artifactByPath)...)
		if ep.CoverArtifact != "" {
			issues = append(issues, validateArtifactRef(ep.CoverArtifact, artifactByPath)...)
		}
		if len(ep.Metadata) > maxMetadataBytes || containsForbiddenMetadataKey(ep.Metadata) {
			issues = append(issues, "episode metadata is invalid")
		}
	}
	return parsed, issues
}

func findBundleArtifact(artifacts []jobs.ImportJobArtifact) (jobs.ImportJobArtifact, bool) {
	for _, artifact := range artifacts {
		if artifact.ArtifactType == "metadata_bundle" {
			return artifact, true
		}
	}
	return jobs.ImportJobArtifact{}, false
}

func artifactMap(artifacts []jobs.ImportJobArtifact) map[string]jobs.ImportJobArtifact {
	items := map[string]jobs.ImportJobArtifact{}
	for _, artifact := range artifacts {
		items[artifact.RelativePath] = artifact
	}
	return items
}

func validateArtifactRef(ref string, artifacts map[string]jobs.ImportJobArtifact) []string {
	if ref == "" || filepath.IsAbs(ref) || strings.Contains(ref, "..") {
		return []string{"artifact reference is invalid"}
	}
	if parsed, err := url.Parse(ref); err == nil && parsed.Scheme != "" {
		return []string{"artifact reference must not be URL"}
	}
	if _, ok := artifacts[ref]; !ok {
		return []string{"artifact reference is not registered"}
	}
	return nil
}

func invalidText(value string, max int) bool {
	value = strings.TrimSpace(value)
	return value == "" || len([]rune(value)) > max
}

func containsForbiddenMetadataKey(raw json.RawMessage) bool {
	lower := strings.ToLower(string(raw))
	for _, marker := range []string{"secret", "token", "cookie", "authorization", "session", "password"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

type FileArtifactReader struct{ Root string }

func (r FileArtifactReader) ReadArtifact(_ context.Context, artifact jobs.ImportJobArtifact) ([]byte, error) {
	if artifact.StorageKey == "" {
		return nil, fmt.Errorf("artifact storage key missing")
	}
	cleaned := filepath.Clean(artifact.StorageKey)
	if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return nil, fmt.Errorf("artifact storage key invalid")
	}
	return osReadFile(filepath.Join(r.Root, cleaned))
}

var osReadFile = func(path string) ([]byte, error) { return os.ReadFile(path) }
