package database

import (
	"fmt"
	"strings"
	"time"
)

// GetToday returns today's date as YYYY-MM-DD.
func GetToday() string {
	return time.Now().Format("2006-01-02")
}

// MakePeriodID creates a period_id from start and end dates.
// If start == end, returns just the date (e.g., "2026-02-06").
// Otherwise returns a range (e.g., "2026-02-01..2026-02-06").
func MakePeriodID(start, end string) string {
	if start == end {
		return start
	}
	return start + ".." + end
}

// FormatPeriodDisplay formats a period_id for human-readable display.
// Single day: "Feb 06, 2026"
// Range: "Feb 01 - Feb 06, 2026"
func FormatPeriodDisplay(periodID string) string {
	if strings.Contains(periodID, "..") {
		parts := strings.SplitN(periodID, "..", 2)
		if len(parts) != 2 {
			return periodID
		}
		start, err := time.Parse("2006-01-02", parts[0])
		if err != nil {
			return periodID
		}
		end, err := time.Parse("2006-01-02", parts[1])
		if err != nil {
			return periodID
		}
		return fmt.Sprintf("%s - %s", start.Format("Jan 02"), end.Format("Jan 02, 2006"))
	}

	d, err := time.Parse("2006-01-02", periodID)
	if err != nil {
		return periodID
	}
	return d.Format("Jan 02, 2006")
}

// PeriodEndDate extracts the end date from a period_id.
// For range periods (YYYY-MM-DD..YYYY-MM-DD), returns the end date.
// For single-day periods, returns the date itself.
func PeriodEndDate(periodID string) string {
	if strings.Contains(periodID, "..") {
		parts := strings.SplitN(periodID, "..", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return periodID
}
