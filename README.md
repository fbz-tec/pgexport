# pgxport

[![CI - Build, Test & Release](https://github.com/fbz-tec/pgxport/actions/workflows/ci.yml/badge.svg)](https://github.com/fbz-tec/pgxport/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fbz-tec/pgxport)](https://goreportcard.com/report/github.com/fbz-tec/pgxport)
[![License](https://img.shields.io/github/license/fbz-tec/pgxport.svg)](LICENSE)

A simple, powerful and efficient CLI tool to export PostgreSQL query results to various formats (CSV, XML, JSON, SQL).

## ✨ Features

- 🚀 Execute SQL queries directly from command line
- 📄 Run SQL queries from files
- 📊 Export to CSV, JSON, XML, and SQL formats
- ⚡ High-performance CSV export using PostgreSQL native COPY mode (`--with-copy`)
- 🔧 Customizable CSV delimiter and header control (`--no-header`) 
- 🗜️ Optional gzip or zip compression for exported files
- ⚙️ Simple configuration via environment variables or `.env` file
- 🔗 Direct connection string support with `--dsn` flag
- 🛡️ Robust error handling and validation
- ⚠️ Optional fail-on-empty mode (`--fail-on-empty`) for scripting and automation
- 🔍 Verbose mode (`--verbose`) with detailed debug information
- ⚡ Optimized for performance with buffered I/O
- 🔄 Batch INSERT statements for SQL exports (`--insert-batch`) for improved import performance
- 🗃️ Clean architecture with separated concerns
- 🎯 Built with [Cobra](https://github.com/spf13/cobra) for a clean CLI experience

## 📦 Installation

### Prerequisites

- Go 1.19 or higher
- PostgreSQL database access

### Build from source

```bash
# Clone the repository
git clone https://github.com/fbz-tec/pgxport.git
cd pgxport

# Install dependencies
go mod download

# Build the binary
go build -o pgxport

# (Optional) Install to your PATH
sudo cp pgxport /usr/local/bin/
```

### Quick install (one-liner)

```bash
go install github.com/fbz-tec/pgxport@latest
```

## ⚙️ Configuration

### Option 1: Using `.env` file (Recommended for daily use)

Create a `.env` file in your project directory or where you run the command:

```env
DB_USER=myuser
DB_PASS=mypassword
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
```

The `.env` file is automatically loaded when you run `pgxport`. No need to export variables!

**Advantages:**
- ✅ No need to export variables every time
- ✅ Credentials stored in one place

### Option 2: Using environment variables

Configure database connection using environment variables:

```bash
export DB_USER=your_username
export DB_PASS=your_password
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=your_database
```

### Option 3: Using `--dsn` flag (Quick override)

Pass the connection string directly via command line:

```bash
pgxport --dsn "postgres://user:pass@host:port/dbname" -s "SELECT * FROM users" -o users.csv

```

### Configuration Priority

The system uses the following priority order:

1. **`--dsn` flag** (highest priority, overrides everything)
2. **Environment variables** (if defined, override `.env`)
3. **`.env` file** (if present)
4. **Default values** (lowest priority)

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_USER` | Database username | `postgres` | No |
| `DB_PASS` | Database password | _(empty)_ | Recommended |
| `DB_HOST` | Database host | `localhost` | No |
| `DB_PORT` | Database port | `5432` | No |
| `DB_NAME` | Database name | `postgres` | No |

**Note**: While default values are provided, it's recommended to explicitly set all variables for production use.

## 📖 Usage

```bash
pgxport [command] [flags]
```

### Commands

| Command | Description |
|---------|-------------|
| `pgxport` | Execute query and export results |
| `pgxport version` | Show version information |
| `pgxport --help` | Show help message |

### Flags

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--sql` | `-s` | SQL query to execute | - | * |
| `--sqlfile` | `-F` | Path to SQL file | - | * |
| `--output` | `-o` | Output file path | - | ✓ |
| `--format` | `-f` | Output format (csv, json, xml, sql) | `csv` | No |
| `--time-format` | `-T` | Custom date/time format | `yyyy-MM-dd HH:mm:ss` | No |
| `--time-zone` | `-Z` | Time zone for date/time conversion | Local | No |
| `--delimiter` | `-d` | CSV delimiter character | `,` | No |
| `--no-header` | - | Skip CSV header row in output | `false` | No |
| `--with-copy` | - | Use PostgreSQL native COPY for CSV export (faster for large datasets) | `false` | No |
| `--xml-root-tag` | - | Sets the root XML element name | `results` | No |
| `--xml-row-tag` | - | Sets the row XML element name | `row` | No |
| `--fail-on-empty` | - | Exit with error if query returns 0 rows | `false` | No |
| `--table` | `-t` | Table name for SQL INSERT exports (supports schema.table) | - | For SQL format |
| `--insert-batch` | - | Number of rows per INSERT statement for SQL exports | `1` | No |
| `--compression` | `-z` | Compression (none, gzip, zip) | `none` | No |
| `--dsn` | - | Database connection string | - | No |
| `--verbose` | `-v` | Enable verbose output with detailed debug information | `false` | No |
| `--help` | `-h` | Show help message | - | No |

_* Either `--sql` or `--sqlfile` must be provided (but not both)_

### Examples

#### Basic Examples

```bash
# Simple query export (uses .env file)
pgxport -s "SELECT * FROM users WHERE active = true" -o users.csv

# Export with semicolon delimiter
pgxport -s "SELECT id, name, email FROM users" -o users.csv -d ';'

# New: Skip header row with --no-header
pgxport -s "SELECT id, name, email FROM users" -o users.csv -f csv --no-header

# Execute query from a SQL file
pgxport -F queries/monthly_report.sql -o report.csv

# Use the high-performance COPY mode for large CSV exports
pgxport -s "SELECT * FROM big_table" -o big_table.csv -f csv --with-copy

# Export to JSON format
pgxport -s "SELECT * FROM products" -o products.json -f json

# Export to XML format
pgxport -s "SELECT * FROM orders" -o orders.xml -f xml

# Export to XML format with custom root and row tags
pgxport -s "SELECT * FROM orders" -o orders.xml -f xml --xml-root-tag="data" --xml-row-tag="record"

# Export to SQL INSERT statements
pgxport -s "SELECT * FROM products" -o products.sql -f sql -t products_backup

# Export to SQL INSERT statements with schema
pgxport -s "SELECT * FROM products" -o products.sql -f sql -t public.products_backup

# Export to different schema
pgxport -s "SELECT * FROM users" -o users.sql -f sql -t backup.users

# Complex schema names with special characters
pgxport -s "SELECT * FROM data" -o data.sql -f sql -t "my-schema"."my-table"

# Export with batch INSERT statements (100 rows per INSERT)
pgxport -s "SELECT * FROM orders" -o orders.sql -f sql -t orders --insert-batch 100


# Export with gzip compression
pgxport -s "SELECT * FROM logs" -o logs.csv.gz -f csv -z gzip

# Export with zip compression (creates logs.zip containing logs.csv)
pgxport -s "SELECT * FROM logs" -o logs.zip -f csv -z zip

# Check version
pgxport version
```

#### Using Connection String

```bash
# Long form
pgxport --dsn "postgres://myuser:mypass@localhost:5432/mydb" \
         -s "SELECT * FROM users LIMIT 5" \
         -o users.csv

# Override .env with different database
pgxport --dsn "postgres://readonly:pass@replica:5432/mydb" \
         -s "SELECT * FROM large_table" \
         -o export.csv
```

#### Handling Empty Results

The `--fail-on-empty` flag is useful for scripting and automation when you want to ensure your query returns data.

```bash
# Default behavior: Warning message but exit code 0
pgxport -s "SELECT * FROM users WHERE 1=0" -o empty.csv
# Output: Warning: Query returned 0 rows. File created at empty.csv but contains no data rows.
# Exit code: 0

# Strict mode: Error and exit code 1
pgxport -s "SELECT * FROM users WHERE 1=0" -o empty.csv --fail-on-empty
# Output: Error: export failed: query returned 0 rows
# Exit code: 1

# Use in shell scripts for validation
if ! pgxport -s "SELECT * FROM critical_data WHERE date = CURRENT_DATE" \
             -o daily_export.csv --fail-on-empty; then
    echo "❌ Export failed or returned no data!"
    # Send alert, log error, etc.
    exit 1
fi
echo "✅ Export successful with data"

# Combine with other flags
pgxport -s "SELECT * FROM orders WHERE status = 'pending'" \
        -o pending_orders.csv \
        --fail-on-empty \
        -z gzip

# Use in CI/CD pipelines
pgxport -F validate_data.sql -o validation.csv --fail-on-empty || exit 1
```

**When to use `--fail-on-empty`:**
- ✅ Data validation scripts
- ✅ ETL pipelines where empty results indicate a problem
- ✅ Automated reporting where no data is an error condition
- ✅ CI/CD data quality checks
- ✅ Scheduled exports that must contain data

**When NOT to use `--fail-on-empty`:**
- ❌ Exploratory queries where empty results are acceptable
- ❌ Optional data exports
- ❌ Queries with filters that may legitimately return no results

#### Date/Time Formatting Examples

```bash
# Export with custom date format (European style)
pgxport -s "SELECT * FROM events" -o events.csv -T "dd/MM/yyyy HH:mm:ss"

# Export with ISO 8601 format with milliseconds
pgxport -s "SELECT * FROM logs" -o logs.csv -T "yyyy-MM-ddTHH:mm:ss.SSS"

# Export with US date format
pgxport -s "SELECT * FROM orders" -o orders.csv -T "MM/dd/yyyy HH:mm:ss"

# Export with timezone conversion to UTC
pgxport -s "SELECT * FROM events" -o events.csv -Z "UTC"

# Export with timezone conversion to America/New_York
pgxport -s "SELECT * FROM events" -o events.csv -Z "America/New_York"

# Combine custom format and timezone
pgxport -s "SELECT created_at FROM users" -o users.csv \
  -T "dd/MM/yyyy HH:mm:ss" -Z "Europe/Paris"

# Export to JSON with custom date format and timezone
pgxport -s "SELECT * FROM products" -o products.json -f json \
  -T "yyyy-MM-dd HH:mm:ss" -Z "America/Los_Angeles"
```

#### Time Format Tokens

The `--time-format` flag accepts the following tokens:

| Token | Description | Example |
|-------|-------------|---------|
| `yyyy` | 4-digit year | 2025 |
| `yy` | 2-digit year | 24 |
| `MM` | Month (01-12) | 03 |
| `dd` | Day (01-31) | 15 |
| `HH` | Hour 24h (00-23) | 14 |
| `mm` | Minute (00-59) | 30 |
| `ss` | Second (00-59) | 45 |
| `SSS` | Milliseconds | 123 |
| `S` | Deciseconds | 1 |

**Common Format Examples:**
- ISO 8601: `yyyy-MM-ddTHH:mm:ss.SSS`
- European: `dd/MM/yyyy HH:mm:ss`
- US: `MM/dd/yyyy HH:mm:ss`
- Date only: `yyyy-MM-dd`
- Time only: `HH:mm:ss`

#### Timezone Support

The `--time-zone` flag accepts standard IANA timezone names:

**Common Timezones:**
- `UTC` - Coordinated Universal Time
- `America/New_York` - US Eastern Time
- `America/Los_Angeles` - US Pacific Time
- `America/Chicago` - US Central Time
- `Europe/London` - UK Time
- `Europe/Paris` - Central European Time
- `Asia/Tokyo` - Japan Standard Time
- `Australia/Sydney` - Australian Eastern Time

**Default Behavior:**
- If `--time-zone` is not specified, the local system timezone is used
- If an invalid timezone is provided, a warning is displayed and local timezone is used

**Full timezone list:** [IANA Time Zone Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

#### Advanced Examples

```bash
# Complex query with joins
pgxport -s "
SELECT 
  u.id, 
  u.username, 
  COUNT(o.id) as order_count,
  SUM(o.total) as total_revenue
FROM users u 
LEFT JOIN orders o ON u.id = o.user_id 
GROUP BY u.id, u.username 
HAVING COUNT(o.id) > 0
ORDER BY total_revenue DESC
" -o user_stats.csv -d ','

# Export with timestamp in filename
pgxport -s "SELECT * FROM logs WHERE created_at > NOW() - INTERVAL '24 hours'" \
         -o "logs_$(date +%Y%m%d).csv"

# Using long-form flags
pgxport --sql "SELECT * FROM stations ORDER BY name" \
         --output stations.csv \
         --format csv \
         --delimiter ';'
```

#### Batch Processing Examples

```bash
# Process multiple queries with a script
for table in users orders products; do
  pgxport -s "SELECT * FROM $table" -o "${table}_export.csv"
done

# Export with error handling
if pgxport -F complex_query.sql -o output.csv; then
  echo "Export successful!"
else
  echo "Export failed!"
  exit 1
fi

# Connect to different environments
pgxport --dsn "$DEV_DATABASE_URL" -s "SELECT * FROM users" -o dev_users.csv
pgxport --dsn "$PROD_DATABASE_URL" -s "SELECT * FROM users" -o prod_users.csv

# Export same data in different formats
pgxport -s "SELECT * FROM products" -o products.csv -f csv
pgxport -s "SELECT * FROM products" -o products.json -f json
pgxport -s "SELECT * FROM products" -o products.xml -f xml
pgxport -s "SELECT * FROM products" -o products.sql -f sql -t products_backup

# Automated validation script
#!/bin/bash
set -e

echo "Exporting daily metrics..."
if ! pgxport -s "SELECT * FROM daily_metrics WHERE date = CURRENT_DATE" \
             -o metrics.csv --fail-on-empty; then
    echo "ERROR: No metrics found for today!"
    # Send notification
    exit 1
fi

echo "✅ Export completed successfully"
```

## 🔍 Verbose Mode

Enable detailed logging for troubleshooting with the `--verbose` (or `-v`) flag:

```bash
# Normal output
pgxport -s "SELECT * FROM users" -o users.csv

# Detailed output with timestamps and debug information
pgxport -s "SELECT * FROM users" -o users.csv --verbose
```

**Verbose mode shows:**
- Configuration details (host, port, database)
- Connection steps and timing
- Query execution time
- Export progress (every 10,000 rows)
- Performance metrics

**Additional diagnostics (CSV format only):**
- Tracks average row fetch time and overall throughput (rows/s)
- Detects slow PostgreSQL streaming when queries stream data gradually
- Displays a performance summary at the end of the export

**Use cases:**
- 🔍 Debugging connection or query issues
- 📊 Analyzing export performance
- 🐛 Troubleshooting errors

**Example output:**
```bash
$ pgxport -s "SELECT * FROM users LIMIT 5" -o users.csv -v

[2025-01-15 14:23:45.258] 🔍 Configuration loaded: host=localhost port=5432 database=mydb
[2025-01-15 14:23:45.258] ℹ Connecting to database...
[2025-01-15 14:23:45.307] 🔍 Connection established, verifying connectivity (ping)...
[2025-01-15 14:23:45.307] ✓ Database connection established
[2025-01-15 14:23:45.308] ℹ Executing query...
[2025-01-15 14:23:45.311] 🔍 Query: SELECT * FROM users LIMIT 5
[2025-01-15 14:23:46.314] 🔍 Query executed successfully in 145ms
[2025-01-15 14:23:46.315] 🔍 CSV export completed successfully: 5 rows written in 120ms
[2025-01-15 14:23:46.315] ✓ Export completed: 5 rows → users.csv
```

**Note:** Sensitive information (passwords) is automatically masked in logs.

## 📊 Output Formats

### CSV

- **Default delimiter**: `,` (comma)
- Headers included automatically
- **Default timestamp format**: `yyyy-MM-dd HH:mm:ss` (customizable with `--time-format`)
- **Timezone**: Local system time (customizable with `--time-zone`)
- NULL values exported as empty strings
- Buffered I/O for optimal performance

**Example output:**
```csv
id;name;email;created_at
1;John Doe;john@example.com;2024-01-15T10:30:00.000
2;Jane Smith;jane@example.com;2024-01-16T14:22:15.000
```

### ⚙️ COPY Mode (High-Performance CSV Export)

The `--with-copy` flag enables PostgreSQL's native COPY TO STDOUT mechanism for CSV exports.
This mode streams data directly from the database server, reducing CPU and memory usage.

**Benefits:**
- 🚀 Up to 10× faster than row-by-row export for large datasets
- 💾 Low memory footprint
- 🗜️ Compatible with compression (gzip, zip)
- 📄 Identical CSV output format

**Limitations:**
- ⚠️ **Ignores `--time-format` and `--time-zone` options**
- ⚠️ Uses PostgreSQL's default date/time formatting
- Only works with CSV format

**When to use:**
- Large datasets (>100k rows)
- Performance is critical
- Default date format is acceptable

**When NOT to use:**
- Need custom date/time formatting
- Need specific timezone conversion
- Working with small datasets (<10k rows)

Example usage:
```bash
pgxport -s "SELECT * FROM analytics_data" -o analytics.csv -f csv --with-copy
```

#### ⚠️ Note:

When using --with-copy, PostgreSQL handles type serialization.
Date and timestamp formats may differ from standard csv export.

### JSON

- Pretty-printed with 2-space indentation
- Array of objects format
- **Default timestamp format**: `yyyy-MM-dd HH:mm:ss` (customizable with `--time-format`)
- **Timezone**: Local system time (customizable with `--time-zone`)
- NULL values preserved as `null`
- Optimized encoding with buffered I/O

**Example output:**
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15 10:30:00"
  },
  {
    "id": 2,
    "name": "Jane Smith",
    "email": "jane@example.com",
    "created_at": "2024-01-16 14:22:15"
  }
]
```

### XML

- Pretty-printed with 2-space indentation
- **Customizable tags** using:
  - `--xml-root-tag` (default: `results`)
  - `--xml-row-tag` (default: `row`)
- Each column becomes a direct XML element (e.g., `<id>`, `<name>`, `<email>`)
- Each column becomes a direct XML element (e.g., `<id>`, `<name>`, `<email>`)
- **Default timestamp format**: `yyyy-MM-dd HH:mm:ss` (customizable with `--time-format`)
- **Timezone**: Local system time (customizable with `--time-zone`)
- NULL values exported as empty strings
- Buffered I/O for optimal performance

**Example output:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<results>
  <row>
    <id>1</id>
    <name>John Doe</name>
    <email>john@example.com</email>
    <created_at>2024-01-15T10:30:00.000</created_at>
  </row>
  <row>
    <id>2</id>
    <name>Jane Smith</name>
    <email>jane@example.com</email>
    <created_at>2024-01-16T14:22:15.000</created_at>
  </row>
</results>
```

### SQL

- INSERT statements format for easy data migration
- Buffered I/O for optimal performance
- **Requires `--table` / `-t` parameter to specify target table name**
- **Batch INSERT support** with `--insert-batch` flag for improved import performance

**Example output:**
```sql
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES (1, 'John Doe', 'john@example.com', '2024-01-15 10:30:00.000');
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES (2, 'Jane Smith', 'jane@example.com', '2024-01-16 14:22:15.000');
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES (3, 'Bob O''Brien', NULL, '2024-01-17 09:15:30.000');

# batch insert
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES
	(1, 'John Doe', 'john@example.com', '2024-01-15 10:30:00.000'),
	(2, 'Jane Smith', 'jane@example.com', '2024-01-16 14:22:15.000'),
	(3, 'Bob O''Brien', NULL, '2024-01-17 09:15:30.000');

# Export to a table in a specific schema
pgxport -s "SELECT * FROM users" -o users.sql -f sql -t public.users

# Export to a backup schema
pgxport -s "SELECT * FROM orders" -o orders.sql -f sql -t backup.orders

# Schema or table names with special characters
pgxport -s "SELECT * FROM data" -o data.sql -f sql -t "\"my-schema\".\"my-table\""

# Export from file to SQL with optimized batch size
pgxport -F query.sql -o output.sql -f sql --table target_table --insert-batch 500
```

**Usage example:**
```bash
# Export to SQL INSERT statements
pgxport -s "SELECT * FROM users WHERE active = true" -o users.sql -f sql -t users_backup

# Export from file to SQL
pgxport -F query.sql -o output.sql -f sql --table target_table

# Complex data types (numbers, booleans, NULL, dates)
pgxport -s "SELECT id, name, price, active, created_at, notes FROM products" \
        -o products.sql -f sql -t products_backup

# Export from file to SQL with optimized batch size
pgxport -F query.sql -o output.sql -f sql --table target_table --insert-batch 500

# Export to a table in a specific schema with batch INSERT
pgxport -s "SELECT * FROM orders" -o orders.sql -f sql -t backup.orders --insert-batch 250        
```

**SQL Format Features:**
- ✅ **Schema-qualified table names**: Supports `schema.table` notation for cross-schema exports
- ✅ **Batch INSERT support**: Use `--insert-batch` to group multiple rows in a single INSERT statement for significantly faster imports
- ✅ **All PostgreSQL data types supported**: integers, floats, strings, booleans, timestamps, NULL, bytea
- ✅ **Automatic escaping**: Single quotes in strings are properly escaped (e.g., `O'Brien` → `'O''Brien'`)
- ✅ **Identifier quoting**: Properly quotes table and column names to handle special characters
- ✅ **Type-aware formatting**: Numbers and booleans without quotes, strings and dates with quotes
- ✅ **NULL handling**: NULL values exported as SQL `NULL` keyword
- ✅ **Ready to import**: Generated SQL can be directly executed on any PostgreSQL database

## 🗃️ Project Structure

```
pgxport/
├── exporters/          # Modular export package
│   ├── exporter.go     # Interface and factory
│   ├── compression.go  # Compression writers (gzip,zip)
│   ├── common.go       # Shared utilities
│   ├── csv_exporter.go # CSV export implementation
│   ├── json_exporter.go# JSON export implementation
│   ├── xml_exporter.go # XML export implementation
│   └── sql_exporter.go # SQL export implementation
├── logger/             # Logging package
│   └── logger.go       # Logger interface and implementation
├── main.go             # CLI entry point and orchestration
├── config.go           # Configuration management with validation
├── store.go            # Database operations (connection, queries)
├── version.go          # Version information
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── LICENSE             # MIT license file
└── README.md           # Documentation
```

### Architecture

The project follows a clean, modular architecture with separated concerns:

- **`exporters/`**: Modular export package with Strategy pattern
  - **`exporter.go`**: Defines the `Exporter` interface and factory
  - **`compression.go`**: Handles output compression (gzip, zip)
  - **`common.go`**: Shared formatting utilities for all exporters
  - **`csv_exporter.go`**: CSV export implementation
  - **`json_exporter.go`**: JSON export implementation
  - **`xml_exporter.go`**: XML export implementation
  - **`sql_exporter.go`**: SQL INSERT export implementation
- **`logger/`**: Logging package with structured output
  - **`logger.go`**: Logger interface and singleton implementation with debug/verbose support
- **`store.go`**: Handles all database operations (connect, query, return results)
- **`main.go`**: Orchestrates the flow between store and exporters
- **`config.go`**: Manages configuration with validation, defaults, and `.env` file loading

Each exporter is isolated in its own file, making the codebase easy to maintain, test, and extend with new formats.

### Building

```bash
# Build for current platform
go build -o pgxport

# Build with version information
go build -ldflags="-X main.Version=1.0.0" -o pgxport

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o pgxport-linux
GOOS=darwin GOARCH=amd64 go build -o pgxport-macos
GOOS=windows GOARCH=amd64 go build -o pgxport.exe
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestValidateExportParams
```

### Dependencies

The project uses the following main dependencies:

- [pgx/v5](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit
- [cobra](https://github.com/spf13/cobra) - Modern CLI framework
- [godotenv](https://github.com/joho/godotenv) - Load environment variables from `.env` file

Install them with:

```bash
go get -u github.com/jackc/pgx/v5
go get -u github.com/spf13/cobra
go get -u github.com/joho/godotenv
```

### Setting up your development environment

1. Clone the repository
2. Create a `.env` file with your local database credentials:

```bash
cat > .env << EOF
DB_USER=postgres
DB_PASS=your_local_password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=testdb
EOF
```

3. Build and test:

```bash
go build -o pgxport
./pgxport -s "SELECT version()" -o version.csv
```

## 🔒 Security Best Practices

1. **Never commit credentials**: 
   - `.env` is already in `.gitignore`
   - Use `.env.example` for documentation
   - For production, use environment variables or secrets management

2. **Avoid passwords in command line**:
   - ❌ Bad: `pgxport --dsn "postgres://user:password123@host/db" ...` (visible in history)
   - ✅ Good: Use `.env` file or environment variables
   - ✅ Good: Store DSN in environment: `export DATABASE_URL="..."` then use `pgxport --dsn "$DATABASE_URL" ...`

3. **Use parameterized queries**: When using dynamic SQL, be aware of SQL injection risks

4. **Limit database permissions**: Use a database user with minimal required privileges (SELECT only for exports)

5. **Secure your output files**: Be careful with sensitive data in exported files

6. **Review queries**: Always review SQL files before execution
   
7. **Verbose mode security**: Remember that `--verbose` logs queries and configuration. Avoid logging sensitive data.

## 🚨 Error Handling

The tool provides clear error messages for common issues:

- **Connection errors**: Check database credentials and network connectivity
- **SQL errors**: Verify your query syntax
- **File errors**: Ensure write permissions for output directory
- **Configuration errors**: Validate all required environment variables
- **Format errors**: Ensure format is one of: csv, json, xml, sql
- **SQL format errors**: Ensure `--table` flag is provided when using SQL format
- **Empty result errors**: Use `--fail-on-empty` to treat 0 rows as an error

Example error output:
```
Error: Invalid format 'txt'. Valid formats are: csv, json, xml, sql
Error: --table (-t) is required when using SQL format
Error: Configuration error: DB_PORT must be a valid port number (1-65535)
Error: export failed: query returned 0 rows
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow Go conventions and use `gofmt`
- Add comments for exported functions
- Keep functions small and focused
- Separate concerns (database vs export logic)
- Write tests for new features

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🗺️ Roadmap

- [x] ~~`.env` file support for easy configuration~~ ✅ Implemented!
- [x] ~~Direct connection string support with `--dsn` flag~~ ✅ Implemented!
- [x] ~~XML export format~~ ✅ Implemented!
- [x] ~~SQL INSERT export format~~ ✅ Implemented!
- [x] ~~High-performance CSV export using PostgreSQL COPY~~ ✅ Implemented!
- [x] ~~Fail-on-empty flag for scripting and automation~~ ✅ Implemented!
- [x] ~~Verbose mode for debugging~~ ✅ Implemented!
- [ ] Excel (XLSX) export format
- [ ] Query pagination for large datasets
- [ ] Progress bar for long-running queries
- [ ] Query result preview before export
- [x] ~~Streaming mode for huge datasets~~ ✅ Implemented!
- [x] ~~Compression support (gzip, zip)~~ ✅ Implemented!
- [x] ~~SQL format with column names: `INSERT INTO table (col1, col2) VALUES ...`~~ ✅ Implemented!
- [x] ~~Batch INSERT statements for better performance~~ ✅ Implemented!

## 💬 Support

If you encounter any issues or have questions:

- 🐛 [Open an issue](https://github.com/fbz-tec/pgxport/issues) on GitHub
- 💡 [Start a discussion](https://github.com/fbz-tec/pgxport/discussions) for feature requests

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- PostgreSQL driver: [pgx](https://github.com/jackc/pgx)
- Environment variables: [godotenv](https://github.com/joho/godotenv)
- Inspired by the need for simple, reliable data exports

## ⭐ Show Your Support

If you find this project helpful, please consider giving it a star on GitHub!

---

**Made with ❤️ for the PostgreSQL community**