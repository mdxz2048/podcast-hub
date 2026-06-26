package connectors

import "errors"

var (
	ErrInvalidConnectorID      = errors.New("invalid_connector_id")
	ErrInvalidVersion          = errors.New("invalid_version")
	ErrConnectorNotFound       = errors.New("connector_not_found")
	ErrVersionNotFound         = errors.New("connector_version_not_found")
	ErrVersionAlreadyExists    = errors.New("connector_version_already_exists")
	ErrConnectorDisabled       = errors.New("connector_disabled")
	ErrVersionNotPendingReview = errors.New("version_not_pending_review")
	ErrVersionNotApproved      = errors.New("version_not_approved")
	ErrInvalidUpload           = errors.New("invalid_upload")
)
