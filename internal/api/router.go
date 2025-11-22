package api

import (
	"net/http"
	"strings"

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
	
	// Device management endpoints
	// POST /api/devices - Create device
	// GET /api/devices - List all devices
	apiHandler.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateDeviceHandler(w, r)
		case http.MethodGet:
			handlers.ListDevicesHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// GET /api/devices/{id} - Get device by ID
	// PUT /api/devices/{id} - Update device
	// DELETE /api/devices/{id} - Delete device
	apiHandler.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a status endpoint
		if strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodGet {
			handlers.GetDeviceStatusHandler(w, r)
			return
		}

		// Regular device CRUD operations
		switch r.Method {
		case http.MethodGet:
			handlers.GetDeviceHandler(w, r)
		case http.MethodPut:
			handlers.UpdateDeviceHandler(w, r)
		case http.MethodDelete:
			handlers.DeleteDeviceHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Melag device operations
	// POST /api/melag/{id}/start - Start cycle
	// GET /api/melag/{id}/status - Get device and cycle status
	// GET /api/melag/{id}/cycles/{cycle_id} - Get cycle details
	apiHandler.HandleFunc("/melag/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/start") && r.Method == http.MethodPost {
			handlers.StartCycleHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodGet {
			handlers.GetMelagStatusHandler(w, r)
			return
		}
		// Check if this is a cycle retrieval endpoint: /melag/{id}/cycles/{cycle_id}
		if strings.Contains(r.URL.Path, "/cycles/") && r.Method == http.MethodGet {
			handlers.GetCycleHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Cycle management endpoints
	// GET /api/cycles - List all cycles with pagination, sorting, and filtering
	apiHandler.HandleFunc("/cycles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.ListCyclesHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Apply authentication middleware to API routes if enabled
	var finalHandler http.Handler = apiHandler
	if cfg.Auth.APIKeyRequired {
		finalHandler = applyAuthMiddleware(apiHandler)
	}

	// Mount API handler
	mux.Handle("/api/", http.StripPrefix("/api", finalHandler))

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

