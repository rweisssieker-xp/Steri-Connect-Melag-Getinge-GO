package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/devices"
	"steri-connect-go/internal/logging"
)

// DiagnosticsResponse represents the device diagnostics response
type DiagnosticsResponse struct {
	DeviceID                    int                      `json:"device_id"`
	DeviceName                  string                   `json:"device_name"`
	Manufacturer                string                   `json:"manufacturer"`
	ConnectionTest              ConnectionTestResult     `json:"connection_test"`
	LastSuccessfulCommunication *time.Time              `json:"last_successful_communication,omitempty"`
	ProtocolInfo                ProtocolDiagnostics      `json:"protocol_info"`
	RecentErrors                []ErrorLogEntry          `json:"recent_errors"`
	ConnectionHistory           []ConnectionHistoryEntry `json:"connection_history"`
	Timestamp                   time.Time                `json:"timestamp"`
}

// ConnectionTestResult represents the result of a connection test
type ConnectionTestResult struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration,omitempty"`
}

// ProtocolDiagnostics contains protocol-specific debugging information
type ProtocolDiagnostics struct {
	Type         string                 `json:"type"` // "Melag" or "Getinge"
	MelagInfo    *MelagDiagnostics     `json:"melag_info,omitempty"`
	GetingeInfo  *GetingeDiagnostics   `json:"getinge_info,omitempty"`
}

// MelagDiagnostics contains Melag-specific diagnostic information
type MelagDiagnostics struct {
	FTPConnected    bool      `json:"ftp_connected"`
	LastFTPAccess   *time.Time `json:"last_ftp_access,omitempty"`
	ProtocolFileAccessible bool `json:"protocol_file_accessible"`
}

// GetingeDiagnostics contains Getinge-specific diagnostic information
type GetingeDiagnostics struct {
	LastPingTime    *time.Time `json:"last_ping_time,omitempty"`
	LastPingSuccess bool      `json:"last_ping_success"`
	PingHistory     []PingEntry `json:"ping_history"`
}

// PingEntry represents a single ping result
type PingEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Reachable bool      `json:"reachable"`
}

// ErrorLogEntry represents an error log entry
type ErrorLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
}

// ConnectionHistoryEntry represents a connection attempt
type ConnectionHistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
}

