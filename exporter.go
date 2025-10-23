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
)

const (
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatXML  = "xml"
	FormatSQL  = "sql"
)

// Exporter interface defines export operations
type Exporter interface {
	Export(result *QueryResult, outputPath string, options ExportOptions) error
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
func (e *dataExporter) Export(result *QueryResult, outputPath string, options ExportOptions) error {
	var rowCount int
	var err error

	switch options.Format {
	case FormatCSV:
		rowCount, err = e.writeCSV(result, outputPath, options.Delimiter)
	case FormatJSON:
		rowCount, err = e.writeJSON(result, outputPath)
	case FormatXML:
		rowCount, err = e.writeXML(result, outputPath)
	case FormatSQL:
		rowCount, err = e.writeSQL(result, outputPath, options.TableName)
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
func (e *dataExporter) writeCSV(result *QueryResult, csvPath string, delimiter rune) (int, error) {
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
func (e *dataExporter) writeJSON(result *QueryResult, jsonPath string) (int, error) {
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

// XMLResults represents the XML root structure
type XMLResults struct {
	XMLName xml.Name `xml:"results"`
	Rows    []XMLRow `xml:"row"`
}

// XMLRow represents a single row with dynamic fields
type XMLRow struct {
	Fields map[string]string
}

// MarshalXML implements custom XML marshaling for XMLRow
func (r XMLRow) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
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
func (e *dataExporter) writeXML(result *QueryResult, xmlPath string) (int, error) {
	file, err := os.Create(xmlPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	// Write XML header
	if _, err := bufferedWriter.WriteString(xml.Header); err != nil {
		return 0, fmt.Errorf("error writing XML header: %w", err)
	}

	// Build XML structure
	xmlResults := XMLResults{
		Rows: make([]XMLRow, 0, len(result.Rows)),
	}

	for _, row := range result.Rows {
		xmlRow := XMLRow{
			Fields: make(map[string]string),
		}

		for i, col := range result.Columns {
			xmlRow.Fields[col] = toString(row[i])
		}

		xmlResults.Rows = append(xmlResults.Rows, xmlRow)
	}

	// Encode to XML with indentation
	encoder := xml.NewEncoder(bufferedWriter)
	encoder.Indent("", "  ")

	if err := encoder.Encode(xmlResults); err != nil {
		return len(xmlResults.Rows), fmt.Errorf("error encoding XML: %w", err)
	}

	// Add final newline
	if _, err := bufferedWriter.WriteString("\n"); err != nil {
		return len(xmlResults.Rows), fmt.Errorf("error writing final newline: %w", err)
	}

	return len(xmlResults.Rows), nil
}

func (e *dataExporter) writeSQL(result *QueryResult, sqlPath string, tableName string) (int, error) {
	file, err := os.Create(sqlPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	for i, row := range result.Rows {
		record := make([]string, len(row))

		//format values
		for j, val := range row {
			record[j] = formatSQLValue(val)
		}

		//Create insert line with values
		line := fmt.Sprintf("INSERT INTO %s VALUES (%s);\n", tableName, strings.Join(record, ", "))

		if _, err := bufferedWriter.WriteString(line); err != nil {
			return i, fmt.Errorf("error writing row %d: %w", i+1, err)
		}
	}

	return len(result.Rows), nil
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
