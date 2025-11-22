package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
)

// TableListResponse represents the list of available tables
type TableListResponse struct {
	Tables []string `json:"tables"`
}

// TableDataResponse represents table data with pagination
type TableDataResponse struct {
	Table      string        `json:"table"`
	Columns    []string      `json:"columns"`
	Rows       []interface{} `json:"rows"`
	TotalCount int           `json:"total_count"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// TableSchemaResponse represents table schema information
type TableSchemaResponse struct {
	Table   string            `json:"table"`
	Columns []ColumnInfo      `json:"columns"`
}

// ColumnInfo represents column schema information
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
}

// ListTablesHandler handles GET /api/test-ui/db/tables
func ListTablesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	tables := []string{"devices", "cycles", "rdg_status", "audit_log"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TableListResponse{Tables: tables})
}

// GetTableDataHandler handles GET /api/test-ui/db/tables/{table}
func GetTableDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	logger := logging.Get()

	// Extract table name from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[0] != "api" || pathParts[1] != "test-ui" || pathParts[2] != "db" || pathParts[3] != "tables" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_path",
			Message: "Invalid path format",
		})
		return
	}

	tableName := ""
	if len(pathParts) > 4 {
		tableName = pathParts[4]
	}

	if tableName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "missing_table",
			Message: "Table name is required",
		})
		return
	}

	// Validate table name (prevent SQL injection)
	validTables := map[string]bool{
		"devices":    true,
		"cycles":     true,
		"rdg_status": true,
		"audit_log":  true,
	}

	if !validTables[tableName] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_table",
			Message: "Invalid table name",
		})
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	deviceID := ""
	startDate := ""
	endDate := ""

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	deviceID = r.URL.Query().Get("device_id")
	startDate = r.URL.Query().Get("start_date")
	endDate = r.URL.Query().Get("end_date")

	// Get table data
	data, columns, totalCount, err := getTableData(tableName, limit, offset, deviceID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get table data", "error", err, "table", tableName)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve table data",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TableDataResponse{
		Table:      tableName,
		Columns:    columns,
		Rows:       data,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	})
}

// GetTableSchemaHandler handles GET /api/test-ui/db/tables/{table}/schema
func GetTableSchemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Extract table name from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 6 || pathParts[5] != "schema" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_path",
			Message: "Invalid path format",
		})
		return
	}

	tableName := pathParts[4]

	// Validate table name
	validTables := map[string]bool{
		"devices":    true,
		"cycles":     true,
		"rdg_status": true,
		"audit_log":  true,
	}

	if !validTables[tableName] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_table",
			Message: "Invalid table name",
		})
		return
	}

	// Get table schema
	schema := getTableSchema(tableName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schema)
}

// ExportTableDataHandler handles GET /api/test-ui/db/tables/{table}/export
func ExportTableDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract table name and format from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 6 || pathParts[5] != "export" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tableName := pathParts[4]
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	// Validate table name
	validTables := map[string]bool{
		"devices":    true,
		"cycles":     true,
		"rdg_status": true,
		"audit_log":  true,
	}

	if !validTables[tableName] {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get all table data (no limit for export)
	data, columns, _, err := getTableData(tableName, 10000, 0, "", "", "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", tableName))
		
		writer := csv.NewWriter(w)
		defer writer.Flush()

		// Write header
		writer.Write(columns)

		// Write data rows
		for _, row := range data {
			rowMap := row.(map[string]interface{})
			record := make([]string, len(columns))
			for i, col := range columns {
				val := rowMap[col]
				if val == nil {
					record[i] = ""
				} else {
					record[i] = fmt.Sprintf("%v", val)
				}
			}
			writer.Write(record)
		}
	} else {
		// JSON export
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", tableName))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"table":   tableName,
			"columns": columns,
			"rows":    data,
		})
	}
}

// getTableData retrieves data from a table with optional filters
func getTableData(tableName string, limit, offset int, deviceID, startDate, endDate string) ([]interface{}, []string, int, error) {
	db := database.DB()
	if db == nil {
		return nil, nil, 0, fmt.Errorf("database not initialized")
	}

	var query string
	var args []interface{}
	var columns []string

	switch tableName {
	case "devices":
		columns = []string{"id", "name", "model", "manufacturer", "ip", "serial", "type", "location", "created", "updated"}
		query = "SELECT id, name, model, manufacturer, ip, serial, type, location, created, updated FROM devices WHERE 1=1"
		if deviceID != "" {
			query += " AND id = ?"
			args = append(args, deviceID)
		}
		query += " ORDER BY id DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

	case "cycles":
		columns = []string{"id", "device_id", "program", "start_ts", "end_ts", "phase", "temperature", "pressure", "progress_percent", "result", "error_code", "error_description", "created", "updated"}
		query = "SELECT id, device_id, program, start_ts, end_ts, phase, temperature, pressure, progress_percent, result, error_code, error_description, created, updated FROM cycles WHERE 1=1"
		if deviceID != "" {
			query += " AND device_id = ?"
			args = append(args, deviceID)
		}
		if startDate != "" {
			query += " AND start_ts >= ?"
			args = append(args, startDate)
		}
		if endDate != "" {
			query += " AND start_ts <= ?"
			args = append(args, endDate)
		}
		query += " ORDER BY id DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

	case "rdg_status":
		columns = []string{"id", "device_id", "timestamp", "reachable"}
		query = "SELECT id, device_id, timestamp, reachable FROM rdg_status WHERE 1=1"
		if deviceID != "" {
			query += " AND device_id = ?"
			args = append(args, deviceID)
		}
		if startDate != "" {
			query += " AND timestamp >= ?"
			args = append(args, startDate)
		}
		if endDate != "" {
			query += " AND timestamp <= ?"
			args = append(args, endDate)
		}
		query += " ORDER BY id DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

	case "audit_log":
		columns = []string{"id", "timestamp", "action", "entity_type", "entity_id", "user", "details", "hash"}
		query = "SELECT id, timestamp, action, entity_type, entity_id, user, details, hash FROM audit_log WHERE 1=1"
		if deviceID != "" {
			query += " AND entity_id = ? AND entity_type = 'device'"
			args = append(args, deviceID)
		}
		if startDate != "" {
			query += " AND timestamp >= ?"
			args = append(args, startDate)
		}
		if endDate != "" {
			query += " AND timestamp <= ?"
			args = append(args, endDate)
		}
		query += " ORDER BY id DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

	default:
		return nil, nil, 0, fmt.Errorf("unknown table: %s", tableName)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var data []interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		data = append(data, row)
	}

	// Get total count
	countQuery := strings.Replace(query, "SELECT "+strings.Join(columns, ", "), "SELECT COUNT(*)", 1)
	countQuery = strings.Split(countQuery, "LIMIT")[0]
	countArgs := args[:len(args)-2] // Remove limit and offset

	var totalCount int
	err = db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		totalCount = len(data)
	}

	return data, columns, totalCount, nil
}

// getTableSchema returns schema information for a table
func getTableSchema(tableName string) TableSchemaResponse {
	schemas := map[string]TableSchemaResponse{
		"devices": {
			Table: "devices",
			Columns: []ColumnInfo{
				{Name: "id", Type: "INTEGER", Nullable: false},
				{Name: "name", Type: "TEXT", Nullable: false},
				{Name: "model", Type: "TEXT", Nullable: true},
				{Name: "manufacturer", Type: "TEXT", Nullable: false},
				{Name: "ip", Type: "TEXT", Nullable: false},
				{Name: "serial", Type: "TEXT", Nullable: true},
				{Name: "type", Type: "TEXT", Nullable: false},
				{Name: "location", Type: "TEXT", Nullable: true},
				{Name: "created", Type: "DATETIME", Nullable: false},
				{Name: "updated", Type: "DATETIME", Nullable: false},
			},
		},
		"cycles": {
			Table: "cycles",
			Columns: []ColumnInfo{
				{Name: "id", Type: "INTEGER", Nullable: false},
				{Name: "device_id", Type: "INTEGER", Nullable: false},
				{Name: "program", Type: "TEXT", Nullable: true},
				{Name: "start_ts", Type: "DATETIME", Nullable: false},
				{Name: "end_ts", Type: "DATETIME", Nullable: true},
				{Name: "phase", Type: "TEXT", Nullable: true},
				{Name: "temperature", Type: "REAL", Nullable: true},
				{Name: "pressure", Type: "REAL", Nullable: true},
				{Name: "progress_percent", Type: "INTEGER", Nullable: true},
				{Name: "result", Type: "TEXT", Nullable: true},
				{Name: "error_code", Type: "TEXT", Nullable: true},
				{Name: "error_description", Type: "TEXT", Nullable: true},
				{Name: "created", Type: "DATETIME", Nullable: false},
				{Name: "updated", Type: "DATETIME", Nullable: false},
			},
		},
		"rdg_status": {
			Table: "rdg_status",
			Columns: []ColumnInfo{
				{Name: "id", Type: "INTEGER", Nullable: false},
				{Name: "device_id", Type: "INTEGER", Nullable: false},
				{Name: "timestamp", Type: "DATETIME", Nullable: false},
				{Name: "reachable", Type: "INTEGER", Nullable: false},
			},
		},
		"audit_log": {
			Table: "audit_log",
			Columns: []ColumnInfo{
				{Name: "id", Type: "INTEGER", Nullable: false},
				{Name: "timestamp", Type: "DATETIME", Nullable: false},
				{Name: "action", Type: "TEXT", Nullable: false},
				{Name: "entity_type", Type: "TEXT", Nullable: false},
				{Name: "entity_id", Type: "INTEGER", Nullable: true},
				{Name: "user", Type: "TEXT", Nullable: false},
				{Name: "details", Type: "TEXT", Nullable: true},
				{Name: "hash", Type: "TEXT", Nullable: false},
			},
		},
	}

	if schema, ok := schemas[tableName]; ok {
		return schema
	}

	return TableSchemaResponse{Table: tableName, Columns: []ColumnInfo{}}
}

