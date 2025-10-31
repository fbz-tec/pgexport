package exporters

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// exportToCSV writes query results to a CSV file with buffered I/O
func (e *dataExporter) writeCSV(rows pgx.Rows, csvPath string, options ExportOptions) (int, error) {
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
	}

	//datetime layout(Golang format) and timezone
	layout, loc := userTimeZoneFormat(options.TimeFormat, options.TimeZone)

	// Write data rows
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
			writer.Flush()
		}

	}

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	return rowCount, nil
}

func (e *dataExporter) writeCopyCSV(conn *pgx.Conn, query string, csvPath string, options ExportOptions) (int, error) {
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

	return int(tag.RowsAffected()), nil

}
