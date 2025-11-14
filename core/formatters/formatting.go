package formatters

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

// formatJSONValue formats a value for JSON export
func FormatJSONValue(v interface{}, layout string, loc *time.Location) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		return val.In(loc).Format(layout)
	case [16]byte:
		// UUID byte array
		return fmt.Sprintf("%x-%x-%x-%x-%x", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
	case []byte:
		return string(val)
	case pgtype.Numeric:
		if !val.Valid {
			return nil
		}
		f, err := val.Float64Value()
		if err != nil || !f.Valid {
			return nil
		}
		return f.Float64
	case pgtype.Interval:
		if !val.Valid {
			return nil
		}
		v, err := val.Value()
		if err != nil {
			return nil
		}
		return fmt.Sprintf("%v", v)
	default:
		return v
	}
}

// formatCSVValue formats a value for CSV export
func FormatCSVValue(v interface{}, layout string, loc *time.Location) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case time.Time:
		return val.In(loc).Format(layout)
	case [16]byte:
		// UUID byte array
		return fmt.Sprintf("%x-%x-%x-%x-%x", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
	case []byte:
		return string(val)
	case pgtype.Numeric:
		if !val.Valid {
			return ""
		}

		f, err := val.Float64Value()
		if err != nil || !f.Valid {
			return ""
		}
		return fmt.Sprintf("%.15g", f.Float64)
	case float32, float64:
		return fmt.Sprintf("%.15g", val)
	case pgtype.Interval:
		if !val.Valid {
			return ""
		}
		strVal, err := val.Value()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%v", strVal)
	case []interface{}:
		if len(val) == 0 {
			return "{}"
		}
		elems := make([]string, len(val))
		for i, elem := range val {
			elems[i] = fmt.Sprintf("%v", elem)
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ","))
	case map[string]interface{}:
		jsonStr, err := json.Marshal(val)
		if err != nil {
			return "{}"
		}
		return string(jsonStr)

	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatXMLValue formats a value for XML export
func FormatXMLValue(v interface{}, layout string, loc *time.Location) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case time.Time:
		return val.In(loc).Format(layout)

	case [16]byte:
		// UUID format
		return fmt.Sprintf("%x-%x-%x-%x-%x", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])

	case []byte:
		return string(val)

	case pgtype.Numeric:
		if !val.Valid {
			return ""
		}
		f, err := val.Float64Value()
		if err != nil || !f.Valid {
			return ""
		}
		return fmt.Sprintf("%.15g", f.Float64)
	case float32, float64:
		return fmt.Sprintf("%.15g", val)
	case pgtype.Interval:
		if !val.Valid {
			return ""
		}
		s, err := val.Value()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%v", s)

	case []interface{}:
		if len(val) == 0 {
			return "{}"
		}
		elems := make([]string, len(val))
		for i, elem := range val {
			elems[i] = fmt.Sprintf("%v", elem)
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ","))

	case map[string]interface{}:
		// Convert JSON object to string for XML
		jsonStr, err := json.Marshal(val)
		if err != nil {
			return ""
		}
		return string(jsonStr)

	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatYAMLValue formats a value for YAML export
func FormatYAMLValue(v interface{}, layout string, loc *time.Location) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {

	case time.Time:
		return val.In(loc).Format(layout)

	case [16]byte: // UUID
		return fmt.Sprintf("%x-%x-%x-%x-%x",
			val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])

	case []byte:
		return string(val)

	case pgtype.Numeric:
		if !val.Valid {
			return nil
		}
		f, err := val.Float64Value()
		if err != nil || !f.Valid {
			return nil
		}
		return f.Float64

	case pgtype.Interval:
		if !val.Valid {
			return nil
		}
		raw, err := val.Value()
		if err != nil {
			return nil
		}
		return fmt.Sprintf("%v", raw)

	default:
		// yaml.Encoder gère tout le reste automatiquement
		return v
	}
}

// formatSQLValue formats a value for SQL export
func FormatSQLValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case [16]byte:
		// UUID byte array
		return fmt.Sprintf("'%x-%x-%x-%x-%x'::uuid", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
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
	case pgtype.Interval:
		if !val.Valid {
			return ""
		}
		strVal, err := val.Value()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("'%v'::interval", strVal)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%.15g", val)
	case pgtype.Numeric:
		if !val.Valid {
			return "NULL"
		}

		f, err := val.Float64Value()
		if err != nil || !f.Valid {
			return "NULL"
		}
		return fmt.Sprintf("%.15g", f.Float64)
	case []interface{}:
		if len(val) == 0 {
			return "{}"
		}
		elems := make([]string, len(val))
		for i, elem := range val {
			elems[i] = fmt.Sprintf("%v", elem)
		}
		return fmt.Sprintf("'{%s}'", strings.Join(elems, ","))
	case map[string]interface{}:
		jsonStr, err := json.Marshal(val)
		if err != nil {
			return "'{}'::jsonb"
		}
		return fmt.Sprintf("'%s'::jsonb", string(jsonStr))
	default:
		str := fmt.Sprintf("%v", val)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

func QuoteIdent(s string) string {
	parts := strings.Split(s, ".")
	for i, part := range parts {
		parts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

func UserTimeZoneFormat(userTimefmt string, timeZone string) (string, *time.Location) {

	layout := ConvertUserTimeFormat(userTimefmt)

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

func ConvertUserTimeFormat(userTimefmt string) string {
	return timeFormatReplacer.Replace(userTimefmt)
}
