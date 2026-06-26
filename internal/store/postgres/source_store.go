package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/sources"
)

type SourceStore struct {
	pool *pgxpool.Pool
}

func NewSourceStore(pool *pgxpool.Pool) *SourceStore {
	return &SourceStore{pool: pool}
}

func (s *SourceStore) ListSources(ctx context.Context) ([]sources.ConnectorSource, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, connector_version_id::text, name, description, status, trigger_type, auth_mode, execution_mode, config_json::text, network_mode, created_by::text, created_at, updated_at
		FROM connector_sources
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()
	items := make([]sources.ConnectorSource, 0)
	for rows.Next() {
		item, scanErr := scanSource(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *SourceStore) GetSource(ctx context.Context, sourceID string) (sources.ConnectorSource, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, connector_version_id::text, name, description, status, trigger_type, auth_mode, execution_mode, config_json::text, network_mode, created_by::text, created_at, updated_at
		FROM connector_sources
		WHERE id=$1::uuid
	`, sourceID)
	item, err := scanSource(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return sources.ConnectorSource{}, false, nil
		}
		return sources.ConnectorSource{}, false, fmt.Errorf("get source: %w", err)
	}
	return item, true, nil
}

func (s *SourceStore) CreateSource(ctx context.Context, source sources.ConnectorSource) (sources.ConnectorSource, error) {
	var createdBy any
	if source.CreatedBy != nil && *source.CreatedBy != "" {
		createdBy = *source.CreatedBy
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO connector_sources(id, connector_version_id, name, description, status, trigger_type, auth_mode, execution_mode, config_json, network_mode, created_by, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11::uuid, $12, $13)
		RETURNING id::text, connector_version_id::text, name, description, status, trigger_type, auth_mode, execution_mode, config_json::text, network_mode, created_by::text, created_at, updated_at
	`, source.ID, source.ConnectorVersionID, source.Name, source.Description, source.Status, source.TriggerType, source.AuthMode, source.ExecutionMode, source.ConfigJSON, source.NetworkMode, createdBy, source.CreatedAt, source.UpdatedAt)
	created, err := scanSource(row)
	if err != nil {
		return sources.ConnectorSource{}, fmt.Errorf("create source: %w", err)
	}
	return created, nil
}

func (s *SourceStore) UpdateSource(ctx context.Context, sourceID string, in sources.UpdateSourceInput) (sources.ConnectorSource, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE connector_sources
		SET name=$1, description=$2, config_json=$3::jsonb, network_mode=$4, updated_at=$5
		WHERE id=$6::uuid
		RETURNING id::text, connector_version_id::text, name, description, status, trigger_type, auth_mode, execution_mode, config_json::text, network_mode, created_by::text, created_at, updated_at
	`, in.Name, in.Description, in.ConfigJSON, in.NetworkMode, time.Now(), sourceID)
	updated, err := scanSource(row)
	if err != nil {
		return sources.ConnectorSource{}, fmt.Errorf("update source: %w", err)
	}
	return updated, nil
}

func (s *SourceStore) SetSourceStatus(ctx context.Context, sourceID string, status sources.SourceStatus) (sources.ConnectorSource, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE connector_sources
		SET status=$1, updated_at=$2
		WHERE id=$3::uuid
		RETURNING id::text, connector_version_id::text, name, description, status, trigger_type, auth_mode, execution_mode, config_json::text, network_mode, created_by::text, created_at, updated_at
	`, status, time.Now(), sourceID)
	updated, err := scanSource(row)
	if err != nil {
		return sources.ConnectorSource{}, fmt.Errorf("set source status: %w", err)
	}
	return updated, nil
}

func (s *SourceStore) ListSecrets(ctx context.Context) ([]sources.SecretRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sr.id::text, sr.name, sr.secret_type, sr.encrypted_payload, sr.encryption_version, sr.created_by::text, sr.created_at, sr.rotated_at, sr.revoked_at, COUNT(ssb.id)::int
		FROM secret_records sr
		LEFT JOIN source_secret_bindings ssb ON ssb.secret_record_id = sr.id
		GROUP BY sr.id
		ORDER BY sr.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()
	items := make([]sources.SecretRecord, 0)
	for rows.Next() {
		secret, scanErr := scanSecret(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, secret)
	}
	return items, nil
}

func (s *SourceStore) GetSecret(ctx context.Context, secretID string) (sources.SecretRecord, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT sr.id::text, sr.name, sr.secret_type, sr.encrypted_payload, sr.encryption_version, sr.created_by::text, sr.created_at, sr.rotated_at, sr.revoked_at, COUNT(ssb.id)::int
		FROM secret_records sr
		LEFT JOIN source_secret_bindings ssb ON ssb.secret_record_id = sr.id
		WHERE sr.id=$1::uuid
		GROUP BY sr.id
	`, secretID)
	secret, err := scanSecret(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return sources.SecretRecord{}, false, nil
		}
		return sources.SecretRecord{}, false, fmt.Errorf("get secret: %w", err)
	}
	return secret, true, nil
}

func (s *SourceStore) CreateSecret(ctx context.Context, secret sources.SecretRecord) (sources.SecretRecord, error) {
	var createdBy any
	if secret.CreatedBy != nil && *secret.CreatedBy != "" {
		createdBy = *secret.CreatedBy
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO secret_records(id, name, secret_type, encrypted_payload, encryption_version, created_by, created_at)
		VALUES ($1::uuid, $2, $3, $4, $5, $6::uuid, $7)
		RETURNING id::text, name, secret_type, encrypted_payload, encryption_version, created_by::text, created_at, rotated_at, revoked_at, 0::int
	`, secret.ID, secret.Name, secret.SecretType, secret.EncryptedPayload, secret.EncryptionVersion, createdBy, secret.CreatedAt)
	created, err := scanSecret(row)
	if err != nil {
		return sources.SecretRecord{}, fmt.Errorf("create secret: %w", err)
	}
	return created, nil
}

