// Package apiclient provides HTTP client types and wrappers for the gtmpc REST API.
// These types match the JSON contract of the backend server and are separate from
// the internal api/types.go which uses native Go types for the local audio engine.
package apiclient

// RegisterRequest is sent to POST /api/auth/register
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse is returned from POST /api/auth/register on success
type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// LoginRequest is sent to POST /api/auth/login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is returned from POST /api/auth/login on success
type LoginResponse struct {
	Token     string `json:"token"`
	Username  string `json:"username"`
	ExpiresAt string `json:"expires_at"`
}

// Track represents a single audio track from the library.
// DurationSeconds is an integer number of seconds to match the JSON API.
type Track struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	Album           string `json:"album"`
	DurationSeconds int    `json:"duration_seconds"`
	Format          string `json:"format"` // mp3 | flac | wav | ogg
	CoverURL        string `json:"cover_url"`
}

// TrackListResponse is returned from GET /api/library/tracks
type TrackListResponse struct {
	Tracks []Track `json:"tracks"`
}

// Playlist represents a playlist entity
type Playlist struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	TrackIDs  []string `json:"track_ids"`
	CreatedAt string   `json:"created_at"`
}

// PlaylistListResponse is returned from GET /api/library/playlists
type PlaylistListResponse struct {
	Playlists []Playlist `json:"playlists"`
}

// CreatePlaylistRequest is sent to POST /api/library/playlists
type CreatePlaylistRequest struct {
	Name     string   `json:"name"`
	TrackIDs []string `json:"track_ids"`
}

// CreatePlaylistResponse is returned from POST /api/library/playlists
type CreatePlaylistResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// HealthResponse is returned from GET /api/health
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// APIError represents an error response from the server
type APIError struct {
	Error string `json:"error"`
}

// ─── Client-side statistics types (no backend required) ───────────────────────

// ArtistBar is one row in the artist breakdown chart.
type ArtistBar struct {
	Artist string
	Count  int
	Bar    string // pre-rendered ASCII bar e.g. "████████"
}

// StatsSummary holds all computed session statistics.
// It is produced by pkg/stats and consumed by TUI and web contexts.
type StatsSummary struct {
	TracksPlayed     int
	TracksLiked      int
	TotalSeconds     int
	FormattedTime    string
	TopArtist        string
	ArtistPlayCounts map[string]int
	MeanDurationSec  float64
	StdDevSec        float64
	FormattedMean    string
	FormattedStdDev  string
	MostPlayedTitle  string
	MostPlayedCount  int
	ArtistChart      []ArtistBar
}