// DiagnosticsHandler handles GET /api/diagnostics/{deviceId} requests
func DiagnosticsHandler(w http.ResponseWriter, r *http.Request) {
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
	deviceIDStr := r.URL.Path[len("/api/diagnostics/"):]
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		logger.Warn("Invalid device ID in diagnostics request", "device_id", deviceIDStr, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_id",
			Message: "Invalid device ID",
		})
		return
	}

	// Get device from database
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
		logger.Error("Failed to get device", "device_id", deviceID, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve device",
		})
		return
	}

	// Perform connection test
	connectionTest := performConnectionTest(deviceID, device.Manufacturer)

	// Get last successful communication timestamp
	lastSuccessfulComm := getLastSuccessfulCommunication(deviceID, device.Manufacturer)

	// Get protocol-specific debugging information
	protocolInfo := getProtocolDiagnostics(deviceID, device.Manufacturer)

	// Get recent error logs for this device
	recentErrors := getRecentErrors(deviceID)

	// Get connection attempt history
	connectionHistory := getConnectionHistory(deviceID)

	response := DiagnosticsResponse{
		DeviceID:                    device.ID,
		DeviceName:                  device.Name,
		Manufacturer:                device.Manufacturer,
		ConnectionTest:              connectionTest,
		LastSuccessfulCommunication: lastSuccessfulComm,
		ProtocolInfo:                protocolInfo,
		RecentErrors:                recentErrors,
		ConnectionHistory:           connectionHistory,
		Timestamp:                   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// performConnectionTest performs a connection test for the device
func performConnectionTest(deviceID int, manufacturer string) ConnectionTestResult {
	startTime := time.Now()
	deviceManager := devices.GetManager()
	if deviceManager == nil {
		return ConnectionTestResult{
			Success:   false,
			Timestamp: time.Now(),
			Error:     "Device manager not initialized",
		}
	}

	adapter := deviceManager.GetAdapter(deviceID)
	if adapter == nil {
		return ConnectionTestResult{
			Success:   false,
			Timestamp: time.Now(),
			Error:     "Device adapter not found",
		}
	}

	// Try to connect
	err := adapter.Connect()
	duration := time.Since(startTime)

	if err != nil {
		return ConnectionTestResult{
			Success:   false,
			Timestamp: time.Now(),
			Error:     err.Error(),
			Duration:  duration.String(),
		}
	}

	// Disconnect after test
	adapter.Disconnect()

	return ConnectionTestResult{
		Success:   true,
		Timestamp: time.Now(),
		Duration:  duration.String(),
	}
}

// getLastSuccessfulCommunication gets the last successful communication timestamp
func getLastSuccessfulCommunication(deviceID int, manufacturer string) *time.Time {
	status, err := database.GetDeviceStatus(deviceID)
	if err != nil {
		return nil
	}

	if status.LastSeen != nil {
		return status.LastSeen
	}

	// For Getinge devices, check last ping time
	if manufacturer == "Getinge" && status.LastPingTime != nil {
		return status.LastPingTime
	}

	return nil
}

// getProtocolDiagnostics gets protocol-specific debugging information
func getProtocolDiagnostics(deviceID int, manufacturer string) ProtocolDiagnostics {
	protocolInfo := ProtocolDiagnostics{
		Type: manufacturer,
	}

	if manufacturer == "Melag" {
		deviceManager := devices.GetManager()
		if deviceManager != nil {
			adapter := deviceManager.GetAdapter(deviceID)
			if adapter != nil {
				connected := adapter.IsConnected()
				protocolInfo.MelagInfo = &MelagDiagnostics{
					FTPConnected:          connected,
					ProtocolFileAccessible: connected, // Simplified - assumes FTP connection means file access
				}
			}
		}
	} else if manufacturer == "Getinge" {
		// Get ping history
		pingHistory, err := database.GetRDGStatusHistory(deviceID, 10)
		if err == nil {
			pingEntries := make([]PingEntry, 0, len(pingHistory))
			for _, status := range pingHistory {
				pingEntries = append(pingEntries, PingEntry{
					Timestamp: status.Timestamp,
					Reachable: status.Reachable,
				})
			}

			lastPingSuccess := false
			var lastPingTime *time.Time
			if len(pingHistory) > 0 {
				lastPingSuccess = pingHistory[0].Reachable
				lastPingTime = &pingHistory[0].Timestamp
			}

			protocolInfo.GetingeInfo = &GetingeDiagnostics{
				LastPingTime:    lastPingTime,
				LastPingSuccess: lastPingSuccess,
				PingHistory:     pingEntries,
			}
		}
	}

	return protocolInfo
}

// getRecentErrors gets recent error logs for the device
func getRecentErrors(deviceID int) []ErrorLogEntry {
	// Get audit logs for this device
	auditLogs, err := database.GetAuditLogs("device", &deviceID, 20)
	if err != nil {
		return []ErrorLogEntry{}
	}

	errors := make([]ErrorLogEntry, 0)
	for _, log := range auditLogs {
		// Filter for error-related actions
		if log.Action == "cycle_failed" || log.Action == "rdg_status_update" {
			errors = append(errors, ErrorLogEntry{
				Timestamp: log.Timestamp,
				Action:    log.Action,
				Details:   log.Details,
			})
		}
	}

	return errors
}

// getConnectionHistory gets connection attempt history
func getConnectionHistory(deviceID int) []ConnectionHistoryEntry {
	// Get audit logs for this device
	auditLogs, err := database.GetAuditLogs("device", &deviceID, 20)
	if err != nil {
		return []ConnectionHistoryEntry{}
	}

	history := make([]ConnectionHistoryEntry, 0)
	for _, log := range auditLogs {
		// Include device-related actions
		if log.Action == "device_added" || log.Action == "device_updated" || 
		   log.Action == "rdg_status_update" || log.Action == "cycle_started" {
			success := true
			if log.Action == "rdg_status_update" {
				// Parse details to determine success
				// Simplified - assume success if action exists
			}
			history = append(history, ConnectionHistoryEntry{
				Timestamp: log.Timestamp,
				Success:   success,
				Action:    log.Action,
				Details:   log.Details,
			})
		}
	}

	return history
}

