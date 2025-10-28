# pgxport

A simple, powerful and efficient CLI tool to export PostgreSQL query results to various formats (CSV, XML, JSON, SQL).

## ✨ Features

- 🚀 Execute SQL queries directly from command line
- 📄 Run SQL queries from files
- 📊 Export to CSV, JSON, XML, and SQL formats
- ⚡ High-performance CSV export using PostgreSQL native COPY mode (`--with-copy`)
- 🔧 Customizable CSV delimiter
- 🗜️ Optional gzip or zip compression for exported files
- ⚙️ Simple configuration via environment variables or `.env` file
- 🔗 Direct connection string support with `--dsn` flag
- 🛡️ Robust error handling and validation
- ⚡ Optimized for performance with buffered I/O
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
# or with short flag
pgxport -c "postgres://user:pass@host:port/dbname" -s "SELECT * FROM users" -o users.csv
```

### Configuration Priority

The system uses the following priority order:

1. **`--dsn` / `-c` flag** (highest priority, overrides everything)
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
| `--delimiter` | `-d` | CSV delimiter character | `,` | No |
| `--with-copy` | - | Use PostgreSQL native COPY for CSV export (faster for large datasets) | `false` | No |
| `--table` | `-t` | Table name for SQL INSERT exports | - | For SQL format |
| `--compression` | `-z` | Compression (none, gzip, zip) | `none` | No |
| `--dsn` | `-c` | Database connection string | - | No |
| `--help` | `-h` | Show help message | - | No |

_* Either `--sql` or `--sqlfile` must be provided (but not both)_

### Examples

#### Basic Examples

```bash
# Simple query export (uses .env file)
pgxport -s "SELECT * FROM users WHERE active = true" -o users.csv

# Export with semicolon delimiter
pgxport -s "SELECT id, name, email FROM users" -o users.csv -d ';'

# Execute query from a SQL file
pgxport -F queries/monthly_report.sql -o report.csv

# Use the high-performance COPY mode for large CSV exports
pgxport -s "SELECT * FROM big_table" -o big_table.csv -f csv --with-copy

# Export to JSON format
pgxport -s "SELECT * FROM products" -o products.json -f json

# Export to XML format
pgxport -s "SELECT * FROM orders" -o orders.xml -f xml

# Export to SQL INSERT statements
pgxport -s "SELECT * FROM products" -o products.sql -f sql -t products_backup

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

# Short form
pgxport -c "postgres://myuser:mypass@prod-server:5432/analytics" \
         -s "SELECT * FROM metrics WHERE date = CURRENT_DATE" \
         -o daily_metrics.csv

# Override .env with different database
pgxport -c "postgres://readonly:pass@replica:5432/mydb" \
         -s "SELECT * FROM large_table" \
         -o export.csv
```

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
pgxport -c "$DEV_DATABASE_URL" -s "SELECT * FROM users" -o dev_users.csv
pgxport -c "$PROD_DATABASE_URL" -s "SELECT * FROM users" -o prod_users.csv

# Export same data in different formats
pgxport -s "SELECT * FROM products" -o products.csv -f csv
pgxport -s "SELECT * FROM products" -o products.json -f json
pgxport -s "SELECT * FROM products" -o products.xml -f xml
pgxport -s "SELECT * FROM products" -o products.sql -f sql -t products_backup
```

## 📊 Output Formats

### CSV

- **Default delimiter**: `,` (comma)
- Headers included automatically
- Timestamps formatted as `2006-01-02T15:04:05.000`
- NULL values exported as empty strings
- Buffered I/O for optimal performance

**Example output:**
```csv
id;name;email;created_at
1;John Doe;john@example.com;2024-01-15T10:30:00.000
2;Jane Smith;jane@example.com;2024-01-16T14:22:15.000
```
### ⚙️ COPY Mode (High-Performance CSV Export)

The `--with-copy` flag enables PostgreSQL’s native COPY TO STDOUT mechanism for CSV exports.
This mode streams data directly from the database server, reducing CPU and memory usage.

Benefits:
- 🚀 Up to 10× faster than row-by-row export for large datasets
- 💾 Low memory footprint
- 🗜️ Compatible with compression (gzip, zip)
- 🔄 Identical CSV output format

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
- Timestamps formatted as `2006-01-02T15:04:05.000`
- NULL values preserved as `null`
- Optimized encoding with buffered I/O

**Example output:**
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00.000"
  },
  {
    "id": 2,
    "name": "Jane Smith",
    "email": "jane@example.com",
    "created_at": "2024-01-16T14:22:15.000"
  }
]
```

### XML

- Pretty-printed with 2-space indentation
- Structured with `<results>` root and `<row>` elements
- Each column becomes a direct XML element (e.g., `<id>`, `<name>`, `<email>`)
- Timestamps formatted as `2006-01-02T15:04:05.000`
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

**Example output:**
```sql
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES (1, 'John Doe', 'john@example.com', '2024-01-15 10:30:00.000');
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES (2, 'Jane Smith', 'jane@example.com', '2024-01-16 14:22:15.000');
INSERT INTO "users" ("id", "name", "email", "created_at") VALUES (3, 'Bob O''Brien', NULL, '2024-01-17 09:15:30.000');
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
```

**SQL Format Features:**
- ✅ **All PostgreSQL data types supported**: integers, floats, strings, booleans, timestamps, NULL, bytea
- ✅ **Automatic escaping**: Single quotes in strings are properly escaped (e.g., `O'Brien` → `'O''Brien'`)
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
│   ├── sql_exporter.go # SQL export implementation
│   └── README.md       # Package documentation
├── main.go             # CLI entry point and orchestration
├── config.go           # Configuration management with validation
├── store.go            # Database operations (connection, queries)
├── version.go          # Version information
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
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
   - ✅ Good: Store DSN in environment: `export DATABASE_URL="..."` then use `pgxport -c "$DATABASE_URL" ...`

3. **Use parameterized queries**: When using dynamic SQL, be aware of SQL injection risks

4. **Limit database permissions**: Use a database user with minimal required privileges (SELECT only for exports)

5. **Secure your output files**: Be careful with sensitive data in exported files

6. **Review queries**: Always review SQL files before execution

## 🚨 Error Handling

The tool provides clear error messages for common issues:

- **Connection errors**: Check database credentials and network connectivity
- **SQL errors**: Verify your query syntax
- **File errors**: Ensure write permissions for output directory
- **Configuration errors**: Validate all required environment variables
- **Format errors**: Ensure format is one of: csv, json, xml, sql
- **SQL format errors**: Ensure `--table` flag is provided when using SQL format

Example error output:
```
Error: Invalid format 'txt'. Valid formats are: csv, json, xml, sql
Error: --table (-t) is required when using SQL format
Error: Configuration error: DB_PORT must be a valid port number (1-65535)
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
- [ ] Interactive password prompt (secure, no history)
- [ ] Excel (XLSX) export format
- [ ] Query pagination for large datasets
- [ ] Progress bar for long-running queries
- [ ] Multiple database support (MySQL, SQLite, SQL Server)
- [ ] Query result preview before export
- [x] ~~Streaming mode for huge datasets~~ ✅ Implemented!
- [x] ~~Compression support (gzip, zip)~~ ✅ Implemented!
- [x] ~~SQL format with column names: `INSERT INTO table (col1, col2) VALUES ...`~~ ✅ Implemented!
- [ ] Batch INSERT statements for better performance

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