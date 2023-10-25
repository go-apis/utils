package xgorm

import (
	"context"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/contextcloud/goutils/xlog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func awsAuthToken(region string, timeout time.Duration) func(ctx context.Context, config *pgx.ConnConfig) error {
	t := time.Now()

	return func(ctx context.Context, config *pgx.ConnConfig) error {
		log := xlog.Logger(ctx)

		if config.Password == "" || time.Since(t) < timeout {
			awscfg, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				log.Error("issue loading aws config", zap.Error(err))
				return err
			}
			dbEndpoint := fmt.Sprintf("%s:%d", config.Host, config.Port)

			authenticationToken, err := auth.BuildAuthToken(ctx, dbEndpoint, region, config.User, awscfg.Credentials)
			if err != nil {
				log.Error("issue building auth token", zap.Error(err))
				return err
			}

			// set the password
			config.Password = authenticationToken

			// restart the time
			t = time.Now()
		}
		return nil
	}
}

func open(ctx context.Context, config *DbConfig, options *Options) (*gorm.DB, error) {
	log := xlog.Logger(ctx)

	str, err := config.DSN(ctx)
	if err != nil {
		log.Error("issue creating dsn", zap.Error(err))
		return nil, err
	}

	dbConfig, err := pgx.ParseConfig(str)
	if err != nil {
		log.Error("issue parsing dsn", zap.Error(err))
		return nil, err
	}

	var openOptions []stdlib.OptionOpenDB
	if config.Password == "" && config.AwsRegion != "" {
		openOptions = append(openOptions, stdlib.OptionBeforeConnect(awsAuthToken(config.AwsRegion, 10*time.Minute)))
	}

	db, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn: stdlib.OpenDB(*dbConfig, openOptions...),
		}),
		&gorm.Config{
			SkipDefaultTransaction:   options.SkipDefaultTransaction,
			DisableNestedTransaction: options.DisableNestedTransaction,
		})
	if err != nil {
		log.Error("issue opening db", zap.Error(err))
		return nil, err
	}

	if options.AutoMigrate && len(options.Models) > 0 {
		if err := db.AutoMigrate(options.Models...); err != nil {
			log.Error("issue automigrating", zap.Error(err))
			return nil, err
		}
	}

	if options.Tracing {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			log.Error("issue adding tracing plugin", zap.Error(err))
			return nil, err
		}
	}
	if options.Logger != nil {
		db.Logger = options.Logger
	}

	inner, err := db.DB()
	if err != nil {
		log.Error("issue getting db", zap.Error(err))
		return nil, err
	}

	inner.SetMaxIdleConns(config.MaxIdleConns)
	inner.SetMaxOpenConns(config.MaxOpenConns)

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
	db, err := open(ctx, master, &Options{})
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

	gdb, err := open(ctx, config, opts)
	if err != nil {
		return nil, err
	}

	return gdb, nil
}
