package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	// Parse flags
	upCmd := flag.NewFlagSet("up", flag.ExitOnError)
	downCmd := flag.NewFlagSet("down", flag.ExitOnError)
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure migrations table exists
	if err := ensureMigrationsTable(ctx, db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	switch os.Args[1] {
	case "up":
		upCmd.Parse(os.Args[2:])
		if err := runUp(ctx, db); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		downCmd.Parse(os.Args[2:])
		if err := runDown(ctx, db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
	case "status":
		statusCmd.Parse(os.Args[2:])
		if err := showStatus(ctx, db); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up      Run all pending migrations")
	fmt.Println("  down    Rollback the last migration")
	fmt.Println("  status  Show migration status")
}

func ensureMigrationsTable(ctx context.Context, db *database.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

func runUp(ctx context.Context, db *database.DB) error {
	migrations, err := getMigrationFiles()
	if err != nil {
		return err
	}

	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		fmt.Printf("Applying migration %s...\n", m.Version)

		content, err := os.ReadFile(m.UpPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", m.UpPath, err)
		}

		tx, err := db.BeginTx(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", m.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m.Version, err)
		}

		fmt.Printf("Applied migration %s\n", m.Version)
	}

	fmt.Println("All migrations applied successfully")
	return nil
}

func runDown(ctx context.Context, db *database.DB) error {
	var lastVersion string
	err := db.GetContext(ctx, &lastVersion, "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1")
	if err != nil {
		return fmt.Errorf("no migrations to rollback")
	}

	migrations, err := getMigrationFiles()
	if err != nil {
		return err
	}

	var migration *Migration
	for _, m := range migrations {
		if m.Version == lastVersion {
			migration = &m
			break
		}
	}

	if migration == nil || migration.DownPath == "" {
		return fmt.Errorf("no down migration found for version %s", lastVersion)
	}

	fmt.Printf("Rolling back migration %s...\n", lastVersion)

	content, err := os.ReadFile(migration.DownPath)
	if err != nil {
		return fmt.Errorf("failed to read down migration file: %w", err)
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute down migration: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = $1", lastVersion); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	fmt.Printf("Rolled back migration %s\n", lastVersion)
	return nil
}

func showStatus(ctx context.Context, db *database.DB) error {
	migrations, err := getMigrationFiles()
	if err != nil {
		return err
	}

	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("-----------------")

	for _, m := range migrations {
		status := "pending"
		if applied[m.Version] {
			status = "applied"
		}
		fmt.Printf("%s: %s\n", m.Version, status)
	}

	return nil
}

type Migration struct {
	Version  string
	UpPath   string
	DownPath string
}

func getMigrationFiles() ([]Migration, error) {
	migrationsDir := "internal/database/migrations"

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationMap := make(map[string]*Migration)

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

		name := f.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		path := filepath.Join(migrationsDir, name)

		if migrationMap[version] == nil {
			migrationMap[version] = &Migration{Version: version}
		}

		if strings.Contains(name, "_down.sql") {
			migrationMap[version].DownPath = path
		} else {
			migrationMap[version].UpPath = path
		}
	}

	var migrations []Migration
	for _, m := range migrationMap {
		if m.UpPath != "" {
			migrations = append(migrations, *m)
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(ctx context.Context, db *database.DB) (map[string]bool, error) {
	var versions []string
	err := db.SelectContext(ctx, &versions, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, v := range versions {
		applied[v] = true
	}

	return applied, nil
}
