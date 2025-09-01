package db

import (
	"context"
	"database/sql"
)

// DBClient interface defines database operations for repository layer.
// Supports queries, transactions, and command execution with context.
type DBClient interface {
	// Exec executes a query that doesn't return rows (INSERT, UPDATE, DELETE).
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryRow executes a query that returns at most one row.
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Query executes a query that returns multiple rows.
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// BeginTx starts a new database transaction with specified options.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// CommitTx commits the specified transaction.
	CommitTx(tx *sql.Tx) error

	// RollbackTx rolls back the specified transaction.
	RollbackTx(tx *sql.Tx) error
}
