package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type QueryResult struct {
	Columns []string
	Rows    [][]interface{}
}

type Store interface {
	Open(dbUrl string) error
	Close() error
	ExecuteQuery(query string) (*QueryResult, error)
}

type dbStore struct {
	conn *pgx.Conn
	ctx  context.Context
}

func NewStore() Store {
	return &dbStore{}
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
	log.Println("Close database connexion ...")

	if store.conn != nil {
		return store.conn.Close(store.ctx)
	}
	return nil
}

func (store *dbStore) ExecuteQuery(query string) (*QueryResult, error) {
	rows, err := store.conn.Query(store.ctx, query)

	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	if len(fieldDescriptions) == 0 {
		return nil, fmt.Errorf("no columns found in query result")
	}

	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Fetch all rows
	var data [][]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("error reading row %d: %w", len(data)+1, err)
		}
		data = append(data, values)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &QueryResult{
		Columns: columns,
		Rows:    data,
	}, nil
}
