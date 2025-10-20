package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type Query interface {
	Open(dbUrl string) error
	Close() error
	ExportQueryToCSV(query string, csvPath string, delimiter rune) error
}

type dbStore struct {
	conn *pgx.Conn
	ctx  context.Context
}

func (store *dbStore) Open(dbUrl string) error {
	store.ctx = context.Background()

	config, err := pgx.ParseConfig(dbUrl)

	if err != nil {
		return fmt.Errorf("unable to parse config: %w", err)
	}

	config.ConnectTimeout = 10 * time.Second

	conn, err := pgx.ConnectConfig(store.ctx, config)

	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Ping the database to verify the connection
	if err := conn.Ping(store.ctx); err != nil {
		conn.Close(store.ctx)
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to DB ...")
	store.conn = conn
	return nil
}

func (store *dbStore) Close() error {
	log.Println("Close DB connexion ...")

	if store.conn != nil {
		return store.conn.Close(store.ctx)
	}
	return nil
}

func (store *dbStore) ExportQueryToCSV(query string, csvPath string, delimiter rune) error {
	rows, err := store.conn.Query(store.ctx, query)

	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	defer rows.Close()
	if err = writePgxRowsToCSV(rows, csvPath, delimiter); err != nil {
		return fmt.Errorf("error writing CSV file: %w", err)
	}
	log.Println("Successfully exported.")
	return nil
}

func writePgxRowsToCSV(rows pgx.Rows, csvPath string, delimiter rune) error {
	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = delimiter
	defer writer.Flush()

	// Récupérer les noms des colonnes
	fieldDescriptions := rows.FieldDescriptions()
	headers := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		headers[i] = string(fd.Name)
	}

	// Écrire les headers
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	// Écrire les données
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("error reading row: %w", err)
		}

		// Convertir les valeurs en strings
		record := make([]string, len(values))
		for i, v := range values {
			if v == nil {
				record[i] = ""
			} else {
				// Formater les timestamps si nécessaire
				switch val := v.(type) {
				case time.Time:
					record[i] = val.Format("2006-01-02T15:04:05.000")
				default:
					record[i] = fmt.Sprintf("%v", v)
				}
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}
