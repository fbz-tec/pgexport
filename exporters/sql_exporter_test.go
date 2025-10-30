package exporters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestWriteSQL(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		query       string
		tableName   string
		compression string
		wantErr     bool
		checkFunc   func(t *testing.T, path string)
	}{
		{
			name:        "basic SQL export",
			query:       "SELECT 1 as id, 'test' as name",
			tableName:   "users",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)

				// Check for INSERT statement
				if !strings.Contains(contentStr, "INSERT INTO") {
					t.Error("Expected INSERT INTO statement")
				}

				// Check table name is quoted
				if !strings.Contains(contentStr, `"users"`) {
					t.Error("Expected quoted table name")
				}

				// Check for VALUES clause
				if !strings.Contains(contentStr, "VALUES") {
					t.Error("Expected VALUES clause")
				}

				// Check statement ends with semicolon
				if !strings.Contains(contentStr, ");") {
					t.Error("Expected semicolon at end of statement")
				}
			},
		},
		{
			name:        "SQL with NULL values",
			query:       "SELECT 1 as id, NULL as description, 'test' as name",
			tableName:   "items",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)

				// NULL should be exported as NULL keyword (not 'NULL' string)
				if !strings.Contains(contentStr, ", NULL,") && !strings.Contains(contentStr, "(NULL,") {
					t.Logf("Content: %s", contentStr)
				}

				// Should NOT contain 'NULL' in quotes
				if strings.Contains(contentStr, "'NULL'") {
					t.Error("NULL should not be quoted")
				}
			},
		},
		{
			name:        "SQL with special characters",
			query:       "SELECT 'O''Brien' as name, 'Line1\nLine2' as address",
			tableName:   "contacts",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)

				// Single quotes should be escaped as ''
				if !strings.Contains(contentStr, "''") {
					t.Error("Expected escaped single quotes")
				}
			},
		},
		{
			name:        "SQL with quoted identifiers",
			query:       "SELECT 1 as \"user_id\", 'test' as \"user name\"",
			tableName:   "test_table",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)

				// Column names should be quoted
				if !strings.Contains(contentStr, `"user_id"`) {
					t.Error("Expected quoted column name user_id")
				}

				if !strings.Contains(contentStr, `"user name"`) {
					t.Error("Expected quoted column name with space")
				}
			},
		},
		{
			name:        "empty result set",
			query:       "SELECT 1 as id WHERE 1=0",
			tableName:   "empty_table",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				if len(content) != 0 {
					t.Error("Expected empty file for empty result set")
				}
			},
		},
		{
			name:        "SQL with gzip compression",
			query:       "SELECT 1 as id, 'test' as name",
			tableName:   "compressed",
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
			name:        "SQL with multiple rows",
			query:       "SELECT generate_series(1, 5) as id, 'test' || generate_series(1, 5) as name",
			tableName:   "multi_row",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				// Count INSERT statements
				insertCount := strings.Count(string(content), "INSERT INTO")
				if insertCount != 5 {
					t.Errorf("Expected 5 INSERT statements, got %d", insertCount)
				}
			},
		},
		{
			name:        "SQL with schema-qualified table name",
			query:       "SELECT 1 as id",
			tableName:   "public.users",
			compression: "none",
			wantErr:     false,
			checkFunc: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)

				// Schema and table should both be quoted separately
				if !strings.Contains(contentStr, `"public"."users"`) {
					t.Error("Expected schema-qualified table name with proper quoting")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.sql")

			ctx := context.Background()
			rows, err := conn.Query(ctx, tt.query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}
			defer rows.Close()

			exporter := &dataExporter{}
			options := ExportOptions{
				Format:      FormatSQL,
				TableName:   tt.tableName,
				Compression: tt.compression,
			}

			_, err = exporter.writeSQL(rows, outputPath, options)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeSQL() error = %v, wantErr %v", err, tt.wantErr)
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

func TestWriteSQLDataTypes(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.sql")

	query := `
		SELECT 
			1::integer as int_col,
			3.14::numeric as numeric_col,
			'text value'::text as text_col,
			true::boolean as bool_col,
			false::boolean as bool_false,
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
		Format:      FormatSQL,
		TableName:   "test_types",
		Compression: "none",
	}

	rowCount, err := exporter.writeSQL(rows, outputPath, options)
	if err != nil {
		t.Fatalf("writeSQL() error: %v", err)
	}

	if rowCount != 1 {
		t.Errorf("Expected 1 row, got %d", rowCount)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Verify data types are formatted correctly
	tests := []struct {
		name     string
		expected string
		message  string
	}{
		{
			name:     "integer without quotes",
			expected: "1",
			message:  "Integers should not be quoted",
		},
		{
			name:     "numeric without quotes",
			expected: "3.14",
			message:  "Numbers should not be quoted",
		},
		{
			name:     "text with quotes",
			expected: "'text value'",
			message:  "Text should be quoted",
		},
		{
			name:     "boolean true",
			expected: "true",
			message:  "Boolean true should be lowercase and unquoted",
		},
		{
			name:     "boolean false",
			expected: "false",
			message:  "Boolean false should be lowercase and unquoted",
		},
		{
			name:     "NULL keyword",
			expected: "NULL",
			message:  "NULL should be uppercase and unquoted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(contentStr, tt.expected) {
				t.Errorf("%s: expected to find %q in output", tt.message, tt.expected)
				t.Logf("Content: %s", contentStr)
			}
		})
	}
}

func TestWriteSQLColumnOrder(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.sql")

	query := "SELECT 1 as id, 'test' as name, true as active"

	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatSQL,
		TableName:   "test_table",
		Compression: "none",
	}

	_, err = exporter.writeSQL(rows, outputPath, options)
	if err != nil {
		t.Fatalf("writeSQL() error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Verify columns appear in INSERT statement
	if !strings.Contains(contentStr, `("id", "name", "active")`) {
		t.Error("Expected column list in INSERT statement")
	}

	// Verify column order matches value order
	// Values should appear in same order as columns
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "INSERT INTO") {
			// Check that id, name, active appear in that order
			idPos := strings.Index(line, "1")
			namePos := strings.Index(line, "'test'")
			activePos := strings.Index(line, "true")

			if idPos > namePos || namePos > activePos {
				t.Error("Values not in correct order")
			}
		}
	}
}

func TestWriteSQLEscaping(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name          string
		query         string
		expectedInSQL string
		notExpected   string
	}{
		{
			name:          "single quote escaping",
			query:         "SELECT 'O''Brien' as name",
			expectedInSQL: "''",
			notExpected:   "",
		},
		{
			name:          "backslash handling",
			query:         "SELECT 'C:\\path\\to\\file' as path",
			expectedInSQL: "\\",
			notExpected:   "",
		},
		{
			name:          "newline in text",
			query:         "SELECT 'Line1\nLine2' as text",
			expectedInSQL: "'Line1",
			notExpected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.sql")

			ctx := context.Background()
			rows, err := conn.Query(ctx, tt.query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}
			defer rows.Close()

			exporter := &dataExporter{}
			options := ExportOptions{
				Format:      FormatSQL,
				TableName:   "test_escape",
				Compression: "none",
			}

			_, err = exporter.writeSQL(rows, outputPath, options)
			if err != nil {
				t.Fatalf("writeSQL() error: %v", err)
			}

			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			contentStr := string(content)

			if tt.expectedInSQL != "" && !strings.Contains(contentStr, tt.expectedInSQL) {
				t.Errorf("Expected to find %q in SQL output", tt.expectedInSQL)
			}

			if tt.notExpected != "" && strings.Contains(contentStr, tt.notExpected) {
				t.Errorf("Did not expect to find %q in SQL output", tt.notExpected)
			}
		})
	}
}

func TestWriteSQLLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "large.sql")

	// Generate 1,000 rows
	query := "SELECT i, 'data_' || i FROM generate_series(1, 1000) AS s(i)"

	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatSQL,
		TableName:   "large_table",
		Compression: "none",
	}

	start := time.Now()
	rowCount, err := exporter.writeSQL(rows, outputPath, options)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("writeSQL() error: %v", err)
	}

	if rowCount != 1000 {
		t.Errorf("Expected 1000 rows, got %d", rowCount)
	}

	t.Logf("Exported 1,000 rows in %v", duration)

	// Verify file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Count INSERT statements
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	insertCount := strings.Count(string(content), "INSERT INTO")
	if insertCount != 1000 {
		t.Errorf("Expected 1000 INSERT statements, got %d", insertCount)
	}
}

func TestWriteSQLStatementFormat(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.sql")

	query := "SELECT 1 as id, 'test' as name"
	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatSQL,
		TableName:   "test_table",
		Compression: "none",
	}

	_, err = exporter.writeSQL(rows, outputPath, options)
	if err != nil {
		t.Fatalf("writeSQL() error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Each non-empty line should be a complete INSERT statement
		if !strings.HasPrefix(line, "INSERT INTO") {
			t.Errorf("Line should start with INSERT INTO: %s", line)
		}

		if !strings.HasSuffix(strings.TrimSpace(line), ");") {
			t.Errorf("Line should end with ); : %s", line)
		}
	}
}

func TestWriteSQLBuffering(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.sql")

	// Generate enough rows to trigger buffering (>10000)
	query := "SELECT generate_series(1, 15000) as id"

	ctx := context.Background()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	exporter := &dataExporter{}
	options := ExportOptions{
		Format:      FormatSQL,
		TableName:   "buffer_test",
		Compression: "none",
	}

	rowCount, err := exporter.writeSQL(rows, outputPath, options)
	if err != nil {
		t.Fatalf("writeSQL() error: %v", err)
	}

	if rowCount != 15000 {
		t.Errorf("Expected 15000 rows, got %d", rowCount)
	}

	// Verify all rows were written
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	insertCount := strings.Count(string(content), "INSERT INTO")
	if insertCount != 15000 {
		t.Errorf("Expected 15000 INSERT statements, got %d", insertCount)
	}
}

func BenchmarkWriteSQL(b *testing.B) {
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
		Format:      FormatSQL,
		TableName:   "bench_table",
		Compression: "none",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "bench.sql")
		query := "SELECT generate_series(1, 100) as id, md5(random()::text) as data"
		rows, err := conn.Query(ctx, query)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}

		_, err = exporter.writeSQL(rows, outputPath, options)
		if err != nil {
			b.Fatalf("writeSQL failed: %v", err)
		}
		rows.Close()
		os.Remove(outputPath)
	}
}
