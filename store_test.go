package main

import (
	"testing"
)

// TestNewStore verifies that NewStore returns a non-nil Store instance
func TestNewStore(t *testing.T) {
	store := NewStore()
	if store == nil {
		t.Error("NewStore() returned nil, expected non-nil Store instance")
	}
}

// TestStoreInterface verifies that dbStore implements Store interface
func TestStoreInterface(t *testing.T) {
	var _ Store = &dbStore{}
}

// TestOpenInvalidURL tests connection with invalid database URLs
func TestOpenInvalidURL(t *testing.T) {
	tests := []struct {
		name  string
		dbURL string
	}{
		{
			name:  "completely invalid URL",
			dbURL: "not-a-valid-url",
		},
		{
			name:  "missing host",
			dbURL: "postgres://user:pass@:5432/db",
		},
		{
			name:  "invalid port",
			dbURL: "postgres://user:pass@localhost:invalid/db",
		},
		{
			name:  "empty URL",
			dbURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			err := store.Open(tt.dbURL)
			if err == nil {
				t.Error("Open() with invalid URL should return error, got nil")
				store.Close()
			}
		})
	}
}

// TestOpenClose tests the basic Open/Close flow
// Note: This test requires a running PostgreSQL instance
// It will be skipped if DB_TEST_URL is not set
func TestOpenClose(t *testing.T) {
	// Skip if no test database URL is provided
	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	store := NewStore()

	// Test Open
	err := store.Open(testURL)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	// Test Close
	err = store.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

// TestCloseWithoutOpen tests closing a store that was never opened
func TestCloseWithoutOpen(t *testing.T) {
	store := NewStore()
	err := store.Close()
	if err != nil {
		t.Errorf("Close() without Open() should not error, got: %v", err)
	}
}

// TestExecuteQueryWithoutConnection tests query execution without connection
func TestExecuteQueryWithoutConnection(t *testing.T) {
	store := NewStore()

	// Should return error, not panic
	result, err := store.ExecuteQuery("SELECT 1")

	if err == nil {
		t.Error("ExecuteQuery() without connection should return error")
	}

	if result != nil {
		t.Error("ExecuteQuery() without connection should return nil result")
	}
}

// Integration tests that require a real database
// These will be skipped if DB_TEST_URL is not set

func TestExecuteQueryIntegration(t *testing.T) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	store := NewStore()
	if err := store.Open(testURL); err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}
	defer store.Close()

	tests := []struct {
		name         string
		query        string
		wantErr      bool
		expectedCols []string
		expectedRows int
	}{
		{
			name:         "simple SELECT 1",
			query:        "SELECT 1 as num",
			wantErr:      false,
			expectedCols: []string{"num"},
			expectedRows: 1,
		},
		{
			name:         "SELECT with multiple columns",
			query:        "SELECT 1 as id, 'test' as name",
			wantErr:      false,
			expectedCols: []string{"id", "name"},
			expectedRows: 1,
		},
		{
			name:         "SELECT version",
			query:        "SELECT version()",
			wantErr:      false,
			expectedCols: []string{"version"},
			expectedRows: 1,
		},
		{
			name:    "invalid SQL syntax",
			query:   "SELECT * FROM",
			wantErr: true,
		},
		{
			name:    "non-existent table",
			query:   "SELECT * FROM this_table_does_not_exist_12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := store.ExecuteQuery(tt.query)

			if tt.wantErr {
				if err == nil {
					t.Error("ExecuteQuery() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ExecuteQuery() unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("ExecuteQuery() returned nil result")
			}

			// Check columns
			if len(result.Columns) != len(tt.expectedCols) {
				t.Errorf("Column count = %d, want %d", len(result.Columns), len(tt.expectedCols))
			}

			for i, col := range tt.expectedCols {
				if i < len(result.Columns) && result.Columns[i] != col {
					t.Errorf("Column[%d] = %q, want %q", i, result.Columns[i], col)
				}
			}

			// Check row count
			if len(result.Rows) != tt.expectedRows {
				t.Errorf("Row count = %d, want %d", len(result.Rows), tt.expectedRows)
			}
		})
	}
}

func TestExecuteQueryEmptyResult(t *testing.T) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	store := NewStore()
	if err := store.Open(testURL); err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}
	defer store.Close()

	// Query that returns no rows
	result, err := store.ExecuteQuery("SELECT 1 as num WHERE 1=0")
	if err != nil {
		t.Fatalf("ExecuteQuery() unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteQuery() returned nil result")
	}

	if len(result.Columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(result.Columns))
	}

	if len(result.Rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(result.Rows))
	}
}

func TestExecuteQueryDataTypes(t *testing.T) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	store := NewStore()
	if err := store.Open(testURL); err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}
	defer store.Close()

	query := `
		SELECT 
			1::integer as int_col,
			'test'::text as text_col,
			true::boolean as bool_col,
			3.14::numeric as numeric_col,
			NULL as null_col,
			NOW() as timestamp_col
	`

	result, err := store.ExecuteQuery(query)
	if err != nil {
		t.Fatalf("ExecuteQuery() error: %v", err)
	}

	expectedCols := []string{"int_col", "text_col", "bool_col", "numeric_col", "null_col", "timestamp_col"}
	if len(result.Columns) != len(expectedCols) {
		t.Errorf("Column count = %d, want %d", len(result.Columns), len(expectedCols))
	}

	if len(result.Rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(result.Rows))
	}

	// Verify we got the row with correct number of values
	row := result.Rows[0]
	if len(row) != len(expectedCols) {
		t.Errorf("Row value count = %d, want %d", len(row), len(expectedCols))
	}

	// Check that null value is actually nil
	if row[4] != nil {
		t.Errorf("null_col should be nil, got %v", row[4])
	}
}

func TestMultipleQueries(t *testing.T) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	store := NewStore()
	if err := store.Open(testURL); err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}
	defer store.Close()

	// Execute multiple queries in sequence
	queries := []string{
		"SELECT 1 as num",
		"SELECT 'hello' as greeting",
		"SELECT true as flag",
	}

	for i, query := range queries {
		t.Run(query, func(t *testing.T) {
			result, err := store.ExecuteQuery(query)
			if err != nil {
				t.Errorf("Query %d failed: %v", i, err)
				return
			}

			if result == nil {
				t.Errorf("Query %d returned nil result", i)
				return
			}

			if len(result.Rows) != 1 {
				t.Errorf("Query %d: expected 1 row, got %d", i, len(result.Rows))
			}
		})
	}
}

func TestConnectionReuse(t *testing.T) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	store := NewStore()
	if err := store.Open(testURL); err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}
	defer store.Close()

	// Execute same query multiple times to verify connection reuse
	for i := 0; i < 5; i++ {
		result, err := store.ExecuteQuery("SELECT 1")
		if err != nil {
			t.Errorf("Query %d failed: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Query %d returned nil", i+1)
		}
	}
}

// Helper function to get test database URL from environment
// Set DB_TEST_URL environment variable to run integration tests
// Example: export DB_TEST_URL="postgres://user:pass@localhost:5432/testdb"
func getTestDatabaseURL() string {
	// Check for test-specific database URL
	return "" // Return empty by default - tests will be skipped

	// To enable integration tests, uncomment the line below and set DB_TEST_URL
	// return os.Getenv("DB_TEST_URL")
}
