package adapters

import (
	"time"
)

// DeviceAdapter defines the interface for device adapters
type DeviceAdapter interface {
	// Connect establishes connection to the device
	Connect() error

	// Disconnect closes connection to the device
	Disconnect() error

	// IsConnected returns true if device is currently connected
	IsConnected() bool

	// GetDeviceID returns the device ID
	GetDeviceID() int

	// GetConnectionState returns current connection state
	GetConnectionState() ConnectionState
}

// ConnectionState represents the connection state of a device
type ConnectionState string

const (
	StateDisconnected ConnectionState = "DISCONNECTED"
	StateConnecting   ConnectionState = "CONNECTING"
	StateConnected    ConnectionState = "CONNECTED"
	StateError        ConnectionState = "ERROR"
)

// ConnectionInfo holds connection status information
type ConnectionInfo struct {
	State     ConnectionState `json:"state"`
	Connected bool           `json:"connected"`
	LastSeen  *time.Time     `json:"last_seen,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// CycleStartParams holds parameters for starting a cycle
type CycleStartParams struct {
	Program     string   `json:"program,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Pressure    *float64 `json:"pressure,omitempty"`
	Duration    *int     `json:"duration,omitempty"` // Duration in minutes
}

// CycleStatus represents the current status of a running cycle
type CycleStatus struct {
	Phase           string        `json:"phase"` // "STARTING", "RUNNING", "Aufheizen", "Sterilisation", "Trocknung", "COMPLETED", "FAILED"
	ProgressPercent int           `json:"progress_percent"` // 0-100
	Temperature     *float64      `json:"temperature,omitempty"`
	Pressure        *float64      `json:"pressure,omitempty"`
	TimeRemaining   *time.Duration `json:"time_remaining,omitempty"` // Time remaining
	IsRunning       bool          `json:"is_running"` // true if cycle is actively running
	Error           string        `json:"error,omitempty"`
}

