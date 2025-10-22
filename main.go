package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	sqlQuery   string
	sqlFile    string
	outputPath string
	format     string
	delimiter  string
	connString string
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime)

	var rootCmd = &cobra.Command{
		Use:   "pgxport",
		Short: "Export PostgreSQL query results to various formats",
		Long: `A powerful CLI tool to export PostgreSQL query results to CSV, JSON, XML and other formats.
Supports direct SQL queries or SQL files, with customizable output options.`,
		Example: `  # Export with inline query
  pgxport -s "SELECT * FROM users" -o users.csv

  # Export from SQL file with custom delimiter
  pgxport -F query.sql -o output.csv -d ","

  # Export to JSON
  pgxport -s "SELECT * FROM products" -o products.json -f json
  
  # Export to XML
  pgxport -s "SELECT * FROM orders" -o orders.xml -f xml`,
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
	rootCmd.Flags().StringVarP(&format, "format", "f", "csv", "Output format (csv, json)")
	rootCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ";", "CSV delimiter character")
	rootCmd.Flags().StringVarP(&connString, "dsn", "c", "", "Database connection string (postgres://user:pass@host:port/dbname)")

	rootCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}

}

func runExport(cmd *cobra.Command, args []string) {

	if sqlQuery == "" && sqlFile == "" {
		log.Fatal("Error: Either --sql or --sqlfile must be provided")
	}

	if sqlQuery != "" && sqlFile != "" {
		log.Fatal("Error: Cannot use both --sql and --sqlfile at the same time")
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

	//close connexion
	defer store.Close()

	log.Println("Executing query...")
	result, err := store.ExecuteQuery(query)

	if err != nil {
		log.Fatalf("Query execution failed: %v", err)
	}

	format = strings.ToLower(strings.TrimSpace(format))

	exporter := NewExporter()

	options := ExportOptions{
		Format:    format,
		Delimiter: rune(delimiter[0]),
	}

	if err := exporter.Export(result, outputPath, options); err != nil {
		log.Fatalf("Export failed: %v", err)
	}

}

func readSQLFromFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %w", err)
	}
	return string(content), nil
}
