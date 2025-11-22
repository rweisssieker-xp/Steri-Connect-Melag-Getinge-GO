package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"steri-connect-go/internal/database"
)

// GenerateCyclesCSV generates a CSV file for cycle data
func GenerateCyclesCSV(cycles []database.CycleWithDevice) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{
		"Cycle ID",
		"Device Name",
		"Device IP",
		"Manufacturer",
		"Program",
		"Start Time",
		"End Time",
		"Phase",
		"Progress (%)",
		"Temperature (°C)",
		"Pressure (bar)",
		"Result",
		"Error Code",
		"Error Description",
	}

	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write cycle data rows
	for _, cycle := range cycles {
		row := []string{
			fmt.Sprintf("%d", cycle.ID),
			cycle.DeviceName,
			cycle.DeviceIP,
			cycle.Manufacturer,
			cycle.Program,
			cycle.StartTS.Format("2006-01-02 15:04:05"),
			formatNullableTime(cycle.EndTS),
			cycle.Phase,
			formatNullableInt(cycle.ProgressPercent),
			formatNullableFloat(cycle.Temperature),
			formatNullableFloat(cycle.Pressure),
			cycle.Result,
			cycle.ErrorCode,
			cycle.ErrorDescription,
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Flush and check for errors
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV: %w", err)
	}

	return buf.Bytes(), nil
}

// formatNullableTime formats a nullable time.Time pointer
func formatNullableTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// formatNullableInt formats a nullable int pointer
func formatNullableInt(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

// formatNullableFloat formats a nullable float64 pointer
func formatNullableFloat(f *float64) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *f)
}

