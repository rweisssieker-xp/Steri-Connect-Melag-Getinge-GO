package database

import "time"

// Device represents a medical sterilization/cleaning device
type Device struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Model        string    `json:"model,omitempty" db:"model"`
	Manufacturer string    `json:"manufacturer" db:"manufacturer"` // "Melag" or "Getinge"
	IP           string    `json:"ip" db:"ip"`
	Serial       string    `json:"serial,omitempty" db:"serial"`
	Type         string    `json:"type" db:"type"` // "Steri" or "RDG"
	Location     string    `json:"location,omitempty" db:"location"`
	Created      time.Time `json:"created" db:"created"`
	Updated      time.Time `json:"updated" db:"updated"`
}

// Cycle represents a sterilization or cleaning cycle
type Cycle struct {
	ID               int        `json:"id" db:"id"`
	DeviceID         int        `json:"device_id" db:"device_id"`
	Program          string     `json:"program,omitempty" db:"program"`
	StartTS          time.Time  `json:"start_ts" db:"start_ts"`
	EndTS            *time.Time `json:"end_ts,omitempty" db:"end_ts"`
	Result           string     `json:"result,omitempty" db:"result"` // "OK" or "NOK"
	ErrorCode        string     `json:"error_code,omitempty" db:"error_code"`
	ErrorDescription string     `json:"error_description,omitempty" db:"error_description"`
	Phase            string     `json:"phase,omitempty" db:"phase"` // "Aufheizen", "Sterilisation", "Trocknung"
	Temperature      *float64   `json:"temperature,omitempty" db:"temperature"`
	Pressure         *float64   `json:"pressure,omitempty" db:"pressure"`
	ProgressPercent  *int       `json:"progress_percent,omitempty" db:"progress_percent"`
}

// RDGStatus represents Getinge device reachability status
type RDGStatus struct {
	ID        int       `json:"id" db:"id"`
	DeviceID  int       `json:"device_id" db:"device_id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Reachable bool      `json:"reachable" db:"reachable"` // true = reachable, false = unreachable
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID         int       `json:"id" db:"id"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Action     string    `json:"action" db:"action"` // "device_added", "cycle_started", etc.
	EntityType string    `json:"entity_type,omitempty" db:"entity_type"` // "device", "cycle", etc.
	EntityID   *int      `json:"entity_id,omitempty" db:"entity_id"`
	User       string    `json:"user,omitempty" db:"user"`
	Details    string    `json:"details,omitempty" db:"details"` // JSON string
	Hash       string    `json:"hash,omitempty" db:"hash"`       // Integrity hash
}

