package exporters

import (
	"bufio"
	"encoding/xml"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	XmlRootElement = "results"
	XmlRowElement  = "row"
)

// exportToXML writes query results to an XML file with buffered I/O
func (e *dataExporter) writeXML(rows pgx.Rows, xmlPath string, options ExportOptions) (int, error) {
	writeCloser, err := createOutputWriter(xmlPath, options, FormatXML)
	if err != nil {
		return 0, err
	}
	defer writeCloser.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(writeCloser)
	defer bufferedWriter.Flush()

	// Encode to XML with indentation
	encoder := xml.NewEncoder(bufferedWriter)
	encoder.Indent("", "  ")

	// Write XML header
	if _, err := bufferedWriter.WriteString(xml.Header); err != nil {
		return 0, fmt.Errorf("error writing XML header: %w", err)
	}

	startResults := xml.StartElement{Name: xml.Name{Local: XmlRootElement}}
	if err := encoder.EncodeToken(startResults); err != nil {
		return 0, fmt.Errorf("error starting <%s>: %w", XmlRootElement, err)
	}

	// get fields names
	fieldDescriptions := rows.FieldDescriptions()
	fields := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		fields[i] = string(fd.Name)
	}

	//datetime layout(Golang format) and timezone
	layout, loc := userTimeZoneFormat(options.TimeFormat, options.TimeZone)

	rowCount := 0
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}

		startRow := xml.StartElement{Name: xml.Name{Local: XmlRowElement}}

		if err := encoder.EncodeToken(startRow); err != nil {
			return rowCount, fmt.Errorf("error starting <%s>: %w", XmlRowElement, err)
		}

		for i, field := range fields {
			elem := xml.StartElement{Name: xml.Name{Local: field}}
			if values[i] == nil {
				if err := encoder.EncodeToken(xml.StartElement{Name: elem.Name}); err != nil {
					return rowCount, fmt.Errorf("error starting <%s>: %w", elem, err)
				}
				if err := encoder.EncodeToken(xml.EndElement{Name: elem.Name}); err != nil {
					return rowCount, fmt.Errorf("error closing </%s>: %w", elem, err)
				}
				continue
			}
			if err := encoder.EncodeElement(formatValue(values[i], layout, loc), elem); err != nil {
				return rowCount, fmt.Errorf("error encoding field %s: %w", field, err)
			}
		}

		// End </row>
		if err := encoder.EncodeToken(xml.EndElement{Name: startRow.Name}); err != nil {
			return rowCount, fmt.Errorf("error closing </%s>: %w", XmlRowElement, err)
		}

		rowCount++

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
		}

	}

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := encoder.EncodeToken(xml.EndElement{Name: startResults.Name}); err != nil {
		return 0, fmt.Errorf("error ending </%s>: %w", XmlRootElement, err)
	}

	// Add final newline
	if _, err := bufferedWriter.WriteString("\n"); err != nil {
		return 0, fmt.Errorf("error writing final newline: %w", err)
	}

	return rowCount, nil
}
