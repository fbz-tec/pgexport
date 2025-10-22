package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSQLFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		expected    string
	}{
		{
			name:        "valid SQL file",
			fileContent: "SELECT * FROM users WHERE id = 1;",
			wantErr:     false,
			expected:    "SELECT * FROM users WHERE id = 1;",
		},
		{
			name:        "empty file",
			fileContent: "",
			wantErr:     false,
			expected:    "",
		},
		{
			name: "multiline SQL",
			fileContent: `SELECT id, name, email
FROM users
WHERE active = true
ORDER BY name;`,
			wantErr: false,
			expected: `SELECT id, name, email
FROM users
WHERE active = true
ORDER BY name;`,
		},
		{
			name: "SQL with comments",
			fileContent: `-- Get all active users
SELECT * FROM users
WHERE active = true; -- filter by status`,
			wantErr: false,
			expected: `-- Get all active users
SELECT * FROM users
WHERE active = true; -- filter by status`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file in OS temp directory
			tmpFile, err := os.CreateTemp("", "test_*.sql")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath)

			// Write content and close
			if _, err := tmpFile.Write([]byte(tt.fileContent)); err != nil {
				tmpFile.Close()
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Test reading the file
			result, err := readSQLFromFile(tmpPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("readSQLFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("readSQLFromFile() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReadSQLFromFileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "non-existent file",
			filepath: "/tmp/this_file_does_not_exist_12345.sql",
		},
		{
			name:     "directory instead of file",
			filepath: "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readSQLFromFile(tt.filepath)
			if err == nil {
				t.Error("readSQLFromFile() expected error for invalid path, got nil")
			}
		})
	}
}

func TestReadSQLFromFilePermissions(t *testing.T) {
	// Create a file with no read permissions
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "noperm.sql")

	err := os.WriteFile(tmpFile, []byte("SELECT 1;"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change permissions to no-read (write-only)
	err = os.Chmod(tmpFile, 0200)
	if err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}

	// Restore permissions in cleanup
	defer os.Chmod(tmpFile, 0644)

	_, err = readSQLFromFile(tmpFile)
	if err == nil {
		t.Error("readSQLFromFile() should fail with no-read permissions")
	}
}

func TestReadSQLFromFileLargeFile(t *testing.T) {
	// Test reading a larger SQL file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "large.sql")

	// Create a large SQL content
	var content string
	for i := 0; i < 1000; i++ {
		content += "SELECT * FROM users WHERE id = " + string(rune(i)) + ";\n"
	}

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := readSQLFromFile(tmpFile)
	if err != nil {
		t.Errorf("readSQLFromFile() unexpected error: %v", err)
	}

	if result != content {
		t.Error("readSQLFromFile() did not return expected content for large file")
	}
}

func TestReadSQLFromFileUTF8(t *testing.T) {
	// Test reading file with UTF-8 characters
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "utf8.sql")

	content := "SELECT * FROM users WHERE name = 'José García';"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := readSQLFromFile(tmpFile)
	if err != nil {
		t.Errorf("readSQLFromFile() unexpected error: %v", err)
	}

	if result != content {
		t.Errorf("readSQLFromFile() = %q, want %q", result, content)
	}
}

// Benchmark tests
func BenchmarkReadSQLFromFile(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "bench.sql")

	content := "SELECT * FROM users WHERE id = 1;"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = readSQLFromFile(tmpFile)
	}
}

func BenchmarkReadSQLFromFileLarge(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "bench_large.sql")

	// Create a 10KB SQL file
	var content string
	for i := 0; i < 100; i++ {
		content += "SELECT * FROM users WHERE id = 1 AND name LIKE '%test%' ORDER BY created_at DESC LIMIT 100;\n"
	}

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = readSQLFromFile(tmpFile)
	}
}
