package exporters

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

// formatValue is kept for backward compatibility (not used in new code)
func formatValue(v interface{}, layout string, loc *time.Location) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case time.Time:
		return val.In(loc).Format(layout)
	case []byte:
		return string(val)
	case float32:
		return fmt.Sprintf("%.15g", val)
	case float64:
		return fmt.Sprintf("%.15g", val)
	default:
		return val
	}
}

// formatValueByOID is the central function that handles all PostgreSQL type conversions
// It returns interface{} for maximum flexibility across different export formats
func formatValueByOID(val interface{}, valueType uint32, userTimefmt string, timeZone string) interface{} {
	if val == nil {
		return nil
	}

	switch valueType {
	case pgtype.DateOID:
		if t, ok := val.(time.Time); ok {
			dateFmt := extractUserDateFormat(userTimefmt)
			layout := convertUserTimeFormat(dateFmt)
			return t.Format(layout)
		}

	case pgtype.TimestampOID:
		if t, ok := val.(time.Time); ok {
			layout := convertUserTimeFormat(userTimefmt)
			return t.Format(layout)
		}

	case pgtype.TimestamptzOID:
		if t, ok := val.(time.Time); ok {
			layout, loc := userTimeZoneFormat(userTimefmt, timeZone)
			return t.In(loc).Format(layout)
		}

	case pgtype.UUIDOID:
		if uuid, ok := val.([16]byte); ok {
			return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
		}

	case pgtype.ByteaOID:
		if bytes, ok := val.([]byte); ok {
			return string(bytes)
		}

	case pgtype.NumericOID:
		if num, ok := val.(pgtype.Numeric); ok {
			if !num.Valid {
				return nil
			}
			f, err := num.Float64Value()
			if err != nil || !f.Valid {
				return nil
			}
			return f.Float64
		}

	case pgtype.IntervalOID:
		if interval, ok := val.(pgtype.Interval); ok {
			if !interval.Valid {
				return nil
			}
			strVal, err := interval.Value()
			if err != nil {
				return nil
			}
			return strVal
		}

	case pgtype.JSONBOID, pgtype.JSONOID:
		// Return as-is for JSON export, will be marshaled for CSV/XML
		return val
	}

	// Return value as-is for generic types
	return val
}

// formatJSONValue formats a value for JSON export
func formatJSONValue(val interface{}, valueType uint32, userTimefmt string, timeZone string) interface{} {
	return formatValueByOID(val, valueType, userTimefmt, timeZone)
}

// formatCSVValue formats a value for CSV export
func formatCSVValue(val interface{}, valueType uint32, userTimefmt string, timeZone string) string {
	result := formatValueByOID(val, valueType, userTimefmt, timeZone)

	if result == nil {
		return ""
	}

	// Handle specific conversions for CSV format
	switch v := result.(type) {
	case string:
		return v

	case float64:
		return fmt.Sprintf("%.15g", v)

	case float32:
		return fmt.Sprintf("%.15g", v)

	case []interface{}:
		if len(v) == 0 {
			return "{}"
		}
		elems := make([]string, len(v))
		for i, elem := range v {
			elems[i] = fmt.Sprintf("%v", elem)
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ","))

	default:
		// Special handling for JSON/JSONB in CSV
		if valueType == pgtype.JSONBOID || valueType == pgtype.JSONOID {
			jsonStr, err := json.Marshal(v)
			if err != nil {
				return "{}"
			}
			return string(jsonStr)
		}
		return fmt.Sprintf("%v", v)
	}
}

// formatXMLValue formats a value for XML export
func formatXMLValue(val interface{}, valueType uint32, userTimefmt string, timeZone string) string {
	result := formatValueByOID(val, valueType, userTimefmt, timeZone)

	if result == nil {
		return ""
	}

	// Handle specific conversions for XML format
	switch v := result.(type) {
	case string:
		return v

	case float64:
		return fmt.Sprintf("%.15g", v)

	case float32:
		return fmt.Sprintf("%.15g", v)

	case []interface{}:
		if len(v) == 0 {
			return "{}"
		}
		elems := make([]string, len(v))
		for i, elem := range v {
			elems[i] = fmt.Sprintf("%v", elem)
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ","))

	default:
		// Special handling for JSON/JSONB in XML
		if valueType == pgtype.JSONBOID || valueType == pgtype.JSONOID {
			jsonStr, err := json.Marshal(v)
			if err != nil {
				return ""
			}
			return string(jsonStr)
		}
		return fmt.Sprintf("%v", v)
	}
}

