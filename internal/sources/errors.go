package sources

import "errors"

var (
	ErrSourceNotFound          = errors.New("source_not_found")
	ErrSecretNotFound          = errors.New("secret_not_found")
	ErrBindingNotFound         = errors.New("source_secret_binding_not_found")
	ErrInvalidInput            = errors.New("invalid_source_input")
	ErrUnsupportedAlphaMode    = errors.New("unsupported_alpha_source_mode")
	ErrConnectorUnavailable    = errors.New("connector_unavailable")
	ErrConnectorVersionInvalid = errors.New("connector_version_invalid")
	ErrMissingRequiredSecrets  = errors.New("missing_required_secrets")
	ErrSecretRevoked           = errors.New("secret_revoked")
	ErrSecretTooLarge          = errors.New("secret_too_large")
)
