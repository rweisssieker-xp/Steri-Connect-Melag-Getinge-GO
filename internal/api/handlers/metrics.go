package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"steri-connect-go/internal/api/middleware"
	"steri-connect-go/internal/config"
	"steri-connect-go/internal/database"
	"steri-connect-go/internal/devices"
)

// MetricsResponse represents the system metrics response
type MetricsResponse struct {
	UptimeSeconds        int64   `json:"uptime_seconds"`
	ActiveConnections    int     `json:"active_device_connections"`
	TotalCycles          int     `json:"total_cycles"`
	CyclesToday          int     `json:"cycles_today"`
	TotalAPIRequests     int64   `json:"total_api_requests"`
	RequestsPerMinute    int     `json:"requests_per_minute"`
	DatabaseSizeMB       float64 `json:"database_size_mb,omitempty"`
	Timestamp            time.Time `json:"timestamp"`
}

var metricsStartTime = time.Now()

// MetricsHandler handles GET /api/metrics requests
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Get uptime in seconds
	uptimeSeconds := int64(time.Since(metricsStartTime).Seconds())

	// Get active device connections
	activeConnections := 0
	deviceManager := devices.GetManager()
	if deviceManager != nil {
		allDevices, err := database.GetAllDevices()
		if err == nil {
			for _, device := range allDevices {
				status, err := database.GetDeviceStatus(device.ID)
				if err == nil && status.Connected {
					activeConnections++
				}
			}
		}
	}

	// Get total cycles
	totalCycles := 0
	_, count, err := database.GetAllCycles(database.CycleListOptions{
		Limit: 10000,
		Offset: 0,
	})
	if err == nil {
		totalCycles = count
	}

	// Get cycles today
	cyclesToday := 0
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24 * time.Hour)
	todayStartTime := todayStart
	todayEndTime := todayEnd
	_, countToday, err := database.GetAllCycles(database.CycleListOptions{
		Limit:     10000,
		Offset:    0,
		StartDate: &todayStartTime,
		EndDate:   &todayEndTime,
	})
	if err == nil {
		cyclesToday = countToday
	}

	// Get API request metrics
	metrics := middleware.GetMetrics()
	totalAPIRequests := metrics.GetTotalRequests()
	requestsPerMinute := metrics.GetRequestsPerMinute()

	// Get database size
	databaseSizeMB := 0.0
	cfg := config.Get()
	if cfg != nil {
		if dbPath := cfg.Database.Path; dbPath != "" {
			if fileInfo, err := os.Stat(dbPath); err == nil {
				databaseSizeMB = float64(fileInfo.Size()) / 1024 / 1024
			}
		}
	}

	response := MetricsResponse{
		UptimeSeconds:     uptimeSeconds,
		ActiveConnections: activeConnections,
		TotalCycles:       totalCycles,
		CyclesToday:       cyclesToday,
		TotalAPIRequests:  totalAPIRequests,
		RequestsPerMinute: requestsPerMinute,
		DatabaseSizeMB:    databaseSizeMB,
		Timestamp:         time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

