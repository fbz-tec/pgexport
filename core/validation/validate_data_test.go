package validation

import (
	"testing"
)

func TestValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "valid ISO format",
			format:  "yyyy-MM-dd HH:mm:ss",
			wantErr: false,
		},
		{
			name:    "valid ISO with milliseconds",
			format:  "yyyy-MM-ddTHH:mm:ss.SSS",
			wantErr: false,
		},
		{
			name:    "valid European format",
			format:  "dd/MM/yyyy HH:mm:ss",
			wantErr: false,
		},
		{
			name:    "valid US format",
			format:  "MM/dd/yyyy HH:mm:ss",
			wantErr: false,
		},
		{
			name:    "valid date only",
			format:  "yyyy-MM-dd",
			wantErr: false,
		},
		{
			name:    "valid time only",
			format:  "HH:mm:ss",
			wantErr: false,
		},
		{
			name:    "invalid format with wrong pattern",
			format:  "yyyy-MM-dd HH:mm:ss:invalid",
			wantErr: false,
		},
		{
			name:    "empty format",
			format:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeFormat(%q) error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimeZone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{
			name:     "valid UTC",
			timezone: "UTC",
			wantErr:  false,
		},
		{
			name:     "valid America/New_York",
			timezone: "America/New_York",
			wantErr:  false,
		},
		{
			name:     "valid Europe/Paris",
			timezone: "Europe/Paris",
			wantErr:  false,
		},
		{
			name:     "valid Asia/Tokyo",
			timezone: "Asia/Tokyo",
			wantErr:  false,
		},
		{
			name:     "empty timezone is valid",
			timezone: "",
			wantErr:  false,
		},
		{
			name:     "invalid timezone",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
		{
			name:     "gibberish timezone",
			timezone: "xyzabc123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeZone(tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeZone(%q) error = %v, wantErr %v", tt.timezone, err, tt.wantErr)
			}
		})
	}
}
