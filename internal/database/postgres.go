package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB wraps sqlx.DB with additional functionality.
type DB struct {
	*sqlx.DB
}

// Config holds database connection settings.
type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns default database configuration.
func DefaultConfig(url string) Config {
	return Config{
		URL:             url,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// New creates a new database connection pool.
func New(cfg Config) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{DB: db}, nil
}

// Connect creates a new database connection with context timeout.
func Connect(ctx context.Context, url string) (*DB, error) {
	cfg := DefaultConfig(url)
	db, err := New(cfg)
	if err != nil {
		return nil, err
	}

	// Verify connection with context
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Health checks the database connection.
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	return db.DB.Close()
}

// BeginTx starts a new transaction with context.
func (db *DB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return db.BeginTxx(ctx, nil)
}
