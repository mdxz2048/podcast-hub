package runner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
)

var (
	ErrProtocolInvalidJSON        = errors.New("invalid JSON Lines event")
	ErrProtocolUnknownEvent       = errors.New("unknown JSON Lines event type")
	ErrProtocolMissingTerminal    = errors.New("missing terminal JSON Lines event")
	ErrProtocolDuplicateTerminal  = errors.New("duplicate terminal JSON Lines event")
	ErrProtocolEventAfterTerminal = errors.New("event after terminal JSON Lines event")
	ErrProtocolExitMismatch       = errors.New("exit code does not match terminal event")
	ErrProtocolLineTooLong        = errors.New("JSON Lines event exceeds maximum line length")
	ErrProtocolTooManyEvents      = errors.New("JSON Lines event count exceeds maximum")
	ErrProtocolTooMuchOutput      = errors.New("JSON Lines output exceeds maximum")
)

type ProtocolLimits struct {
	MaxLineBytes  int
	MaxTotalBytes int
	MaxEventCount int
}

func DefaultProtocolLimits() ProtocolLimits {
	return ProtocolLimits{MaxLineBytes: 64 * 1024, MaxTotalBytes: 2 * 1024 * 1024, MaxEventCount: 1000}
}

type ProtocolEvent struct {
	Type         string         `json:"type"`
	Level        string         `json:"level,omitempty"`
	Message      string         `json:"message,omitempty"`
	ArtifactType string         `json:"artifact_type,omitempty"`
	Path         string         `json:"path,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type DeclaredArtifact struct {
	ArtifactType string
	RelativePath string
}

type ProtocolResult struct {
	Events            []jobs.ImportJobEvent
	DeclaredArtifacts []DeclaredArtifact
	TerminalType      string
}

func ParseJSONLines(jobID string, stdout io.Reader, exitCode int, limits ProtocolLimits) (ProtocolResult, error) {
	if limits.MaxLineBytes <= 0 || limits.MaxTotalBytes <= 0 || limits.MaxEventCount <= 0 {
		limits = DefaultProtocolLimits()
	}
	reader := bufio.NewReader(stdout)
	result := ProtocolResult{}
	total := 0
	terminalSeen := false

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			total += len(line)
			if total > limits.MaxTotalBytes {
				return result, ErrProtocolTooMuchOutput
			}
			if len(line) > limits.MaxLineBytes {
				return result, ErrProtocolLineTooLong
			}
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) == 0 {
				if err == io.EOF {
					break
				}
				continue
			}
			if len(result.Events) >= limits.MaxEventCount {
				return result, ErrProtocolTooManyEvents
			}
			event, parseErr := parseProtocolEvent(trimmed)
			if parseErr != nil {
				return result, parseErr
			}
			if terminalSeen {
				if event.Type == "completed" || event.Type == "failed" {
					return result, ErrProtocolDuplicateTerminal
				}
				return result, ErrProtocolEventAfterTerminal
			}
			persisted, convErr := event.toJobEvent(jobID)
			if convErr != nil {
				return result, convErr
			}
			result.Events = append(result.Events, persisted)
			if event.Type == "artifact_ready" {
				result.DeclaredArtifacts = append(result.DeclaredArtifacts, DeclaredArtifact{ArtifactType: event.ArtifactType, RelativePath: event.Path})
			}
			if event.Type == "completed" || event.Type == "failed" {
				if terminalSeen {
					return result, ErrProtocolDuplicateTerminal
				}
				terminalSeen = true
				result.TerminalType = event.Type
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, fmt.Errorf("read JSON Lines: %w", err)
		}
	}
	if result.TerminalType == "" {
		return result, ErrProtocolMissingTerminal
	}
	if (result.TerminalType == "completed" && exitCode != 0) || (result.TerminalType == "failed" && exitCode == 0) {
		return result, ErrProtocolExitMismatch
	}
	return result, nil
}

func parseProtocolEvent(line []byte) (ProtocolEvent, error) {
	var event ProtocolEvent
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return ProtocolEvent{}, ErrProtocolInvalidJSON
	}
	switch event.Type {
	case "log", "progress", "artifact_ready", "completed", "failed":
	default:
		return ProtocolEvent{}, ErrProtocolUnknownEvent
	}
	if event.Type == "artifact_ready" && (strings.TrimSpace(event.ArtifactType) == "" || strings.TrimSpace(event.Path) == "") {
		return ProtocolEvent{}, ErrProtocolUnknownEvent
	}
	return event, nil
}

func (e ProtocolEvent) toJobEvent(jobID string) (jobs.ImportJobEvent, error) {
	level := strings.TrimSpace(e.Level)
	if level == "" {
		level = "info"
	}
	metadata := "{}"
	if len(e.Metadata) > 0 {
		body, err := json.Marshal(e.Metadata)
		if err != nil {
			return jobs.ImportJobEvent{}, fmt.Errorf("encode event metadata: %w", err)
		}
		metadata = string(body)
	}
	if e.Type == "artifact_ready" {
		body, _ := json.Marshal(map[string]string{"artifact_type": e.ArtifactType, "relative_path": e.Path})
		metadata = string(body)
	}
	return jobs.ImportJobEvent{
		ID:               uuid.NewString(),
		ImportJobID:      jobID,
		EventType:        "connector." + e.Type,
		Level:            level,
		MessageRedacted:  redact(e.Message),
		MetadataRedacted: redact(metadata),
		CreatedAt:        time.Now(),
	}, nil
}
