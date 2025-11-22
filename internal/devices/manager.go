package devices

import (
	"fmt"
	"sync"
	"time"

	"steri-connect-go/internal/adapters"
	"steri-connect-go/internal/adapters/getinge"
	"steri-connect-go/internal/adapters/melag"
	"steri-connect-go/internal/config"
	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
	"steri-connect-go/internal/api/websocket"
)

// Manager manages device connections
type Manager struct {
	adapters         map[int]adapters.DeviceAdapter
	adaptersMutex    sync.RWMutex
	logger           *logging.Logger
	retryInterval    time.Duration
	maxRetries       int
	pingMonitors     map[int]chan bool // Channel to stop ping monitoring for each device
	pingMonitorsMutex sync.RWMutex
}

var globalManager *Manager
var managerMutex sync.RWMutex

// GetManager returns the global device manager instance
func GetManager() *Manager {
	managerMutex.RLock()
	defer managerMutex.RUnlock()
	return globalManager
}

// SetManager sets the global device manager instance
func SetManager(manager *Manager) {
	managerMutex.Lock()
	defer managerMutex.Unlock()
	globalManager = manager
}

// NewManager creates a new device manager
func NewManager() *Manager {
	return &Manager{
		adapters:      make(map[int]adapters.DeviceAdapter),
		logger:        logging.Get(),
		retryInterval: 5 * time.Second,
		maxRetries:    3,
		pingMonitors:  make(map[int]chan bool),
	}
}

// LoadDevices loads all devices from database and initializes connections
func (m *Manager) LoadDevices() error {
	m.logger.Info("Loading devices from database")

	devices, err := database.GetAllDevices()
	if err != nil {
		return fmt.Errorf("failed to load devices: %w", err)
	}

	m.logger.Info("Found devices in database", "count", len(devices))

	// Initialize adapters for Melag and Getinge devices
	for _, device := range devices {
		if device.Manufacturer == "Melag" || device.Manufacturer == "Getinge" {
			if err := m.AddDevice(&device); err != nil {
				m.logger.Warn("Failed to initialize device adapter",
					"device_id", device.ID,
					"device_name", device.Name,
					"manufacturer", device.Manufacturer,
					"error", err)
				// Continue with other devices
				continue
			}
		}
	}

	return nil
}

// AddDevice adds a device and initializes its adapter
func (m *Manager) AddDevice(device *database.Device) error {
	if device == nil {
		return fmt.Errorf("device cannot be nil")
	}

	m.adaptersMutex.Lock()
	defer m.adaptersMutex.Unlock()

	// Check if adapter already exists
	if _, exists := m.adapters[device.ID]; exists {
		return fmt.Errorf("adapter for device %d already exists", device.ID)
	}

	// Create adapter based on manufacturer
	var adapter adapters.DeviceAdapter
	var err error

	switch device.Manufacturer {
	case "Melag":
		adapter, err = melag.NewMelagAdapter(device)
		if err != nil {
			return fmt.Errorf("failed to create Melag adapter: %w", err)
		}
	case "Getinge":
		cfg := config.Get()
		pingTimeout := time.Duration(cfg.Devices.Getinge.PingTimeout) * time.Second
		if pingTimeout == 0 {
			pingTimeout = 5 * time.Second // Default
		}
		adapter, err = getinge.NewGetingeAdapter(device, pingTimeout)
		if err != nil {
			return fmt.Errorf("failed to create Getinge adapter: %w", err)
		}
		// Start ping monitoring for Getinge devices
		m.startPingMonitoring(device.ID, adapter)
	default:
		return fmt.Errorf("unsupported manufacturer: %s", device.Manufacturer)
	}

	m.adapters[device.ID] = adapter

	m.logger.Info("Device adapter created",
		"device_id", device.ID,
		"device_name", device.Name,
		"manufacturer", device.Manufacturer)

	// Attempt to connect in background
	go m.connectDeviceWithRetry(adapter)

	return nil
}

// RemoveDevice removes a device and disconnects it
func (m *Manager) RemoveDevice(deviceID int) error {
	m.adaptersMutex.Lock()
	defer m.adaptersMutex.Unlock()

	// Stop ping monitoring if active
	m.stopPingMonitoring(deviceID)

	adapter, exists := m.adapters[deviceID]
	if !exists {
		return fmt.Errorf("adapter for device %d not found", deviceID)
	}

	// Disconnect before removing
	if err := adapter.Disconnect(); err != nil {
		m.logger.Warn("Error disconnecting device during removal",
			"device_id", deviceID,
			"error", err)
	}

	delete(m.adapters, deviceID)

	m.logger.Info("Device adapter removed",
		"device_id", deviceID)

	return nil
}

// ConnectDevice connects to a device
func (m *Manager) ConnectDevice(deviceID int) error {
	adapter := m.getAdapter(deviceID)
	if adapter == nil {
		return fmt.Errorf("device %d not found", deviceID)
	}

	if err := adapter.Connect(); err != nil {
		m.broadcastStatusChange(deviceID, adapter.GetConnectionState())
		return err
	}

	m.broadcastStatusChange(deviceID, adapter.GetConnectionState())
	return nil
}

