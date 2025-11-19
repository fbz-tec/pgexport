package exporters

import (
	"github.com/jackc/pgx/v5"
)

const (
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatXML  = "xml"
	FormatSQL  = "sql"
	FormatYAML = "yaml"
	FormatXLSX = "xlsx"
)

// ExportOptions holds export configuration
type ExportOptions struct {
	Format          string
	Delimiter       rune
	TableName       string
	Compression     string
	TimeFormat      string
	TimeZone        string
	NoHeader        bool
	XmlRootElement  string
	XmlRowElement   string
	RowPerStatement int
}

// Exporter interface defines export operations
type Exporter interface {
	Export(rows pgx.Rows, outputPath string, options ExportOptions) (int, error)
}

// Optional capability interface for exporters that can use PostgreSQL COPY
type CopyCapable interface {
	ExportCopy(conn *pgx.Conn, query string, outputPath string, options ExportOptions) (int, error)
}
