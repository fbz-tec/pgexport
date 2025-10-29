package exporters

import (
	"testing"
	"time"
)

func TestConvertUserTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ISO 8601 format",
			input:    "yyyy-MM-dd HH:mm:ss",
			expected: "2006-01-02 15:04:05",
		},
		{
			name:     "ISO 8601 with milliseconds",
			input:    "yyyy-MM-ddTHH:mm:ss.SSS",
			expected: "2006-01-02T15:04:05.000",
		},
		{
			name:     "European format",
			input:    "dd/MM/yyyy HH:mm:ss",
			expected: "02/01/2006 15:04:05",
		},
		{
			name:     "US format",
			input:    "MM/dd/yyyy HH:mm:ss",
			expected: "01/02/2006 15:04:05",
		},
		{
			name:     "Short year format",
			input:    "yy-MM-dd",
			expected: "06-01-02",
		},
		{
			name:     "With deciseconds",
			input:    "yyyy-MM-dd HH:mm:ss.S",
			expected: "2006-01-02 15:04:05.0",
		},
		{
			name:     "Date only",
			input:    "yyyy-MM-dd",
			expected: "2006-01-02",
		},
		{
			name:     "Time only",
			input:    "HH:mm:ss",
			expected: "15:04:05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertUserTimeFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertUserTimeFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 123000000, time.UTC)
	layout := "2006-01-02 15:04:05"
	loc := time.UTC

	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: nil,
		},
		{
			name:     "time.Time value",
			value:    testTime,
			expected: "2024-03-15 14:30:45",
		},
		{
			name:     "[]byte value",
			value:    []byte("test data"),
			expected: "test data",
		},
		{
			name:     "float32 value",
			value:    float32(3.14159),
			expected: "3.1415901184082",
		},
		{
			name:     "float64 value",
			value:    float64(2.718281828459045),
			expected: "2.71828182845905",
		},
		{
			name:     "string value",
			value:    "test string",
			expected: "test string",
		},
		{
			name:     "int value",
			value:    42,
			expected: 42,
		},
		{
			name:     "bool value",
			value:    true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value, layout, loc)

			// For floats, we need to compare strings since the formatting might differ slightly
			if _, ok := tt.value.(float32); ok {
				if result.(string) != tt.expected.(string) {
					t.Errorf("formatValue(%v) = %v, want %v", tt.value, result, tt.expected)
				}
			} else if _, ok := tt.value.(float64); ok {
				if result.(string) != tt.expected.(string) {
					t.Errorf("formatValue(%v) = %v, want %v", tt.value, result, tt.expected)
				}
			} else if result != tt.expected {
				t.Errorf("formatValue(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormatValueWithDifferentTimezones(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	layout := "2006-01-02 15:04:05"

	tests := []struct {
		name     string
		timezone string
		expected string
	}{
		{
			name:     "UTC timezone",
			timezone: "UTC",
			expected: "2024-03-15 14:30:45",
		},
		{
			name:     "America/New_York timezone",
			timezone: "America/New_York",
			expected: "2024-03-15 10:30:45", // UTC-4 in March (EDT)
		},
		{
			name:     "Europe/Paris timezone",
			timezone: "Europe/Paris",
			expected: "2024-03-15 15:30:45", // UTC+1 in March (CET)
		},
		{
			name:     "Asia/Tokyo timezone",
			timezone: "Asia/Tokyo",
			expected: "2024-03-15 23:30:45", // UTC+9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Skipf("Timezone %s not available: %v", tt.timezone, err)
			}

			result := formatValue(testTime, layout, loc)
			if result != tt.expected {
				t.Errorf("formatValue with timezone %s = %v, want %v", tt.timezone, result, tt.expected)
			}
		})
	}
}

