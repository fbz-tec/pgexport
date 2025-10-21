# pgexport

A simple, powerful and efficient CLI tool to export PostgreSQL query results to various formats (CSV, JSON).

## ✨ Features

- 🚀 Execute SQL queries directly from command line
- 📄 Run SQL queries from files
- 📊 Export to CSV and JSON formats
- 🔧 Customizable CSV delimiter
- ⚙️ Simple configuration via environment variables
- 🛡️ Robust error handling and validation
- ⚡ Optimized for performance with buffered I/O
- 🏗️ Clean architecture with separated concerns
- 🎯 Built with [Cobra](https://github.com/spf13/cobra) for a clean CLI experience

## 📦 Installation

### Prerequisites

- Go 1.19 or higher
- PostgreSQL database access

### Build from source

```bash
# Clone the repository
git clone https://github.com/fbz-tec/pgexport.git
cd pgexport

# Install dependencies
go mod download

# Build the binary
go build -o pgexport

# (Optional) Install to your PATH
sudo cp pgexport /usr/local/bin/
```

### Quick install (one-liner)

```bash
go install github.com/fbz-tec/pgexport@latest
```

## ⚙️ Configuration

Configure database connection using environment variables:

```bash
export DB_USER=your_username
export DB_PASS=your_password
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=your_database
```

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_USER` | Database username | `postgres` | No |
| `DB_PASS` | Database password | _(empty)_ | Recommended |
| `DB_HOST` | Database host | `localhost` | No |
| `DB_PORT` | Database port | `5432` | No |
| `DB_NAME` | Database name | `postgres` | No |

**Note**: While default values are provided, it's recommended to explicitly set all variables for production use.

### Using .env file (optional)

Create a `.env` file in your project directory:

```env
DB_USER=myuser
DB_PASS=mypassword
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
```

## 📖 Usage

```bash
pgexport [command] [flags]
```

### Commands

| Command | Description |
|---------|-------------|
| `pgexport` | Execute query and export results |
| `pgexport version` | Show version information |
| `pgexport --help` | Show help message |

### Flags

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--sql` | `-s` | SQL query to execute | - | * |
| `--sqlfile` | `-F` | Path to SQL file | - | * |
| `--output` | `-o` | Output file path | - | ✓ |
| `--format` | `-f` | Output format (csv, json) | `csv` | No |
| `--delimiter` | `-d` | CSV delimiter character | `;` | No |
| `--help` | `-h` | Show help message | - | No |

_* Either `--sql` or `--sqlfile` must be provided (but not both)_

### Examples

#### Basic Examples

```bash
# Simple query export
pgexport -s "SELECT * FROM users WHERE active = true" -o users.csv

# Export with comma delimiter
pgexport -s "SELECT id, name, email FROM users" -o users.csv -d ','

# Execute query from a SQL file
pgexport -F queries/monthly_report.sql -o report.csv

# Export to JSON format
pgexport -s "SELECT * FROM products" -o products.json -f json

# Check version
pgexport version
```

#### Advanced Examples

```bash
# Complex query with joins
pgexport -s "
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
pgexport -s "SELECT * FROM logs WHERE created_at > NOW() - INTERVAL '24 hours'" \
         -o "logs_$(date +%Y%m%d).csv"

# Using long-form flags
pgexport --sql "SELECT * FROM stations ORDER BY name" \
         --output stations.csv \
         --format csv \
         --delimiter ';'
```

#### Batch Processing Examples

```bash
# Process multiple queries with a script
for table in users orders products; do
  pgexport -s "SELECT * FROM $table" -o "${table}_export.csv"
done

# Export with error handling
if pgexport -F complex_query.sql -o output.csv; then
  echo "Export successful!"
else
  echo "Export failed!"
  exit 1
fi
```

## 📊 Output Formats

### CSV

- **Default delimiter**: `;` (semicolon)
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

## 🏗️ Project Structure

```
pgexport/
├── main.go           # CLI entry point and orchestration
├── config.go         # Configuration management with validation
├── store.go          # Database operations (connection, queries)
├── exporter.go       # Export operations (CSV, JSON formatting)
├── version.go        # Version information
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
└── README.md         # Documentation
```

### Architecture

The project follows a clean architecture with separated concerns:

- **`store.go`**: Handles all database operations (connect, query, return results)
- **`exporter.go`**: Handles all export operations (format data, write files)
- **`main.go`**: Orchestrates the flow between store and exporter
- **`config.go`**: Manages configuration with validation and defaults


## 🔧 Development

### Building

```bash
# Build for current platform
go build -o pgexport

# Build with version information
go build -ldflags="-X main.Version=1.0.0" -o pgexport

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o pgexport-linux
GOOS=darwin GOARCH=amd64 go build -o pgexport-macos
GOOS=windows GOARCH=amd64 go build -o pgexport.exe
```

### Dependencies

The project uses the following main dependencies:

- [pgx/v5](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit
- [cobra](https://github.com/spf13/cobra) - Modern CLI framework

Install them with:

```bash
go get -u github.com/jackc/pgx/v5
go get -u github.com/spf13/cobra
```

## 🔒 Security Best Practices

1. **Never commit credentials**: Use environment variables or `.env` files (add `.env` to `.gitignore`)
2. **Use parameterized queries**: When using dynamic SQL, be aware of SQL injection risks
3. **Limit database permissions**: Use a database user with minimal required privileges
4. **Secure your output files**: Be careful with sensitive data in exported files
5. **Review queries**: Always review SQL files before execution

## 🚨 Error Handling

The tool provides clear error messages for common issues:

- **Connection errors**: Check database credentials and network connectivity
- **SQL errors**: Verify your query syntax
- **File errors**: Ensure write permissions for output directory
- **Configuration errors**: Validate all required environment variables

Example error output:
```
2024/01/15 10:30:00 Configuration error: DB_PORT must be a valid port number (1-65535), got: abc
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

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🗺️ Roadmap

- [ ] XML export format
- [ ] Excel (XLSX) export format
- [ ] Query pagination for large datasets
- [ ] Progress bar for long-running queries
- [ ] Multiple database support (MySQL, SQLite, SQL Server)
- [ ] Connection string support as alternative to env vars
- [ ] Query result preview before export
- [ ] Streaming mode for huge datasets
- [ ] Compression support (gzip, zip)

## 💬 Support

If you encounter any issues or have questions:

- 🐛 [Open an issue](https://github.com/fbz-tec/pgexport/issues) on GitHub
- 💡 [Start a discussion](https://github.com/fbz-tec/pgexport/discussions) for feature requests

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- PostgreSQL driver: [pgx](https://github.com/jackc/pgx)
- Inspired by the need for simple, reliable data exports

## ⭐ Show Your Support

If you find this project helpful, please consider giving it a star on GitHub!

---

**Made with ❤️ for the PostgreSQL community**