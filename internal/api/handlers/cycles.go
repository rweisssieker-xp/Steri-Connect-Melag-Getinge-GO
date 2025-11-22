package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
	"steri-connect-go/internal/pdf"
	"steri-connect-go/internal/csv"
)

// ListCyclesResponse represents the response for listing cycles
type ListCyclesResponse struct {
	Cycles     []database.CycleWithDevice `json:"cycles"`
	TotalCount int                        `json:"total_count"`
	Limit      int                        `json:"limit,omitempty"`
	Offset     int                        `json:"offset,omitempty"`
}

// ListCyclesHandler handles GET /api/cycles requests
func ListCyclesHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Parse query parameters
	options := database.CycleListOptions{}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_limit",
				Message: "Limit must be a positive integer",
			})
			return
		}
		options.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_offset",
				Message: "Offset must be a non-negative integer",
			})
			return
		}
		options.Offset = offset
	}

	// Sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		options.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		options.SortOrder = sortOrder
	}

	// Filtering by device ID
	if deviceIDStr := r.URL.Query().Get("device_id"); deviceIDStr != "" {
		deviceID, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_device_id",
				Message: "Device ID must be an integer",
			})
			return
		}
		options.DeviceID = &deviceID
	}

	// Filtering by date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			// Try alternative format
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_start_date",
					Message: "Start date must be in RFC3339 or YYYY-MM-DD format",
				})
				return
			}
		}
		options.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			// Try alternative format
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_end_date",
					Message: "End date must be in RFC3339 or YYYY-MM-DD format",
				})
				return
			}
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		options.EndDate = &endDate
	}

	// Filtering by result
	if result := r.URL.Query().Get("result"); result != "" {
		if result != "OK" && result != "NOK" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_result",
				Message: "Result must be 'OK' or 'NOK'",
			})
			return
		}
		options.Result = &result
	}

	// Retrieve cycles from database
	cycles, totalCount, err := database.GetAllCycles(options)
	if err != nil {
		logger.Error("Failed to retrieve cycles", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycles",
		})
		return
	}

	// Build response
	response := ListCyclesResponse{
		Cycles:     cycles,
		TotalCount: totalCount,
	}

	if options.Limit > 0 {
		response.Limit = options.Limit
		response.Offset = options.Offset
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetCycleHandler handles GET /api/cycles/{id} requests
func GetCycleHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Extract cycle ID from URL path
	cycleID, err := extractCycleID(r.URL.Path)
	if err != nil {
		logger.Warn("Failed to extract cycle ID from path", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_cycle_id",
			Message: "Invalid cycle ID in URL path",
		})
		return
	}

	// Get cycle from database with device information
	cycle, err := database.GetCycleWithDevice(cycleID)
	if err != nil {
		if err == database.ErrCycleNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "cycle_not_found",
				Message: fmt.Sprintf("Cycle with ID %d not found", cycleID),
			})
			return
		}

		logger.Error("Failed to get cycle", "error", err, "cycle_id", cycleID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycle",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cycle)
}

// extractCycleID extracts cycle ID from URL path like "/cycles/5"
func extractCycleID(path string) (int, error) {
	// Path will be "/cycles/5" after StripPrefix
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] != "cycles" {
		return 0, fmt.Errorf("invalid path format: expected /cycles/{id}")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid cycle ID: %w", err)
	}

	return id, nil
}

// ExportCyclePDFHandler handles GET /api/cycles/{id}/export/pdf requests
func ExportCyclePDFHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Extract cycle ID from URL path
	// Path format: /cycles/{id}/export/pdf
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[0] != "cycles" || pathParts[2] != "export" || (len(pathParts) > 3 && pathParts[3] != "pdf") {
		logger.Warn("Invalid path format for PDF export", "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_path",
			Message: "Invalid path format: expected /cycles/{id}/export/pdf",
		})
		return
	}

	cycleID, err := strconv.Atoi(pathParts[1])
	if err != nil {
		logger.Warn("Failed to extract cycle ID from path", "error", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_cycle_id",
			Message: "Invalid cycle ID in URL path",
		})
		return
	}

	// Get cycle from database with device information
	cycle, err := database.GetCycleWithDevice(cycleID)
	if err != nil {
		if err == database.ErrCycleNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "cycle_not_found",
				Message: fmt.Sprintf("Cycle with ID %d not found", cycleID),
			})
			return
		}

		logger.Error("Failed to get cycle", "error", err, "cycle_id", cycleID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycle",
		})
		return
	}

	// Generate PDF
	pdfBytes, err := pdf.GenerateCyclePDF(cycle)
	if err != nil {
		logger.Error("Failed to generate PDF", "error", err, "cycle_id", cycleID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate PDF",
		})
		return
	}

	// Set response headers for PDF download
	filename := fmt.Sprintf("cycle-%d-protocol.pdf", cycleID)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Write PDF to response
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// ExportCyclesCSVHandler handles GET /api/cycles/export/csv requests
func ExportCyclesCSVHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Parse query parameters (same as ListCyclesHandler)
	options := database.CycleListOptions{}

	// Pagination (optional for CSV export)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_limit",
				Message: "Limit must be a positive integer",
			})
			return
		}
		options.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_offset",
				Message: "Offset must be a non-negative integer",
			})
			return
		}
		options.Offset = offset
	}

	// Sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		options.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		options.SortOrder = sortOrder
	}

	// Filtering by device ID
	if deviceIDStr := r.URL.Query().Get("device_id"); deviceIDStr != "" {
		deviceID, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_device_id",
				Message: "Device ID must be an integer",
			})
			return
		}
		options.DeviceID = &deviceID
	}

	// Filtering by date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			// Try alternative format
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_start_date",
					Message: "Start date must be in RFC3339 or YYYY-MM-DD format",
				})
				return
			}
		}
		options.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			// Try alternative format
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_end_date",
					Message: "End date must be in RFC3339 or YYYY-MM-DD format",
				})
				return
			}
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		options.EndDate = &endDate
	}

	// Filtering by result
	if result := r.URL.Query().Get("result"); result != "" {
		if result != "OK" && result != "NOK" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_result",
				Message: "Result must be 'OK' or 'NOK'",
			})
			return
		}
		options.Result = &result
	}

	// Retrieve cycles from database
	cycles, _, err := database.GetAllCycles(options)
	if err != nil {
		logger.Error("Failed to retrieve cycles", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycles",
		})
		return
	}

	// Generate CSV
	csvBytes, err := csv.GenerateCyclesCSV(cycles)
	if err != nil {
		logger.Error("Failed to generate CSV", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate CSV",
		})
		return
	}

	// Set response headers for CSV download
	filename := "cycles-export.csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvBytes)))

	// Write CSV to response
	w.WriteHeader(http.StatusOK)
	w.Write(csvBytes)
}

