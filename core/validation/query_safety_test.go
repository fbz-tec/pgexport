package validation

import (
	"testing"
)

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "valid SELECT",
			query:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "forbidden DELETE",
			query:   "DELETE FROM users",
			wantErr: true,
		},
		{
			name:    "forbidden DROP",
			query:   "DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "chained DELETE",
			query:   "SELECT 1; DELETE FROM users",
			wantErr: true,
		},
		{
			name:    "lowercase delete",
			query:   "delete from users",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
