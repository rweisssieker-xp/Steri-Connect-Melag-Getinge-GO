package melag

import (
	"fmt"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"

	"steri-connect-go/internal/adapters"
	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
)

// MelagAdapter implements the DeviceAdapter interface for Melag devices
type MelagAdapter struct {
	deviceID     int
	device       *database.Device
	ftpClient    *ftp.ServerConn
	state        adapters.ConnectionState
	stateMutex   sync.RWMutex
	ftpHost      string
	ftpUsername  string
	ftpPassword  string
	ftpTimeout   time.Duration
	logger       *logging.Logger
}

// NewMelagAdapter creates a new Melag adapter instance
func NewMelagAdapter(device *database.Device) (*MelagAdapter, error) {
	if device == nil {
		return nil, fmt.Errorf("device cannot be nil")
	}
	if device.Manufacturer != "Melag" {
		return nil, fmt.Errorf("device manufacturer must be Melag, got %s", device.Manufacturer)
	}

	logger := logging.Get()

	// Default FTP credentials (can be configured later)
	ftpUsername := "melanet"
	ftpPassword := "melanet"

	adapter := &MelagAdapter{
		deviceID:    device.ID,
		device:      device,
		state:       adapters.StateDisconnected,
		ftpHost:     device.IP + ":21", // Default FTP port
		ftpUsername: ftpUsername,
		ftpPassword: ftpPassword,
		ftpTimeout:  10 * time.Second,
		logger:      logger,
	}

	return adapter, nil
}

// Connect establishes FTP connection to MELAnet Box
func (a *MelagAdapter) Connect() error {
	a.setState(adapters.StateConnecting)

	a.logger.Info("Connecting to Melag device",
		"device_id", a.deviceID,
		"device_name", a.device.Name,
		"ip", a.device.IP,
		"host", a.ftpHost)

	// Create FTP connection
	conn, err := ftp.Dial(a.ftpHost, ftp.DialWithTimeout(a.ftpTimeout))
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to connect to FTP server: %v", err)
		a.setStateWithError(adapters.StateError, errorMsg)
		a.logger.Error("FTP connection failed",
			"device_id", a.deviceID,
			"ip", a.device.IP,
			"error", err)
		return fmt.Errorf("ftp connection failed: %w", err)
	}

	// Authenticate
	if err := conn.Login(a.ftpUsername, a.ftpPassword); err != nil {
		conn.Quit()
		errorMsg := fmt.Sprintf("FTP authentication failed: %v", err)
		a.setStateWithError(adapters.StateError, errorMsg)
		a.logger.Error("FTP authentication failed",
			"device_id", a.deviceID,
			"ip", a.device.IP,
			"error", err)
		return fmt.Errorf("ftp authentication failed: %w", err)
	}

	// Store connection
	a.stateMutex.Lock()
	a.ftpClient = conn
	a.stateMutex.Unlock()

	a.setState(adapters.StateConnected)

	a.logger.Info("Successfully connected to Melag device",
		"device_id", a.deviceID,
		"device_name", a.device.Name,
		"ip", a.device.IP)

	return nil
}

// Disconnect closes the FTP connection
func (a *MelagAdapter) Disconnect() error {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	if a.ftpClient == nil {
		return nil // Already disconnected
	}

	a.logger.Info("Disconnecting from Melag device",
		"device_id", a.deviceID,
		"device_name", a.device.Name)

	if err := a.ftpClient.Quit(); err != nil {
		a.logger.Warn("Error during FTP disconnect",
			"device_id", a.deviceID,
			"error", err)
		// Continue anyway
	}

	a.ftpClient = nil
	a.setStateUnsafe(adapters.StateDisconnected)

	a.logger.Info("Disconnected from Melag device",
		"device_id", a.deviceID,
		"device_name", a.device.Name)

	return nil
}

// IsConnected returns true if device is currently connected
func (a *MelagAdapter) IsConnected() bool {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.state == adapters.StateConnected && a.ftpClient != nil
}

// GetDeviceID returns the device ID
func (a *MelagAdapter) GetDeviceID() int {
	return a.deviceID
}

// GetConnectionState returns current connection state
func (a *MelagAdapter) GetConnectionState() adapters.ConnectionState {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.state
}

// GetFTPClient returns the FTP client (for internal use)
func (a *MelagAdapter) GetFTPClient() *ftp.ServerConn {
	a.stateMutex.RLock()
	defer a.stateMutex.RUnlock()
	return a.ftpClient
}

// setState updates the connection state (thread-safe)
func (a *MelagAdapter) setState(state adapters.ConnectionState) {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	a.state = state
}

