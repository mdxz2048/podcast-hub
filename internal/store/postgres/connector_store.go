package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/connectors"
)

type ConnectorStore struct {
	pool *pgxpool.Pool
}

func NewConnectorStore(pool *pgxpool.Pool) *ConnectorStore {
	return &ConnectorStore{pool: pool}
}

func (s *ConnectorStore) FindConnectorBySlug(ctx context.Context, slug string) (connectors.Connector, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, slug, name, description, status, created_by::text, created_at, updated_at
		FROM connectors
		WHERE slug=$1
	`, slug)
	connector, err := scanConnector(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return connectors.Connector{}, false, nil
		}
		return connectors.Connector{}, false, fmt.Errorf("find connector by slug: %w", err)
	}
	return connector, true, nil
}

func (s *ConnectorStore) CreateConnector(ctx context.Context, connector connectors.Connector) (connectors.Connector, error) {
	var createdBy any
	if connector.CreatedBy != nil && *connector.CreatedBy != "" {
		createdBy = *connector.CreatedBy
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO connectors(id, slug, name, description, status, created_by, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, $6::uuid, $7, $8)
		RETURNING id::text, slug, name, description, status, created_by::text, created_at, updated_at
	`, connector.ID, connector.Slug, connector.Name, connector.Description, connector.Status, createdBy, connector.CreatedAt, connector.UpdatedAt)
	created, err := scanConnector(row)
	if err != nil {
		return connectors.Connector{}, fmt.Errorf("create connector: %w", err)
	}
	return created, nil
}

func (s *ConnectorStore) UpdateConnectorStatus(ctx context.Context, connectorID string, status connectors.ConnectorStatus) (connectors.Connector, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE connectors
		SET status=$1, updated_at=$2
		WHERE id=$3::uuid
		RETURNING id::text, slug, name, description, status, created_by::text, created_at, updated_at
	`, status, time.Now(), connectorID)
	updated, err := scanConnector(row)
	if err != nil {
		return connectors.Connector{}, fmt.Errorf("update connector status: %w", err)
	}
	return updated, nil
}

func (s *ConnectorStore) GetConnector(ctx context.Context, connectorID string) (connectors.Connector, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, slug, name, description, status, created_by::text, created_at, updated_at
		FROM connectors
		WHERE id=$1::uuid
	`, connectorID)
	connector, err := scanConnector(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return connectors.Connector{}, false, nil
		}
		return connectors.Connector{}, false, fmt.Errorf("get connector: %w", err)
	}
	return connector, true, nil
}

func (s *ConnectorStore) ListConnectors(ctx context.Context) ([]connectors.Connector, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, slug, name, description, status, created_by::text, created_at, updated_at
		FROM connectors
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list connectors: %w", err)
	}
	defer rows.Close()
	items := make([]connectors.Connector, 0)
	for rows.Next() {
		connector, scanErr := scanConnector(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan connector: %w", scanErr)
		}
		items = append(items, connector)
	}
	return items, nil
}

func (s *ConnectorStore) ListConnectorVersions(ctx context.Context, connectorID string) ([]connectors.ConnectorVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, connector_id::text, version, review_status, runtime_profile, entrypoint, manifest_json::text, package_sha256, package_size_bytes, package_storage_key,
		       validation_summary_json::text, uploaded_by::text, reviewed_by::text, reviewed_at, created_at
		FROM connector_versions
		WHERE connector_id=$1::uuid
		ORDER BY created_at DESC
	`, connectorID)
	if err != nil {
		return nil, fmt.Errorf("list connector versions: %w", err)
	}
	defer rows.Close()
	items := make([]connectors.ConnectorVersion, 0)
	for rows.Next() {
		version, scanErr := scanConnectorVersion(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan connector version: %w", scanErr)
		}
		items = append(items, version)
	}
	return items, nil
}

func (s *ConnectorStore) GetConnectorVersion(ctx context.Context, versionID string) (connectors.ConnectorVersion, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, connector_id::text, version, review_status, runtime_profile, entrypoint, manifest_json::text, package_sha256, package_size_bytes, package_storage_key,
		       validation_summary_json::text, uploaded_by::text, reviewed_by::text, reviewed_at, created_at
		FROM connector_versions
		WHERE id=$1::uuid
	`, versionID)
	version, err := scanConnectorVersion(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return connectors.ConnectorVersion{}, false, nil
		}
		return connectors.ConnectorVersion{}, false, fmt.Errorf("get connector version: %w", err)
	}
	return version, true, nil
}

