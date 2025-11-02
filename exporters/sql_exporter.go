package exporters

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/fbz-tec/pgexport/logger"
	"github.com/jackc/pgx/v5"
)

func (e *dataExporter) writeSQL(rows pgx.Rows, sqlPath string, options ExportOptions) (int, error) {

	start := time.Now()
	logger.Debug("Preparing SQL export (table=%s, compression=%s)", options.TableName, options.Compression)

	writeCloser, err := createOutputWriter(sqlPath, options, FormatSQL)
	if err != nil {
		return 0, err
	}
	defer writeCloser.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(writeCloser)
	defer bufferedWriter.Flush()

	rowCount := 0

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, fd := range fields {
		columns[i] = quoteIdent(fd.Name)
	}
	size := len(columns)

	logger.Debug("Starting to write SQL INSERT statements...")

	for rows.Next() {
		record := make([]string, size)

		values, err := rows.Values()
		if err != nil {
			return 0, fmt.Errorf("error reading row: %w", err)
		}

		//format values
		for i, val := range values {
			record[i] = formatSQLValue(val)
		}

		rowCount++
		//Create insert line with values
		line := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n", quoteIdent(options.TableName), strings.Join(columns, ", "), strings.Join(record, ", "))

		if _, err := bufferedWriter.WriteString(line); err != nil {
			return 0, fmt.Errorf("error writing row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
			logger.Debug("%d INSERT statements written...", rowCount)
		}
	}

	logger.Debug("Flushing remaining SQL statements to disk...")
	bufferedWriter.Flush()

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	logger.Debug("SQL export completed successfully: %d rows written in %v", rowCount, time.Since(start))

	return rowCount, nil
}
