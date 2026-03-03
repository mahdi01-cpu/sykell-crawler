package db

import (
	"database/sql"
	"errors"
	"fmt"

	"embed"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	mmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var MigrationFiles embed.FS

func RunMigrations(dsn string) error {
	if dsn == "" {
		return fmt.Errorf("DB_DSN is required")
	}

	// golang-migrate uses database/sql
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer sqlDB.Close()

	driver, err := mmysql.WithInstance(sqlDB, &mmysql.Config{})
	if err != nil {
		return fmt.Errorf("mysql driver: %w", err)
	}

	src, err := iofs.New(MigrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}
