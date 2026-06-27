package publication

import "time"

type ProgramAccessStatus string
type FeedStatus string

const (
	ProgramAccessActive  ProgramAccessStatus = "active"
	ProgramAccessRevoked ProgramAccessStatus = "revoked"
	FeedStatusActive     FeedStatus          = "active"
	FeedStatusRevoked    FeedStatus          = "revoked"
	FeedStatusExpired    FeedStatus          = "expired"
)

type ProgramAccessGrant struct {
	ID        string              `json:"id"`
	UserID    string              `json:"user_id"`
	ProgramID string              `json:"program_id"`
	Status    ProgramAccessStatus `json:"status"`
	Reason    string              `json:"reason"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	RevokedAt *time.Time          `json:"revoked_at,omitempty"`
}

type RSSFeed struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Status      FeedStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	RotatedAt   *time.Time `json:"rotated_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type AdminRSSFeed struct {
	RSSFeed
	UserEmailHint string `json:"user_email_hint"`
}

type AuthorizedMedia struct {
	EpisodeID       string
	ProgramID       string
	ContentType     string
	SizeBytes       int64
	SHA256          string
	PublishedAt     time.Time
	PublishedKey    string
	FeedID          *string
	FeedOwnerUserID *string
	EpisodeTitle    string
	ProgramTitle    string
}

type FeedEpisode struct {
	EpisodeID     string
	ProgramID     string
	ProgramTitle  string
	ProgramAuthor string
	Language      string
	Title         string
	Description   string
	PublishedAt   time.Time
	ContentType   string
	SizeBytes     int64
	SHA256        string
	UpdatedAt     time.Time
}

type FeedDocument struct {
	Feed      RSSFeed
	UserID    string
	Items     []FeedEpisode
	ETag      string
	UpdatedAt time.Time
}

type FeedTokenResult struct {
	Feed    RSSFeed `json:"feed"`
	Token   string  `json:"token"`
	FeedURL string  `json:"feed_url"`
}
