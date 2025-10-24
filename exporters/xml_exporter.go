package exporters

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

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
func (e *dataExporter) writeXML(rows pgx.Rows, xmlPath string, option ExportOptions) (int, error) {
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
