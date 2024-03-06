package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	migrate "github.com/rubenv/sql-migrate"
)

const RDBDNSForTest = "postgresql://root:root@127.0.0.1:5432/postgres?sslmode=disable"

func MigrateRDBPostgres(ctx context.Context, db *sql.DB) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to Getwd: %v", err)
	}

	projectRoot, err := GetProjectRootDir(cwd, 10)
	if err != nil {
		return fmt.Errorf("failed to GetProjectRootDir: %v", err)
	}

	migrateDir := filepath.Join(projectRoot, "misc/db/migrations/postgres")
	mig := migrate.FileMigrationSource{
		Dir: migrateDir,
	}

	if _, err := migrate.ExecContext(ctx, db, "postgres", mig, migrate.Up); err != nil {
		return fmt.Errorf("failed to migrate.ExecContext: %v", err)
	}
	return nil
}

func UintPtr(v uint) *uint {
	var value uint
	value = v
	return &value
}
