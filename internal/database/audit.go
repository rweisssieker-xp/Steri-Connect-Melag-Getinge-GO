package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// AuditAction represents an audit log action
type AuditAction string

const (
	ActionDeviceAdded     AuditAction = "device_added"
	ActionDeviceUpdated   AuditAction = "device_updated"
	ActionDeviceDeleted   AuditAction = "device_deleted"
	ActionCycleStarted    AuditAction = "cycle_started"
	ActionCycleUpdated    AuditAction = "cycle_updated"
	ActionCycleCompleted  AuditAction = "cycle_completed"
	ActionCycleFailed     AuditAction = "cycle_failed"
	ActionRDGStatusUpdate AuditAction = "rdg_status_update"
)

// LogAudit writes an audit log entry to the database
func LogAudit(action AuditAction, entityType string, entityID *int, user string, details map[string]interface{}) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Convert details to JSON
	detailsJSON := ""
	if details != nil {
		jsonBytes, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
		detailsJSON = string(jsonBytes)
	}

	// Calculate hash for integrity verification
	hash := calculateAuditHash(action, entityType, entityID, user, detailsJSON, time.Now())

	// Insert audit log entry
	query := `
		INSERT INTO audit_log (timestamp, action, entity_type, entity_id, user, details, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query, time.Now(), string(action), entityType, entityID, user, detailsJSON, hash)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	// Verify insertion (for immutability check)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify audit log insertion: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("unexpected rows affected: %d", rowsAffected)
	}

	return nil
}

// calculateAuditHash calculates a hash for audit log integrity verification
func calculateAuditHash(action AuditAction, entityType string, entityID *int, user string, details string, timestamp time.Time) string {
	// Create hash input from all fields
	hashInput := fmt.Sprintf("%s|%s|%v|%s|%s|%s",
		string(action),
		entityType,
		entityID,
		user,
		details,
		timestamp.Format(time.RFC3339Nano))

	// Calculate SHA256 hash
	hasher := sha256.New()
	hasher.Write([]byte(hashInput))
	hashBytes := hasher.Sum(nil)

	// Return hex-encoded hash
	return hex.EncodeToString(hashBytes)
}

// GetAuditLogs retrieves audit logs with optional filters
func GetAuditLogs(entityType string, entityID *int, limit int) ([]AuditLog, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, timestamp, action, entity_type, entity_id, user, details, hash
		FROM audit_log
		WHERE 1=1
	`

	args := []interface{}{}

	if entityType != "" {
		query += " AND entity_type = ?"
		args = append(args, entityType)
	}

	if entityID != nil {
		query += " AND entity_id = ?"
		args = append(args, *entityID)
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.User,
			&log.Details,
			&log.Hash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, nil
}