// setStateUnsafe updates the connection state without lock (use only when lock is already held)
func (a *MelagAdapter) setStateUnsafe(state adapters.ConnectionState) {
	a.state = state
}

// setStateWithError updates the connection state with error message
func (a *MelagAdapter) setStateWithError(state adapters.ConnectionState, errorMsg string) {
	a.setState(state)
	// Error message can be logged but not stored in state
	// In future, we might want to store last error
	a.logger.Error("Connection state changed to error",
		"device_id", a.deviceID,
		"state", state,
		"error", errorMsg)
}

// StartCycle starts a sterilization cycle on the Melag device
// Note: This is an MVP implementation with placeholder for actual FTP protocol command
// The actual protocol format will need to be determined from MELAnet Box documentation
func (a *MelagAdapter) StartCycle(params adapters.CycleStartParams) error {
	a.stateMutex.RLock()
	if a.ftpClient == nil || a.state != adapters.StateConnected {
		a.stateMutex.RUnlock()
		return fmt.Errorf("device is not connected")
	}
	ftpClient := a.ftpClient
	a.stateMutex.RUnlock()

	a.logger.Info("Starting cycle on Melag device",
		"device_id", a.deviceID,
		"device_name", a.device.Name,
		"program", params.Program)

	// MVP: Placeholder for actual FTP command
	// TODO: Implement actual MELAnet Box FTP protocol for cycle start
	// This will require:
	// 1. MELAnet Box documentation for cycle start command format
	// 2. FTP file upload or command execution protocol
	// 3. Protocol file format understanding

	// For MVP, we'll simulate a successful cycle start
	// In production, this would be something like:
	// - Upload a cycle start protocol file via FTP
	// - Send a command via FTP protocol
	// - Monitor FTP directory for cycle status files

	// Example placeholder implementation:
	// This is a placeholder - actual implementation requires MELAnet Box protocol docs
	_ = ftpClient // Use ftpClient to avoid unused variable error

	// Simulate successful cycle start for MVP
	// In production, replace this with actual FTP command:
	// err := ftpClient.UploadProtocolFile(params)
	// if err != nil {
	//     return fmt.Errorf("failed to start cycle: %w", err)
	// }

	a.logger.Info("Cycle start command sent successfully",
		"device_id", a.deviceID,
		"program", params.Program,
		"note", "MVP placeholder - actual protocol implementation required")

	return nil
}

// GetCycleStatus retrieves the current status of a running cycle
// Note: This is an MVP implementation with placeholder for actual FTP protocol status retrieval
// The actual protocol format will need to be determined from MELAnet Box documentation
func (a *MelagAdapter) GetCycleStatus() (adapters.CycleStatus, error) {
	a.stateMutex.RLock()
	if a.ftpClient == nil || a.state != adapters.StateConnected {
		a.stateMutex.RUnlock()
		return adapters.CycleStatus{}, fmt.Errorf("device is not connected")
	}
	ftpClient := a.ftpClient
	a.stateMutex.RUnlock()

	// MVP: Placeholder for actual FTP status retrieval
	// TODO: Implement actual MELAnet Box FTP protocol for cycle status
	// This will require:
	// 1. MELAnet Box documentation for status file format
	// 2. FTP file download or status command protocol
	// 3. Protocol file parsing (likely XML or similar format)

	// For MVP, we'll return a placeholder status
	// In production, this would be something like:
	// - Download status file via FTP
	// - Parse status file format
	// - Extract phase, temperature, pressure, progress

	// Example placeholder implementation:
	// This is a placeholder - actual implementation requires MELAnet Box protocol docs
	_ = ftpClient // Use ftpClient to avoid unused variable error

	// Simulate cycle status for MVP
	// In production, replace this with actual FTP status retrieval:
	// statusFile, err := ftpClient.Retr("status.xml")
	// if err != nil {
	//     return nil, fmt.Errorf("failed to retrieve status: %w", err)
	// }
	// defer statusFile.Close()
	// status, err := parseStatusFile(statusFile)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to parse status: %w", err)
	// }

	// Placeholder status - will be replaced with real implementation
	status := adapters.CycleStatus{
		Phase:           "RUNNING",
		ProgressPercent: 50, // Placeholder
		IsRunning:       true,
	}

	a.logger.Debug("Retrieved cycle status (MVP placeholder)",
		"device_id", a.deviceID,
		"phase", status.Phase,
		"progress", status.ProgressPercent,
		"note", "MVP placeholder - actual protocol implementation required")

	return status, nil
}

