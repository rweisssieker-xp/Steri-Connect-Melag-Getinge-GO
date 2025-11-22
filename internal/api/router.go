package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"steri-connect-go/internal/api/handlers"
	"steri-connect-go/internal/api/middleware"
	"steri-connect-go/internal/api/websocket"
	"steri-connect-go/internal/config"
	"steri-connect-go/internal/testui"
)

// SetupRouter configures all API routes
func SetupRouter() http.Handler {
	mux := http.NewServeMux()
	cfg := config.Get()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/api/health", handlers.HealthHandler)

	// Metrics endpoint (no auth required)
	mux.HandleFunc("/api/metrics", handlers.MetricsHandler)

	// Diagnostics endpoint (no auth required)
	mux.HandleFunc("/api/diagnostics/", handlers.DiagnosticsHandler)

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
			handlers.GetMelagCycleHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Cycle management endpoints
	// GET /api/cycles - List all cycles with pagination, sorting, and filtering
	// GET /api/cycles/{id} - Get cycle details
	// GET /api/cycles/running - Get all running cycles
	// GET /api/cycles/{id}/export/pdf - Export cycle protocol as PDF
	// GET /api/cycles/export/csv - Export cycles as CSV
	// GET /api/cycles/export/json - Export cycles as JSON
	apiHandler.HandleFunc("/cycles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			// Check if this is a running cycles endpoint: /cycles/running
			if len(pathParts) >= 2 && pathParts[0] == "cycles" && pathParts[1] == "running" {
				handlers.GetRunningCyclesHandler(w, r)
				return
			}
			// Check if this is a JSON export endpoint: /cycles/export/json
			if len(pathParts) >= 3 && pathParts[0] == "cycles" && pathParts[1] == "export" && pathParts[2] == "json" {
				handlers.ExportCyclesJSONHandler(w, r)
				return
			}
			// Check if this is a CSV export endpoint: /cycles/export/csv
			if len(pathParts) >= 3 && pathParts[0] == "cycles" && pathParts[1] == "export" && pathParts[2] == "csv" {
				handlers.ExportCyclesCSVHandler(w, r)
				return
			}
			// Check if this is a PDF export endpoint: /cycles/{id}/export/pdf
			if len(pathParts) >= 4 && pathParts[0] == "cycles" && pathParts[2] == "export" && pathParts[3] == "pdf" {
				handlers.ExportCyclePDFHandler(w, r)
				return
			}
			// Check if this is a detail endpoint: /cycles/{id}
			if len(pathParts) > 1 && pathParts[0] == "cycles" && pathParts[1] != "export" && pathParts[1] != "running" {
				handlers.GetCycleHandler(w, r)
				return
			}
			handlers.ListCyclesHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Apply metrics middleware to track API requests
	// Apply authentication middleware to API routes if enabled
	var finalHandler http.Handler = middleware.MetricsMiddleware(apiHandler)
	if cfg.Auth.APIKeyRequired {
		finalHandler = applyAuthMiddleware(finalHandler)
	}

	// Mount API handler
	mux.Handle("/api/", http.StripPrefix("/api", finalHandler))

	// Test UI routes (only if enabled)
	if cfg.TestUI.Enabled {
		// Initialize Test UI templates
		if err := testui.InitTemplates(); err != nil {
			// Log error but don't fail - Test UI will show error when accessed
		}

		// Test UI main page
		mux.HandleFunc("/test-ui", testui.TestUIHandler)

		// Test UI static assets (CSS, JS)
		staticDir := http.Dir(filepath.Join("web", "testui"))
		mux.Handle("/test-ui/static/", http.StripPrefix("/test-ui/static/", http.FileServer(staticDir)))

		// Database inspection endpoints (read-only, Test UI only)
		apiHandler.HandleFunc("/test-ui/db/tables", handlers.ListTablesHandler)
		apiHandler.HandleFunc("/test-ui/db/tables/", func(w http.ResponseWriter, r *http.Request) {
			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(pathParts) >= 6 && pathParts[5] == "schema" {
				handlers.GetTableSchemaHandler(w, r)
			} else if len(pathParts) >= 6 && pathParts[5] == "export" {
				handlers.ExportTableDataHandler(w, r)
			} else if len(pathParts) >= 5 {
				handlers.GetTableDataHandler(w, r)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		})

		// Log viewing endpoints (Test UI only)
		apiHandler.HandleFunc("/test-ui/logs", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				handlers.GetLogsHandler(w, r)
			} else if r.Method == http.MethodDelete {
				handlers.ClearLogsHandler(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
	}

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

