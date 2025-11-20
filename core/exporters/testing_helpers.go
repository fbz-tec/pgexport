package exporters

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

// setupTestDB creates a connection to the test database for integration tests.
// It returns the connection and a cleanup function.
// If DB_TEST_URL environment variable is not set, the test is skipped.
func setupTestDB(t *testing.T) (*pgx.Conn, func()) {
	t.Helper() // Mark this as a test helper

	// Check if test database URL is provided
	testURL := os.Getenv("DB_TEST_URL")
	if testURL == "" {
		t.Skip("Skipping integration test: DB_TEST_URL not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	cleanup := func() {
		conn.Close(ctx)
	}

	return conn, cleanup
}
