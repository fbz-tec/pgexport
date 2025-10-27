package exporters

import (
	"fmt"
	"strings"
	"time"
)

// formatValue formats a value for export (unified function)
func formatValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case time.Time:
		return val.Format("2006-01-02T15:04:05.000")
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

// toString converts any value to string for CSV export
func toString(v interface{}) string {
	formatted := formatValue(v)
	if formatted == nil {
		return ""
	}
	return fmt.Sprintf("%v", formatted)
}
