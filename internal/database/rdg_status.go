package database

import (
	"fmt"
	"time"
)

// CreateRDGStatus creates a new RDG status entry
func CreateRDGStatus(deviceID int, reachable bool) (*RDGStatus, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		INSERT INTO rdg_status (device_id, timestamp, reachable)
		VALUES (?, ?, ?)
	`

	result, err := db.Exec(query, deviceID, time.Now(), reachable)
	if err != nil {
		return nil, fmt.Errorf("failed to create rdg status: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &RDGStatus{
		ID:        int(id),
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Reachable: reachable,
	}, nil
}

// GetLatestRDGStatus retrieves the latest RDG status for a device
func GetLatestRDGStatus(deviceID int) (*RDGStatus, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, device_id, timestamp, reachable
		FROM rdg_status
		WHERE device_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var status RDGStatus
	var reachableInt int

	err := db.QueryRow(query, deviceID).Scan(
		&status.ID,
		&status.DeviceID,
		&status.Timestamp,
		&reachableInt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get rdg status: %w", err)
	}

	status.Reachable = reachableInt == 1

	return &status, nil
}

// GetRDGStatusHistory retrieves RDG status history for a device
func GetRDGStatusHistory(deviceID int, limit int) ([]RDGStatus, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 100 // Default limit
	}

	query := `
		SELECT id, device_id, timestamp, reachable
		FROM rdg_status
		WHERE device_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.Query(query, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query rdg status history: %w", err)
	}
	defer rows.Close()

	var statuses []RDGStatus
	for rows.Next() {
		var status RDGStatus
		var reachableInt int

		err := rows.Scan(
			&status.ID,
			&status.DeviceID,
			&status.Timestamp,
			&reachableInt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rdg status: %w", err)
		}

		status.Reachable = reachableInt == 1
		statuses = append(statuses, status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rdg status rows: %w", err)
	}

	return statuses, nil
}