// DisconnectDevice disconnects from a device
func (m *Manager) DisconnectDevice(deviceID int) error {
	adapter := m.getAdapter(deviceID)
	if adapter == nil {
		return fmt.Errorf("device %d not found", deviceID)
	}

	if err := adapter.Disconnect(); err != nil {
		return err
	}

	m.broadcastStatusChange(deviceID, adapter.GetConnectionState())
	return nil
}

// GetAdapter returns the adapter for a device
func (m *Manager) GetAdapter(deviceID int) adapters.DeviceAdapter {
	return m.getAdapter(deviceID)
}

// GetConnectionInfo returns connection information for a device
func (m *Manager) GetConnectionInfo(deviceID int) (*adapters.ConnectionInfo, error) {
	adapter := m.getAdapter(deviceID)
	if adapter == nil {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}

	state := adapter.GetConnectionState()
	connected := adapter.IsConnected()

	info := &adapters.ConnectionInfo{
		State:     state,
		Connected: connected,
	}

	if connected {
		now := time.Now()
		info.LastSeen = &now
	}

	return info, nil
}

// connectDeviceWithRetry attempts to connect with retry logic
func (m *Manager) connectDeviceWithRetry(adapter adapters.DeviceAdapter) {
	deviceID := adapter.GetDeviceID()
	retries := 0

	for retries < m.maxRetries {
		err := adapter.Connect()
		if err == nil {
			m.broadcastStatusChange(deviceID, adapter.GetConnectionState())
			return
		}

		retries++
		m.logger.Warn("Connection attempt failed, retrying",
			"device_id", deviceID,
			"attempt", retries,
			"max_retries", m.maxRetries,
			"error", err)

		if retries < m.maxRetries {
			time.Sleep(m.retryInterval)
		}
	}

	m.logger.Error("Failed to connect after all retries",
		"device_id", deviceID,
		"max_retries", m.maxRetries)

	m.broadcastStatusChange(deviceID, adapter.GetConnectionState())
}

// broadcastStatusChange broadcasts device status change via WebSocket
func (m *Manager) broadcastStatusChange(deviceID int, state adapters.ConnectionState) {
	event := websocket.Event{
		Event: "device_status_change",
		Data: map[string]interface{}{
			"device_id": deviceID,
			"state":     string(state),
			"connected": state == adapters.StateConnected,
		},
	}

	if err := websocket.BroadcastEvent(event); err != nil {
		m.logger.Warn("Failed to broadcast device status change",
			"device_id", deviceID,
			"error", err)
	}

	m.logger.Debug("Broadcasted device status change",
		"device_id", deviceID,
		"state", state)
}

// getAdapter returns the adapter for a device (thread-safe)
func (m *Manager) getAdapter(deviceID int) adapters.DeviceAdapter {
	m.adaptersMutex.RLock()
	defer m.adaptersMutex.RUnlock()
	return m.adapters[deviceID]
}

