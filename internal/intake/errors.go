package intake

import "errors"

var (
	ErrJobNotCompleted = errors.New("import job is not completed")
	ErrBundleInvalid   = errors.New("metadata bundle is invalid")
	ErrBundleMissing   = errors.New("metadata bundle artifact is missing")
)
