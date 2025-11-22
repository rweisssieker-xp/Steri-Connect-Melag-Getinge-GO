package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/devices"
	"steri-connect-go/internal/api/websocket"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status            string                 `json:"status"`
	Timestamp         time.Time              `json:"timestamp"`
	Version           string                 `json:"version"`
	Uptime            string                 `json:"uptime,omitempty"`
	Database          DatabaseStatus         `json:"database"`
	Devices           DeviceStatusSummary    `json:"devices"`
	WebSocket         WebSocketStatus        `json:"websocket"`
	Memory            MemoryStatus           `json:"memory,omitempty"`
}

// DatabaseStatus represents database connection status
type DatabaseStatus struct {
	Connected bool   `json:"connected"`
	Status    string `json:"status"`
}

// DeviceStatusSummary represents device connectivity summary
type DeviceStatusSummary struct {
	Total    int `json:"total"`
	Online   int `json:"online"`
	Offline  int `json:"offline"`
	Error    int `json:"error"`
}

// WebSocketStatus represents WebSocket connection status
type WebSocketStatus struct {
	Connections int `json:"connections"`
}

// MemoryStatus represents memory usage information
type MemoryStatus struct {
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
}

var startTime = time.Now()

// HealthHandler handles GET /api/health requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check database connection
	dbStatus := DatabaseStatus{Connected: false, Status: "disconnected"}
	if db := database.DB(); db != nil {
		if err := db.Ping(); err == nil {
			dbStatus.Connected = true
			dbStatus.Status = "connected"
		}
	}

	// Get device connectivity summary
	deviceSummary := DeviceStatusSummary{}
	deviceManager := devices.GetManager()
	if deviceManager != nil {
		allDevices, err := database.GetAllDevices()
		if err == nil {
			deviceSummary.Total = len(allDevices)
			for _, device := range allDevices {
				status, err := database.GetDeviceStatus(device.ID)
				if err == nil {
					if status.Connected {
						deviceSummary.Online++
					} else {
						deviceSummary.Offline++
					}
				} else {
					deviceSummary.Offline++
				}
			}
		}
	}

	// Get WebSocket connection count
	wsConnections := 0
	if hub := websocket.GetHub(); hub != nil {
		wsConnections = hub.GetConnectionCount()
	}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memory := MemoryStatus{
		AllocMB:      float64(memStats.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(memStats.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(memStats.Sys) / 1024 / 1024,
		NumGC:        memStats.NumGC,
	}

	// Determine overall status
	status := "ok"
	if !dbStatus.Connected {
		status = "error"
	} else if deviceSummary.Total > 0 && deviceSummary.Offline > deviceSummary.Online {
		status = "degraded"
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Database:  dbStatus,
		Devices:   deviceSummary,
		WebSocket: WebSocketStatus{Connections: wsConnections},
		Memory:    memory,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

