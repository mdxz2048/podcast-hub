package auth

import "errors"

var (
	ErrInvalidEmail            = errors.New("invalid_email")
	ErrWeakPassword            = errors.New("weak_password")
	ErrPasswordMismatch        = errors.New("password_mismatch")
	ErrRateLimited             = errors.New("rate_limited")
	ErrTurnstileRequired       = errors.New("turnstile_required")
	ErrTurnstileFailed         = errors.New("turnstile_failed")
	ErrInvalidOrExpiredCode    = errors.New("invalid_or_expired_code")
	ErrVerificationCodeExpired = errors.New("verification_code_expired")
	ErrTooManyAttempts         = errors.New("too_many_attempts")
	ErrInvalidCredentials      = errors.New("invalid_credentials")
	ErrAccountUnavailable      = errors.New("account_unavailable")
	ErrNotAuthenticated        = errors.New("not_authenticated")
	ErrSessionExpired          = errors.New("session_expired")
	ErrInvalidOrExpiredProof   = errors.New("invalid_or_expired_proof")
	ErrInvalidResetProof       = errors.New("invalid_reset_proof")
	ErrResetProofExpired       = errors.New("reset_proof_expired")
	ErrTemporarilyUnavailable  = errors.New("temporarily_unavailable")
	ErrForbiddenCSRF           = errors.New("csrf_invalid")
	ErrUnsupportedContentType  = errors.New("unsupported_content_type")
	ErrInvalidJSONBody         = errors.New("invalid_json_body")
	ErrInternal                = errors.New("internal_error")
)
