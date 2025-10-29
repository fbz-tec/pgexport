package exporters

import (
	"fmt"
	"log"
	"strings"
	"time"
)

var timeFormatReplacer = strings.NewReplacer(
	"yyyy", "2006",
	"yy", "06",
	"MM", "01",
	"dd", "02",
	"HH", "15",
	"mm", "04",
	"ss", "05",
	"SSS", "000", // Milliseconds
	"S", "0", // Deciseconds
)

// formatValue formats a value for export (unified function)
func formatValue(v interface{}, layout string, loc *time.Location) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		return val.In(loc).Format(layout)
	case []byte:
		return string(val)
	case float32, float64:
		return fmt.Sprintf("%.15g", val)
	default:
		return v
	}
}

func formatSQLValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case string:
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		str := string(val)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case time.Time:
		return fmt.Sprintf("'%s'", val.Format("2006-01-02T15:04:05.000"))
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%.15g", val)
	default:
		str := fmt.Sprintf("%v", val)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

func quoteIdent(s string) string {
	parts := strings.Split(s, ".")
	for i, part := range parts {
		parts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

func userTimeZoneFormat(userTimefmt string, timeZone string) (string, *time.Location) {

	layout := convertUserTimeFormat(userTimefmt)

	if timeZone == "" {
		return layout, time.Local
	}

	loc, err := time.LoadLocation(timeZone)

	if err != nil {
		log.Printf("Warning: Invalid timezone %q, using local time: %v", timeZone, err)
		return layout, time.Local
	}

	return layout, loc
}

func convertUserTimeFormat(userTimefmt string) string {
	return timeFormatReplacer.Replace(userTimefmt)
}

// ValidateTimeFormat validates that a time format is valid by testing it
func ValidateTimeFormat(format string) error {
	// Test the format with a known time
	testTime := time.Date(2006, 1, 2, 15, 4, 5, 123456789, time.UTC)
	layout := convertUserTimeFormat(format)

	// Try to format and parse back
	formatted := testTime.Format(layout)
	_, err := time.Parse(layout, formatted)

	if err != nil {
		return fmt.Errorf("invalid time format %q: %w", format, err)
	}

	return nil
}

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
