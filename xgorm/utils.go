package xgorm

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func open(dsn string, options *Options) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}

	if options.AutoMigrate && len(options.Models) > 0 {
		if err := db.AutoMigrate(options.Models...); err != nil {
			return nil, err
		}
	}

	if options.Tracing {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			return nil, err
		}
	}
	if options.Logger != nil {
		db.Logger = options.Logger
	}

	return db, nil
}

func recreate(ctx context.Context, dsn string) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return err
	}

	databaseName := cfg.Database
	cfg.Database = "postgres"

	db, err := open(cfg.ConnString(), &Options{})
	if err != nil {
		return err
	}

	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = ? and pid <> pg_backend_pid()`
	if err := db.Exec(query, databaseName).Error; err != nil {
		return err
	}

	q1 := fmt.Sprintf(`drop database if exists %s`, databaseName)
	if err := db.Exec(q1).Error; err != nil {
		return err
	}

	q2 := fmt.Sprintf(`create database %s`, databaseName)
	if err := db.Exec(q2).Error; err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func NewDb(ctx context.Context, config *DbConfig, opt ...Option) (*gorm.DB, error) {
	opts := NewOptions()
	for _, o := range opt {
		o(opts)
	}

	dsn, err := config.DSN(ctx)
	if err != nil {
		return nil, err
	}

	if opts.Recreate {
		if err := recreate(ctx, dsn); err != nil {
			return nil, err
		}
	}

	return open(dsn, opts)
}
