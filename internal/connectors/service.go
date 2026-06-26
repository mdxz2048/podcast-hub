package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store        Store
	packageStore ConnectorPackageStore
	limits       ValidationLimits
}

func NewService(store Store, packageStore ConnectorPackageStore) *Service {
	return &Service{
		store:        store,
		packageStore: packageStore,
		limits:       DefaultValidationLimits(),
	}
}

type UploadInput struct {
	ConnectorID string
	Version     string
	PackageName string
	Content     io.Reader
	UploadedBy  *string
}

type UploadResult struct {
	Connector Connector
	Version   ConnectorVersion
	Summary   ValidationSummary
}

func (s *Service) ListConnectors(ctx context.Context) ([]Connector, error) {
	return s.store.ListConnectors(ctx)
}

func (s *Service) GetConnector(ctx context.Context, connectorID string) (Connector, error) {
	connector, found, err := s.store.GetConnector(ctx, connectorID)
	if err != nil {
		return Connector{}, fmt.Errorf("get connector: %w", err)
	}
	if !found {
		return Connector{}, ErrConnectorNotFound
	}
	return connector, nil
}

func (s *Service) ListVersions(ctx context.Context, connectorID string) ([]ConnectorVersion, error) {
	if _, found, err := s.store.GetConnector(ctx, connectorID); err != nil {
		return nil, fmt.Errorf("check connector: %w", err)
	} else if !found {
		return nil, ErrConnectorNotFound
	}
	return s.store.ListConnectorVersions(ctx, connectorID)
}

func (s *Service) GetVersion(ctx context.Context, versionID string) (ConnectorVersion, error) {
	version, found, err := s.store.GetConnectorVersion(ctx, versionID)
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("get connector version: %w", err)
	}
	if !found {
		return ConnectorVersion{}, ErrVersionNotFound
	}
	return version, nil
}

