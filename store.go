package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/fbz-tec/pgexport/logger"
	"github.com/jackc/pgx/v5"
)

type Store interface {
	Open(dbUrl string) error
	Close() error
	GetConnection() *pgx.Conn
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

	logger.Debug("Connection timeout: 10s")
	logger.Debug("Attempting to connect to database host: %s", sanitizeURL(dbUrl))

	conn, err := pgx.Connect(ctx, dbUrl)

	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	logger.Debug("Connection established, verifying connectivity (ping)...")

	// Ping the database to verify the connection
	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Debug("Database ping successful")
	store.conn = conn
	return nil
}

func (store *dbStore) Close() error {
	logger.Debug("Closing database connection...")

	if store.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := store.conn.Close(ctx)
		if err != nil {
			logger.Debug("Error closing database connection: %v", err)
		} else {
			logger.Debug("Database connection closed successfully")
		}
		return err
	}
	return nil
}

func (store *dbStore) ExecuteQuery(ctx context.Context, sql string) (pgx.Rows, error) {

	if store.conn == nil {
		logger.Debug("No active database connection; query cannot be executed")
		return nil, fmt.Errorf("no connection to database")
	}

	logger.Debug("Executing SQL query...")
	logger.Debug("Query: %s", sql)

	startTime := time.Now()
	rows, err := store.conn.Query(ctx, sql)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	logger.Debug("Query executed successfully in %v", duration)
	return rows, nil
}

func (store *dbStore) GetConnection() *pgx.Conn {
	return store.conn
}

// sanitizeURL removes the password part from a PostgreSQL DSN before logging.
func sanitizeURL(dbUrl string) string {
	u, err := url.Parse(dbUrl)
	if err != nil {
		return "<invalid-url>"
	}

	var userInfo string

	if u.User != nil {
		username := u.User.Username()
		if _, hasPwd := u.User.Password(); hasPwd {
			userInfo = fmt.Sprintf("%s:***@", username)
		} else {
			userInfo = fmt.Sprintf("%s@", username)
		}
	}

	path := u.Path
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s://%s%s%s", u.Scheme, userInfo, u.Host, path)
}
