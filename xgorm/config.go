package xgorm

import (
	"context"
	"fmt"
)

type DbConfig struct {
	AwsRegion string
	Username  string
	Password  string
	Host      string
	Port      int
	Database  string
	SSLMode   string

	MaxIdleConns int
	MaxOpenConns int
}

func (cfg *DbConfig) DSN(ctx context.Context) (string, error) {
	dbHost := cfg.Host
	dbPort := cfg.Port
	dbName := cfg.Database
	dbUser := cfg.Username
	dbPass := cfg.Password
	dbSslMode := cfg.SSLMode

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", dbHost, dbPort, dbUser, dbPass, dbName)
	if len(dbSslMode) > 0 {
		dsn += fmt.Sprintf(" sslmode=%s", dbSslMode)
	}
	return dsn, nil
}
