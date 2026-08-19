package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func OpenDatabase(ctx context.Context, cfg DatabaseConfig) (*sql.DB, error) {
	if cfg.Driver == "" {
		cfg.Driver = "sqlite"
	}
	if cfg.Driver == "sqlite" {
		if dir := filepath.Dir(cfg.DSN); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create sqlite directory: %w", err)
			}
		}
	}
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.Driver, err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	if cfg.Driver == "sqlite" {
		if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
		}
		if _, err := db.ExecContext(ctx, `PRAGMA busy_timeout=5000`); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set sqlite busy timeout: %w", err)
		}
	}
	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB, cfg DatabaseConfig) error {
	if strings.EqualFold(cfg.Driver, "postgres") || strings.Contains(cfg.Driver, "postgres") {
		return migratePostgres(ctx, db, cfg.MigrationsDir)
	}
	return migrateSQLite(ctx, db)
}

func migratePostgres(ctx context.Context, db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		version := strings.TrimSuffix(entry.Name(), ".sql")
		if applied(ctx, db, version) {
			continue
		}
		if _, err := db.ExecContext(ctx, string(raw)); err != nil {
			return fmt.Errorf("apply postgres migration %s: %w", entry.Name(), err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func applied(ctx context.Context, db *sql.DB, version string) bool {
	var one int
	err := db.QueryRowContext(ctx, `SELECT 1 FROM schema_migrations WHERE version=$1`, version).Scan(&one)
	return err == nil
}

func migrateSQLite(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return err
	}
	for _, migration := range sqliteMigrations {
		var one int
		err := db.QueryRowContext(ctx, `SELECT 1 FROM schema_migrations WHERE version=?`, migration.version).Scan(&one)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		if _, err := db.ExecContext(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply sqlite migration %s: %w", migration.version, err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, migration.version, time.Now().UTC()); err != nil {
			return fmt.Errorf("record sqlite migration %s: %w", migration.version, err)
		}
	}
	return nil
}

type sqlMigration struct {
	version string
	sql     string
}

var sqliteMigrations = []sqlMigration{
	{version: "001_init", sql: sqliteSchema},
}
