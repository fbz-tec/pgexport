package validation

import (
	"fmt"
	"strings"
)

// ValidateQuery checks if the query is safe for export (read-only)
func ValidateQuery(query string) error {
	// Normalize query to uppercase for checking
	normalized := strings.ToUpper(strings.TrimSpace(query))

	// List of forbidden SQL commands
	forbiddenCommands := []string{
		"DELETE",
		"DROP",
		"TRUNCATE",
		"INSERT",
		"UPDATE",
		"ALTER",
		"CREATE",
		"GRANT",
		"REVOKE",
	}

	// Check if query starts with forbidden command
	for _, cmd := range forbiddenCommands {
		if strings.HasPrefix(normalized, cmd) {
			return fmt.Errorf("forbidden SQL command detected: %s (read-only mode)", cmd)
		}
	}

	// Additional check: detect forbidden keywords anywhere in query
	for _, cmd := range forbiddenCommands {
		if strings.Contains(normalized, cmd+" ") || strings.Contains(normalized, cmd+";") {
			return fmt.Errorf("forbidden SQL command detected in query: %s", cmd)
		}
	}

	return nil
}
