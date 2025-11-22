package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrCycleNotFound = errors.New("cycle not found")
)

// CreateCycle creates a new cycle in the database
func CreateCycle(cycle *Cycle) (*Cycle, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		INSERT INTO cycles (device_id, program, start_ts, phase, progress_percent)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		cycle.DeviceID,
		cycle.Program,
		cycle.StartTS,
		cycle.Phase,
		cycle.ProgressPercent,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create cycle: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle ID: %w", err)
	}

	// Retrieve the created cycle to get database defaults
	createdCycle, err := GetCycle(int(id))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created cycle: %w", err)
	}

	return createdCycle, nil
}

// GetCycle retrieves a cycle by ID
func GetCycle(id int) (*Cycle, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, device_id, program, start_ts, end_ts, result, error_code,
		       error_description, phase, temperature, pressure, progress_percent
		FROM cycles
		WHERE id = ?
	`

	cycle := &Cycle{}
	var endTS sql.NullTime
	var result sql.NullString
	var errorCode sql.NullString
	var errorDesc sql.NullString
	var phase sql.NullString
	var temp sql.NullFloat64
	var pressure sql.NullFloat64
	var progress sql.NullInt64

	err := db.QueryRow(query, id).Scan(
		&cycle.ID,
		&cycle.DeviceID,
		&cycle.Program,
		&cycle.StartTS,
		&endTS,
		&result,
		&errorCode,
		&errorDesc,
		&phase,
		&temp,
		&pressure,
		&progress,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCycleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle: %w", err)
	}

	// Handle nullable fields
	if endTS.Valid {
		cycle.EndTS = &endTS.Time
	}
	if result.Valid {
		cycle.Result = result.String
	}
	if errorCode.Valid {
		cycle.ErrorCode = errorCode.String
	}
	if errorDesc.Valid {
		cycle.ErrorDescription = errorDesc.String
	}
	if phase.Valid {
		cycle.Phase = phase.String
	}
	if temp.Valid {
		cycle.Temperature = &[]float64{temp.Float64}[0]
	}
	if pressure.Valid {
		cycle.Pressure = &[]float64{pressure.Float64}[0]
	}
	if progress.Valid {
		progressVal := int(progress.Int64)
		cycle.ProgressPercent = &progressVal
	}

	return cycle, nil
}

// UpdateCycleStatus updates cycle status and phase information
func UpdateCycleStatus(id int, phase string, progress *int, temperature *float64, pressure *float64) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	setParts := []string{}
	args := []interface{}{}

	if phase != "" {
		setParts = append(setParts, "phase = ?")
		args = append(args, phase)
	}

	if progress != nil {
		setParts = append(setParts, "progress_percent = ?")
		args = append(args, *progress)
	}

	if temperature != nil {
		setParts = append(setParts, "temperature = ?")
		args = append(args, *temperature)
	}

	if pressure != nil {
		setParts = append(setParts, "pressure = ?")
		args = append(args, *pressure)
	}

	if len(setParts) == 0 {
		return nil // Nothing to update
	}

	// Build query properly
	query := "UPDATE cycles SET "
	for i, part := range setParts {
		if i > 0 {
			query += ", "
		}
		query += part
	}
	query += " WHERE id = ?"

	// Add cycle ID to args for WHERE clause
	args = append(args, id)

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cycle: %w", err)
	}

	return nil
}

// UpdateCycleResult updates cycle final result
func UpdateCycleResult(id int, result string, endTS time.Time, errorCode *string, errorDesc *string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
		UPDATE cycles
		SET result = ?, end_ts = ?, error_code = ?, error_description = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, result, endTS, errorCode, errorDesc, id)
	if err != nil {
		return fmt.Errorf("failed to update cycle result: %w", err)
	}

	return nil
}

// GetDeviceCycles retrieves all cycles for a device
func GetDeviceCycles(deviceID int) ([]Cycle, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, device_id, program, start_ts, end_ts, result, error_code,
		       error_description, phase, temperature, pressure, progress_percent
		FROM cycles
		WHERE device_id = ?
		ORDER BY start_ts DESC
	`

	rows, err := db.Query(query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cycles: %w", err)
	}
	defer rows.Close()

	var cycles []Cycle
	for rows.Next() {
		var cycle Cycle
		var endTS sql.NullTime
		var result sql.NullString
		var errorCode sql.NullString
		var errorDesc sql.NullString
		var phase sql.NullString
		var temp sql.NullFloat64
		var pressure sql.NullFloat64
		var progress sql.NullInt64

		err := rows.Scan(
			&cycle.ID,
			&cycle.DeviceID,
			&cycle.Program,
			&cycle.StartTS,
			&endTS,
			&result,
			&errorCode,
			&errorDesc,
			&phase,
			&temp,
			&pressure,
			&progress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cycle: %w", err)
		}

		// Handle nullable fields
		if endTS.Valid {
			cycle.EndTS = &endTS.Time
		}
		if result.Valid {
			cycle.Result = result.String
		}
		if errorCode.Valid {
			cycle.ErrorCode = errorCode.String
		}
		if errorDesc.Valid {
			cycle.ErrorDescription = errorDesc.String
		}
		if phase.Valid {
			cycle.Phase = phase.String
		}
		if temp.Valid {
			tempVal := temp.Float64
			cycle.Temperature = &tempVal
		}
		if pressure.Valid {
			pressureVal := pressure.Float64
			cycle.Pressure = &pressureVal
		}
		if progress.Valid {
			progressVal := int(progress.Int64)
			cycle.ProgressPercent = &progressVal
		}

		cycles = append(cycles, cycle)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cycles: %w", err)
	}

	return cycles, nil
}

