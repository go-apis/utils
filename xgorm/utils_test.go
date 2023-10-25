package xgorm

import (
	"context"
	"testing"
)

func TestConnect(t *testing.T) {
	ctx := context.Background()
	cfg := &DbConfig{
		AwsRegion: "us-east-1",
		Username:  "identity",
		Password:  "",
		Host:      "localhost",
		Port:      5432,
		Database:  "identity-db",
		SSLMode:   "disable",
	}

	db, err := NewDb(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if db.Error != nil {
		t.Fatal(db.Error)
	}
}
