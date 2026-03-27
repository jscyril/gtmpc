// Package apiclient provides an HTTP client for the gtmpc REST API.
// It wraps all network calls, injects JWT tokens, and maps HTTP errors to
// typed Go errors.
package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrUnauthorized is returned when the server responds with 401
var ErrUnauthorized = fmt.Errorf("unauthorized: please log in again")

// ErrConflict is returned when the server responds with 409 (e.g. username taken)
var ErrConflict = fmt.Errorf("conflict: resource already exists")

// ErrNotFound is returned when the server responds with 404
var ErrNotFound = fmt.Errorf("not found")

// ErrServerError is returned for 5xx responses
var ErrServerError = fmt.Errorf("server error")

// APIClient is the HTTP client for the gtmpc server.
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// NewAPIClient creates a new APIClient with a 30-second timeout.
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken stores the JWT token for subsequent authenticated requests.
func (c *APIClient) SetToken(token string) {
	c.Token = token
}

// doRequest performs an HTTP request, injecting Authorization header when a token is set.
// body may be nil for GET requests. Returns the raw response for the caller to decode.
func (c *APIClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}

	return resp, nil
}

// doJSONRequest performs an HTTP request and decodes the JSON response into result.
// It maps HTTP status codes to typed errors before attempting JSON decode.
func (c *APIClient) doJSONRequest(method, path string, body interface{}, result interface{}) error {
	resp, err := c.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// Map error status codes
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("%w: %s", ErrUnauthorized, apiErr.Error)
		}
		return ErrUnauthorized
	case http.StatusConflict:
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("%w: %s", ErrConflict, apiErr.Error)
		}
		return ErrConflict
	case http.StatusNotFound:
		return ErrNotFound
	}

	if resp.StatusCode >= 500 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("%w: %s", ErrServerError, apiErr.Error)
		}
		return ErrServerError
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("request failed (%d): %s", resp.StatusCode, apiErr.Error)
		}
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Decode success response if a result target was provided
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// CheckHealth calls GET /api/health and returns the health response.
// Returns an error if the server is unreachable or responds with an error.
func (c *APIClient) CheckHealth() (*HealthResponse, error) {
	var health HealthResponse
	if err := c.doJSONRequest("GET", "/api/health", nil, &health); err != nil {
		return nil, err
	}
	return &health, nil
}

// StreamRequest returns a configured *http.Request for streaming audio.
// The caller is responsible for closing the response body.
// Uses no timeout so large streams are not cut off.
func (c *APIClient) StreamRequest(streamURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create stream request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return req, nil
}
