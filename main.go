package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fbz-tec/pgexport/exporters"
	"github.com/spf13/cobra"
)

var (
	sqlQuery    string
	sqlFile     string
	outputPath  string
	format      string
	delimiter   string
	connString  string
	tableName   string
	compression string
	timeFormat  string
	timeZone    string
	withCopy    bool
	failOnEmpty bool
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime)

	var rootCmd = &cobra.Command{
		Use:   "pgxport",
		Short: "Export PostgreSQL query results to CSV, JSON, XML, or SQL formats",
		Long: `A powerful CLI tool to export PostgreSQL query results.
It supports direct SQL queries or SQL files, with customizable output options.
		
Supported output formats:
 • CSV  — standard text export with customizable delimiter
 • JSON — structured export for API or data processing
 • XML  — hierarchical export for interoperability
 • SQL  — generate INSERT statements`,
		Example: `  # Export with inline query
  pgxport -s "SELECT * FROM users" -o users.csv

  # Export from SQL file with custom delimiter
  pgxport -F query.sql -o output.csv -d ";"

  # Use the high-performance COPY mode for large CSV exports
  pgxport -s "SELECT * FROM events" -o events.csv -f csv --with-copy

  # Export to JSON
  pgxport -s "SELECT * FROM products" -o products.json -f json
  
  # Export to XML
  pgxport -s "SELECT * FROM orders" -o orders.xml -f xml

   # Export to SQL insert statements
  pgxport -s "SELECT * FROM orders" -o orders.sql -f sql -t orders_table`,
		RunE: runExport,
	}

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(GetVersionInfo())
		},
	}

	rootCmd.Flags().StringVarP(&sqlQuery, "sql", "s", "", "SQL query to execute")
	rootCmd.Flags().StringVarP(&sqlFile, "sqlfile", "F", "", "Path to SQL file containing the query")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (required)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "csv", "Output format (csv, json, xml, sql)")
	rootCmd.Flags().StringVarP(&timeFormat, "time-format", "T", "yyyy-MM-dd HH:mm:ss", "Custom time format (e.g. yyyy-MM-ddTHH:mm:ss.SSS)")
	rootCmd.Flags().StringVarP(&timeZone, "time-zone", "Z", "", "Time zone for date/time formatting (e.g. UTC, Europe/Paris). Defaults to local time zone.")
	rootCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "CSV delimiter character")
	rootCmd.Flags().StringVarP(&connString, "dsn", "", "", "Database connection string (postgres://user:pass@host:port/dbname)")
	rootCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table name for SQL insert exports")
	rootCmd.Flags().StringVarP(&compression, "compression", "z", "none", "Compression to apply to the output file (none, gzip, zip)")
	rootCmd.Flags().BoolVar(&withCopy, "with-copy", false, "Use PostgreSQL native COPY for CSV export (faster for large datasets)")
	rootCmd.Flags().BoolVar(&failOnEmpty, "fail-on-empty", false, "Exit with error if query returns 0 rows")

	rootCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

}

func runExport(cmd *cobra.Command, args []string) error {
	if err := validateExportParams(); err != nil {
		return err
	}

	var dbUrl string
	if connString != "" {
		dbUrl = connString
	} else {
		config := LoadConfig()
		if err := config.Validate(); err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		dbUrl = config.GetConnectionString()
	}

	var query string
	var err error

	if sqlFile != "" {
		query, err = readSQLFromFile(sqlFile)
		if err != nil {
			return fmt.Errorf("error reading SQL file: %w", err)
		}
	} else {
		query = sqlQuery
	}

	format = strings.ToLower(strings.TrimSpace(format))

	var delimRune rune = ','
	if format == "csv" {
		delimRune, err = parseDelimiter(delimiter)
		if err != nil {
			return fmt.Errorf("invalid delimiter: %w", err)
		}
	}

	store := NewStore()
	if err := store.Open(dbUrl); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer store.Close()

	options := exporters.ExportOptions{
		Format:      format,
		Delimiter:   delimRune,
		TableName:   tableName,
		Compression: compression,
		TimeFormat:  timeFormat,
		TimeZone:    timeZone,
	}

	log.Println("Executing query...")

	if format == "csv" && withCopy {

		log.Println("Using PostgreSQL native COPY mode for CSV export...")
		exporter := exporters.NewCopyExporter()

		rowCount, err := exporter.ExportCopy(store.GetConnection(), query, outputPath, options)

		if err != nil {
			return fmt.Errorf("error exporting to %s: %w", options.Format, err)
		}

		return handleExportResult(rowCount, outputPath)
	}

	rows, err := store.ExecuteQuery(context.Background(), query)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	exporter := exporters.NewExporter()

	rowCount, err := exporter.Export(rows, outputPath, options)

	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	return handleExportResult(rowCount, outputPath)
}

func validateExportParams() error {
	// Validate SQL query source
	if sqlQuery == "" && sqlFile == "" {
		return fmt.Errorf("Error: Either --sql or --sqlfile must be provided")
	}

	if sqlQuery != "" && sqlFile != "" {
		return fmt.Errorf("Error: Cannot use both --sql and --sqlfile at the same time")
	}

	// Normalize and validate format
	format = strings.ToLower(strings.TrimSpace(format))
	validFormats := []string{"csv", "json", "xml", "sql"}

	isValid := false
	for _, f := range validFormats {
		if format == f {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("Error: Invalid format '%s'. Valid formats are: %s",
			format, strings.Join(validFormats, ", "))
	}

	compression = strings.ToLower(strings.TrimSpace(compression))
	if compression == "" {
		compression = "none"
	}
	validCompressions := []string{"none", "gzip", "zip"}
	compressionValid := false
	for _, c := range validCompressions {
		if compression == c {
			compressionValid = true
			break
		}
	}

	if !compressionValid {
		return fmt.Errorf("Error: Invalid compression '%s'. Valid options are: %s",
			compression, strings.Join(validCompressions, ", "))
	}

	// Validate table name for SQL format
	if format == "sql" && strings.TrimSpace(tableName) == "" {
		return fmt.Errorf("Error: --table (-t) is required when using SQL format")
	}

	// Validate time format if provided
	if timeFormat != "" {
		if err := exporters.ValidateTimeFormat(timeFormat); err != nil {
			return fmt.Errorf("Error: Invalid time format '%s'. Use format like 'yyyy-MM-dd HH:mm:ss'", timeFormat)
		}
	}

	// Validate timezone if provided
	if timeZone != "" {
		if err := exporters.ValidateTimeZone(timeZone); err != nil {
			return fmt.Errorf("Error: Invalid timezone '%s'. Use format like 'UTC' or 'Europe/Paris'", timeZone)
		}
	}

	return nil
}

func readSQLFromFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %w", err)
	}
	return string(content), nil
}

func parseDelimiter(delim string) (rune, error) {
	delim = strings.TrimSpace(delim)

	if delim == "" {
		return 0, fmt.Errorf("delimiter cannot be empty")
	}

	if delim == `\t` {
		return '\t', nil
	}

	runes := []rune(delim)

	if len(runes) != 1 {
		return 0, fmt.Errorf("delimiter must be a single character (use \\t for tab)")
	}

	return runes[0], nil
}

func handleExportResult(rowCount int, outputPath string) error {
	if rowCount == 0 {

		if failOnEmpty {
			return fmt.Errorf("export failed: query returned 0 rows")
		}

		log.Printf("Warning: Query returned 0 rows. File created at %s but contains no data rows.", outputPath)

	} else {
		log.Printf("Successfully exported %d rows to %s", rowCount, outputPath)
	}

	return nil
}
