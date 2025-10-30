package exporters

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestWriteCSV(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		query       string
		delimiter   rune
		compression string
		wantErr     bool
		checkFunc   func(t *testing.T, path string)
	}{
		{
			name:        "basic CSV export",
			query:       "SELECT 1 as id, 'test' as name, true as active",
			delimiter:   ',',
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				if len(lines) != 2 { // header + 1 data row
					t.Errorf("Expected 2 lines, got %d", len(lines))
				}

				// Check header
				if !strings.Contains(lines[0], "id") || !strings.Contains(lines[0], "name") {
					t.Errorf("Header missing expected columns: %s", lines[0])
				}
			},
		},
		{
			name:        "CSV with custom delimiter",
			query:       "SELECT 1 as id, 'test' as name",
			delimiter:   ';',
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				if !strings.Contains(lines[0], ";") {
					t.Errorf("Expected semicolon delimiter, got: %s", lines[0])
				}
			},
		},
		{
			name:        "CSV with NULL values",
			query:       "SELECT 1 as id, NULL as description, 'test' as name",
			delimiter:   ',',
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				// NULL should be exported as empty string in CSV
				if !strings.Contains(string(content), ",,") && !strings.Contains(string(content), ",\"\",") {
					t.Logf("Content: %s", string(content))
				}
			},
		},
		{
			name:        "CSV with special characters",
			query:       "SELECT 'O''Brien' as name, 'Line1\nLine2' as address",
			delimiter:   ',',
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				f, err := os.Open(path)
				if err != nil {
					t.Fatalf("Failed to open file: %v", err)
				}
				defer f.Close()

				reader := csv.NewReader(f)
				records, err := reader.ReadAll()
				if err != nil {
					t.Fatalf("Failed to parse CSV: %v", err)
				}

				if len(records) != 2 { // header + 1 row
					t.Errorf("Expected 2 records, got %d", len(records))
				}
			},
		},
		{
			name:        "empty result set",
			query:       "SELECT 1 as id WHERE 1=0",
			delimiter:   ',',
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				if len(lines) != 1 { // only header
					t.Errorf("Expected only header line, got %d lines", len(lines))
				}
			},
		},
		{
			name:        "CSV with gzip compression",
			query:       "SELECT 1 as id, 'test' as name",
			delimiter:   ',',
			compression: "gzip",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				if !strings.HasSuffix(path, ".gz") {
					t.Errorf("Expected .gz extension, got: %s", path)
				}

				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("Failed to stat file: %v", err)
				}
				if info.Size() == 0 {
					t.Error("Compressed file is empty")
				}
			},
		},
		{
			name:        "CSV with multiple rows",
			query:       "SELECT generate_series(1, 100) as id, 'test' as name",
			delimiter:   ',',
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				f, err := os.Open(path)
				if err != nil {
					t.Fatalf("Failed to open file: %v", err)
				}
				defer f.Close()

				reader := csv.NewReader(f)
				records, err := reader.ReadAll()
				if err != nil {
					t.Fatalf("Failed to parse CSV: %v", err)
				}

				if len(records) != 101 { // header + 100 rows
					t.Errorf("Expected 101 records, got %d", len(records))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.csv")

			ctx := context.Background()
			rows, err := conn.Query(ctx, tt.query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}
			defer rows.Close()

			exporter := &dataExporter{}
			options := ExportOptions{
				Format:      FormatCSV,
				Delimiter:   tt.delimiter,
				Compression: tt.compression,
				TimeFormat:  "yyyy-MM-dd HH:mm:ss",
				TimeZone:    "",
			}

			_, err = exporter.writeCSV(rows, outputPath, options)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				finalPath := outputPath
				if tt.compression == "gzip" && !strings.HasSuffix(outputPath, ".gz") {
					finalPath = outputPath + ".gz"
				}
				tt.checkFunc(t, finalPath)
			}
		})
	}
}

func TestWriteCSVTimeFormatting(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name       string
		timeFormat string
		timeZone   string
		checkFunc  func(t *testing.T, content string)
	}{
		{
			name:       "default time format",
			timeFormat: "yyyy-MM-dd HH:mm:ss",
			timeZone:   "",
			checkFunc: func(t *testing.T, content string) {
				if !strings.Contains(content, "-") || !strings.Contains(content, ":") {
					t.Errorf("Expected date-time format, got: %s", content)
				}
			},
		},
		{
			name:       "custom time format",
			timeFormat: "dd/MM/yyyy",
			timeZone:   "",
			checkFunc: func(t *testing.T, content string) {
				if !strings.Contains(content, "/") {
					t.Errorf("Expected custom date format with /, got: %s", content)
				}
			},
		},
		{
			name:       "UTC timezone",
			timeFormat: "yyyy-MM-dd HH:mm:ss",
			timeZone:   "UTC",
			checkFunc: func(t *testing.T, content string) {
				// Just verify it doesn't crash and produces output
				if len(content) == 0 {
					t.Error("Expected non-empty output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.csv")

			query := "SELECT NOW() as created_at"
			ctx := context.Background()
			rows, err := conn.Query(ctx, query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}
			defer rows.Close()

			exporter := &dataExporter{}
			options := ExportOptions{
				Format:      FormatCSV,
				Delimiter:   ',',
				Compression: "none",
				TimeFormat:  tt.timeFormat,
				TimeZone:    tt.timeZone,
			}

			_, err = exporter.writeCSV(rows, outputPath, options)
			if err != nil {
				t.Fatalf("writeCSV() error: %v", err)
			}

			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			tt.checkFunc(t, string(content))
		})
	}
}

