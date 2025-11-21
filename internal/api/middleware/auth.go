package middleware

import (
	"net"
	"net/http"
	"strings"

	"steri-connect-go/internal/config"
)

// APIKeyAuthMiddleware handles API key authentication
func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()

		// Skip authentication if disabled
		if !cfg.Auth.APIKeyRequired {
			next.ServeHTTP(w, r)
			return
		}

		// Allow localhost bypass if configured
		if isLocalhost(r.RemoteAddr) {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "API key required", "message": "X-API-Key header is missing"}`))
			return
		}

		// Validate API key
		if apiKey != cfg.Auth.APIKey || cfg.Auth.APIKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid API key", "message": "The provided API key is invalid"}`))
			return
		}

		// API key is valid, proceed
		next.ServeHTTP(w, r)
	})
}

// isLocalhost checks if the remote address is localhost
func isLocalhost(remoteAddr string) bool {
	// Remove port if present
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If parsing fails, try using the address directly
		host = remoteAddr
	}

	// Check for localhost addresses
	return host == "127.0.0.1" ||
		host == "localhost" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.") ||
		strings.HasPrefix(host, "::ffff:127.")
}

// ShouldSkipAuth checks if a path should skip authentication
func ShouldSkipAuth(path string) bool {
	// Health check endpoint should always be accessible
	if path == "/api/health" {
		return true
	}
	// WebSocket endpoint (optional - can require auth in future)
	if path == "/ws" {
		return true
	}
	return false
}

