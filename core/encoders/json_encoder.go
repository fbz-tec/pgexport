package encoders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fbz-tec/pgxport/core/formatters"
)

// OrderedJsonEncoder encodes JSON while preserving key order
type OrderedJsonEncoder struct {
	timeLayout string
	timezone   *time.Location
}

// NewOrderedJsonEncoder creates a new ordered JSON encoder with time formatting options
func NewOrderedJsonEncoder(timeFormat, timeZone string) OrderedJsonEncoder {
	layout, loc := formatters.UserTimeZoneFormat(timeFormat, timeZone)
	return OrderedJsonEncoder{
		timeLayout: layout,
		timezone:   loc,
	}
}

// EncodeJSONWithOrder encodes a map to JSON preserving the order with proper indentation
func (o OrderedJsonEncoder) EncodeRow(keys []string, values []interface{}) ([]byte, error) {
	if len(keys) == 0 {
		return []byte("{}"), nil
	}

	var row bytes.Buffer

	// Pre-allocate memory to avoid reallocation
	row.Grow(len(keys) * 32)

	row.WriteString("{\n")

	for i, key := range keys {
		if i > 0 {
			row.WriteString(",\n")
		}

		// Add indentation (4 spaces for inner content)
		row.WriteString("    ")

		row.WriteString(fmt.Sprintf("%q", key))
		row.WriteString(": ")

		// value
		formattedValue := formatters.FormatJSONValue(values[i], o.timeLayout, o.timezone)

		// Marshal formatted value with HTML escaping disabled
		valueJSON, err := marshalWithoutHTMLEscape(formattedValue)
		if err != nil {
			return nil, fmt.Errorf("error marshaling value for key %q: %w", key, err)
		}

		row.Write(valueJSON)
	}

	row.WriteString("\n  }")
	return row.Bytes(), nil
}

func marshalWithoutHTMLEscape(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if _, ok := v.(map[string]interface{}); ok {
		encoder.SetIndent("    ", "  ")
	}

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	result := buf.Bytes()
	return bytes.TrimSuffix(result, []byte("\n")), nil
}
