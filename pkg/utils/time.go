package utils

import (
	"errors"
	"time"
)

// ParseFlexibleDate tries to parse a date string using various common formats
func ParseFlexibleDate(dateStr string) (time.Time, error) {
	// 1. Try RFC3339 (Standard ISO8601 with Z or ±HH:MM)
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	// 2. Try ISO8601 with milliseconds but no timezone
	if t, err := time.Parse("2006-01-02T15:04:05.000", dateStr); err == nil {
		return t, nil
	}

	// 3. Try ISO8601 without milliseconds or timezone
	if t, err := time.Parse("2006-01-02T15:04:05", dateStr); err == nil {
		return t, nil
	}

	// 4. Try just the date
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}

	return time.Time{}, errors.New("invalid date format")
}
