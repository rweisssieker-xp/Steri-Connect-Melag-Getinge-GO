package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDuplicateDevice = errors.New("device with same IP and manufacturer already exists")
	ErrDeviceNotFound  = errors.New("device not found")
)

// CreateDevice creates a new device in the database
func CreateDevice(device *Device) (*Device, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Check for duplicate IP+manufacturer combination
	var existingID int
	checkQuery := `SELECT id FROM devices WHERE ip = ? AND manufacturer = ?`
	err := db.QueryRow(checkQuery, device.IP, device.Manufacturer).Scan(&existingID)
	if err == nil {
		return nil, ErrDuplicateDevice
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for duplicate device: %w", err)
	}

	// Insert new device
	now := time.Now()
	query := `
		INSERT INTO devices (name, model, manufacturer, ip, serial, type, location, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		device.Name,
		device.Model,
		device.Manufacturer,
		device.IP,
		device.Serial,
		device.Type,
		device.Location,
		now,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get device ID: %w", err)
	}

	// Retrieve the created device to get timestamps from database
	createdDevice, err := GetDevice(int(id))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created device: %w", err)
	}

	return createdDevice, nil
}

// GetDevice retrieves a device by ID
func GetDevice(id int) (*Device, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, name, model, manufacturer, ip, serial, type, location, created, updated
		FROM devices
		WHERE id = ?
	`

	device := &Device{}
	err := db.QueryRow(query, id).Scan(
		&device.ID,
		&device.Name,
		&device.Model,
		&device.Manufacturer,
		&device.IP,
		&device.Serial,
		&device.Type,
		&device.Location,
		&device.Created,
		&device.Updated,
	)

	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return device, nil
}

// GetAllDevices retrieves all devices from the database
func GetAllDevices() ([]Device, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, name, model, manufacturer, ip, serial, type, location, created, updated
		FROM devices
		ORDER BY created DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var device Device
		err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.Model,
			&device.Manufacturer,
			&device.IP,
			&device.Serial,
			&device.Type,
			&device.Location,
			&device.Created,
			&device.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating devices: %w", err)
	}

	return devices, nil
}

// UpdateDevice updates a device in the database (partial update)
func UpdateDevice(id int, updates *Device) (*Device, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// First, get existing device to ensure it exists and preserve manufacturer
	existingDevice, err := GetDevice(id)
	if err != nil {
		return nil, err // Returns ErrDeviceNotFound if not found
	}

	// Build update query dynamically based on provided fields
	setParts := []string{}
	args := []interface{}{}

	// Only update fields that are provided (non-zero values)
	if updates.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, updates.Name)
	}
	if updates.Model != "" || updates.Model == "" && updates.Model != existingDevice.Model {
		// Allow empty string to clear model
		setParts = append(setParts, "model = ?")
		args = append(args, updates.Model)
	}
	if updates.IP != "" {
		// Check for duplicate IP+manufacturer if IP is being changed
		if updates.IP != existingDevice.IP {
			var existingID int
			checkQuery := `SELECT id FROM devices WHERE ip = ? AND manufacturer = ? AND id != ?`
			err := db.QueryRow(checkQuery, updates.IP, existingDevice.Manufacturer, id).Scan(&existingID)
			if err == nil {
				return nil, ErrDuplicateDevice
			} else if err != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to check for duplicate device: %w", err)
			}
		}
		setParts = append(setParts, "ip = ?")
		args = append(args, updates.IP)
	}
	if updates.Serial != "" || updates.Serial == "" && updates.Serial != existingDevice.Serial {
		// Allow empty string to clear serial
		setParts = append(setParts, "serial = ?")
		args = append(args, updates.Serial)
	}
	if updates.Type != "" {
		setParts = append(setParts, "type = ?")
		args = append(args, updates.Type)
	}
	if updates.Location != "" || updates.Location == "" && updates.Location != existingDevice.Location {
		// Allow empty string to clear location
		setParts = append(setParts, "location = ?")
		args = append(args, updates.Location)
	}

	// If no fields to update, return existing device
	if len(setParts) == 0 {
		return existingDevice, nil
	}

	// Always update the updated timestamp
	setParts = append(setParts, "updated = ?")
	args = append(args, time.Now())

	// Add device ID to args for WHERE clause
	args = append(args, id)

	// Build and execute update query
	query := fmt.Sprintf(`
		UPDATE devices
		SET %s
		WHERE id = ?
	`, strings.Join(setParts, ", "))

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Retrieve updated device
	return GetDevice(id)
}

// DeleteDevice deletes a device from the database
func DeleteDevice(id int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check if device exists before deletion
	_, err := GetDevice(id)
	if err != nil {
		return err // Returns ErrDeviceNotFound if not found
	}

	// Delete device (CASCADE will handle cycles and rdg_status)
	query := `DELETE FROM devices WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	// Verify deletion
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify deletion: %w", err)
	}
	if rowsAffected == 0 {
		// This shouldn't happen since we checked existence above, but handle it anyway
		return ErrDeviceNotFound
	}

	return nil
}

// GetDeviceStatus retrieves the health and connection status of a device
func GetDeviceStatus(id int) (*DeviceStatus, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Get device from database
	device, err := GetDevice(id)
	if err != nil {
		return nil, err // Returns ErrDeviceNotFound if not found
	}

	// For MVP: Placeholder status calculation
	// Real connection state will be implemented in Epic 3 (Device Integration)
	status := &DeviceStatus{
		DeviceID:     device.ID,
		Manufacturer: device.Manufacturer,
		IP:           device.IP,
		Connected:    false,       // Placeholder - will be real in Epic 3
		HealthStatus: "unhealthy", // Default, will be calculated below
	}

	// Calculate health status based on last communication
	// For MVP: Use device.Updated as proxy for last_seen
	// In Epic 3, this will be actual last communication time
	lastSeen := device.Updated
	status.LastSeen = &lastSeen

	// Calculate health status based on time since last update
	timeSinceUpdate := time.Since(device.Updated)
	if timeSinceUpdate < 5*time.Minute {
		status.HealthStatus = "healthy"
	} else if timeSinceUpdate < 15*time.Minute {
		status.HealthStatus = "degraded"
	} else {
		status.HealthStatus = "unhealthy"
	}

	// Manufacturer-specific fields
	if device.Manufacturer == "Melag" {
		// Placeholder for Melag-specific status
		// Will be populated in Epic 3 with actual MELAnet connection status
		status.ConnectionType = "MELAnet" // Default assumption
		status.LastCycleStatus = ""       // Will be populated from cycles table in future
	} else if device.Manufacturer == "Getinge" {
		// Placeholder for Getinge-specific status
		// Will be populated in Epic 3 with actual ICMP ping results
		status.ICMPReachable = false // Default, will be real in Epic 3
		// LastPingTime will be populated from rdg_status table in Epic 3
	}

	return status, nil
}

// UpdateDeviceConnectionStatus updates the connection status for a device (status stored in rdg_status table for Getinge)
// This function updates the device's last_seen timestamp
func UpdateDeviceConnectionStatus(deviceID int, connected bool) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Update last_seen timestamp in devices table
	// Connection status for Getinge devices is stored in rdg_status table
	query := `
		UPDATE devices
		SET updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device connection status: %w", err)
	}
	return nil
}

// UpdateDeviceLastSeen updates the last seen timestamp for a device
func UpdateDeviceLastSeen(deviceID int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
		UPDATE devices
		SET last_seen = ?, updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, time.Now(), deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device last seen: %w", err)
	}
	return nil
}