// Shutdown disconnects all devices
func (m *Manager) Shutdown() error {
	m.logger.Info("Shutting down device manager")

	m.adaptersMutex.Lock()
	defer m.adaptersMutex.Unlock()

	var errors []error
	for deviceID, adapter := range m.adapters {
		if err := adapter.Disconnect(); err != nil {
			m.logger.Warn("Error disconnecting device during shutdown",
				"device_id", deviceID,
				"error", err)
			errors = append(errors, fmt.Errorf("device %d: %w", deviceID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	m.logger.Info("Device manager shutdown complete")
	return nil
}

// ActiveCycle represents an active cycle being monitored
type ActiveCycle struct {
	CycleID     int
	DeviceID    int
	StopPolling chan bool
}

// StartCyclePolling starts polling for a cycle's status
func (m *Manager) StartCyclePolling(cycleID int, deviceID int) {
	m.logger.Info("Starting cycle status polling",
		"cycle_id", cycleID,
		"device_id", deviceID)

	stopChan := make(chan bool)

	// Start polling goroutine
	go m.pollCycleStatus(cycleID, deviceID, stopChan)
}

// StopCyclePolling stops polling for a cycle's status
func (m *Manager) StopCyclePolling(cycleID int) {
	m.logger.Info("Stopping cycle status polling",
		"cycle_id", cycleID)
	// Polling will stop automatically when cycle completes
}

// pollCycleStatus polls the device for cycle status updates
func (m *Manager) pollCycleStatus(cycleID int, deviceID int, stopChan chan bool) {
	ticker := time.NewTicker(2 * time.Second) // Poll every 2 seconds
	defer ticker.Stop()

	m.logger.Info("Cycle status polling started",
		"cycle_id", cycleID,
		"device_id", deviceID,
		"interval", "2s")

	for {
		select {
		case <-stopChan:
			m.logger.Info("Cycle status polling stopped",
				"cycle_id", cycleID)
			return

		case <-ticker.C:
			// Get device adapter
			adapter := m.GetAdapter(deviceID)
			if adapter == nil {
				m.logger.Warn("Adapter not found for cycle polling",
					"cycle_id", cycleID,
					"device_id", deviceID)
				continue
			}

			// Cast to MelagAdapter
			melagAdapter, ok := adapter.(*melag.MelagAdapter)
			if !ok {
				m.logger.Warn("Adapter is not a MelagAdapter",
					"cycle_id", cycleID,
					"device_id", deviceID)
				continue
			}

			// Get cycle status
			status, err := melagAdapter.GetCycleStatus()
			if err != nil {
				m.logger.Warn("Failed to get cycle status",
					"cycle_id", cycleID,
					"device_id", deviceID,
					"error", err)
				continue
			}

			// Check if cycle is still running
			if !status.IsRunning && status.Phase == "COMPLETED" {
				m.logger.Info("Cycle completed, stopping polling",
					"cycle_id", cycleID,
					"device_id", deviceID)

				// Update cycle result and end timestamp
				endTime := time.Now()
				err = database.UpdateCycleResult(cycleID, "OK", endTime, nil, nil)
				if err != nil {
					m.logger.Error("Failed to update cycle result",
						"cycle_id", cycleID,
						"error", err)
				}

				// Broadcast cycle_completed event
				event := websocket.Event{
					Event: "cycle_completed",
					Data: map[string]interface{}{
						"cycle_id":  cycleID,
						"device_id": deviceID,
						"result":    "OK",
						"end_ts":    endTime.Format(time.RFC3339),
					},
				}
				if err := websocket.BroadcastEvent(event); err != nil {
					m.logger.Warn("Failed to broadcast cycle_completed event",
						"cycle_id", cycleID,
						"error", err)
				}

				// Log audit entry
				details := map[string]interface{}{
					"cycle_id": cycleID,
					"device_id": deviceID,
					"result":   "OK",
				}
				if err := database.LogAudit(database.ActionCycleCompleted, "cycle", &cycleID, "", details); err != nil {
					m.logger.Warn("Failed to log cycle completion audit",
						"cycle_id", cycleID,
						"error", err)
				}

				return
			}

			if !status.IsRunning && status.Phase == "FAILED" {
				m.logger.Info("Cycle failed, stopping polling",
					"cycle_id", cycleID,
					"device_id", deviceID)

				// Update cycle result and end timestamp
				endTime := time.Now()
				errorDesc := "Cycle failed - see device logs"
				err = database.UpdateCycleResult(cycleID, "NOK", endTime, nil, &errorDesc)
				if err != nil {
					m.logger.Error("Failed to update cycle result",
						"cycle_id", cycleID,
						"error", err)
				}

				// Broadcast cycle_failed event
				event := websocket.Event{
					Event: "cycle_failed",
					Data: map[string]interface{}{
						"cycle_id":        cycleID,
						"device_id":       deviceID,
						"result":          "NOK",
						"error_description": errorDesc,
						"end_ts":          endTime.Format(time.RFC3339),
					},
				}
				if err := websocket.BroadcastEvent(event); err != nil {
					m.logger.Warn("Failed to broadcast cycle_failed event",
						"cycle_id", cycleID,
						"error", err)
				}

				// Log audit entry
				details := map[string]interface{}{
					"cycle_id": cycleID,
					"device_id": deviceID,
					"result":   "NOK",
					"error_description": errorDesc,
				}
				if err := database.LogAudit(database.ActionCycleFailed, "cycle", &cycleID, "", details); err != nil {
					m.logger.Warn("Failed to log cycle failure audit",
						"cycle_id", cycleID,
						"error", err)
				}

				return
			}

			// Update cycle status in database
			var progress *int
			if status.ProgressPercent >= 0 {
				progressVal := status.ProgressPercent
				progress = &progressVal
			}

			err = database.UpdateCycleStatus(
				cycleID,
				status.Phase,
				progress,
				status.Temperature,
				status.Pressure,
			)

			if err != nil {
				m.logger.Error("Failed to update cycle status in database",
					"cycle_id", cycleID,
					"error", err)
				continue
			}

			// Broadcast status update via WebSocket
			event := websocket.Event{
				Event: "cycle_status_update",
				Data: map[string]interface{}{
					"cycle_id":        cycleID,
					"device_id":       deviceID,
					"phase":           status.Phase,
					"progress_percent": status.ProgressPercent,
					"temperature":     status.Temperature,
					"pressure":        status.Pressure,
					"time_remaining":  status.TimeRemaining,
					"is_running":      status.IsRunning,
				},
			}

			if err := websocket.BroadcastEvent(event); err != nil {
				m.logger.Warn("Failed to broadcast cycle status update",
					"cycle_id", cycleID,
					"error", err)
				// Continue even if broadcast fails
			}

			m.logger.Debug("Cycle status updated",
				"cycle_id", cycleID,
				"phase", status.Phase,
				"progress", status.ProgressPercent)
		}
	}
}

