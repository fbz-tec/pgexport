package exporters

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fbz-tec/pgexport/logger"
	"github.com/jackc/pgx/v5"
)

// exportToJSON writes query results to a JSON file with buffered I/O
func (e *dataExporter) writeJSON(rows pgx.Rows, jsonPath string, options ExportOptions) (int, error) {

	start := time.Now()
	logger.Debug("Preparing JSON export (indent=2 spaces, compression=%s)", options.Compression)

	writeCloser, err := createOutputWriter(jsonPath, options, FormatJSON)
	if err != nil {
		return 0, err
	}

	defer writeCloser.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(writeCloser)
	defer bufferedWriter.Flush()

	// Get object keys names
	fieldDescriptions := rows.FieldDescriptions()
	keys := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		keys[i] = string(fd.Name)
	}

	// Write opening bracket
	if _, err := bufferedWriter.WriteString("[\n"); err != nil {
		return 0, fmt.Errorf("error writing start of JSON array: %w", err)
	}

	//datetime layout(Golang format) and timezone
	layout, loc := userTimeZoneFormat(options.TimeFormat, options.TimeZone)

	rowCount := 0

	logger.Debug("Starting to write JSON objects...")

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}

		//Convert row to map
		entry := make(map[string]interface{}, len(keys))
		for i, key := range keys {
			entry[key] = formatJSONValue(values[i], layout, loc)
		}
		rowCount++

		// Write comma separator for subsequent entries
		if rowCount > 1 {
			if _, err := bufferedWriter.WriteString(",\n"); err != nil {
				return 0, fmt.Errorf("error writing comma for row %d: %w", rowCount, err)
			}
		}

		// Encode JSON to buffer with proper HTML escaping disabled
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("  ", "  ")

		if err := encoder.Encode(entry); err != nil {
			return 0, fmt.Errorf("error encoding JSON: %w", err)
		}

		// Write the indented JSON object (trim the trailing newline from Encode)
		jsonStr := strings.TrimSuffix(buf.String(), "\n")

		if _, err := bufferedWriter.WriteString("  "); err != nil {
			return 0, fmt.Errorf("error writing indentation for row %d: %w", rowCount, err)
		}

		if _, err := bufferedWriter.WriteString(jsonStr); err != nil {
			return 0, fmt.Errorf("error writing JSON object for row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
			logger.Debug("%d JSON objects written...", rowCount)
		}
	}

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	// Write closing bracket
	if _, err := bufferedWriter.WriteString("\n]\n"); err != nil {
		return 0, fmt.Errorf("error writing end of JSON array: %w", err)
	}

	logger.Debug("JSON export completed successfully: %d rows written in %v", rowCount, time.Since(start))

	return rowCount, nil
}
