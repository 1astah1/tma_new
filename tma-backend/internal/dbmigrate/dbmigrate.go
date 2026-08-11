package dbmigrate

import (
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"

	"tma-backend/migrations"
)

// Run накатывает вшитые миграции на подключённую базу.
func Run(db *sqlx.DB) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Migrations up to date")
			return nil
		}
		return err
	}

	version, dirty, err := m.Version()
	if err == nil {
		slog.Info("Migrations applied", slog.Uint64("version", uint64(version)), slog.Bool("dirty", dirty))
	}
	return nil
}
