package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
)

// CreateDeviceRequest represents the request body for creating a device
type CreateDeviceRequest struct {
	Name         string `json:"name"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer"` // "Melag" or "Getinge"
	IP           string `json:"ip"`
	Serial       string `json:"serial,omitempty"`
	Type         string `json:"type"` // "Steri" or "RDG"
	Location     string `json:"location,omitempty"`
}

// CreateDeviceResponse represents the response for creating a device
type CreateDeviceResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Model        string    `json:"model,omitempty"`
	Manufacturer string    `json:"manufacturer"`
	IP           string    `json:"ip"`
	Serial       string    `json:"serial,omitempty"`
	Type         string    `json:"type"`
	Location     string    `json:"location,omitempty"`
	Created      string    `json:"created"`
	Updated      string    `json:"updated"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// CreateDeviceHandler handles POST /api/devices requests
func CreateDeviceHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req CreateDeviceRequest
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

	// Validate required fields
	if err := validateCreateDeviceRequest(&req); err != nil {
		logger.Warn("Device creation validation failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Create device in database
	device := &database.Device{
		Name:         req.Name,
		Model:        req.Model,
		Manufacturer: req.Manufacturer,
		IP:           req.IP,
		Serial:       req.Serial,
		Type:         req.Type,
		Location:     req.Location,
	}

	createdDevice, err := database.CreateDevice(device)
	if err != nil {
		logger.Error("Failed to create device", "error", err, "device_name", req.Name, "ip", req.IP)

		// Check for duplicate device error
		if err == database.ErrDuplicateDevice {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "duplicate_device",
				Message: "Device with same IP and manufacturer already exists",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create device",
		})
		return
	}

	// Log audit entry
	deviceID := createdDevice.ID
	details := map[string]interface{}{
		"name":         createdDevice.Name,
		"manufacturer": createdDevice.Manufacturer,
		"ip":           createdDevice.IP,
		"type":         createdDevice.Type,
	}

	if err := database.LogAudit(database.ActionDeviceAdded, "device", &deviceID, "", details); err != nil {
		logger.Warn("Failed to create audit log", "error", err)
		// Continue even if audit log fails
	}

	logger.Info("Device created successfully",
		"device_id", createdDevice.ID,
		"name", createdDevice.Name,
		"manufacturer", createdDevice.Manufacturer,
		"ip", createdDevice.IP)

	// Return created device
	response := CreateDeviceResponse{
		ID:           createdDevice.ID,
		Name:         createdDevice.Name,
		Model:        createdDevice.Model,
		Manufacturer: createdDevice.Manufacturer,
		IP:           createdDevice.IP,
		Serial:       createdDevice.Serial,
		Type:         createdDevice.Type,
		Location:     createdDevice.Location,
		Created:      createdDevice.Created.Format("2006-01-02T15:04:05Z07:00"),
		Updated:      createdDevice.Updated.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateDeviceRequest represents the request body for updating a device
type UpdateDeviceRequest struct {
	Name     string `json:"name,omitempty"`
	Model    string `json:"model,omitempty"`
	IP       string `json:"ip,omitempty"`
	Serial   string `json:"serial,omitempty"`
	Type     string `json:"type,omitempty"` // "Steri" or "RDG"
	Location string `json:"location,omitempty"`
}

// GetDeviceHandler handles GET /api/devices/{id} requests
func GetDeviceHandler(w http.ResponseWriter, r *http.Request) {
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
	// Path will be like "/devices/1" after StripPrefix
	deviceID, err := extractDeviceID(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract device ID", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_id",
			Message: "Invalid device ID in URL path",
		})
		return
	}

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

	response := CreateDeviceResponse{
		ID:           device.ID,
		Name:         device.Name,
		Model:        device.Model,
		Manufacturer: device.Manufacturer,
		IP:           device.IP,
		Serial:       device.Serial,
		Type:         device.Type,
		Location:     device.Location,
		Created:      device.Created.Format("2006-01-02T15:04:05Z07:00"),
		Updated:      device.Updated.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateDeviceHandler handles PUT /api/devices/{id} requests
func UpdateDeviceHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodPut {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only PUT method is allowed",
		})
		return
	}

	// Extract device ID from URL path
	deviceID, err := extractDeviceID(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract device ID", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_id",
			Message: "Invalid device ID in URL path",
		})
		return
	}

	// Parse request body
	var req UpdateDeviceRequest
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

	// Validate updatable fields
	if err := validateUpdateDeviceRequest(&req); err != nil {
		logger.Warn("Device update validation failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Build update struct
	updates := &database.Device{
		Name:     req.Name,
		Model:    req.Model,
		IP:       req.IP,
		Serial:   req.Serial,
		Type:     req.Type,
		Location: req.Location,
	}

	// Update device in database
	updatedDevice, err := database.UpdateDevice(deviceID, updates)
	if err != nil {
		logger.Error("Failed to update device", "error", err, "device_id", deviceID)

		if err == database.ErrDeviceNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "device_not_found",
				Message: fmt.Sprintf("Device with ID %d not found", deviceID),
			})
			return
		}

		if err == database.ErrDuplicateDevice {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "duplicate_device",
				Message: "Device with same IP and manufacturer already exists",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update device",
		})
		return
	}

	// Log audit entry
	details := map[string]interface{}{
		"device_id":    updatedDevice.ID,
		"name":         updatedDevice.Name,
		"manufacturer": updatedDevice.Manufacturer,
		"ip":           updatedDevice.IP,
		"type":         updatedDevice.Type,
	}

	if err := database.LogAudit(database.ActionDeviceUpdated, "device", &deviceID, "", details); err != nil {
		logger.Warn("Failed to create audit log", "error", err)
		// Continue even if audit log fails
	}

	logger.Info("Device updated successfully",
		"device_id", updatedDevice.ID,
		"name", updatedDevice.Name,
		"manufacturer", updatedDevice.Manufacturer,
		"ip", updatedDevice.IP)

	// Return updated device
	response := CreateDeviceResponse{
		ID:           updatedDevice.ID,
		Name:         updatedDevice.Name,
		Model:        updatedDevice.Model,
		Manufacturer: updatedDevice.Manufacturer,
		IP:           updatedDevice.IP,
		Serial:       updatedDevice.Serial,
		Type:         updatedDevice.Type,
		Location:     updatedDevice.Location,
		Created:      updatedDevice.Created.Format("2006-01-02T15:04:05Z07:00"),
		Updated:      updatedDevice.Updated.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractDeviceID extracts device ID from URL path like "/devices/1"
func extractDeviceID(path string) (int, error) {
	// Path will be "/devices/1" after StripPrefix
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] != "devices" {
		return 0, fmt.Errorf("invalid path format: expected /devices/{id}")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid device ID: %w", err)
	}

	return id, nil
}

// validateUpdateDeviceRequest validates the update device request
func validateUpdateDeviceRequest(req *UpdateDeviceRequest) error {
	// Validate type if provided
	if req.Type != "" && req.Type != "Steri" && req.Type != "RDG" {
		return fmt.Errorf("type must be 'Steri' or 'RDG'")
	}

	// Validate IP address format if provided
	if req.IP != "" {
		if ip := net.ParseIP(req.IP); ip == nil {
			return fmt.Errorf("invalid IP address format: %s", req.IP)
		}
	}

	return nil
}

// ListDevicesHandler handles GET /api/devices requests
func ListDevicesHandler(w http.ResponseWriter, r *http.Request) {
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

	devices, err := database.GetAllDevices()
	if err != nil {
		logger.Error("Failed to get devices", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve devices",
		})
		return
	}

	// Convert to response format
	response := make([]CreateDeviceResponse, 0, len(devices))
	for _, device := range devices {
		response = append(response, CreateDeviceResponse{
			ID:           device.ID,
			Name:         device.Name,
			Model:        device.Model,
			Manufacturer: device.Manufacturer,
			IP:           device.IP,
			Serial:       device.Serial,
			Type:         device.Type,
			Location:     device.Location,
			Created:      device.Created.Format("2006-01-02T15:04:05Z07:00"),
			Updated:      device.Updated.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteDeviceHandler handles DELETE /api/devices/{id} requests
func DeleteDeviceHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only DELETE method is allowed",
		})
		return
	}

	// Extract device ID from URL path
	deviceID, err := extractDeviceID(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract device ID", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_device_id",
			Message: "Invalid device ID in URL path",
		})
		return
	}

	// Get device before deletion for audit log
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

		logger.Error("Failed to get device for deletion", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve device",
		})
		return
	}

	// Delete device from database
	if err := database.DeleteDevice(deviceID); err != nil {
		logger.Error("Failed to delete device", "error", err, "device_id", deviceID)

		if err == database.ErrDeviceNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "device_not_found",
				Message: fmt.Sprintf("Device with ID %d not found", deviceID),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete device",
		})
		return
	}

	// Log audit entry (using device data retrieved before deletion)
	details := map[string]interface{}{
		"device_id":    device.ID,
		"name":         device.Name,
		"manufacturer": device.Manufacturer,
		"ip":           device.IP,
		"type":         device.Type,
	}

	if err := database.LogAudit(database.ActionDeviceDeleted, "device", &deviceID, "", details); err != nil {
		logger.Warn("Failed to create audit log", "error", err)
		// Continue even if audit log fails
	}

	logger.Info("Device deleted successfully",
		"device_id", deviceID,
		"name", device.Name,
		"manufacturer", device.Manufacturer,
		"ip", device.IP)

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}

// validateCreateDeviceRequest validates the create device request
func validateCreateDeviceRequest(req *CreateDeviceRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Manufacturer == "" {
		return fmt.Errorf("manufacturer is required")
	}
	if req.Manufacturer != "Melag" && req.Manufacturer != "Getinge" {
		return fmt.Errorf("manufacturer must be 'Melag' or 'Getinge'")
	}

	if req.IP == "" {
		return fmt.Errorf("ip is required")
	}
	// Validate IP address format
	if ip := net.ParseIP(req.IP); ip == nil {
		return fmt.Errorf("invalid IP address format: %s", req.IP)
	}

	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	if req.Type != "Steri" && req.Type != "RDG" {
		return fmt.Errorf("type must be 'Steri' or 'RDG'")
	}

	return nil
}

// DeviceStatusResponse represents the response for device status
type DeviceStatusResponse struct {
	DeviceID      int        `json:"device_id"`
	Connected     bool       `json:"connected"`
	LastSeen      *string    `json:"last_seen,omitempty"`
	HealthStatus  string     `json:"health_status"`
	Manufacturer  string     `json:"manufacturer"`
	IP            string     `json:"ip"`
	ConnectionType   *string `json:"connection_type,omitempty"`
	LastCycleStatus  *string `json:"last_cycle_status,omitempty"`
	ICMPReachable *bool      `json:"icmp_reachable,omitempty"`
	LastPingTime  *string    `json:"last_ping_time,omitempty"`
}

// GetDeviceStatusHandler handles GET /api/devices/{id}/status requests
func GetDeviceStatusHandler(w http.ResponseWriter, r *http.Request) {
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
	// Path will be like "/devices/1/status" after StripPrefix
	deviceID, err := extractDeviceIDFromStatusPath(r.URL.Path)
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

	// Get device status
	status, err := database.GetDeviceStatus(deviceID)
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

		logger.Error("Failed to get device status", "error", err, "device_id", deviceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve device status",
		})
		return
	}

	// Convert to response format
	response := DeviceStatusResponse{
		DeviceID:     status.DeviceID,
		Connected:    status.Connected,
		HealthStatus: status.HealthStatus,
		Manufacturer: status.Manufacturer,
		IP:           status.IP,
	}

	// Format LastSeen timestamp
	if status.LastSeen != nil {
		lastSeenStr := status.LastSeen.Format("2006-01-02T15:04:05Z07:00")
		response.LastSeen = &lastSeenStr
	}

	// Melag-specific fields
	if status.Manufacturer == "Melag" {
		if status.ConnectionType != "" {
			response.ConnectionType = &status.ConnectionType
		}
		if status.LastCycleStatus != "" {
			response.LastCycleStatus = &status.LastCycleStatus
		}
	}

	// Getinge-specific fields
	if status.Manufacturer == "Getinge" {
		response.ICMPReachable = &status.ICMPReachable
		if status.LastPingTime != nil {
			lastPingStr := status.LastPingTime.Format("2006-01-02T15:04:05Z07:00")
			response.LastPingTime = &lastPingStr
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractDeviceIDFromStatusPath extracts device ID from URL path like "/devices/1/status"
func extractDeviceIDFromStatusPath(path string) (int, error) {
	// Path will be "/devices/1/status" after StripPrefix
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "devices" || parts[2] != "status" {
		return 0, fmt.Errorf("invalid path format: expected /devices/{id}/status")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid device ID: %w", err)
	}

	return id, nil
}
