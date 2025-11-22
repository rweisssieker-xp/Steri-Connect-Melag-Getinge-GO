package devices

import (
	"time"

	"steri-connect-go/internal/adapters"
	"steri-connect-go/internal/adapters/getinge"
	"steri-connect-go/internal/config"
	"steri-connect-go/internal/database"
)

// startPingMonitoring starts ping monitoring for a Getinge device
func (m *Manager) startPingMonitoring(deviceID int, adapter adapters.DeviceAdapter) {
	cfg := config.Get()
	pingInterval := time.Duration(cfg.Devices.Getinge.PingInterval) * time.Second
	if pingInterval == 0 {
		pingInterval = 15 * time.Second // Default
	}

	m.pingMonitorsMutex.Lock()
	stopChan := make(chan bool)
	m.pingMonitors[deviceID] = stopChan
	m.pingMonitorsMutex.Unlock()

	m.logger.Info("Starting ping monitoring for Getinge device",
		"device_id", deviceID,
		"ping_interval", pingInterval)

	go m.pingDeviceLoop(deviceID, adapter, pingInterval, stopChan)
}

// stopPingMonitoring stops ping monitoring for a device
func (m *Manager) stopPingMonitoring(deviceID int) {
	m.pingMonitorsMutex.Lock()
	defer m.pingMonitorsMutex.Unlock()

	if stopChan, exists := m.pingMonitors[deviceID]; exists {
		close(stopChan)
		delete(m.pingMonitors, deviceID)
		m.logger.Info("Stopped ping monitoring for device",
			"device_id", deviceID)
	}
}

// pingDeviceLoop continuously pings a Getinge device
func (m *Manager) pingDeviceLoop(deviceID int, adapter adapters.DeviceAdapter, interval time.Duration, stopChan chan bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial ping
	m.performPing(deviceID, adapter)

	for {
		select {
		case <-stopChan:
			m.logger.Info("Ping monitoring stopped",
				"device_id", deviceID)
			return
		case <-ticker.C:
			m.performPing(deviceID, adapter)
		}
	}
}

// performPing performs a single ping and updates status
func (m *Manager) performPing(deviceID int, adapter adapters.DeviceAdapter) {
	getingeAdapter, ok := adapter.(*getinge.GetingeAdapter)
	if !ok {
		m.logger.Warn("Adapter is not a GetingeAdapter",
			"device_id", deviceID)
		return
	}

	reachable, err := getingeAdapter.Ping()
	if err != nil {
		m.logger.Warn("Ping failed",
			"device_id", deviceID,
			"error", err)
		reachable = false
	}

	// Update last ping time and reachability in adapter
	now := time.Now()
	getingeAdapter.SetLastPing(&now)
	getingeAdapter.SetLastReachable(reachable)

	// Update connection state
	if reachable {
		getingeAdapter.SetConnectionState(adapters.StateConnected)
	} else {
		getingeAdapter.SetConnectionState(adapters.StateError)
	}

	// Save RDG status to database
	_, err = database.CreateRDGStatus(deviceID, reachable)
	if err != nil {
		m.logger.Error("Failed to save RDG status",
			"device_id", deviceID,
			"error", err)
	}

	// Update device connection status in database
	err = database.UpdateDeviceConnectionStatus(deviceID, reachable)
	if err != nil {
		m.logger.Error("Failed to update device connection status",
			"device_id", deviceID,
			"error", err)
	}

	// Log audit entry
	details := map[string]interface{}{
		"device_id": deviceID,
		"reachable": reachable,
	}
	if err := database.LogAudit(database.ActionRDGStatusUpdate, "device", &deviceID, "system", details); err != nil {
		m.logger.Warn("Failed to log RDG status update audit",
			"device_id", deviceID,
			"error", err)
	}

	// Broadcast status change event
	m.broadcastStatusChange(deviceID, adapter.GetConnectionState())
}

