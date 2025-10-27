package exporters

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (e *dataExporter) writeSQL(rows pgx.Rows, sqlPath string, options ExportOptions) (int, error) {
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
		columns[i] = string(fd.Name)
	}
	size := len(columns)

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
		line := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n", options.TableName, strings.Join(columns, ", "), strings.Join(record, ", "))

		if _, err := bufferedWriter.WriteString(line); err != nil {
			return 0, fmt.Errorf("error writing row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			bufferedWriter.Flush()
		}
	}

	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}

	return rowCount, nil
}
