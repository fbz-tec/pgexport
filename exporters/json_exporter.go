package exporters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// exportToJSON writes query results to a JSON file with buffered I/O
func (e *dataExporter) writeJSON(rows pgx.Rows, jsonPath string, options ExportOptions) (int, error) {
	file, err := os.Create(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	// Encode to JSON with indentation
	encoder := json.NewEncoder(bufferedWriter)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	// Get object keys names
	fieldDescriptions := rows.FieldDescriptions()
	keys := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		keys[i] = string(fd.Name)
	}

	if _, err := bufferedWriter.WriteString("[\n"); err != nil {
		return 0, fmt.Errorf("error writing start of JSON array: %w", err)
	}

	rowCount := 0

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}

		//Convert row to map
		entry := make(map[string]interface{}, len(keys))
		for i, key := range keys {
			entry[key] = formatValue(values[i])
		}
		rowCount++

		if rowCount > 1 {
			if _, err := bufferedWriter.WriteString(",\n"); err != nil {
				return 0, fmt.Errorf("error writing comma for row %d: %w", rowCount, err)
			}
		}

		if err := encoder.Encode(entry); err != nil {
			return 0, fmt.Errorf("error encoding JSON: %w", err)
		}

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
		}
	}

	if _, err := bufferedWriter.WriteString("]\n"); err != nil {
		return 0, fmt.Errorf("error writing end of JSON array: %w", err)
	}

	return rowCount, nil
}
