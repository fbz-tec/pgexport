# pgexport

A simple and powerful CLI tool to export PostgreSQL query results to various formats (CSV, JSON, and more).

## Features

- 🚀 Execute SQL queries directly from command line
- 📄 Run SQL queries from files
- 📊 Export to CSV, JSON format (XML coming soon)
- 🔧 Customizable CSV delimiter
- ⚙️ Simple configuration via environment variables
- 🎯 Built with [Cobra](https://github.com/spf13/cobra) for a clean CLI experience

## Installation

### Prerequisites

- Go 1.19 or higher
- PostgreSQL database access

### Build from source

```bash
# Clone the repository
git clone https://github.com/yourusername/pgexport.git
cd pgexport

# Install dependencies
go mod download

# Build the binary
go build -o pgexport

# (Optional) Install to your PATH
cp pgexport ~/bin/
# or
sudo cp pgexport /usr/local/bin/
```

## Configuration

Configure database connection using environment variables:

```bash
export DB_USER=your_username
export DB_PASS=your_password
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=your_database
```

**Optional**: You can use a `.env` file for development or to secure sensitive information like passwords. The tool will automatically load it if present.

```env
# .env file (optional)
DB_USER=your_username
DB_PASS=your_password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=your_database
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_USER` | Database username | `postgres` |
| `DB_PASS` | Database password | _(empty)_ |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_NAME` | Database name | `postgres` |

## Usage

```bash
pgexport [flags]
```

### Flags

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--sql` | `-s` | SQL query to execute | - | * |
| `--sqlfile` | `-F` | Path to SQL file | - | * |
| `--output` | `-o` | Output file path | - | ✓ |
| `--format` | `-f` | Output format (csv, json) | `csv` | |
| `--delimiter` | `-d` | CSV delimiter character | `;` | |
| `--help` | `-h` | Show help message | - | |

_* Either `--sql` or `--sqlfile` must be provided_

### Examples

#### Execute a simple query

```bash
pgexport -s "SELECT * FROM users WHERE active = true" -o users.csv
```

#### Export with comma delimiter

```bash
pgexport -s "SELECT id, name, email FROM users" -o users.csv -d ','
```

#### Execute query from a SQL file

```bash
pgexport -F queries/report.sql -o report.csv
```

#### Using long-form flags

```bash
pgexport --sql "SELECT * FROM stations ORDER BY name" \
         --output stations.csv \
         --format csv \
         --delimiter ';'
```

#### Complex query example

```bash
pgexport -s "SELECT 
  u.id, 
  u.username, 
  COUNT(o.id) as order_count 
FROM users u 
LEFT JOIN orders o ON u.id = o.user_id 
GROUP BY u.id, u.username 
ORDER BY order_count DESC" -o user_stats.csv -d ','
```

## Output Formats

### CSV (Current)

- Default delimiter: `;` (semicolon)
- Headers included automatically
- Timestamps formatted as `2006-01-02T15:04:05.000`
- NULL values exported as empty strings

### JSON 

JSON export

### XML (Coming Soon)

JSON export functionality is planned for future releases.

## Development

### Project Structure

```
pgexport/
├── main.go           # CLI entry point with Cobra
├── config.go         # Configuration loader
├── store.go          # Database connection and export logic
├── .env             # Environment configuration (optional)
├── go.mod           # Go module definition
└── README.md        # This file
```

### Dependencies

```bash
go get -u github.com/spf13/cobra
go get -u github.com/jackc/pgx/v5
go get -u github.com/joho/godotenv
```

### Building

```bash
# Build for current platform
go build -o pgexport

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o pgexport-linux

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o pgexport.exe

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o pgexport-macos
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] XML export format
- [ ] Excel export format
- [ ] Query pagination for large datasets
- [ ] Progress bar for long-running queries
- [ ] Multiple database support (MySQL, SQLite)
- [ ] Connection string support as alternative to env vars
- [ ] Query result preview before export

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/yourusername/pgexport/issues) on GitHub.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- PostgreSQL driver: [pgx](https://github.com/jackc/pgx)