func TestWriteCSVDataTypes(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	query := `
		SELECT 
			1::integer as int_col,
			3.14::numeric as numeric_col,
			'text value'::text as text_col,
			true::boolean as bool_col,
			NULL as null_col,
			NOW() as timestamp_col,
			'2024-01-15'::date as date_col
	`

	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatCSV,
		Delimiter:   ',',
		Compression: "none",
		TimeFormat:  "yyyy-MM-dd HH:mm:ss",
		TimeZone:    "",
	}

	rowCount, err := exporter.writeCSV(rows, outputPath, options)
	if err != nil {
		t.Fatalf("writeCSV() error: %v", err)
	}

	if rowCount != 1 {
		t.Errorf("Expected 1 row, got %d", rowCount)
	}

	// Verify the file can be parsed
	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records (header + 1 row), got %d", len(records))
	}

	if len(records[0]) != 7 {
		t.Errorf("Expected 7 columns, got %d", len(records[0]))
	}
}

func TestWriteCopyCSV(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name      string
		query     string
		delimiter rune
		wantErr   bool
		minRows   int
		checkFunc func(t *testing.T, path string)
	}{
		{
			name:      "basic COPY export",
			query:     "SELECT 1 as id, 'test' as name",
			delimiter: ',',
			wantErr:   false,
			minRows:   1,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output: %v", err)
				}

				if !strings.Contains(string(content), "id") {
					t.Error("Expected header with 'id' column")
				}
			},
		},
		{
			name:      "COPY with multiple rows",
			query:     "SELECT generate_series(1, 50) as id",
			delimiter: ',',
			wantErr:   false,
			minRows:   50,
			checkFunc: func(t *testing.T, path string) {
				f, err := os.Open(path)
				if err != nil {
					t.Fatalf("Failed to open file: %v", err)
				}
				defer f.Close()

				reader := csv.NewReader(f)
				records, err := reader.ReadAll()
				if err != nil {
					t.Fatalf("Failed to parse CSV: %v", err)
				}

				if len(records) != 51 { // header + 50 rows
					t.Errorf("Expected 51 records, got %d", len(records))
				}
			},
		},
		{
			name:      "COPY with custom delimiter",
			query:     "SELECT 1 as id, 'test' as name",
			delimiter: ';',
			wantErr:   false,
			minRows:   1,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output: %v", err)
				}

				if !strings.Contains(string(content), ";") {
					t.Error("Expected semicolon delimiter")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.csv")

			exporter := &dataExporter{}
			options := ExportOptions{
				Format:      FormatCSV,
				Delimiter:   tt.delimiter,
				Compression: "none",
			}

			rowCount, err := exporter.writeCopyCSV(conn, tt.query, outputPath, options)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeCopyCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if rowCount < tt.minRows {
					t.Errorf("Expected at least %d rows, got %d", tt.minRows, rowCount)
				}

				if tt.checkFunc != nil {
					tt.checkFunc(t, outputPath)
				}
			}
		})
	}
}

func TestWriteCSVLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "large.csv")

	// Generate 10,000 rows
	query := "SELECT generate_series(1, 10000) as id, md5(random()::text) as data"

	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatCSV,
		Delimiter:   ',',
		Compression: "none",
		TimeFormat:  "yyyy-MM-dd HH:mm:ss",
		TimeZone:    "",
	}

	start := time.Now()
	rowCount, err := exporter.writeCSV(rows, outputPath, options)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("writeCSV() error: %v", err)
	}

	if rowCount != 10000 {
		t.Errorf("Expected 10000 rows, got %d", rowCount)
	}

	t.Logf("Exported 10,000 rows in %v", duration)

	// Verify file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}
}

func BenchmarkWriteCSV(b *testing.B) {
	testURL := os.Getenv("DB_TEST_URL")
	if testURL == "" {
		b.Skip("Skipping benchmark: DB_TEST_URL not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testURL)
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	tmpDir := b.TempDir()
	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatCSV,
		Delimiter:   ',',
		Compression: "none",
		TimeFormat:  "yyyy-MM-dd HH:mm:ss",
		TimeZone:    "",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "bench.csv")
		query := "SELECT generate_series(1, 1000) as id, md5(random()::text) as data"
		rows, err := conn.Query(ctx, query)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}

		_, err = exporter.writeCSV(rows, outputPath, options)
		if err != nil {
			b.Fatalf("writeCSV failed: %v", err)
		}
		rows.Close()
		os.Remove(outputPath)
	}
}
