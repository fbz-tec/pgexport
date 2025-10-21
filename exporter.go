package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	FormatCSV  = "csv"
	FormatJSON = "json"
)

// Exporter interface defines export operations
type Exporter interface {
	Export(result *QueryResult, outputPath string, options ExportOptions) error
}

// ExportOptions holds export configuration
type ExportOptions struct {
	Format    string
	Delimiter rune
}

// dataExporter implements Exporter interface
type dataExporter struct{}

// NewExporter creates a new exporter instance
func NewExporter() Exporter {
	return &dataExporter{}
}

// Export exports query results to the specified format
func (e *dataExporter) Export(result *QueryResult, outputPath string, options ExportOptions) error {
	var rowCount int
	var err error

	switch options.Format {
	case FormatCSV:
		rowCount, err = e.exportToCSV(result, outputPath, options.Delimiter)
	case FormatJSON:
		rowCount, err = e.exportToJSON(result, outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", options.Format)
	}

	if err != nil {
		return fmt.Errorf("error exporting to %s: %w", options.Format, err)
	}

	log.Printf("Successfully exported %d rows to %s", rowCount, outputPath)
	return nil
}

// exportToCSV writes query results to a CSV file with buffered I/O
func (e *dataExporter) exportToCSV(result *QueryResult, csvPath string, delimiter rune) (int, error) {
	file, err := os.Create(csvPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	writer := csv.NewWriter(bufferedWriter)
	writer.Comma = delimiter
	defer writer.Flush()

	// Write headers
	if err := writer.Write(result.Columns); err != nil {
		return 0, fmt.Errorf("error writing headers: %w", err)
	}

	// Write data rows
	for i, row := range result.Rows {
		record := make([]string, len(row))
		for j, v := range row {
			record[j] = toString(v)
		}

		if err := writer.Write(record); err != nil {
			return i, fmt.Errorf("error writing row %d: %w", i+1, err)
		}
	}

	return len(result.Rows), nil
}

// exportToJSON writes query results to a JSON file with buffered I/O
func (e *dataExporter) exportToJSON(result *QueryResult, jsonPath string) (int, error) {
	file, err := os.Create(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	// Convert rows to array of maps
	var results []map[string]interface{}
	for _, row := range result.Rows {
		entry := make(map[string]interface{})
		for i, col := range result.Columns {
			entry[col] = formatValue(row[i])
		}
		results = append(results, entry)
	}

	// Encode to JSON with indentation
	encoder := json.NewEncoder(bufferedWriter)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(results); err != nil {
		return len(results), fmt.Errorf("error encoding JSON: %w", err)
	}

	return len(results), nil
}

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
	default:
		return v
	}
}

// toString converts any value to string for CSV export
func toString(v interface{}) string {
	formatted := formatValue(v)
	if formatted == nil {
		return ""
	}
	return fmt.Sprintf("%v", formatted)
}