func TestFormatSQLValue(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 123000000, time.UTC)

	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: "NULL",
		},
		{
			name:     "string value",
			value:    "test string",
			expected: "'test string'",
		},
		{
			name:     "string with single quote",
			value:    "O'Brien",
			expected: "'O''Brien'",
		},
		{
			name:     "string with multiple quotes",
			value:    "It's a 'test'",
			expected: "'It''s a ''test'''",
		},
		{
			name:     "[]byte value",
			value:    []byte("binary data"),
			expected: "'binary data'",
		},
		{
			name:     "[]byte with quotes",
			value:    []byte("O'Connor"),
			expected: "'O''Connor'",
		},
		{
			name:     "time.Time value",
			value:    testTime,
			expected: "'2024-03-15T14:30:45.123'",
		},
		{
			name:     "bool true",
			value:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			value:    false,
			expected: "false",
		},
		{
			name:     "int value",
			value:    int(42),
			expected: "42",
		},
		{
			name:     "int8 value",
			value:    int8(127),
			expected: "127",
		},
		{
			name:     "int16 value",
			value:    int16(32767),
			expected: "32767",
		},
		{
			name:     "int32 value",
			value:    int32(2147483647),
			expected: "2147483647",
		},
		{
			name:     "int64 value",
			value:    int64(9223372036854775807),
			expected: "9223372036854775807",
		},
		{
			name:     "uint value",
			value:    uint(42),
			expected: "42",
		},
		{
			name:     "uint8 value",
			value:    uint8(255),
			expected: "255",
		},
		{
			name:     "uint16 value",
			value:    uint16(65535),
			expected: "65535",
		},
		{
			name:     "uint32 value",
			value:    uint32(4294967295),
			expected: "4294967295",
		},
		{
			name:     "uint64 value",
			value:    uint64(18446744073709551615),
			expected: "18446744073709551615",
		},
		{
			name:     "float32 value",
			value:    float32(3.14159),
			expected: "3.1415901184082",
		},
		{
			name:     "float64 value",
			value:    float64(2.718281828459045),
			expected: "2.71828182845905",
		},
		{
			name:     "negative int",
			value:    int(-42),
			expected: "-42",
		},
		{
			name:     "negative float",
			value:    float64(-3.14),
			expected: "-3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSQLValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatSQLValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestQuoteIdent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple identifier",
			input:    "users",
			expected: `"users"`,
		},
		{
			name:     "schema.table",
			input:    "public.users",
			expected: `"public"."users"`,
		},
		{
			name:     "schema.table.column",
			input:    "public.users.id",
			expected: `"public"."users"."id"`,
		},
		{
			name:     "identifier with quotes",
			input:    `table"name`,
			expected: `"table""name"`,
		},
		{
			name:     "schema with quotes",
			input:    `schema"1.table"2`,
			expected: `"schema""1"."table""2"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "special characters",
			input:    "table-name",
			expected: `"table-name"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdent(tt.input)
			if result != tt.expected {
				t.Errorf("quoteIdent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUserTimeZoneFormat(t *testing.T) {
	tests := []struct {
		name            string
		userTimefmt     string
		timeZone        string
		expectedLayout  string
		expectedLocName string
	}{
		{
			name:            "ISO format with UTC",
			userTimefmt:     "yyyy-MM-dd HH:mm:ss",
			timeZone:        "UTC",
			expectedLayout:  "2006-01-02 15:04:05",
			expectedLocName: "UTC",
		},
		{
			name:            "European format with Paris timezone",
			userTimefmt:     "dd/MM/yyyy HH:mm:ss",
			timeZone:        "Europe/Paris",
			expectedLayout:  "02/01/2006 15:04:05",
			expectedLocName: "Europe/Paris",
		},
		{
			name:            "Empty timezone defaults to Local",
			userTimefmt:     "yyyy-MM-dd",
			timeZone:        "",
			expectedLayout:  "2006-01-02",
			expectedLocName: "Local",
		},
		{
			name:            "Invalid timezone defaults to Local",
			userTimefmt:     "yyyy-MM-dd HH:mm:ss",
			timeZone:        "Invalid/Timezone",
			expectedLayout:  "2006-01-02 15:04:05",
			expectedLocName: "Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout, loc := userTimeZoneFormat(tt.userTimefmt, tt.timeZone)

			if layout != tt.expectedLayout {
				t.Errorf("userTimeZoneFormat() layout = %q, want %q", layout, tt.expectedLayout)
			}

			if loc.String() != tt.expectedLocName {
				t.Errorf("userTimeZoneFormat() location = %q, want %q", loc.String(), tt.expectedLocName)
			}
		})
	}
}

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

// Benchmark tests
func BenchmarkFormatValue(b *testing.B) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	layout := "2006-01-02 15:04:05"
	loc := time.UTC

	b.Run("time", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatValue(testTime, layout, loc)
		}
	})

	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatValue("test string", layout, loc)
		}
	})

	b.Run("float", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatValue(3.14159, layout, loc)
		}
	})
}

func BenchmarkFormatSQLValue(b *testing.B) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 123000000, time.UTC)

	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatSQLValue("test string")
		}
	})

	b.Run("string_with_quotes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatSQLValue("O'Brien's test")
		}
	})

	b.Run("int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatSQLValue(42)
		}
	})

	b.Run("float", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatSQLValue(3.14159)
		}
	})

	b.Run("time", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatSQLValue(testTime)
		}
	})
}

func BenchmarkQuoteIdent(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			quoteIdent("users")
		}
	})

	b.Run("schema_table", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			quoteIdent("public.users")
		}
	})

	b.Run("with_quotes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			quoteIdent(`table"name`)
		}
	})
}
