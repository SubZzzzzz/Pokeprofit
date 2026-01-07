package repository

import (
	"context"

	"github.com/SubZzzzzz/pokeprofit/internal/database"
	"github.com/jmoiron/sqlx"
)

// Base provides common repository functionality.
type Base struct {
	db *database.DB
}

// NewBase creates a new Base repository.
func NewBase(db *database.DB) *Base {
	return &Base{db: db}
}

// DB returns the underlying database connection.
func (b *Base) DB() *database.DB {
	return b.db
}

// Tx runs a function within a transaction.
func (b *Base) Tx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := b.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// QueryRow executes a query that returns a single row.
func (b *Base) QueryRow(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return b.db.GetContext(ctx, dest, query, args...)
}

// QueryRows executes a query that returns multiple rows.
func (b *Base) QueryRows(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return b.db.SelectContext(ctx, dest, query, args...)
}

// Exec executes a query that doesn't return rows.
func (b *Base) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := b.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ExecNamed executes a named query that doesn't return rows.
func (b *Base) ExecNamed(ctx context.Context, query string, arg interface{}) (int64, error) {
	result, err := b.db.NamedExecContext(ctx, query, arg)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// InsertNamed inserts a record using named parameters.
func (b *Base) InsertNamed(ctx context.Context, query string, arg interface{}) error {
	_, err := b.db.NamedExecContext(ctx, query, arg)
	return err
}