func (s *Service) Upload(ctx context.Context, in UploadInput) (UploadResult, error) {
	connectorID := strings.TrimSpace(in.ConnectorID)
	version := strings.TrimSpace(in.Version)
	if !slugPattern.MatchString(connectorID) {
		return UploadResult{}, ErrInvalidConnectorID
	}
	if !semverPattern.MatchString(version) {
		return UploadResult{}, ErrInvalidVersion
	}
	if in.Content == nil {
		return UploadResult{}, ErrInvalidUpload
	}
	connector, found, err := s.store.FindConnectorBySlug(ctx, connectorID)
	if err != nil {
		return UploadResult{}, fmt.Errorf("find connector by slug: %w", err)
	}
	if found {
		if _, exists, err := s.store.GetConnectorVersionByVersion(ctx, connector.ID, version); err != nil {
			return UploadResult{}, fmt.Errorf("find existing version: %w", err)
		} else if exists {
			return UploadResult{}, ErrVersionAlreadyExists
		}
	}
	packageName := strings.TrimSpace(in.PackageName)
	if packageName == "" {
		packageName = fmt.Sprintf("%s-%s-%s.zip", connectorID, version, uuid.NewString())
	}
	packageRef, err := s.packageStore.PutQuarantine(ctx, packageName, in.Content)
	if err != nil {
		return UploadResult{}, fmt.Errorf("store uploaded package: %w", err)
	}
	keepPackage := false
	defer func() {
		if !keepPackage {
			_ = s.packageStore.Delete(ctx, packageRef)
		}
	}()
	packageReader, err := s.packageStore.Read(ctx, packageRef)
	if err != nil {
		return UploadResult{}, fmt.Errorf("read uploaded package: %w", err)
	}
	defer packageReader.Close()
	zipBytes, err := io.ReadAll(packageReader)
	if err != nil {
		return UploadResult{}, fmt.Errorf("read package bytes: %w", err)
	}
	validation := ValidatePackageZip(zipBytes, connectorID, version, s.limits)
	if hasValidationIssue(validation.Summary, "zip_invalid") {
		return UploadResult{}, ErrInvalidUpload
	}
	validationSummaryJSON, err := json.Marshal(validation.Summary)
	if err != nil {
		return UploadResult{}, fmt.Errorf("marshal validation summary: %w", err)
	}
	now := time.Now()
	if !found {
		name := strings.TrimSpace(validation.Manifest.Name)
		if name == "" {
			name = connectorID
		}
		connector, err = s.store.CreateConnector(ctx, Connector{
			ID:          uuid.NewString(),
			Slug:        connectorID,
			Name:        name,
			Description: validation.Manifest.Description,
			Status:      ConnectorStatusActive,
			CreatedBy:   in.UploadedBy,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			return UploadResult{}, fmt.Errorf("create connector: %w", err)
		}
	}
	if _, exists, err := s.store.GetConnectorVersionByVersion(ctx, connector.ID, version); err != nil {
		return UploadResult{}, fmt.Errorf("find existing version: %w", err)
	} else if exists {
		return UploadResult{}, ErrVersionAlreadyExists
	}

	reviewStatus := ReviewStatusPendingReview
	if !validation.Summary.IsValid {
		reviewStatus = ReviewStatusRejected
	}
	manifestJSON := validation.ManifestJSON
	if strings.TrimSpace(manifestJSON) == "" {
		manifestJSON = "{}"
	}
	versionModel, err := s.store.CreateConnectorVersion(ctx, ConnectorVersion{
		ID:                    uuid.NewString(),
		ConnectorID:           connector.ID,
		Version:               version,
		ReviewStatus:          reviewStatus,
		RuntimeProfile:        validation.Manifest.Runtime.Profile,
		Entrypoint:            validation.Manifest.Runtime.Entrypoint,
		ManifestJSON:          manifestJSON,
		PackageSHA256:         packageRef.SHA256,
		PackageSizeBytes:      packageRef.SizeBytes,
		PackageStorageKey:     packageRef.StorageKey,
		ValidationSummaryJSON: string(validationSummaryJSON),
		UploadedBy:            in.UploadedBy,
		CreatedAt:             now,
	})
	if err != nil {
		return UploadResult{}, fmt.Errorf("create connector version: %w", err)
	}
	connectorIDCopy := connector.ID
	versionIDCopy := versionModel.ID
	_ = s.store.InsertConnectorEvent(ctx, ConnectorEvent{
		ConnectorID:        &connectorIDCopy,
		ConnectorVersionID: &versionIDCopy,
		ActorUserID:        in.UploadedBy,
		EventType:          "connector.version_uploaded",
		DetailJSON:         string(validationSummaryJSON),
	})
	keepPackage = true
	return UploadResult{
		Connector: connector,
		Version:   versionModel,
		Summary:   validation.Summary,
	}, nil
}

func (s *Service) ApproveVersion(ctx context.Context, versionID string, actorUserID *string) (ConnectorVersion, error) {
	version, found, err := s.store.GetConnectorVersion(ctx, versionID)
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("get connector version: %w", err)
	}
	if !found {
		return ConnectorVersion{}, ErrVersionNotFound
	}
	if version.ReviewStatus != ReviewStatusPendingReview {
		return ConnectorVersion{}, ErrVersionNotPendingReview
	}
	connector, found, err := s.store.GetConnector(ctx, version.ConnectorID)
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("get connector: %w", err)
	}
	if !found {
		return ConnectorVersion{}, ErrConnectorNotFound
	}
	if connector.Status == ConnectorStatusDisabled {
		return ConnectorVersion{}, ErrConnectorDisabled
	}
	newRef, err := s.packageStore.PromoteApproved(ctx, PackageRef{
		StorageKey: version.PackageStorageKey,
		SizeBytes:  version.PackageSizeBytes,
		SHA256:     version.PackageSHA256,
	})
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("promote package: %w", err)
	}
	now := time.Now()
	updated, err := s.store.UpdateConnectorVersionReview(ctx, versionID, UpdateVersionReviewInput{
		ReviewStatus:      ReviewStatusApproved,
		ReviewedBy:        actorUserID,
		ReviewedAt:        &now,
		PackageStorageKey: &newRef.StorageKey,
	})
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("update connector version: %w", err)
	}
	versionIDCopy := updated.ID
	connectorIDCopy := connector.ID
	_ = s.store.InsertConnectorEvent(ctx, ConnectorEvent{
		ConnectorID:        &connectorIDCopy,
		ConnectorVersionID: &versionIDCopy,
		ActorUserID:        actorUserID,
		EventType:          "connector.version_approved",
		DetailJSON:         `{"review_status":"approved"}`,
	})
	return updated, nil
}

