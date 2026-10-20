// Package migrate applies embedded SQL migrations using golang-migrate.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"harns-platform/migrations"
	"harns-platform/pkg/logging"
)

// Run applies pending migrations from the embedded filesystem.
func Run(ctx context.Context, dsn string) error {
	migrationDSN := dsn
	if !strings.Contains(migrationDSN, "multiStatements=") {
		migrationDSN += "&multiStatements=true"
	}

	db, err := sql.Open("mysql", migrationDSN)
	if err != nil {
		return fmt.Errorf("open migration database failed: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect migration database failed: %w", err)
	}

	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("init migration driver failed: %w", err)
	}
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("init migration source failed: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("init migrator failed: %w", err)
	}

	if version, dirty, verr := m.Version(); verr == nil {
		logging.V(1).InfoS("Migration state", "version", version, "dirty", dirty)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations failed: %w", err)
	}

	if version, _, verr := m.Version(); verr == nil {
		logging.V(1).InfoS("Migrations applied", "version", version)
	}
	return nil
}
