package exporters

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/fbz-tec/pgexport/logger"
	"github.com/jackc/pgx/v5"
)

// exportToCSV writes query results to a CSV file with buffered I/O
func (e *dataExporter) writeCSV(rows pgx.Rows, csvPath string, options ExportOptions) (int, error) {
	start := time.Now()

	logger.Debug("Preparing CSV export (delimiter=%q, noHeader=%v, compression=%s)",
		string(options.Delimiter), options.NoHeader, options.Compression)

	writerCloser, err := createOutputWriter(csvPath, options, FormatCSV)
	if err != nil {
		return 0, err
	}

	defer writerCloser.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(writerCloser)
	defer bufferedWriter.Flush()

	writer := csv.NewWriter(bufferedWriter)
	writer.Comma = options.Delimiter
	defer writer.Flush()

	// Write headers
	if !options.NoHeader {
		fieldDescriptions := rows.FieldDescriptions()
		headers := make([]string, len(fieldDescriptions))
		for i, fd := range fieldDescriptions {
			headers[i] = string(fd.Name)
		}

		if err := writer.Write(headers); err != nil {
			return 0, fmt.Errorf("error writing headers: %w", err)
		}
		logger.Debug("CSV headers written: %s", strings.Join(headers, string(options.Delimiter)))
	}

	//datetime layout(Golang format) and timezone
	layout, loc := userTimeZoneFormat(options.TimeFormat, options.TimeZone)

	// Write data rows
	logger.Debug("Starting to write CSV rows...")

	rowCount := 0

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return rowCount, fmt.Errorf("error reading row: %w", err)
		}
		//format values to strings
		record := make([]string, len(values))
		for i, v := range values {
			record[i] = fmt.Sprintf("%v", formatValue(v, layout, loc))
		}

		rowCount++

		if err := writer.Write(record); err != nil {
			return 0, fmt.Errorf("error writing row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			logger.Debug("%d rows written...", rowCount)
			writer.Flush()
		}

	}

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	logger.Debug("Flushing CSV buffers to disk...")
	writer.Flush()

	if err := writer.Error(); err != nil {
		return rowCount, fmt.Errorf("error flushing CSV: %w", err)
	}

	logger.Debug("CSV export completed successfully: %d rows written in %v", rowCount, time.Since(start))

	return rowCount, nil
}

func (e *dataExporter) writeCopyCSV(conn *pgx.Conn, query string, csvPath string, options ExportOptions) (int, error) {

	start := time.Now()
	logger.Debug("Starting PostgreSQL COPY export (noHeader=%v, compression=%s)", options.NoHeader, options.Compression)

	writerCloser, err := createOutputWriter(csvPath, options, FormatCSV)
	if err != nil {
		return 0, err
	}

	defer writerCloser.Close()

	copySql := fmt.Sprintf("COPY (%s) TO STDOUT WITH (FORMAT csv, HEADER %t, DELIMITER '%c')", query, !options.NoHeader, options.Delimiter)

	tag, err := conn.PgConn().CopyTo(context.Background(), writerCloser, copySql)
	if err != nil {
		return 0, fmt.Errorf("COPY TO STDOUT failed: %w", err)
	}

	rowCount := int(tag.RowsAffected())
	logger.Debug("COPY export completed successfully: %d rows written in %v", rowCount, time.Since(start))

	return rowCount, nil

}