// CycleListOptions holds options for listing cycles
type CycleListOptions struct {
	Limit     int       // Number of cycles to return (0 = no limit)
	Offset    int       // Number of cycles to skip
	SortBy    string    // Sort field: "start_ts", "device_id", "result" (default: "start_ts")
	SortOrder string    // Sort order: "ASC" or "DESC" (default: "DESC")
	DeviceID  *int      // Filter by device ID (nil = all devices)
	StartDate *time.Time // Filter by start date (from)
	EndDate   *time.Time // Filter by end date (to)
	Result    *string   // Filter by result: "OK" or "NOK" (nil = all)
}

// CycleWithDevice represents a cycle with device information
type CycleWithDevice struct {
	Cycle
	DeviceName     string `json:"device_name"`
	DeviceIP       string `json:"device_ip"`
	Manufacturer   string `json:"manufacturer"`
}

// GetAllCycles retrieves all cycles with pagination, sorting, and filtering
func GetAllCycles(options CycleListOptions) ([]CycleWithDevice, int, error) {
	if db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	// Build WHERE clause
	whereParts := []string{}
	args := []interface{}{}

	if options.DeviceID != nil {
		whereParts = append(whereParts, "c.device_id = ?")
		args = append(args, *options.DeviceID)
	}

	if options.StartDate != nil {
		whereParts = append(whereParts, "c.start_ts >= ?")
		args = append(args, *options.StartDate)
	}

	if options.EndDate != nil {
		whereParts = append(whereParts, "c.start_ts <= ?")
		args = append(args, *options.EndDate)
	}

	if options.Result != nil {
		whereParts = append(whereParts, "c.result = ?")
		args = append(args, *options.Result)
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// Build ORDER BY clause
	sortBy := options.SortBy
	if sortBy == "" {
		sortBy = "start_ts"
	}
	// Validate sort field
	validSortFields := map[string]bool{
		"start_ts":  true,
		"device_id": true,
		"result":    true,
	}
	if !validSortFields[sortBy] {
		sortBy = "start_ts"
	}

	sortOrder := options.SortOrder
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM cycles c
		%s
	`, whereClause)

	var totalCount int
	err := db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cycle count: %w", err)
	}

	// Build main query with JOIN to devices table
	query := fmt.Sprintf(`
		SELECT c.id, c.device_id, c.program, c.start_ts, c.end_ts, c.result, c.error_code,
		       c.error_description, c.phase, c.temperature, c.pressure, c.progress_percent,
		       d.name as device_name, d.ip as device_ip, d.manufacturer
		FROM cycles c
		LEFT JOIN devices d ON c.device_id = d.id
		%s
		ORDER BY c.%s %s
	`, whereClause, sortBy, sortOrder)

	// Add LIMIT and OFFSET
	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", options.Limit)
		if options.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", options.Offset)
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query cycles: %w", err)
	}
	defer rows.Close()

	var cycles []CycleWithDevice
	for rows.Next() {
		var cycle CycleWithDevice
		var endTS sql.NullTime
		var result sql.NullString
		var errorCode sql.NullString
		var errorDesc sql.NullString
		var phase sql.NullString
		var temp sql.NullFloat64
		var pressure sql.NullFloat64
		var progress sql.NullInt64

		err := rows.Scan(
			&cycle.ID,
			&cycle.DeviceID,
			&cycle.Program,
			&cycle.StartTS,
			&endTS,
			&result,
			&errorCode,
			&errorDesc,
			&phase,
			&temp,
			&pressure,
			&progress,
			&cycle.DeviceName,
			&cycle.DeviceIP,
			&cycle.Manufacturer,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan cycle: %w", err)
		}

		// Handle nullable fields
		if endTS.Valid {
			cycle.EndTS = &endTS.Time
		}
		if result.Valid {
			cycle.Result = result.String
		}
		if errorCode.Valid {
			cycle.ErrorCode = errorCode.String
		}
		if errorDesc.Valid {
			cycle.ErrorDescription = errorDesc.String
		}
		if phase.Valid {
			cycle.Phase = phase.String
		}
		if temp.Valid {
			tempVal := temp.Float64
			cycle.Temperature = &tempVal
		}
		if pressure.Valid {
			pressureVal := pressure.Float64
			cycle.Pressure = &pressureVal
		}
		if progress.Valid {
			progressVal := int(progress.Int64)
			cycle.ProgressPercent = &progressVal
		}

		cycles = append(cycles, cycle)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating cycles: %w", err)
	}

	return cycles, totalCount, nil
}

