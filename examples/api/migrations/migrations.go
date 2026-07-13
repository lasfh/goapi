package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

var (
	//go:embed *.up.sql
	files embed.FS

	ErrMigrationInterruptedByContext = errors.New("migration interrupted by context")
)

type MigrationManager struct {
	migrate        *migrate.Migrate
	sourceInstance source.Driver
}

func NewMigrationManager(
	db *sql.DB,
	databaseName string,
) (*MigrationManager, error) {
	sourceInstance, err := iofs.New(files, ".")
	if err != nil {
		return nil, fmt.Errorf("source[iofs]: %w", err)
	}

	dbInstance, err := postgres.WithInstance(
		db,
		&postgres.Config{
			MigrationsTable: postgres.DefaultMigrationsTable,
			DatabaseName:    databaseName,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("db[sqlserver]: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceInstance,
		databaseName,
		dbInstance,
	)
	if err != nil {
		return nil, fmt.Errorf("new migration instance: %w", err)
	}

	return &MigrationManager{
		migrate:        m,
		sourceInstance: sourceInstance,
	}, nil
}

func (m *MigrationManager) Up(ctx context.Context, ignoreVersion ...uint) error {
	firstMigration := false

	curVersion, dirty, err := m.migrate.Version()
	if err != nil {
		if !errors.Is(err, migrate.ErrNilVersion) {
			return fmt.Errorf("current version: %w", err)
		}

		firstMigration = true

		curVersion, err = m.sourceInstance.First()
		if err != nil {
			return fmt.Errorf("first version: %w", err)
		}
	}

	if dirty {
		return migrate.ErrDirty{Version: int(curVersion)}
	}

	for {
		select {
		case <-ctx.Done():
			return ErrMigrationInterruptedByContext
		default:
			if !firstMigration {
				curVersion, err = m.sourceInstance.Next(curVersion)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						slog.Info("No pending migrations to apply")

						return nil
					}

					return fmt.Errorf("next version: %w", err)
				}
			} else {
				firstMigration = false
			}

			r, identifier, err := m.sourceInstance.ReadUp(curVersion)
			if err != nil {
				return err
			}

			migr, err := migrate.NewMigration(r, identifier, curVersion, int(curVersion))
			if err != nil {
				return err
			}

			slog.Info(
				"Sending file (up)",
				slog.Uint64("version", uint64(migr.Version)),
				slog.String("identifier", identifier),
			)

			// Validação para pular migrações, se necessário.
			if slices.Contains(ignoreVersion, migr.Version) {
				continue
			}

			err = m.migrate.Run(migr)
			if err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					slog.Info("No pending migrations to apply")

					return nil
				}

				return fmt.Errorf("migration[%d] %s: %w", migr.Version, identifier, err)
			}
		}
	}
}
