package connectors

import (
	"context"
	"io"
	"time"
)

type Store interface {
	FindConnectorBySlug(ctx context.Context, slug string) (Connector, bool, error)
	CreateConnector(ctx context.Context, connector Connector) (Connector, error)
	UpdateConnectorStatus(ctx context.Context, connectorID string, status ConnectorStatus) (Connector, error)
	GetConnector(ctx context.Context, connectorID string) (Connector, bool, error)
	ListConnectors(ctx context.Context) ([]Connector, error)
	ListConnectorVersions(ctx context.Context, connectorID string) ([]ConnectorVersion, error)
	GetConnectorVersion(ctx context.Context, versionID string) (ConnectorVersion, bool, error)
	GetConnectorVersionByVersion(ctx context.Context, connectorID, version string) (ConnectorVersion, bool, error)
	CreateConnectorVersion(ctx context.Context, version ConnectorVersion) (ConnectorVersion, error)
	UpdateConnectorVersionReview(ctx context.Context, versionID string, in UpdateVersionReviewInput) (ConnectorVersion, error)
	InsertConnectorEvent(ctx context.Context, event ConnectorEvent) error
}

type ConnectorEvent struct {
	ConnectorID        *string
	ConnectorVersionID *string
	ActorUserID        *string
	EventType          string
	DetailJSON         string
}

type UpdateVersionReviewInput struct {
	ReviewStatus      ReviewStatus
	ReviewedBy        *string
	ReviewedAt        *time.Time
	PackageStorageKey *string
}

type ConnectorPackageStore interface {
	PutQuarantine(ctx context.Context, packageName string, content io.Reader) (PackageRef, error)
	Read(ctx context.Context, ref PackageRef) (io.ReadCloser, error)
	PromoteApproved(ctx context.Context, ref PackageRef) (PackageRef, error)
	Delete(ctx context.Context, ref PackageRef) error
}