func (s *SourceStore) RevokeSecret(ctx context.Context, secretID string, revokedAt time.Time) (sources.SecretRecord, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE secret_records
		SET revoked_at=$1
		WHERE id=$2::uuid
		RETURNING id::text, name, secret_type, encrypted_payload, encryption_version, created_by::text, created_at, rotated_at, revoked_at, 0::int
	`, revokedAt, secretID)
	secret, err := scanSecret(row)
	if err != nil {
		return sources.SecretRecord{}, fmt.Errorf("revoke secret: %w", err)
	}
	return secret, nil
}

func (s *SourceStore) ListSourceSecretBindings(ctx context.Context, sourceID string) ([]sources.SourceSecretBinding, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, connector_source_id::text, secret_name, secret_record_id::text, created_at
		FROM source_secret_bindings
		WHERE connector_source_id=$1::uuid
		ORDER BY created_at ASC
	`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("list source secret bindings: %w", err)
	}
	defer rows.Close()
	items := make([]sources.SourceSecretBinding, 0)
	for rows.Next() {
		binding, scanErr := scanBinding(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, binding)
	}
	return items, nil
}

func (s *SourceStore) CreateSourceSecretBinding(ctx context.Context, binding sources.SourceSecretBinding) (sources.SourceSecretBinding, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO source_secret_bindings(id, connector_source_id, secret_name, secret_record_id, created_at)
		VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5)
		ON CONFLICT (connector_source_id, secret_name)
		DO UPDATE SET secret_record_id=EXCLUDED.secret_record_id, created_at=EXCLUDED.created_at
		RETURNING id::text, connector_source_id::text, secret_name, secret_record_id::text, created_at
	`, binding.ID, binding.ConnectorSourceID, binding.SecretName, binding.SecretRecordID, binding.CreatedAt)
	created, err := scanBinding(row)
	if err != nil {
		return sources.SourceSecretBinding{}, fmt.Errorf("create source secret binding: %w", err)
	}
	return created, nil
}

func (s *SourceStore) DeleteSourceSecretBinding(ctx context.Context, bindingID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM source_secret_bindings WHERE id=$1::uuid`, bindingID)
	if err != nil {
		return fmt.Errorf("delete source secret binding: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return sources.ErrBindingNotFound
	}
	return nil
}

func (s *SourceStore) InsertSourceEvent(ctx context.Context, event sources.SourceEvent) error {
	var sourceID any
	var secretID any
	var actorUserID any
	if event.SourceID != nil && *event.SourceID != "" {
		sourceID = *event.SourceID
	}
	if event.SecretID != nil && *event.SecretID != "" {
		secretID = *event.SecretID
	}
	if event.ActorUserID != nil && *event.ActorUserID != "" {
		actorUserID = *event.ActorUserID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO source_events(id, connector_source_id, secret_record_id, actor_user_id, event_type, detail_json, created_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::jsonb, $7)
	`, uuid.NewString(), sourceID, secretID, actorUserID, event.EventType, defaultJSON(event.DetailJSON), time.Now())
	if err != nil {
		return fmt.Errorf("insert source event: %w", err)
	}
	return nil
}

func scanSource(row interface{ Scan(dest ...any) error }) (sources.ConnectorSource, error) {
	var model sources.ConnectorSource
	var createdBy *string
	err := row.Scan(&model.ID, &model.ConnectorVersionID, &model.Name, &model.Description, &model.Status, &model.TriggerType, &model.AuthMode, &model.ExecutionMode, &model.ConfigJSON, &model.NetworkMode, &createdBy, &model.CreatedAt, &model.UpdatedAt)
	if err != nil {
		return sources.ConnectorSource{}, err
	}
	model.CreatedBy = createdBy
	return model, nil
}

func scanSecret(row interface{ Scan(dest ...any) error }) (sources.SecretRecord, error) {
	var model sources.SecretRecord
	var createdBy *string
	err := row.Scan(&model.ID, &model.Name, &model.SecretType, &model.EncryptedPayload, &model.EncryptionVersion, &createdBy, &model.CreatedAt, &model.RotatedAt, &model.RevokedAt, &model.BindingCount)
	if err != nil {
		return sources.SecretRecord{}, err
	}
	model.CreatedBy = createdBy
	return model, nil
}

func scanBinding(row interface{ Scan(dest ...any) error }) (sources.SourceSecretBinding, error) {
	var model sources.SourceSecretBinding
	err := row.Scan(&model.ID, &model.ConnectorSourceID, &model.SecretName, &model.SecretRecordID, &model.CreatedAt)
	if err != nil {
		return sources.SourceSecretBinding{}, err
	}
	return model, nil
}
