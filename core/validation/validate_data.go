package validation

import (
	"fmt"
	"time"

	"github.com/fbz-tec/pgxport/core/formatters"
)

// ValidateTimeZone checks if a timezone string is valid
func ValidateTimeZone(timezone string) error {
	if timezone == "" {
		return nil // Empty is valid (uses Local)
	}

	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}

	return nil
}

// ValidateTimeFormat validates that a time format is valid by testing it
func ValidateTimeFormat(format string) error {

	// Empty format is invalid
	if format == "" {
		return fmt.Errorf("time format cannot be empty")
	}

	// Test the format with a known time
	testTime := time.Date(2006, 1, 2, 15, 4, 5, 123456789, time.UTC)
	layout := formatters.ConvertUserTimeFormat(format)

	// Try to format and parse back
	formatted := testTime.Format(layout)
	_, err := time.Parse(layout, formatted)

	if err != nil {
		return fmt.Errorf("invalid time format %q: %w", format, err)
	}

	return nil
}
