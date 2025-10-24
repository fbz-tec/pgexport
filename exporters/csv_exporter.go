package exporters

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// exportToCSV writes query results to a CSV file with buffered I/O
func (e *dataExporter) writeCSV(rows pgx.Rows, csvPath string, options ExportOptions) (int, error) {
	file, err := os.Create(csvPath)
	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	writer := csv.NewWriter(bufferedWriter)
	writer.Comma = options.Delimiter
	defer writer.Flush()

	// Write headers
	fieldDescriptions := rows.FieldDescriptions()
	headers := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		headers[i] = string(fd.Name)
	}

	if err := writer.Write(headers); err != nil {
		return 0, fmt.Errorf("error writing headers: %w", err)
	}

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
			record[i] = toString(v)
		}

		rowCount++

		if err := writer.Write(record); err != nil {
			return 0, fmt.Errorf("error writing row %d: %w", rowCount, err)
		}

		if rowCount%10000 == 0 {
			writer.Flush()
		}

	}

	return rowCount, nil
}
