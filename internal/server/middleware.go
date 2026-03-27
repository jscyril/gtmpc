package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jscyril/golang_music_player/internal/auth"
)

// --- Middleware Chain ---

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies a sequence of middleware to a handler (left-to-right execution)
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// LoggingMiddleware logs every incoming HTTP request with method, path, and latency.
// Demonstrates: goroutine-safe request logging and performance monitoring.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[HTTP] --> %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("[HTTP] <-- %s %s (%v)", r.Method, r.URL.Path, time.Since(start))
	})
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates the JWT Bearer token from the Authorization header.
// Protected routes use this to enforce authentication.
func AuthMiddleware(secret []byte) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"success":false,"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ValidateToken(token, secret)
			if err != nil {
				http.Error(w, `{"success":false,"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Inject claims into request context for downstream handlers
			r.Header.Set("X-User", claims.Username)
			r.Header.Set("X-Role", claims.Role)
			next.ServeHTTP(w, r)
		})
	}
}
