package jobs

import "errors"

var (
	ErrJobNotFound       = errors.New("job_not_found")
	ErrSourceNotRunnable = errors.New("source_not_runnable")
	ErrInvalidJobState   = errors.New("invalid_job_state")
	ErrActiveJobExists   = errors.New("active_job_exists")
)
