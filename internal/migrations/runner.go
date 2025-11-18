package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"go.uber.org/zap"
)

//go:embed *.sql
var migrationFiles embed.FS

// migration represents a single DB migration.
type migration struct {
	name string
	sql  string
}

// Run applies all migrations to the database.
// It reads migration files from the embedded filesystem and executes them in order.
// If any migration fails, it returns an error.
func Run(ctx context.Context, db *sql.DB, logger *zap.Logger) error {
	migs, err := loadMigrations()
	if err != nil {
		return err
	}

	for _, m := range migs {
		logger.Info("Applying migration", zap.String("migration", m.name))
		if _, err := db.ExecContext(ctx, m.sql); err != nil {
			logger.Error("Migration failed", zap.String("migration", m.name), zap.Error(err))
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}
		logger.Info("Migration applied", zap.String("migration", m.name))
	}
	return nil
}

// loadMigrations reads migration files from the embedded filesystem
// and returns a slice of migration structs.
// If reading fails, it returns an error.
// Migrations files are expected to be prefixed with timestamp for ordering.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	migs := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		migs = append(migs, migration{
			name: entry.Name(),
			sql:  string(content),
		})
	}
	return migs, nil
}
