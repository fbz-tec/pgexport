# Changelog

All notable changes to pgxport will be documented in this file.

## [v1.0.0-rc1] - 2025-11-10

### First Pre-Release

This is the first pre-release of pgxport.

#### Features

- Export PostgreSQL queries to CSV, JSON, XML, and SQL formats
- High-performance CSV export with PostgreSQL native COPY mode (`--with-copy`)
- Compression support (gzip, zip)
- Flexible configuration (`.env` file, environment variables, or DSN)
- Customizable CSV delimiter and header control
- Custom XML tags with `--xml-root-tag` and `--xml-row-tag` flags
- Verbose mode with performance diagnostics (`--verbose`)
- Fail-on-empty mode for automation (`--fail-on-empty`)
- Custom date/time formats and timezone support
- SQL export with schema-qualified table names
- Batch INSERT statements for SQL exports (`--insert-batch`) for improved import performance

#### Installation

Download the pre-built binary for your platform from the [releases page](https://github.com/fbz-tec/pgxport/releases/tag/v1.0.0-rc1):

- **Linux (x86_64)**: `pgxport-linux-amd64.tar.gz`
- **Linux (ARM64)**: `pgxport-linux-arm64.tar.gz`
- **macOS (Intel)**: `pgxport-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `pgxport-darwin-arm64.tar.gz`
- **Windows (x86_64)**: `pgxport-windows-amd64.zip`
- **Windows (ARM64)**: `pgxport-windows-arm64.zip`

Extract and use immediately - **no installation required!**

**For Go developers:**
```bash
go install github.com/fbz-tec/pgxport@v1.0.0
```


---

For detailed usage, see [README.md](README.md)