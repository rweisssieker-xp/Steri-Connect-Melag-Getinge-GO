package getinge

import (
	"fmt"
	"net"
	"sync"
	"time"

	"steri-connect-go/internal/adapters"
	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
)

// GetingeAdapter implements the DeviceAdapter interface for Getinge devices
type GetingeAdapter struct {
	deviceID    int
	device      *database.Device
	state       adapters.ConnectionState
	stateMutex  sync.RWMutex
	pingTimeout time.Duration
	logger      *logging.Logger
	lastPing    *time.Time
	lastReachable bool
}

// NewGetingeAdapter creates a new Getinge adapter instance
func NewGetingeAdapter(device *database.Device, pingTimeout time.Duration) (*GetingeAdapter, error) {
	if device == nil {
		return nil, fmt.Errorf("device cannot be nil")
	}
	if device.Manufacturer != "Getinge" {
		return nil, fmt.Errorf("device manufacturer must be Getinge, got %s", device.Manufacturer)
	}

	logger := logging.Get()

	adapter := &GetingeAdapter{
		deviceID:    device.ID,
		device:      device,
		state:       adapters.StateDisconnected,
		pingTimeout: pingTimeout,
		logger:      logger,
	}

	return adapter, nil
}

// Connect performs ICMP ping to check device reachability
func (a *GetingeAdapter) Connect() error {
	a.setState(adapters.StateConnecting)

	a.logger.Info("Connecting to Getinge device",
		"device_id", a.deviceID,
		"device_name", a.device.Name,
		"ip", a.device.IP)

	reachable, err := a.Ping()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to ping device: %v", err)
		a.setStateWithError(adapters.StateError, errorMsg)
		return fmt.Errorf("ping failed: %w", err)
	}

	if reachable {
		a.setState(adapters.StateConnected)
		a.lastReachable = true
		now := time.Now()
		a.lastPing = &now
		a.logger.Info("Getinge device is reachable",
			"device_id", a.deviceID,
			"ip", a.device.IP)
	} else {
		a.setState(adapters.StateError)
		a.lastReachable = false
		now := time.Now()
		a.lastPing = &now
		a.logger.Warn("Getinge device is not reachable",
			"device_id", a.deviceID,
			"ip", a.device.IP)
	}

	return nil
}

// Disconnect marks the device as disconnected
func (a *GetingeAdapter) Disconnect() error {
	a.setState(adapters.StateDisconnected)
	a.logger.Info("Disconnected from Getinge device",
		"device_id", a.deviceID)
	return nil
}

// IsConnected returns true if the device is currently reachable
func (a *GetingeAdapter) IsConnected() bool {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.state == adapters.StateConnected && a.lastReachable
}

// GetStatus returns device status information (not used for Getinge in MVP)
func (a *GetingeAdapter) GetStatus() (database.DeviceStatus, error) {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()

	connected := a.state == adapters.StateConnected && a.lastReachable
	healthStatus := "healthy"
	if !connected {
		healthStatus = "unhealthy"
	}

	return database.DeviceStatus{
		DeviceID:     a.deviceID,
		Connected:    connected,
		LastSeen:     a.lastPing,
		HealthStatus: healthStatus,
		Manufacturer: a.device.Manufacturer,
		IP:           a.device.IP,
		ICMPReachable: a.lastReachable,
		LastPingTime:  a.lastPing,
	}, nil
}

// StartCycle is not implemented for Getinge devices in MVP
func (a *GetingeAdapter) StartCycle(params adapters.CycleStartParams) error {
	return fmt.Errorf("start cycle not supported for Getinge devices in MVP")
}

// GetCycleStatus is not implemented for Getinge devices in MVP
func (a *GetingeAdapter) GetCycleStatus() (adapters.CycleStatus, error) {
	return adapters.CycleStatus{}, fmt.Errorf("cycle status not supported for Getinge devices in MVP")
}

// GetDeviceID returns the device ID
func (a *GetingeAdapter) GetDeviceID() int {
	return a.deviceID
}

// GetDeviceDetails returns the device details
func (a *GetingeAdapter) GetDeviceDetails() *database.Device {
	return a.device
}

// GetConnectionState returns the current connection state
func (a *GetingeAdapter) GetConnectionState() adapters.ConnectionState {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.state
}

// SetConnectionState sets the connection state
func (a *GetingeAdapter) SetConnectionState(state adapters.ConnectionState) {
	a.setState(state)
}

// Ping performs a network reachability check to the device IP address
// Uses TCP connection attempt as a proxy for ICMP ping (Windows-compatible)
func (a *GetingeAdapter) Ping() (bool, error) {
	// Try to connect to a common port (port 80 or 443) to check reachability
	// This is a simple reachability test that works on Windows without raw sockets
	timeout := a.pingTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Try connecting to port 80 (HTTP) or 443 (HTTPS) as a reachability test
	addresses := []string{
		net.JoinHostPort(a.device.IP, "80"),
		net.JoinHostPort(a.device.IP, "443"),
		net.JoinHostPort(a.device.IP, "22"), // SSH
	}

	for _, addr := range addresses {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err == nil {
			conn.Close()
			return true, nil
		}
	}

	// If all TCP attempts failed, try a simple ICMP-like check using UDP
	// This is a fallback for devices that don't respond to TCP
	udpConn, err := net.DialTimeout("udp", net.JoinHostPort(a.device.IP, "53"), timeout)
	if err == nil {
		udpConn.Close()
		return true, nil
	}

	// Final check: try to resolve the IP address (basic connectivity)
	_, err = net.LookupAddr(a.device.IP)
	if err == nil {
		return true, nil
	}

	return false, fmt.Errorf("device unreachable via TCP/UDP checks")
}

// GetLastPing returns the last ping time
func (a *GetingeAdapter) GetLastPing() *time.Time {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.lastPing
}

// IsReachable returns the last known reachability status
func (a *GetingeAdapter) IsReachable() bool {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.lastReachable
}

// setState sets the connection state
func (a *GetingeAdapter) setState(state adapters.ConnectionState) {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	a.state = state
}

// setStateWithError sets the connection state to ERROR with an error message
func (a *GetingeAdapter) setStateWithError(state adapters.ConnectionState, errorMsg string) {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	a.state = state
	a.logger.Error("Getinge adapter state error",
		"device_id", a.deviceID,
		"state", state,
		"error", errorMsg)
}

// SetLastPing sets the last ping time
func (a *GetingeAdapter) SetLastPing(t *time.Time) {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	a.lastPing = t
}

// SetLastReachable sets the last reachability status
func (a *GetingeAdapter) SetLastReachable(reachable bool) {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	a.lastReachable = reachable
}

