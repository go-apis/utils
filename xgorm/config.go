package xgorm

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
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
	region := cfg.AwsRegion
	dbHost := cfg.Host
	dbPort := cfg.Port
	dbName := cfg.Database
	dbUser := cfg.Username
	dbPass := cfg.Password
	dbEndpoint := fmt.Sprintf("%s:%d", dbHost, dbPort)
	dbSslMode := cfg.SSLMode

	// if the password doesn't exist lets try using AWS directly
	if len(dbPass) == 0 && len(region) > 0 {
		awscfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return "", err
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, dbEndpoint, region, dbUser, awscfg.Credentials)
		if err != nil {
			return "", err
		}

		// set the pass
		dbPass = authenticationToken
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", dbHost, dbPort, dbUser, dbPass, dbName)
	if len(dbSslMode) > 0 {
		dsn += fmt.Sprintf(" sslmode=%s", dbSslMode)
	}
	return dsn, nil
}
