package exporters

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/fbz-tec/pgxport/internal/logger"
	"github.com/jackc/pgx/v5"
)

type xmlExporter struct{}

// writes query results to an XML file with buffered I/O
func (e *xmlExporter) Export(rows pgx.Rows, xmlPath string, options ExportOptions) (int, error) {

	start := time.Now()
	logger.Debug("Preparing XML export (indent=2 spaces, compression=%s)", options.Compression)

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

	logger.Debug("XML header written")

	startResults := xml.StartElement{Name: xml.Name{Local: options.XmlRootElement}}
	if err := encoder.EncodeToken(startResults); err != nil {
		return 0, fmt.Errorf("error starting <%s>: %w", options.XmlRootElement, err)
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

	logger.Debug("Starting to write XML rows...")

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}

		startRow := xml.StartElement{Name: xml.Name{Local: options.XmlRowElement}}

		if err := encoder.EncodeToken(startRow); err != nil {
			return rowCount, fmt.Errorf("error opening <%s>: %w", options.XmlRowElement, err)
		}

		for i, field := range fields {
			elem := xml.StartElement{Name: xml.Name{Local: field}}
			val := formatXMLValue(values[i], layout, loc)
			if val == "" {
				if err := encoder.EncodeToken(xml.StartElement{Name: elem.Name}); err != nil {
					return rowCount, fmt.Errorf("error opening <%s>: %w", elem, err)
				}
				if err := encoder.EncodeToken(xml.EndElement{Name: elem.Name}); err != nil {
					return rowCount, fmt.Errorf("error closing </%s>: %w", elem, err)
				}
				continue
			}
			isJSONLike := strings.HasPrefix(val, "{") || strings.HasPrefix(val, "[") || strings.Contains(val, "\":")
			if isJSONLike {
				if err := encoder.EncodeToken(elem); err != nil {
					return rowCount, fmt.Errorf("error opening <%s>: %w", field, err)
				}
				if _, err := bufferedWriter.WriteString(val); err != nil {
					return rowCount, fmt.Errorf("error writing raw value for <%s>: %w", field, err)
				}
				if err := encoder.EncodeToken(xml.EndElement{Name: elem.Name}); err != nil {
					return rowCount, fmt.Errorf("error closing </%s>: %w", field, err)
				}
			} else {
				if err := encoder.EncodeElement(val, elem); err != nil {
					return rowCount, fmt.Errorf("error encoding field %s: %w", field, err)
				}
			}

		}

		// End </row>
		if err := encoder.EncodeToken(xml.EndElement{Name: startRow.Name}); err != nil {
			return rowCount, fmt.Errorf("error closing </%s>: %w", options.XmlRowElement, err)
		}

		rowCount++

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
			logger.Debug("%d XML rows written...", rowCount)
		}

	}

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	if err := encoder.EncodeToken(xml.EndElement{Name: startResults.Name}); err != nil {
		return 0, fmt.Errorf("error ending </%s>: %w", options.XmlRootElement, err)
	}

	// Add final newline
	if _, err := bufferedWriter.WriteString("\n"); err != nil {
		return 0, fmt.Errorf("error writing final newline: %w", err)
	}

	logger.Debug("XML export completed successfully: %d rows written in %v", rowCount, time.Since(start))

	return rowCount, nil
}

func init() {
	MustRegisterExporter(FormatXML, func() Exporter { return &xmlExporter{} })
}
