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
	sqlQuery   string
	sqlFile    string
	outputPath string
	format     string
	delimiter  string
	connString string
	tableName  string
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime)

	var rootCmd = &cobra.Command{
		Use:   "pgxport",
		Short: "Export PostgreSQL query results to various formats",
		Long: `A powerful CLI tool to export PostgreSQL query results to CSV, JSON, XML, SQL(insert) and other formats.
Supports direct SQL queries or SQL files, with customizable output options.`,
		Example: `  # Export with inline query
  pgxport -s "SELECT * FROM users" -o users.csv

  # Export from SQL file with custom delimiter
  pgxport -F query.sql -o output.csv -d ";"

  # Export to JSON
  pgxport -s "SELECT * FROM products" -o products.json -f json
  
  # Export to XML
  pgxport -s "SELECT * FROM orders" -o orders.xml -f xml

   # Export to SQL insert statements
  pgxport -s "SELECT * FROM orders" -o orders.sql -f sql -t orders_table`,
		Run: runExport,
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
	rootCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "CSV delimiter character")
	rootCmd.Flags().StringVarP(&connString, "dsn", "c", "", "Database connection string (postgres://user:pass@host:port/dbname)")
	rootCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table name for SQL insert exports")

	rootCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}

}

func runExport(cmd *cobra.Command, args []string) {

	if err := validateExportParams(); err != nil {
		log.Fatal(err)
	}

	var dbUrl string

	if connString != "" {
		dbUrl = connString
	} else {
		config := LoadConfig()

		if err := config.Validate(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		dbUrl = config.GetConnectionString()

	}

	var query string
	var err error

	if sqlFile != "" {
		query, err = readSQLFromFile(sqlFile)
		if err != nil {
			log.Fatalf("Error reading SQL file: %v\n", err)
		}
	} else {
		query = sqlQuery
	}

	store := NewStore()

	// connect to DB
	if err := store.Open(dbUrl); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	//Close connexion
	defer store.Close()

	log.Println("Executing query...")
	rows, err := store.ExecuteQuery(context.Background(), query)

	if err != nil {
		log.Fatalf("Query execution failed: %v", err)
	}

	defer rows.Close()

	format = strings.ToLower(strings.TrimSpace(format))

	exporter := exporters.NewExporter()

	options := exporters.ExportOptions{
		Format:    format,
		Delimiter: rune(delimiter[0]),
		TableName: tableName,
	}

	if err := exporter.Export(rows, outputPath, options); err != nil {
		log.Fatalf("Export failed: %v", err)
	}

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

	// Validate table name for SQL format
	if format == "sql" && strings.TrimSpace(tableName) == "" {
		return fmt.Errorf("Error: --table (-t) is required when using SQL format")
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
