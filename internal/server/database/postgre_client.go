package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config contains PostgreSQL database connection configuration parameters.
type Config struct {
	// Host specifies the PostgreSQL server hostname or IP address.
	Host string
	// User specifies the database username for authentication.
	User string
	// Password specifies the database password for authentication (sensitive data).
	Password string
	// DBName specifies the target database name to connect to.
	DBName string
	// SSLMode specifies the SSL connection mode (disable, require, verify-ca, verify-full).
	SSLMode string
	// Port specifies the PostgreSQL server port number (typically 5432).
	Port int
	// Timeout specifies the maximum duration for connection attempts and pings.
	Timeout time.Duration
}

// Client provides a PostgreSQL database client with connection management and query execution.
type Client struct {
	// db is the underlying SQL database connection.
	db *sql.DB
	// pingTimeout specifies the timeout duration for health check operations.
	pingTimeout time.Duration
}

// NewClient creates a new PostgreSQL client with the provided configuration.
// It establishes a connection and verifies connectivity with a ping operation.
func NewClient(cfg *Config) (*Client, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection opening failed: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err = dbConn.PingContext(pingCtx); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Client{db: dbConn, pingTimeout: cfg.Timeout}, nil
}

// Exec executes a query that doesn't return rows (INSERT, UPDATE, DELETE).
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query %q execution failed: %w", query, err)
	}
	return result, nil
}

// QueryRow executes a query that returns at most one row and returns a *sql.Row.
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// Query executes a query that returns multiple rows and returns a *sql.Rows result set.
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query %q execution failed: %w", query, err)
	}
	return rows, nil
}

// BeginTx starts a new database transaction with the specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("transaction start failed: %w", err)
	}
	return tx, nil
}

// CommitTx commits the specified transaction, making all changes permanent.
func (c *Client) CommitTx(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}

// RollbackTx rolls back the specified transaction, discarding all changes.
func (c *Client) RollbackTx(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("transaction rollback failed: %w", err)
	}
	return nil
}

// Ping verifies the database connection is still alive and functioning.
func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, c.pingTimeout)
	defer cancel()

	if err := c.db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database connection open failed: %w", err)
	}
	return nil
}

// Close gracefully closes the database connection with context cancellation support.
func (c *Client) Close(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- c.db.Close()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled while closing database connection: %w", ctx.Err())
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("database connection closure failed: %w", err)
		}
		return nil
	}
}
