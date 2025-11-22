package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"steri-connect-go/internal/database"
	"steri-connect-go/internal/logging"
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

