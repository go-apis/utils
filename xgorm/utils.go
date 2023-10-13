package xgorm

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func open(dsn string, options *Options) (*gorm.DB, error) {
	dr := postgres.Open(dsn)

	db, err := gorm.Open(dr, &gorm.Config{
		SkipDefaultTransaction:   options.SkipDefaultTransaction,
		DisableNestedTransaction: options.DisableNestedTransaction,
	})
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

func recreate(ctx context.Context, config *DbConfig) error {
	databaseName := config.Database

	master := &DbConfig{
		AwsRegion: config.AwsRegion,
		Username:  config.Username,
		Password:  config.Password,
		Host:      config.Host,
		Port:      config.Port,
		Database:  "postgres",
		SSLMode:   config.SSLMode,
	}
	dsn, err := master.DSN(ctx)
	if err != nil {
		return err
	}

	db, err := open(dsn, &Options{})
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

	if opts.Recreate {
		if err := recreate(ctx, config); err != nil {
			return nil, err
		}
	}

	dsn, err := config.DSN(ctx)
	if err != nil {
		return nil, err
	}

	gdb, err := open(dsn, opts)
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	return gdb, nil
}