// formatSQLValue formats a value for SQL export
func formatSQLValue(val interface{}, valueType uint32) string {
	if val == nil {
		return "NULL"
	}

	switch valueType {
	case pgtype.DateOID:
		if t, ok := val.(time.Time); ok {
			return fmt.Sprintf("'%s'::date", t.Format("2006-01-02"))
		}

	case pgtype.TimestampOID:
		if t, ok := val.(time.Time); ok {
			return fmt.Sprintf("'%s'::timestamp", t.Format("2006-01-02 15:04:05"))
		}

	case pgtype.TimestamptzOID:
		if t, ok := val.(time.Time); ok {
			return fmt.Sprintf("'%s'::timestamptz", t.Format("2006-01-02 15:04:05"))
		}

	case pgtype.UUIDOID:
		if uuid, ok := val.([16]byte); ok {
			return fmt.Sprintf("'%x-%x-%x-%x-%x'::uuid", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
		}

	case pgtype.ByteaOID:
		if bytes, ok := val.([]byte); ok {
			escaped := strings.ReplaceAll(string(bytes), "'", "''")
			return fmt.Sprintf("'%s'::bytea", escaped)
		}

	case pgtype.BoolOID:
		if b, ok := val.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}

	case pgtype.NumericOID:
		if num, ok := val.(pgtype.Numeric); ok {
			if !num.Valid {
				return "NULL"
			}
			f, err := num.Float64Value()
			if err != nil {
				return "NULL"
			}
			return fmt.Sprintf("%.15g", f.Float64)
		}

	case pgtype.IntervalOID:
		if interval, ok := val.(pgtype.Interval); ok {
			if !interval.Valid {
				return "NULL"
			}
			strVal, err := interval.Value()
			if err != nil {
				return "NULL"
			}
			return fmt.Sprintf("'%v'::interval", strVal)
		}

	case pgtype.JSONBOID:
		jsonStr, err := json.Marshal(val)
		if err != nil {
			return "'{}'::jsonb"
		}
		return fmt.Sprintf("'%s'::jsonb", string(jsonStr))

	case pgtype.JSONOID:
		jsonStr, err := json.Marshal(val)
		if err != nil {
			return "'{}'::json"
		}
		return fmt.Sprintf("'%s'::json", string(jsonStr))
	}

	// Generic SQL value formatting
	switch v := val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)

	case float32, float64:
		return fmt.Sprintf("%.15g", val)

	case []interface{}:
		if len(v) == 0 {
			return "'{}'"
		}
		elems := make([]string, len(v))
		for i, elem := range v {
			elems[i] = fmt.Sprintf("%v", elem)
		}
		return fmt.Sprintf("'{%s}'", strings.Join(elems, ","))

	default:
		str := fmt.Sprintf("%v", val)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// quoteIdent quotes a SQL identifier (table or column name)
func quoteIdent(s string) string {
	parts := strings.Split(s, ".")
	for i, part := range parts {
		parts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

// userTimeZoneFormat returns the Go time layout and location for the user's format and timezone
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

// convertUserTimeFormat converts user-friendly format to Go time format
func convertUserTimeFormat(userTimefmt string) string {
	return timeFormatReplacer.Replace(userTimefmt)
}

// extractUserDateFormat extracts only the date portion from a datetime format string.
// For example, "yyyy-MM-dd HH:mm:ss" becomes "yyyy-MM-dd".
// This ensures DATE columns are exported without time components.
func extractUserDateFormat(userFmt string) string {
	dateTokens := []string{"yyyy", "yy", "MM", "dd"}
	last := -1
	for _, tok := range dateTokens {
		idx := strings.LastIndex(userFmt, tok)
		if idx != -1 {
			end := idx + len(tok)
			if end > last {
				last = end
			}
		}
	}

	if last == -1 {
		// No date tokens found, return original
		return userFmt
	}
	return strings.TrimSpace(userFmt[:last])
}

// ValidateTimeFormat validates that a time format is valid by testing it
func ValidateTimeFormat(format string) error {
	// Empty format is invalid
	if format == "" {
		return fmt.Errorf("time format cannot be empty")
	}

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
