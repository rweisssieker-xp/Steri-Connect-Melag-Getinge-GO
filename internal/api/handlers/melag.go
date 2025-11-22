package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/devices"
	"steri-connect-go/internal/logging"
	"steri-connect-go/internal/api/websocket"
	"steri-connect-go/internal/adapters/melag"
)

// StartCycleRequest represents the request body for starting a cycle
type StartCycleRequest struct {
	Program     string   `json:"program,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Pressure    *float64 `json:"pressure,omitempty"`
	Duration    *int     `json:"duration,omitempty"` // Duration in minutes
}

// StartCycleResponse represents the response for starting a cycle
type StartCycleResponse struct {
	CycleID     int       `json:"cycle_id"`
	DeviceID    int       `json:"device_id"`
	Program     string    `json:"program,omitempty"`
	Status      string    `json:"status"`
	Phase       string    `json:"phase,omitempty"`
	StartTime   string    `json:"start_time"`
}

// StartCycleHandler handles POST /api/melag/{id}/start requests
func StartCycleHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only POST method is allowed",
		})
		return
	}

	// Extract device ID from URL path
	deviceID, err := extractMelagDeviceID(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract device ID from path", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_id",
			Message: "Invalid device ID in URL path",
		})
		return
	}

	// Parse request body
	var req StartCycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Failed to parse request body", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body: " + err.Error(),
		})
		return
	}

	// Get device to verify it exists and is a Melag device
	device, err := database.GetDevice(deviceID)
	if err != nil {
		if err == database.ErrDeviceNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "device_not_found",
				Message: fmt.Sprintf("Device with ID %d not found", deviceID),
			})
			return
		}

		logger.Error("Failed to get device", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve device",
		})
		return
	}

	// Verify device is a Melag device
	if device.Manufacturer != "Melag" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_type",
			Message: fmt.Sprintf("Device %d is not a Melag device", deviceID),
		})
		return
	}

	// Get device manager (global instance)
	deviceManager := devices.GetManager()
	if deviceManager == nil {
		logger.Error("Device manager not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "service_unavailable",
			Message: "Device manager not initialized",
		})
		return
	}

	adapter := deviceManager.GetAdapter(deviceID)
	if adapter == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "device_not_connected",
			Message: fmt.Sprintf("Device %d is not connected", deviceID),
		})
		return
	}

	// Cast to MelagAdapter
	melagAdapter, ok := adapter.(*melag.MelagAdapter)
	if !ok {
		logger.Error("Failed to cast adapter to MelagAdapter", "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid adapter type",
		})
		return
	}

	// Verify device is connected
	if !melagAdapter.IsConnected() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "device_not_connected",
			Message: fmt.Sprintf("Device %d is not connected", deviceID),
		})
		return
	}

	// Start cycle on device
	startParams := melag.CycleStartParams{
		Program:     req.Program,
		Temperature: req.Temperature,
		Pressure:    req.Pressure,
		Duration:    req.Duration,
	}

	if err := melagAdapter.StartCycle(startParams); err != nil {
		logger.Error("Failed to start cycle", "error", err, "device_id", deviceID)

		// Create cycle record with FAILED status
		cycle := &database.Cycle{
			DeviceID: deviceID,
			Program:  req.Program,
			StartTS:  time.Now(),
			Phase:    "FAILED",
			Result:   "NOK",
			ErrorDescription: err.Error(),
		}

		if _, createErr := database.CreateCycle(cycle); createErr != nil {
			logger.Error("Failed to create cycle record", "error", createErr)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "cycle_start_failed",
			Message: "Failed to start cycle: " + err.Error(),
		})
		return
	}

	// Create cycle record in database
	now := time.Now()
	cycle := &database.Cycle{
		DeviceID:      deviceID,
		Program:       req.Program,
		StartTS:       now,
		Phase:         "STARTING",
		ProgressPercent: func() *int { v := 0; return &v }(),
	}

	createdCycle, err := database.CreateCycle(cycle)
	if err != nil {
		logger.Error("Failed to create cycle record", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create cycle record",
		})
		return
	}

	// Broadcast cycle_started event
	event := websocket.Event{
		Event: "cycle_started",
		Data: map[string]interface{}{
			"cycle_id":  createdCycle.ID,
			"device_id": deviceID,
			"program":   req.Program,
			"phase":     "STARTING",
		},
	}

	if err := websocket.BroadcastEvent(event); err != nil {
		logger.Warn("Failed to broadcast cycle_started event", "error", err)
		// Continue even if broadcast fails
	}

	// Log audit entry
	cycleID := createdCycle.ID
	details := map[string]interface{}{
		"cycle_id":  cycleID,
		"device_id": deviceID,
		"program":   req.Program,
		"phase":     "STARTING",
	}

	if err := database.LogAudit(database.ActionCycleStarted, "cycle", &cycleID, "", details); err != nil {
		logger.Warn("Failed to create audit log", "error", err)
		// Continue even if audit log fails
	}

	logger.Info("Cycle started successfully",
		"cycle_id", createdCycle.ID,
		"device_id", deviceID,
		"program", req.Program)

	// Start cycle status polling
	deviceManager.StartCyclePolling(createdCycle.ID, deviceID)

	// Return cycle information
	response := StartCycleResponse{
		CycleID:   createdCycle.ID,
		DeviceID:  deviceID,
		Program:   req.Program,
		Status:    "STARTING",
		Phase:     "STARTING",
		StartTime: createdCycle.StartTS.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// extractMelagDeviceID extracts device ID from URL path like "/melag/1/start"
func extractMelagDeviceID(path string) (int, error) {
	// Path will be "/melag/1/start" after StripPrefix
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "melag" || parts[2] != "start" {
		return 0, fmt.Errorf("invalid path format: expected /melag/{id}/start")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid device ID: %w", err)
	}

	return id, nil
}

// GetMelagStatusResponse represents the response for Melag device status
type GetMelagStatusResponse struct {
	DeviceID int                      `json:"device_id"`
	Connected bool                    `json:"connected"`
	CurrentCycle *CycleStatusInfo     `json:"current_cycle,omitempty"`
	LastCycle   *CycleStatusInfo      `json:"last_cycle,omitempty"`
}

// CycleStatusInfo represents cycle status information
type CycleStatusInfo struct {
	CycleID        int       `json:"cycle_id"`
	Program        string    `json:"program,omitempty"`
	Phase          string    `json:"phase,omitempty"`
	ProgressPercent *int     `json:"progress_percent,omitempty"`
	Temperature    *float64  `json:"temperature,omitempty"`
	Pressure       *float64  `json:"pressure,omitempty"`
	StartTime      string    `json:"start_time"`
	EndTime        *string   `json:"end_time,omitempty"`
	Result         string    `json:"result,omitempty"`
}

// GetMelagStatusHandler handles GET /api/melag/{id}/status requests
func GetMelagStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Extract device ID from URL path
	deviceID, err := extractMelagDeviceIDFromStatusPath(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract device ID from status path", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_id",
			Message: "Invalid device ID in URL path",
		})
		return
	}

	// Get device to verify it exists and is a Melag device
	device, err := database.GetDevice(deviceID)
	if err != nil {
		if err == database.ErrDeviceNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "device_not_found",
				Message: fmt.Sprintf("Device with ID %d not found", deviceID),
			})
			return
		}

		logger.Error("Failed to get device", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve device",
		})
		return
	}

	// Verify device is a Melag device
	if device.Manufacturer != "Melag" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_type",
			Message: fmt.Sprintf("Device %d is not a Melag device", deviceID),
		})
		return
	}

	// Get device manager
	deviceManager := devices.GetManager()
	if deviceManager == nil {
		logger.Error("Device manager not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "service_unavailable",
			Message: "Device manager not initialized",
		})
		return
	}

	// Get device adapter to check connection status
	adapter := deviceManager.GetAdapter(deviceID)
	connected := false
	if adapter != nil {
		connected = adapter.IsConnected()
	}

	// Get device cycles
	cycles, err := database.GetDeviceCycles(deviceID)
	if err != nil {
		logger.Error("Failed to get device cycles", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycles",
		})
		return
	}

	// Build response
	response := GetMelagStatusResponse{
		DeviceID:  deviceID,
		Connected: connected,
	}

	// Find current cycle (running cycle)
	for _, cycle := range cycles {
		if cycle.EndTS == nil && cycle.Phase != "FAILED" && cycle.Phase != "COMPLETED" {
			// This is a running cycle
			startTime := cycle.StartTS.Format("2006-01-02T15:04:05Z07:00")
			cycleInfo := CycleStatusInfo{
				CycleID:        cycle.ID,
				Program:        cycle.Program,
				Phase:          cycle.Phase,
				ProgressPercent: cycle.ProgressPercent,
				Temperature:    cycle.Temperature,
				Pressure:       cycle.Pressure,
				StartTime:      startTime,
				Result:         cycle.Result,
			}
			response.CurrentCycle = &cycleInfo
			break // Only one running cycle expected
		}
	}

	// Find last cycle (most recent completed cycle)
	for _, cycle := range cycles {
		if cycle.EndTS != nil {
			startTime := cycle.StartTS.Format("2006-01-02T15:04:05Z07:00")
			endTime := cycle.EndTS.Format("2006-01-02T15:04:05Z07:00")
			cycleInfo := CycleStatusInfo{
				CycleID:        cycle.ID,
				Program:        cycle.Program,
				Phase:          cycle.Phase,
				ProgressPercent: cycle.ProgressPercent,
				Temperature:    cycle.Temperature,
				Pressure:       cycle.Pressure,
				StartTime:      startTime,
				EndTime:        &endTime,
				Result:         cycle.Result,
			}
			response.LastCycle = &cycleInfo
			break // Most recent completed cycle
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractMelagDeviceIDFromStatusPath extracts device ID from URL path like "/melag/1/status"
func extractMelagDeviceIDFromStatusPath(path string) (int, error) {
	// Path will be "/melag/1/status" after StripPrefix
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "melag" || parts[2] != "status" {
		return 0, fmt.Errorf("invalid path format: expected /melag/{id}/status")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid device ID: %w", err)
	}

	return id, nil
}

// GetCycleResponse represents the response for retrieving a cycle
type GetCycleResponse struct {
	CycleID          int       `json:"cycle_id"`
	DeviceID         int       `json:"device_id"`
	Program          string    `json:"program,omitempty"`
	Phase            string    `json:"phase,omitempty"`
	ProgressPercent  *int      `json:"progress_percent,omitempty"`
	Temperature      *float64  `json:"temperature,omitempty"`
	Pressure         *float64  `json:"pressure,omitempty"`
	StartTime        string    `json:"start_time"`
	EndTime          *string   `json:"end_time,omitempty"`
	Result           *string   `json:"result,omitempty"`
	ErrorDescription *string   `json:"error_description,omitempty"`
}

// GetCycleHandler handles GET /api/melag/{id}/cycles/{cycle_id} requests
func GetCycleHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Extract device ID and cycle ID from URL path
	deviceID, cycleID, err := extractMelagDeviceIDAndCycleID(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract device ID and cycle ID from path", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_path",
			Message: "Invalid path format: expected /melag/{device_id}/cycles/{cycle_id}",
		})
		return
	}

	// Get device to verify it exists and is a Melag device
	device, err := database.GetDevice(deviceID)
	if err != nil {
		if err == database.ErrDeviceNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "device_not_found",
				Message: fmt.Sprintf("Device with ID %d not found", deviceID),
			})
			return
		}

		logger.Error("Failed to get device", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve device",
		})
		return
	}

	// Verify device is a Melag device
	if device.Manufacturer != "Melag" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_type",
			Message: fmt.Sprintf("Device %d is not a Melag device", deviceID),
		})
		return
	}

	// Get cycle from database
	cycle, err := database.GetCycle(cycleID)
	if err != nil {
		if err == database.ErrCycleNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "cycle_not_found",
				Message: fmt.Sprintf("Cycle with ID %d not found", cycleID),
			})
			return
		}

		logger.Error("Failed to get cycle", "error", err, "cycle_id", cycleID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycle",
		})
		return
	}

	// Verify cycle belongs to the specified device
	if cycle.DeviceID != deviceID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "cycle_device_mismatch",
			Message: fmt.Sprintf("Cycle %d does not belong to device %d", cycleID, deviceID),
		})
		return
	}

	// Build response
	startTime := cycle.StartTS.Format("2006-01-02T15:04:05Z07:00")
	response := GetCycleResponse{
		CycleID:         cycle.ID,
		DeviceID:        cycle.DeviceID,
		Program:         cycle.Program,
		Phase:           cycle.Phase,
		ProgressPercent: cycle.ProgressPercent,
		Temperature:     cycle.Temperature,
		Pressure:        cycle.Pressure,
		StartTime:       startTime,
	}

	if cycle.EndTS != nil {
		endTime := cycle.EndTS.Format("2006-01-02T15:04:05Z07:00")
		response.EndTime = &endTime
	}

	if cycle.Result != "" {
		response.Result = &cycle.Result
	}

	if cycle.ErrorDescription != "" {
		response.ErrorDescription = &cycle.ErrorDescription
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractMelagDeviceIDAndCycleID extracts device ID and cycle ID from URL path like "/melag/1/cycles/5"
func extractMelagDeviceIDAndCycleID(path string) (int, int, error) {
	// Path will be "/melag/1/cycles/5" after StripPrefix
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[0] != "melag" || parts[2] != "cycles" {
		return 0, 0, fmt.Errorf("invalid path format: expected /melag/{device_id}/cycles/{cycle_id}")
	}

	deviceID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid device ID: %w", err)
	}

	cycleID, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid cycle ID: %w", err)
	}

	return deviceID, cycleID, nil
}

