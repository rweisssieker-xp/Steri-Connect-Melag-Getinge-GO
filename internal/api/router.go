package api

import (
	"net/http"

	"steri-connect-go/internal/api/handlers"
	"steri-connect-go/internal/api/middleware"
	"steri-connect-go/internal/api/websocket"
	"steri-connect-go/internal/config"
)

// SetupRouter configures all API routes
func SetupRouter() http.Handler {
	mux := http.NewServeMux()
	cfg := config.Get()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/api/health", handlers.HealthHandler)

	// WebSocket endpoint (no auth required by default)
	mux.HandleFunc("/ws", websocket.HandleWebSocket)

	// Create API router with authentication for all /api/* routes (except health)
	apiHandler := http.NewServeMux()
	
	// API routes (future endpoints will be added here)
	// apiHandler.HandleFunc("/api/devices", ...)
	// apiHandler.HandleFunc("/api/cycles", ...)

	// Apply authentication middleware to API routes if enabled
	if cfg.Auth.APIKeyRequired {
		apiHandler = applyAuthMiddleware(apiHandler)
	}

	// Mount API handler
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Apply CORS middleware to all routes
	handler := middleware.CORSMiddleware(mux)

	return handler
}

// applyAuthMiddleware applies authentication middleware to API routes
func applyAuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check
		if r.URL.Path == "/health" || r.URL.Path == "/" {
			handler.ServeHTTP(w, r)
			return
		}

		// Apply auth middleware
		middleware.APIKeyAuthMiddleware(handler).ServeHTTP(w, r)
	})
}

