// Package apiclient provides authentication API calls for the gtmpc REST API.
package apiclient

import "fmt"

// Register creates a new user account via POST /api/auth/register.
// Returns an error if the username is already taken (ErrConflict) or other failure.
func (c *APIClient) Register(req RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse
	if err := c.doJSONRequest("POST", "/api/auth/register", req, &resp); err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	return &resp, nil
}

// Login authenticates the user via POST /api/auth/login.
// On success, automatically calls SetToken() with the returned JWT.
// Returns ErrUnauthorized if credentials are invalid.
func (c *APIClient) Login(req LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse
	if err := c.doJSONRequest("POST", "/api/auth/login", req, &resp); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	// Automatically store the token for subsequent requests
	c.SetToken(resp.Token)
	return &resp, nil
}
