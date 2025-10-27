package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type Store interface {
	Open(dbUrl string) error
	Close() error
	ExecuteQuery(ctx context.Context, sql string) (pgx.Rows, error)
}

type dbStore struct {
	conn *pgx.Conn
}

func NewStore() Store {
	return &dbStore{}
}

func (store *dbStore) Open(dbUrl string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbUrl)

	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Ping the database to verify the connection
	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Database connection established")
	store.conn = conn
	return nil
}

func (store *dbStore) Close() error {
	log.Println("Closing database connection ...")

	if store.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return store.conn.Close(ctx)
	}
	return nil
}

func (store *dbStore) ExecuteQuery(ctx context.Context, sql string) (pgx.Rows, error) {

	if store.conn == nil {
		return nil, fmt.Errorf("no connection to database")
	}

	rows, err := store.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return rows, nil
}
