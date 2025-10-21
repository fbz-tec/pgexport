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
)

func main() {

	var rootCmd = &cobra.Command{
		Use:   "pgexport",
		Short: "A tool to export database query results to CSV, JSON and other formats",
		Long:  `A CLI tool to export PostgreSQL query results to CSV, JSON and other formats.`,
		Run:   runExport,
	}

	rootCmd.Flags().StringVarP(&sqlQuery, "sql", "s", "", "SQL query to execute")
	rootCmd.Flags().StringVarP(&sqlFile, "sqlfile", "F", "", "Path to SQL file containing the query")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (required)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "csv", "Output format (csv, json)")
	rootCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ";", "CSV delimiter character")

	rootCmd.MarkFlagRequired("output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func runExport(cmd *cobra.Command, args []string) {

	if sqlQuery == "" && sqlFile == "" {
		log.Fatal("Error: Either --sql or --sqlfile must be provided")
	}

	if sqlQuery != "" && sqlFile != "" {
		log.Fatal("Error: Cannot use both --sql and --sqlfile at the same time")
	}

	config := LoadConfig()

	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
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

	store := &dbStore{}

	dbUrl := config.GetConnectionString()

	log.Println("Connecting to database...")
	// connect to DB
	if err := store.Open(dbUrl); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	//close connexion
	defer store.Close()

	format = strings.ToLower(format)
	switch format {

	case "csv":
		log.Printf("Exporting query results to CSV: %s\n", outputPath)
		if err := store.ExportQueryToCSV(query, outputPath, rune(delimiter[0])); err != nil {
			log.Fatalf("Export failed: %v", err)
		}

	case "json":
		log.Printf("Exporting query results to JSON: %s\n", outputPath)
		if err := store.ExportQueryToJSON(query, outputPath); err != nil {
			log.Fatalf("Export failed: %v", err)
		}

	default:
		log.Fatalf("Unsupported format: %s. Supported formats: csv, json", format)
	}
	log.Println("Export completed successfully!")
}

func readSQLFromFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %w", err)
	}
	return string(content), nil
}
