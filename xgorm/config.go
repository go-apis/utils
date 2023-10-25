package xgorm

import (
	"context"
	"fmt"
	"strings"
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
	var parts []string
	if len(cfg.Host) > 0 {
		parts = append(parts, fmt.Sprintf("host=%s", cfg.Host))
	}
	if cfg.Port > 0 {
		parts = append(parts, fmt.Sprintf("port=%d", cfg.Port))
	}
	if len(cfg.Database) > 0 {
		parts = append(parts, fmt.Sprintf("dbname=%s", cfg.Database))
	}
	if len(cfg.Username) > 0 {
		parts = append(parts, fmt.Sprintf("user=%s", cfg.Username))
	}
	if len(cfg.Password) > 0 {
		parts = append(parts, fmt.Sprintf("password=%s", cfg.Password))
	}
	if len(cfg.SSLMode) > 0 {
		parts = append(parts, fmt.Sprintf("sslmode=%s", cfg.SSLMode))
	}
	return strings.Join(parts, " "), nil
}
