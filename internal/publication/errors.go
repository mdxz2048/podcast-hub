package publication

import "errors"

var (
	ErrProgramAccessNotFound = errors.New("program_access_not_found")
	ErrFeedNotFound          = errors.New("rss_feed_not_found")
	ErrFeedForbidden         = errors.New("rss_feed_forbidden")
	ErrFeedExpired           = errors.New("rss_feed_expired")
	ErrFeedRevoked           = errors.New("rss_feed_revoked")
	ErrFeedTokenInvalid      = errors.New("rss_feed_token_invalid")
	ErrMediaNotAvailable     = errors.New("media_not_available")
	ErrUserNotEligible       = errors.New("user_not_eligible")
	ErrInvalidFeedName       = errors.New("invalid_feed_name")
	ErrCollectionNotFound    = errors.New("collection_not_found")
	ErrProgramNotAvailable   = errors.New("program_not_available")
	ErrInvalidCollection     = errors.New("invalid_collection")
)