func (s *ConnectorStore) GetConnectorVersionByVersion(ctx context.Context, connectorID, version string) (connectors.ConnectorVersion, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, connector_id::text, version, review_status, runtime_profile, entrypoint, manifest_json::text, package_sha256, package_size_bytes, package_storage_key,
		       validation_summary_json::text, uploaded_by::text, reviewed_by::text, reviewed_at, created_at
		FROM connector_versions
		WHERE connector_id=$1::uuid AND version=$2
	`, connectorID, version)
	model, err := scanConnectorVersion(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return connectors.ConnectorVersion{}, false, nil
		}
		return connectors.ConnectorVersion{}, false, fmt.Errorf("get connector version by version: %w", err)
	}
	return model, true, nil
}

func (s *ConnectorStore) CreateConnectorVersion(ctx context.Context, version connectors.ConnectorVersion) (connectors.ConnectorVersion, error) {
	var uploadedBy any
	if version.UploadedBy != nil && *version.UploadedBy != "" {
		uploadedBy = *version.UploadedBy
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO connector_versions(id, connector_id, version, review_status, runtime_profile, entrypoint, manifest_json, package_sha256, package_size_bytes, package_storage_key,
		                               validation_summary_json, uploaded_by, created_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11::jsonb, $12::uuid, $13)
		RETURNING id::text, connector_id::text, version, review_status, runtime_profile, entrypoint, manifest_json::text, package_sha256, package_size_bytes, package_storage_key,
		          validation_summary_json::text, uploaded_by::text, reviewed_by::text, reviewed_at, created_at
	`, version.ID, version.ConnectorID, version.Version, version.ReviewStatus, version.RuntimeProfile, version.Entrypoint, version.ManifestJSON,
		version.PackageSHA256, version.PackageSizeBytes, version.PackageStorageKey, version.ValidationSummaryJSON, uploadedBy, version.CreatedAt)
	created, err := scanConnectorVersion(row)
	if err != nil {
		return connectors.ConnectorVersion{}, fmt.Errorf("create connector version: %w", err)
	}
	return created, nil
}

func (s *ConnectorStore) UpdateConnectorVersionReview(ctx context.Context, versionID string, in connectors.UpdateVersionReviewInput) (connectors.ConnectorVersion, error) {
	var reviewedBy any
	if in.ReviewedBy != nil && *in.ReviewedBy != "" {
		reviewedBy = *in.ReviewedBy
	}
	var reviewedAt any
	if in.ReviewedAt != nil {
		reviewedAt = *in.ReviewedAt
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE connector_versions
		SET review_status=$1,
		    reviewed_by=$2::uuid,
		    reviewed_at=$3,
		    package_storage_key=COALESCE($4, package_storage_key)
		WHERE id=$5::uuid
		RETURNING id::text, connector_id::text, version, review_status, runtime_profile, entrypoint, manifest_json::text, package_sha256, package_size_bytes, package_storage_key,
		          validation_summary_json::text, uploaded_by::text, reviewed_by::text, reviewed_at, created_at
	`, in.ReviewStatus, reviewedBy, reviewedAt, in.PackageStorageKey, versionID)
	updated, err := scanConnectorVersion(row)
	if err != nil {
		return connectors.ConnectorVersion{}, fmt.Errorf("update connector version review: %w", err)
	}
	return updated, nil
}

func (s *ConnectorStore) InsertConnectorEvent(ctx context.Context, event connectors.ConnectorEvent) error {
	var connectorID any
	var connectorVersionID any
	var actorUserID any
	if event.ConnectorID != nil && *event.ConnectorID != "" {
		connectorID = *event.ConnectorID
	}
	if event.ConnectorVersionID != nil && *event.ConnectorVersionID != "" {
		connectorVersionID = *event.ConnectorVersionID
	}
	if event.ActorUserID != nil && *event.ActorUserID != "" {
		actorUserID = *event.ActorUserID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO connector_events(id, connector_id, connector_version_id, actor_user_id, event_type, detail_json, created_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::jsonb, $7)
	`, uuid.New(), connectorID, connectorVersionID, actorUserID, event.EventType, defaultJSON(event.DetailJSON), time.Now())
	if err != nil {
		return fmt.Errorf("insert connector event: %w", err)
	}
	return nil
}

func scanConnector(row interface{ Scan(dest ...any) error }) (connectors.Connector, error) {
	var model connectors.Connector
	var createdBy *string
	err := row.Scan(&model.ID, &model.Slug, &model.Name, &model.Description, &model.Status, &createdBy, &model.CreatedAt, &model.UpdatedAt)
	if err != nil {
		return connectors.Connector{}, err
	}
	model.CreatedBy = createdBy
	return model, nil
}

func scanConnectorVersion(row interface{ Scan(dest ...any) error }) (connectors.ConnectorVersion, error) {
	var model connectors.ConnectorVersion
	var uploadedBy *string
	var reviewedBy *string
	err := row.Scan(
		&model.ID, &model.ConnectorID, &model.Version, &model.ReviewStatus, &model.RuntimeProfile, &model.Entrypoint,
		&model.ManifestJSON, &model.PackageSHA256, &model.PackageSizeBytes, &model.PackageStorageKey,
		&model.ValidationSummaryJSON, &uploadedBy, &reviewedBy, &model.ReviewedAt, &model.CreatedAt,
	)
	if err != nil {
		return connectors.ConnectorVersion{}, err
	}
	model.UploadedBy = uploadedBy
	model.ReviewedBy = reviewedBy
	return model, nil
}

func defaultJSON(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}
