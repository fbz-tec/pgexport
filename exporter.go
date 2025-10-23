package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatXML  = "xml"
	FormatSQL  = "sql"
)

// Exporter interface defines export operations
type Exporter interface {
	Export(rows pgx.Rows, outputPath string, options ExportOptions) error
}

// ExportOptions holds export configuration
type ExportOptions struct {
	Format    string
	Delimiter rune
	TableName string
}

// dataExporter implements Exporter interface
type dataExporter struct{}

// NewExporter creates a new exporter instance
func NewExporter() Exporter {
	return &dataExporter{}
}

// Export exports query results to the specified format
func (e *dataExporter) Export(rows pgx.Rows, outputPath string, options ExportOptions) error {

	defer rows.Close()

	var rowCount int
	var err error

	switch options.Format {
	case FormatCSV:
		rowCount, err = e.writeCSV(rows, outputPath, options.Delimiter)
	case FormatJSON:
		rowCount, err = e.writeJSON(rows, outputPath)
	case FormatXML:
		rowCount, err = e.writeXML(rows, outputPath)
	case FormatSQL:
		rowCount, err = e.writeSQL(rows, outputPath, options.TableName)
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
func (e *dataExporter) writeCSV(rows pgx.Rows, csvPath string, delimiter rune) (int, error) {
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
	fieldDescriptions := rows.FieldDescriptions()
	headers := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		headers[i] = string(fd.Name)
	}

	if err := writer.Write(headers); err != nil {
		return 0, fmt.Errorf("error writing headers: %w", err)
	}

	// Write data rows
	rowCount := 0
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return rowCount, fmt.Errorf("error reading row: %w", err)
		}
		//format values to strings
		record := make([]string, len(values))
		for i, v := range values {
			record[i] = toString(v)
		}

		rowCount++

		if err := writer.Write(record); err != nil {
			return 0, fmt.Errorf("error writing row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			writer.Flush()
		}

	}

	return rowCount, nil
}

// exportToJSON writes query results to a JSON file with buffered I/O
func (e *dataExporter) writeJSON(rows pgx.Rows, jsonPath string) (int, error) {
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

// XMLRow represents a single row with dynamic fields
type XMLRow struct {
	Fields map[string]string
}

// MarshalXML implements custom XML marshaling for XMLRow
func (r XMLRow) MarshalXML(e *xml.Encoder, start xml.StartElement) error {

	start.Name.Local = "row"
	// Start the row element
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// Encode each field as a separate XML element
	for key, value := range r.Fields {
		elem := xml.StartElement{Name: xml.Name{Local: key}}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	// End the row element
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// exportToXML writes query results to an XML file with buffered I/O
func (e *dataExporter) writeXML(rows pgx.Rows, xmlPath string) (int, error) {
	file, err := os.Create(xmlPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	// Encode to XML with indentation
	encoder := xml.NewEncoder(bufferedWriter)
	encoder.Indent("", "  ")

	// Write XML header
	if _, err := bufferedWriter.WriteString(xml.Header); err != nil {
		return 0, fmt.Errorf("error writing XML header: %w", err)
	}

	startResults := xml.StartElement{Name: xml.Name{Local: "results"}}
	if err := encoder.EncodeToken(startResults); err != nil {
		return 0, fmt.Errorf("error starting <results>: %w", err)
	}

	// get fields names
	fieldDescriptions := rows.FieldDescriptions()
	fields := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		fields[i] = string(fd.Name)
	}

	rowCount := 0
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}
		xmlRow := XMLRow{
			Fields: make(map[string]string),
		}
		for i, field := range fields {
			xmlRow.Fields[field] = toString(values[i])
		}

		rowCount++

		if err := encoder.Encode(xmlRow); err != nil {
			return 0, fmt.Errorf("error encoding XML: %w", err)
		}
		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
		}

	}

	if err := encoder.EncodeToken(xml.EndElement{Name: startResults.Name}); err != nil {
		return 0, fmt.Errorf("error ending </results>: %w", err)
	}

	// Add final newline
	if _, err := bufferedWriter.WriteString("\n"); err != nil {
		return 0, fmt.Errorf("error writing final newline: %w", err)
	}

	return rowCount, nil
}

func (e *dataExporter) writeSQL(rows pgx.Rows, sqlPath string, tableName string) (int, error) {
	file, err := os.Create(sqlPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	rowCount := 0

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, fd := range fields {
		columns[i] = string(fd.Name)
	}
	size := len(columns)

	for rows.Next() {
		record := make([]string, size)

		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}

		//format values
		for i, val := range values {
			record[i] = formatSQLValue(val)
		}

		rowCount++
		//Create insert line with values
		line := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n", tableName, strings.Join(columns, ", "), strings.Join(record, ", "))

		if _, err := bufferedWriter.WriteString(line); err != nil {
			return 0, fmt.Errorf("error writing row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
		}
	}
	return rowCount, nil
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
		return fmt.Sprintf("%v", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		str := fmt.Sprintf("%v", val)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
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
