package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// JWT errors
var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// jwtHeader is the fixed header for our HS256 tokens
var jwtHeader = base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))

// Claims represents the JWT payload
type Claims struct {
	Username string `json:"sub"`
	Role     string `json:"role"`
	IssuedAt int64  `json:"iat"`
	ExpAt    int64  `json:"exp"`
}

// GenerateToken creates a signed JWT for the given user.
// Uses HMAC-SHA256 for signing — no external JWT library needed.
func GenerateToken(username, role string, secret []byte, ttl time.Duration) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		IssuedAt: time.Now().Unix(),
		ExpAt:    time.Now().Add(ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("jwt: failed to marshal claims: %w", err)
	}

	encodedPayload := base64URLEncode(payload)
	signingInput := jwtHeader + "." + encodedPayload
	signature := sign([]byte(signingInput), secret)

	return signingInput + "." + signature, nil
}

// ValidateToken parses and validates a JWT, returning the claims.
func ValidateToken(tokenStr string, secret []byte) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Verify signature using constant-time comparison to prevent timing attacks
	signingInput := parts[0] + "." + parts[1]
	expectedSig := sign([]byte(signingInput), secret)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, ErrInvalidToken
	}

	// Decode payload
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiry
	if time.Now().Unix() > claims.ExpAt {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// --- Helpers ---

func sign(data, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return base64URLEncode(h.Sum(nil))
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding back
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