func hasValidationIssue(summary ValidationSummary, code string) bool {
	for _, issue := range summary.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func (s *Service) RejectVersion(ctx context.Context, versionID string, actorUserID *string) (ConnectorVersion, error) {
	version, found, err := s.store.GetConnectorVersion(ctx, versionID)
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("get connector version: %w", err)
	}
	if !found {
		return ConnectorVersion{}, ErrVersionNotFound
	}
	if version.ReviewStatus != ReviewStatusPendingReview {
		return ConnectorVersion{}, ErrVersionNotPendingReview
	}
	now := time.Now()
	updated, err := s.store.UpdateConnectorVersionReview(ctx, versionID, UpdateVersionReviewInput{
		ReviewStatus: ReviewStatusRejected,
		ReviewedBy:   actorUserID,
		ReviewedAt:   &now,
	})
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("update connector version: %w", err)
	}
	versionIDCopy := updated.ID
	connectorIDCopy := updated.ConnectorID
	_ = s.store.InsertConnectorEvent(ctx, ConnectorEvent{
		ConnectorID:        &connectorIDCopy,
		ConnectorVersionID: &versionIDCopy,
		ActorUserID:        actorUserID,
		EventType:          "connector.version_rejected",
		DetailJSON:         `{"review_status":"rejected"}`,
	})
	return updated, nil
}

func (s *Service) DisableVersion(ctx context.Context, versionID string, actorUserID *string) (ConnectorVersion, error) {
	version, found, err := s.store.GetConnectorVersion(ctx, versionID)
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("get connector version: %w", err)
	}
	if !found {
		return ConnectorVersion{}, ErrVersionNotFound
	}
	if version.ReviewStatus != ReviewStatusApproved {
		return ConnectorVersion{}, ErrVersionNotApproved
	}
	now := time.Now()
	updated, err := s.store.UpdateConnectorVersionReview(ctx, versionID, UpdateVersionReviewInput{
		ReviewStatus: ReviewStatusDisabled,
		ReviewedBy:   actorUserID,
		ReviewedAt:   &now,
	})
	if err != nil {
		return ConnectorVersion{}, fmt.Errorf("update connector version: %w", err)
	}
	versionIDCopy := updated.ID
	connectorIDCopy := updated.ConnectorID
	_ = s.store.InsertConnectorEvent(ctx, ConnectorEvent{
		ConnectorID:        &connectorIDCopy,
		ConnectorVersionID: &versionIDCopy,
		ActorUserID:        actorUserID,
		EventType:          "connector.version_disabled",
		DetailJSON:         `{"review_status":"disabled"}`,
	})
	return updated, nil
}

func (s *Service) SetConnectorStatus(ctx context.Context, connectorID string, status ConnectorStatus, actorUserID *string) (Connector, error) {
	connector, found, err := s.store.GetConnector(ctx, connectorID)
	if err != nil {
		return Connector{}, fmt.Errorf("get connector: %w", err)
	}
	if !found {
		return Connector{}, ErrConnectorNotFound
	}
	if connector.Status == status {
		return connector, nil
	}
	updated, err := s.store.UpdateConnectorStatus(ctx, connectorID, status)
	if err != nil {
		return Connector{}, fmt.Errorf("update connector status: %w", err)
	}
	connectorIDCopy := updated.ID
	_ = s.store.InsertConnectorEvent(ctx, ConnectorEvent{
		ConnectorID: &connectorIDCopy,
		ActorUserID: actorUserID,
		EventType:   "connector.status_changed",
		DetailJSON:  fmt.Sprintf(`{"status":"%s"}`, status),
	})
	return updated, nil
}
