# Changelog

All notable changes to pgxport will be documented in this file.

## [1.0.0] - 2025-11-04

### First Stable Release

This is the first production-ready release of pgxport.

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
- Comprehensive test coverage
- CI/CD pipeline with automated builds

#### Installation

```bash
go install github.com/fbz-tec/pgxport@v1.0.0
```

Or download pre-built binaries from [GitHub Releases](https://github.com/fbz-tec/pgxport/releases/tag/v1.0.0).

---

For detailed usage, see [README.md](README.md)