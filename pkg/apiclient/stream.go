// Package apiclient provides stream URL helpers for the gtmpc REST API.
package apiclient

// StreamURL returns the full URL for streaming a track by ID.
// GET /api/stream/{trackId} with Authorization header.
func (c *APIClient) StreamURL(trackID string) string {
	return c.BaseURL + "/api/stream/" + trackID
}
