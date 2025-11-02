package exporters

import (
	"fmt"

	"github.com/fbz-tec/pgexport/logger"
	"github.com/jackc/pgx/v5"
)

const (
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatXML  = "xml"
	FormatSQL  = "sql"
)

// ExportOptions holds export configuration
type ExportOptions struct {
	Format      string
	Delimiter   rune
	TableName   string
	Compression string
	TimeFormat  string
	TimeZone    string
	NoHeader    bool
}

// Exporter interface defines export operations
type Exporter interface {
	Export(rows pgx.Rows, outputPath string, options ExportOptions) (int, error)
}

// Optional capability interface for exporters that can use PostgreSQL COPY
type CopyCapable interface {
	ExportCopy(conn *pgx.Conn, query string, outputPath string, options ExportOptions) (int, error)
}

// dataExporter implements Exporter & CopyCapable interfaces
type dataExporter struct{}

// NewExporter creates a new exporter instance
func NewExporter() Exporter {
	return &dataExporter{}
}

func NewCopyExporter() CopyCapable {
	return &dataExporter{}
}

// Export exports query results to the specified format
func (e *dataExporter) Export(rows pgx.Rows, outputPath string, options ExportOptions) (int, error) {

	logger.Debug("Starting export to %s format → output: %s", options.Format, outputPath)

	var rowCount int
	var err error

	switch options.Format {
	case FormatCSV:
		rowCount, err = e.writeCSV(rows, outputPath, options)
	case FormatJSON:
		rowCount, err = e.writeJSON(rows, outputPath, options)
	case FormatXML:
		rowCount, err = e.writeXML(rows, outputPath, options)
	case FormatSQL:
		rowCount, err = e.writeSQL(rows, outputPath, options)
	default:
		return 0, fmt.Errorf("unsupported format: %s", options.Format)
	}

	if err != nil {
		return rowCount, fmt.Errorf("error exporting to %s: %w", options.Format, err)
	}

	return rowCount, nil
}

func (e *dataExporter) ExportCopy(conn *pgx.Conn, query string, outputPath string, options ExportOptions) (int, error) {

	logger.Debug("Starting COPY export for %s format → output: %s", options.Format, outputPath)

	var rowCount int
	var err error
	switch options.Format {
	case FormatCSV:
		rowCount, err = e.writeCopyCSV(conn, query, outputPath, options)
	default:
		return 0, fmt.Errorf("unsupported format: %s", options.Format)
	}

	if err != nil {
		return rowCount, fmt.Errorf("error exporting to %s: %w", options.Format, err)
	}

	return rowCount, nil
}