// ExportCyclesJSONHandler handles GET /api/cycles/export/json requests
func ExportCyclesJSONHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Parse query parameters (same as ListCyclesHandler)
	options := database.CycleListOptions{}

	// Pagination (optional for JSON export)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_limit",
				Message: "Limit must be a positive integer",
			})
			return
		}
		options.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_offset",
				Message: "Offset must be a non-negative integer",
			})
			return
		}
		options.Offset = offset
	}

	// Sorting
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		options.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		options.SortOrder = sortOrder
	}

	// Filtering by device ID
	if deviceIDStr := r.URL.Query().Get("device_id"); deviceIDStr != "" {
		deviceID, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_device_id",
				Message: "Device ID must be an integer",
			})
			return
		}
		options.DeviceID = &deviceID
	}

	// Filtering by date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			// Try alternative format
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_start_date",
					Message: "Start date must be in RFC3339 or YYYY-MM-DD format",
				})
				return
			}
		}
		options.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			// Try alternative format
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_end_date",
					Message: "End date must be in RFC3339 or YYYY-MM-DD format",
				})
				return
			}
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		options.EndDate = &endDate
	}

	// Filtering by result
	if result := r.URL.Query().Get("result"); result != "" {
		if result != "OK" && result != "NOK" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_result",
				Message: "Result must be 'OK' or 'NOK'",
			})
			return
		}
		options.Result = &result
	}

	// Retrieve cycles from database
	cycles, totalCount, err := database.GetAllCycles(options)
	if err != nil {
		logger.Error("Failed to retrieve cycles", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve cycles",
		})
		return
	}

	// Check if file download is requested
	download := r.URL.Query().Get("download") == "true"

	// Build response
	response := map[string]interface{}{
		"cycles":      cycles,
		"total_count": totalCount,
	}

	if options.Limit > 0 {
		response["limit"] = options.Limit
		response["offset"] = options.Offset
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if download {
		filename := "cycles-export.json"
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RunningCycleResponse represents a running cycle with estimated completion time
type RunningCycleResponse struct {
	database.CycleWithDevice
	EstimatedCompletionTime *time.Time `json:"estimated_completion_time,omitempty"`
	ElapsedTime             string      `json:"elapsed_time"`
}

// GetRunningCyclesHandler handles GET /api/cycles/running requests
func GetRunningCyclesHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Only GET method is allowed",
		})
		return
	}

	// Get running cycles from database
	cycles, err := database.GetRunningCycles()
	if err != nil {
		logger.Error("Failed to retrieve running cycles", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve running cycles",
		})
		return
	}

	// Build response with estimated completion times
	response := make([]RunningCycleResponse, 0, len(cycles))
	now := time.Now()

	for _, cycle := range cycles {
		runningCycle := RunningCycleResponse{
			CycleWithDevice: cycle,
		}

		// Calculate elapsed time
		elapsed := now.Sub(cycle.StartTS)
		runningCycle.ElapsedTime = formatDuration(elapsed)

		// Estimate completion time if progress is available
		if cycle.ProgressPercent != nil && *cycle.ProgressPercent > 0 && *cycle.ProgressPercent < 100 {
			// Estimate: elapsed_time / progress_percent * 100
			estimatedTotalDuration := time.Duration(float64(elapsed) / float64(*cycle.ProgressPercent) * 100.0)
			estimatedCompletion := cycle.StartTS.Add(estimatedTotalDuration)
			runningCycle.EstimatedCompletionTime = &estimatedCompletion
		}

		response = append(response, runningCycle)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// formatDuration formats a duration as a human-readable string
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

