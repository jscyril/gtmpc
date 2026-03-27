// Package apiclient provides library API calls for the gtmpc REST API.
package apiclient

import "fmt"

// GetTracks fetches the full track library via GET /api/library/tracks.
// Requires authentication (token must be set via SetToken).
func (c *APIClient) GetTracks() (*TrackListResponse, error) {
	var resp TrackListResponse
	if err := c.doJSONRequest("GET", "/api/library/tracks", nil, &resp); err != nil {
		return nil, fmt.Errorf("get tracks: %w", err)
	}
	return &resp, nil
}

// GetPlaylists fetches all playlists via GET /api/library/playlists.
// Requires authentication.
func (c *APIClient) GetPlaylists() (*PlaylistListResponse, error) {
	var resp PlaylistListResponse
	if err := c.doJSONRequest("GET", "/api/library/playlists", nil, &resp); err != nil {
		return nil, fmt.Errorf("get playlists: %w", err)
	}
	return &resp, nil
}

// CreatePlaylist creates a new playlist via POST /api/library/playlists.
// Requires authentication.
func (c *APIClient) CreatePlaylist(req CreatePlaylistRequest) (*CreatePlaylistResponse, error) {
	var resp CreatePlaylistResponse
	if err := c.doJSONRequest("POST", "/api/library/playlists", req, &resp); err != nil {
		return nil, fmt.Errorf("create playlist: %w", err)
	}
	return &resp, nil
}

// CoverURL returns the URL for fetching cover art for a given track ID.
// Returns an empty string if trackID is empty.
func (c *APIClient) CoverURL(trackID string) string {
	if trackID == "" {
		return ""
	}
	return c.BaseURL + "/api/library/cover/" + trackID
}
