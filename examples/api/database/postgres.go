package database

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"

	"github.com/lasfh/goapi/examples/api/internal/config"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	onceDB sync.Once
)

func NewClient(cfg config.PostgresConfig) (*gorm.DB, error) {
	var err error

	onceDB.Do(func() {
		db, err = NewPostgresClient(cfg, cfg.DSN())
	})

	return db, err
}

func NewPostgresClient(cfg config.PostgresConfig, customDSN ...string) (*gorm.DB, error) {
	dsn := cfg.DSN()
	if len(customDSN) > 0 {
		dsn = customDSN[0]
	}

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: func() logger.Interface {
				if cfg.Debug {
					return logger.Default.LogMode(logger.Info)
				}

				return logger.Default
			}(),
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func NewPostgresClientMigration(cfg config.PostgresConfig) (*sql.DB, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.UserMigration, cfg.UserMigration),
		Host:   net.JoinHostPort(cfg.Host, strconv.FormatUint(uint64(cfg.Port), 10)),
		RawQuery: url.Values{
			"database": []string{cfg.Database},
		}.Encode(),
	}

	db, err := sql.Open("postgres", u.String())
	if err != nil {
		return nil, fmt.Errorf("open[postgres]: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping[postgres]: %w", err)
	}

	return db, nil
}
